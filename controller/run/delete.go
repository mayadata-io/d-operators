/*
Copyright 2020 The MayaData Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package run

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	types "mayadata.io/d-operators/types/run"
)

// DeleteRequest holds the config to delete the
// observed state(s) from cluster
type DeleteRequest struct {
	IncludeInfo       map[types.IncludeInfoKey]bool
	Run               *unstructured.Unstructured
	Watch             *unstructured.Unstructured
	TaskKey           string
	Apply             map[string]interface{}
	DeleteTemplate    *unstructured.Unstructured
	ObservedResources []*unstructured.Unstructured
}

// DeleteResponse holds the observed states that
// needs to be deleted from the cluster
type DeleteResponse struct {
	ExplicitDeletes []*unstructured.Unstructured
	Result          *types.Result
}

// DeleteStateBuilder holds the observed resource(s)
// to be deleted
//
// NOTE:
// 	A resource that is owned by this controller & needs to
// be deleted will not be added to the response. Skipping
// an observed resource implies this resource should get
// deleted at the cluster.
type DeleteStatesBuilder struct {
	Request DeleteRequest
	Result  *types.Result

	deleteTemplate     *unstructured.Unstructured
	deleteResourceName string
	desiredDeletes     []*unstructured.Unstructured
	explicitDeletes    []*unstructured.Unstructured

	err error
}

func (r *DeleteStatesBuilder) evalDeleteTemplate() {
	if r.Request.DeleteTemplate != nil {
		r.deleteTemplate = r.Request.DeleteTemplate
		return
	}
	obj := &unstructured.Unstructured{}
	obj.SetUnstructuredContent(r.Request.Apply)
	// verify if its a proper unstructured instance
	_, r.err = obj.MarshalJSON()
	if r.err != nil {
		r.err = errors.Wrapf(
			r.err,
			"Invalid delete state: %q",
			r.Request.TaskKey,
		)
		return
	}
	if len(obj.Object) == 0 {
		r.err = errors.Errorf(
			"Empty delete state: %q",
			r.Request.TaskKey,
		)
		return
	}
	if obj.GetAPIVersion() == "" {
		r.err = errors.Errorf(
			"Invalid delete state: Missing apiversion: %q",
			r.Request.TaskKey,
		)
		return
	}
	if obj.GetKind() == "" {
		r.err = errors.Errorf(
			"Invalid delete state: Missing kind: %q",
			r.Request.TaskKey,
		)
		return
	}
	r.deleteTemplate = obj
}

func (r *DeleteStatesBuilder) evalDeleteResourceName() {
	// start with generate name if any
	name := r.deleteTemplate.GetGenerateName()
	if name == "" {
		// use name from specified desired state
		name = r.deleteTemplate.GetName()
	}
	r.deleteResourceName = name
}

func (r *DeleteStatesBuilder) includeDesiredInfoIfEnabled(
	resource *unstructured.Unstructured,
	message string,
) {
	if r.Request.IncludeInfo == nil {
		return
	}
	if !r.Request.IncludeInfo[types.IncludeDesiredInfo] &&
		!r.Request.IncludeInfo[types.IncludeAllInfo] {
		return
	}
	r.Result.DesiredInfo = append(
		r.Result.DesiredInfo,
		fmt.Sprintf(
			"%s: %q / %q: %s",
			message,
			resource.GetNamespace(),
			resource.GetName(),
			resource.GroupVersionKind().String(),
		),
	)
}

func (r *DeleteStatesBuilder) includeExplicitInfoIfEnabled(
	resource *unstructured.Unstructured,
	message string,
) {
	if r.Request.IncludeInfo == nil {
		return
	}
	if !r.Request.IncludeInfo[types.IncludeExplicitInfo] &&
		!r.Request.IncludeInfo[types.IncludeAllInfo] {
		return
	}
	r.Result.ExplicitInfo = append(
		r.Result.ExplicitInfo,
		fmt.Sprintf(
			"%s: %q / %q: %s",
			message,
			resource.GetNamespace(),
			resource.GetName(),
			resource.GroupVersionKind().String(),
		),
	)
}

func (r *DeleteStatesBuilder) includeSkippedInfoIfEnabled(
	resource *unstructured.Unstructured,
	message string,
) {
	if r.Request.IncludeInfo == nil {
		return
	}
	if !r.Request.IncludeInfo[types.IncludeSkippedInfo] &&
		!r.Request.IncludeInfo[types.IncludeAllInfo] {
		return
	}
	r.Result.SkippedInfo = append(
		r.Result.SkippedInfo,
		fmt.Sprintf(
			"%s: %q / %q: %s",
			message,
			resource.GetNamespace(),
			resource.GetName(),
			resource.GroupVersionKind().String(),
		),
	)
}

func (r *DeleteStatesBuilder) markResourcesForExplicitDelete() {
	var watchuid = string(r.Request.Watch.GetUID())
	for _, observed := range r.Request.ObservedResources {
		if observed == nil || observed.Object == nil {
			r.err = errors.Errorf(
				"Can't mark for delete: Nil observed object",
			)
			return
		}
		if r.deleteTemplate == nil || r.deleteTemplate.Object == nil {
			r.err = errors.Errorf(
				"Can't mark for delete: Nil delete template",
			)
			return
		}
		if !strings.HasPrefix(observed.GetName(), r.deleteResourceName) ||
			r.deleteTemplate.GetKind() != observed.GetKind() ||
			r.deleteTemplate.GetAPIVersion() != observed.GetAPIVersion() ||
			r.deleteTemplate.GetNamespace() != observed.GetNamespace() {
			// add this as skip info, for debugging purposes
			r.includeSkippedInfoIfEnabled(observed, "Skipped for delete")
			// not a match, so nothing to do
			continue
		}
		observedAnns := observed.GetAnnotations()
		if len(observedAnns) == 0 ||
			observedAnns["metac.openebs.io/created-due-to-watch"] != watchuid {
			// Observed resource was not created by this watch resource.
			// Hence, this needs to be deleted explicitly by metac
			r.explicitDeletes = append(
				r.explicitDeletes,
				observed.DeepCopy(),
			)
			// add this as desired info, for debugging purposes
			r.includeExplicitInfoIfEnabled(observed, "Marked for explicit delete")
		} else {
			// Observed resource was created by this watch resource.
			// Hence, this needs to be avoided adding to the response
			// thereby enabling metac to delete the resource
			r.desiredDeletes = append(
				r.desiredDeletes,
				observed.DeepCopy(),
			)
			// add this as desired info, for debugging purposes
			r.includeDesiredInfoIfEnabled(observed, "Marked for desired delete")
		}
	}
}

// Build returns the resource(s) that needs to be deleted
func (r *DeleteStatesBuilder) Build() (*DeleteResponse, error) {
	if r.Request.TaskKey == "" {
		return nil, errors.Errorf(
			"Can't delete: Empty task key",
		)
	}
	if r.Request.Run == nil {
		return nil, errors.Errorf(
			"Can't delete: Nil run: %q",
			r.Request.TaskKey,
		)
	}
	if r.Request.Watch == nil {
		return nil, errors.Errorf(
			"Can't delete: Nil watch: %q",
			r.Request.TaskKey,
		)
	}
	if len(r.Request.Apply) == 0 &&
		(r.Request.DeleteTemplate == nil || r.Request.DeleteTemplate.Object == nil) {
		return nil, errors.Errorf(
			"Can't delete: Nil delete state found: %q",
			r.Request.TaskKey,
		)
	}
	fns := []func(){
		r.evalDeleteTemplate,
		r.evalDeleteResourceName,
		r.markResourcesForExplicitDelete,
	}
	for _, fn := range fns {
		fn()
		if r.err != nil {
			return nil, r.err
		}
	}
	return &DeleteResponse{
		ExplicitDeletes: r.explicitDeletes,
		Result: &types.Result{
			SkippedInfo:  r.Result.SkippedInfo,  // is set if enabled
			DesiredInfo:  r.Result.DesiredInfo,  // is set if enabled
			ExplicitInfo: r.Result.ExplicitInfo, // is set if enabled
			Phase:        types.ResultPhaseOnline,
			Message: fmt.Sprintf(
				"Delete action was successful: Desired deletes %d: Explicit deletes %d",
				len(r.desiredDeletes),
				len(r.explicitDeletes),
			),
		},
	}, nil
}

// BuildDeleteStates returns the resources to be deleted
// from the k8s cluster
func BuildDeleteStates(request DeleteRequest) (*DeleteResponse, error) {
	r := &DeleteStatesBuilder{
		Request: request,
		Result:  &types.Result{}, // initialize
	}
	return r.Build()
}
