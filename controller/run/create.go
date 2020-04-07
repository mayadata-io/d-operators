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
)

// CreateRequest holds the config required to
// build the desired state(s) that need to be created
type CreateRequest struct {
	IncludeInfo       map[types.IncludeInfoKey]bool
	Run               *unstructured.Unstructured
	Watch             *unstructured.Unstructured
	TaskKey           string
	Apply             map[string]interface{}
	DesiredTemplate   *unstructured.Unstructured
	Replicas          int
	ObservedResources []*unstructured.Unstructured
}

// CreateResponse holds the desired states that
// need to be created at the cluster
type CreateResponse struct {
	DesiredResources []*unstructured.Unstructured
	Result           *types.Result
}

// CreateStateBuilder builds the desired resource(s)
type CreateStatesBuilder struct {
	Request CreateRequest
	Result  *types.Result

	desiredName      string
	desiredTemplate  *unstructured.Unstructured
	desiredResources []*unstructured.Unstructured

	err error
}

func (r *CreateStatesBuilder) setDesiredTemplate() {
	if r.Request.DesiredTemplate != nil {
		r.desiredTemplate = r.Request.DesiredTemplate
		return
	}
	obj := &unstructured.Unstructured{}
	obj.SetUnstructuredContent(r.Request.Apply)
	// verify if its a proper unstructured instance
	_, r.err = obj.MarshalJSON()
	if r.err != nil {
		r.err = errors.Wrapf(
			r.err,
			"Invalid desired state: %q",
			r.Request.TaskKey,
		)
		return
	}
	if len(obj.Object) == 0 {
		r.err = errors.Errorf(
			"Empty desired state: %q",
			r.Request.TaskKey,
		)
		return
	}
	if obj.GetAPIVersion() == "" {
		r.err = errors.Errorf(
			"Invalid desired state: Missing apiversion: %q",
			r.Request.TaskKey,
		)
		return
	}
	if obj.GetKind() == "" {
		r.err = errors.Errorf(
			"Invalid desired state: Missing kind: %q",
			r.Request.TaskKey,
		)
		return
	}
	r.desiredTemplate = obj
}

func (r *CreateStatesBuilder) evalDesiredName() {
	// start with generate name if any
	name := r.desiredTemplate.GetGenerateName()
	// **reset** desired state's generate name to empty
	r.desiredTemplate.SetGenerateName("")
	if name == "" {
		// use name from specified desired state
		name = r.desiredTemplate.GetName()
	}
	if name == "" {
		r.err = errors.Errorf(
			"Invalid desired state: Missing name: %q",
			r.Request.TaskKey,
		)
		return
	}
	// final desired name
	r.desiredName = name
}

func (r *CreateStatesBuilder) updateDesiredTemplateAnnotations() {
	anns := r.desiredTemplate.GetAnnotations()
	if anns == nil {
		anns = make(map[string]string)
	}
	// add various details as annotations
	anns[types.AnnotationKeyRunUID] = string(r.Request.Run.GetUID())
	anns[types.AnnotationKeyRunName] = r.Request.Run.GetName()
	anns[types.AnnotationKeyWatchUID] = string(r.Request.Watch.GetUID())
	anns[types.AnnotationKeyWatchName] = r.Request.Watch.GetName()
	anns[types.AnnotationKeyTaskKey] = r.Request.TaskKey
	// set the updated annotations
	r.desiredTemplate.SetAnnotations(anns)
}

func (r *CreateStatesBuilder) includeDesiredInfoIfEnabled(
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

func (r *CreateStatesBuilder) buildStateFromTemplateWithName(
	name string,
) *unstructured.Unstructured {
	obj := r.desiredTemplate.DeepCopy()
	obj.SetName(name)
	return obj
}

func (r *CreateStatesBuilder) buildDesiredStates() {
	if r.Request.Replicas == 1 {
		desired := r.buildStateFromTemplateWithName(r.desiredName)
		r.desiredResources = append(
			r.desiredResources,
			desired,
		)
		r.includeDesiredInfoIfEnabled(desired, "Marked for create")
		return
	}
	for i := 0; i < r.Request.Replicas; i++ {
		name := fmt.Sprintf("%s-%d", r.desiredName, i)
		desired := r.buildStateFromTemplateWithName(name)
		r.desiredResources = append(
			r.desiredResources,
			desired,
		)
		r.includeDesiredInfoIfEnabled(desired, "Marked for create")
	}
}

// Build returns the built up desired resource
func (r *CreateStatesBuilder) Build() (*CreateResponse, error) {
	if r.Request.TaskKey == "" {
		return nil, errors.Errorf(
			"Can't create: Empty task key",
		)
	}
	if r.Request.Run == nil || r.Request.Run.Object == nil {
		return nil, errors.Errorf(
			"Can't create: Nil run: %q",
			r.Request.TaskKey,
		)
	}
	if r.Request.Watch == nil || r.Request.Watch.Object == nil {
		return nil, errors.Errorf(
			"Can't create: Nil watch: %q",
			r.Request.TaskKey,
		)
	}
	if len(r.Request.Apply) == 0 &&
		(r.Request.DesiredTemplate == nil || r.Request.DesiredTemplate.Object == nil) {
		return nil, errors.Errorf(
			"Can't create: Nil desired state found: %q",
			r.Request.TaskKey,
		)
	}
	if r.Request.Replicas == 0 {
		return nil, errors.Errorf(
			"Can't create: Replicas can't be 0: %q",
			r.Request.TaskKey,
		)
	}
	fns := []func(){
		r.setDesiredTemplate,
		r.evalDesiredName,
		r.updateDesiredTemplateAnnotations,
		r.buildDesiredStates,
	}
	for _, fn := range fns {
		fn()
		if r.err != nil {
			return nil, r.err
		}
	}
	return &CreateResponse{
		DesiredResources: r.desiredResources,
		Result: &types.Result{
			DesiredResourcesInfo: r.Result.DesiredResourcesInfo, // is set if enabled
			SkippedResourcesInfo: r.Result.SkippedResourcesInfo, // is set if enabled
			Phase:                types.ResultPhaseOnline,
			Message: fmt.Sprintf(
				"Create action was successful for %d resource(s)",
				len(r.desiredResources),
			),
		},
	}, nil
}

// BuildCreateStates returns the desired resources
// that need to either applied or deleted from the k8s cluster
func BuildCreateStates(request CreateRequest) (*CreateResponse, error) {
	r := &CreateStatesBuilder{
		Request: request,
		Result:  &types.Result{}, // initialize
	}
	return r.Build()
}
