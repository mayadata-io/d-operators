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

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	types "mayadata.io/d-operators/types/run"
	"openebs.io/metac/controller/common/selector"
	"openebs.io/metac/dynamic/apply"
)

// UpdateRequest provides various data required to
// build the update state(s)
type UpdateRequest struct {
	IncludeInfo       map[types.IncludeInfoKey]bool
	Run               *unstructured.Unstructured
	Watch             *unstructured.Unstructured
	TaskKey           string
	Apply             map[string]interface{}
	TargetSelector    types.TargetSelector
	ObservedResources []*unstructured.Unstructured
}

// UpdateResponse holds the updated states that
// needs to be applied against the cluster
type UpdateResponse struct {
	DesiredUpdates  []*unstructured.Unstructured
	ExplicitUpdates []*unstructured.Unstructured
	Result          *types.Result
}

// UpdateStatesBuilder builds the desired resource(s) to
// be updated at the cluster
type UpdateStatesBuilder struct {
	Request UpdateRequest
	Result  *types.Result

	// filtered resources that need to be updated
	filteredResources []*unstructured.Unstructured

	// resources created due to this controller & are
	// marked to be updated
	markedForDesiredUpdates []*unstructured.Unstructured

	// resources that are not created by this controller
	// & are marked to be updated
	markedForExplicitUpdates []*unstructured.Unstructured

	desiredUpdates  []*unstructured.Unstructured
	explicitUpdates []*unstructured.Unstructured

	// flags if the update process should be skipped
	isSkip bool

	err error
}

func (r *UpdateStatesBuilder) includeDesiredInfoIfEnabled(
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
	r.Result.DesiredResourcesInfo = append(
		r.Result.DesiredResourcesInfo,
		fmt.Sprintf(
			"%s: %q / %q: %s",
			message,
			resource.GetNamespace(),
			resource.GetName(),
			resource.GroupVersionKind().String(),
		),
	)
}

func (r *UpdateStatesBuilder) includeExplicitInfoIfEnabled(
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
	r.Result.ExplicitResourcesInfo = append(
		r.Result.ExplicitResourcesInfo,
		fmt.Sprintf(
			"%s: %q / %q: %s",
			message,
			resource.GetNamespace(),
			resource.GetName(),
			resource.GroupVersionKind().String(),
		),
	)
}

func (r *UpdateStatesBuilder) includeSkippedInfoIfEnabled(
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
	r.Result.SkippedResourcesInfo = append(
		r.Result.SkippedResourcesInfo,
		fmt.Sprintf(
			"%s: %q / %q: %s",
			message,
			resource.GetNamespace(),
			resource.GetName(),
			resource.GroupVersionKind().String(),
		),
	)
}

func (r *UpdateStatesBuilder) filterResources() {
	for _, observed := range r.Request.ObservedResources {
		if observed == nil || observed.Object == nil {
			r.err = errors.Errorf(
				"Can't filter resources: Nil observed object",
			)
			return
		}
		e := selector.Evaluation{
			Terms:     r.Request.TargetSelector.SelectorTerms,
			Target:    observed,
			Reference: r.Request.Watch,
		}
		isMatch, err := e.RunMatch()
		if err != nil {
			r.err = err
			return
		}
		if !isMatch {
			r.includeSkippedInfoIfEnabled(observed, "Skipped for update")
			continue
		}
		r.filteredResources = append(
			r.filteredResources,
			observed,
		)
	}
}

func (r *UpdateStatesBuilder) isSkipUpdate() {
	if len(r.filteredResources) == 0 {
		// no resources matched i.e. none were filtered
		// so the entire update task will be skipped
		r.isSkip = true
	}
}

// group resources by explicit updates or desired updates
func (r *UpdateStatesBuilder) groupResourcesByUpdateType() {
	var watchuid = string(r.Request.Watch.GetUID())
	for _, filtered := range r.filteredResources {
		observedAnns := filtered.GetAnnotations()
		if len(observedAnns) != 0 &&
			observedAnns[types.AnnotationKeyMetacCreatedDueToWatch] == watchuid {
			// This observed resource was created by this controller
			// Hence this can be **updated** by metac without any special needs
			r.markedForDesiredUpdates = append(
				r.markedForDesiredUpdates,
				filtered,
			)
			r.includeDesiredInfoIfEnabled(filtered, "Marked for desired update")
		} else {
			// This observed resource was not created by this controller.
			// Hence, this resource needs to be **updated explicitly** by metac.
			r.markedForExplicitUpdates = append(
				r.markedForExplicitUpdates,
				filtered,
			)
			r.includeExplicitInfoIfEnabled(filtered, "Marked for explicit update")
		}
	}
}

// A 3-way merge is performed against the resource to arrive
// at the final resource state. This final state gets updated
// against the cluster. This **avoids** the need to parse
// desired state & observed states individually to update the
// desired against observed.
//
// NOTE:
// 	A 3-way merge helps us in applying the desired state against
// the observed state without the need to have this desired state
// with all the identifying fields. In other words, this desired
// state need not have name, uid, apiVersion, kind at all & yet
// this logic will be able to merge the desired state against
// the observed state & arrive at the final merge state.
//
// NOTE:
//	Last applied state & desired state are applied against the
// observed state to derive the final state.
//
// NOTE:
//	This logic is essentially a 2 way merge since last applied &
// desired state are same. Both last applied & desired state
// equals the desired state.
func (r *UpdateStatesBuilder) runApplyForDesiredUpdates() {
	for _, obj := range r.markedForDesiredUpdates {
		final, err := apply.Merge(
			obj.UnstructuredContent(), // observed
			r.Request.Apply,           // last applied
			r.Request.Apply,           // desired update content
		)
		if err != nil {
			r.err = errors.Wrapf(
				err,
				"Can't update %q / %q: %s",
				obj.GetNamespace(),
				obj.GetName(),
				obj.GroupVersionKind().String(),
			)
			return
		}
		r.desiredUpdates = append(
			r.desiredUpdates,
			&unstructured.Unstructured{
				Object: final,
			},
		)
	}
}

// A 3-way merge is performed against the resource to arrive
// at the final resource state. This final state gets updated
// against the cluster. This **avoids** the need to parse
// desired state & observed states individually to update the
// desired against observed.
//
// NOTE:
// 	A 3-way merge helps us in applying the desired state against
// the observed state without the need to have this desired state
// with all the identifying fields. In other words, this desired
// state need not have name, uid, apiVersion, kind at all & yet
// this logic will be able to merge the desired state against
// the observed state & arrive at the final merge state.
//
// NOTE:
//	Last applied state & desired state are applied against the
// observed state to derive the final state.
//
// NOTE:
//	This logic is essentially a 2 way merge since last applied &
// desired state are same. Both last applied & desired state
// equals the desired state.
func (r *UpdateStatesBuilder) runApplyForExplicitUpdates() {
	for _, obj := range r.markedForExplicitUpdates {
		final, err := apply.Merge(
			obj.UnstructuredContent(), // observed
			r.Request.Apply,           // last applied
			r.Request.Apply,           // desired update content
		)
		if err != nil {
			r.err = errors.Wrapf(
				err,
				"Can't update %q / %q: %s",
				obj.GetNamespace(),
				obj.GetName(),
				obj.GroupVersionKind().String(),
			)
			return
		}
		r.explicitUpdates = append(
			r.explicitUpdates,
			&unstructured.Unstructured{
				Object: final,
			},
		)
	}
}

// Build builds the desired resources to be updated
// at the cluster
func (r *UpdateStatesBuilder) Build() (*UpdateResponse, error) {
	if r.Request.TaskKey == "" {
		return nil, errors.Errorf(
			"Can't update: Missing task key",
		)
	}
	if r.Request.Run == nil {
		return nil, errors.Errorf(
			"Can't update: Missing run: %q",
			r.Request.TaskKey,
		)
	}
	if r.Request.Watch == nil {
		return nil, errors.Errorf(
			"Can't update: Missing watch: %q",
			r.Request.TaskKey,
		)
	}
	if len(r.Request.Apply) == 0 {
		return nil, errors.Errorf(
			"Can't update: Missing update state: %q",
			r.Request.TaskKey,
		)
	}
	if len(r.Request.TargetSelector.SelectorTerms) == 0 {
		return nil, errors.Errorf(
			"Can't update: Missing target selector: %q",
			r.Request.TaskKey,
		)
	}
	fns := []func(){
		r.filterResources,
		r.isSkipUpdate,
		r.groupResourcesByUpdateType,
		r.runApplyForDesiredUpdates,
		r.runApplyForExplicitUpdates,
	}
	for _, fn := range fns {
		fn()
		if r.err != nil {
			return nil, r.err
		}
		if r.isSkip {
			return &UpdateResponse{
				Result: &types.Result{
					Phase:   types.ResultPhaseSkipped,
					Message: "No eligible resources found to update",
				},
			}, nil
		}
	}
	return &UpdateResponse{
		ExplicitUpdates: r.explicitUpdates,
		DesiredUpdates:  r.desiredUpdates,
		Result: &types.Result{
			SkippedResourcesInfo:  r.Result.SkippedResourcesInfo,  // is set if enabled
			DesiredResourcesInfo:  r.Result.DesiredResourcesInfo,  // is set if enabled
			ExplicitResourcesInfo: r.Result.ExplicitResourcesInfo, // is set if enabled
			Phase:                 types.ResultPhaseOnline,
			Message: fmt.Sprintf(
				"Update action was successful: Desired updates %d: Explicit updates %d",
				len(r.desiredUpdates),
				len(r.explicitUpdates),
			),
		},
	}, nil
}

// BuildUpdateStates returns the desired resources
// that need to be updated at the k8s cluster
func BuildUpdateStates(request UpdateRequest) (*UpdateResponse, error) {
	r := &UpdateStatesBuilder{
		Request: request,
		Result:  &types.Result{},
	}
	return r.Build()
}
