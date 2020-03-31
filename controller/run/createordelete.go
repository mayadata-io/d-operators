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

// CreateOrDeleteRequest provides various data required to
// build the desired state(s)
type CreateOrDeleteRequest struct {
	Run               *unstructured.Unstructured
	Watch             *unstructured.Unstructured
	TaskKey           string
	Apply             map[string]interface{}
	Replicas          *int
	ObservedResources []*unstructured.Unstructured
}

// CreateOrDeleteResponse holds the desired states that
// needs to be applied against the cluster
type CreateOrDeleteResponse struct {
	DesiredResources []*unstructured.Unstructured
	ExplicitDeletes  []*unstructured.Unstructured
	Message          string
	Phase            types.TaskResultPhase
}

// CreateOrDeleteBuilder builds the desired resource(s)
type CreateOrDeleteBuilder struct {
	Request CreateOrDeleteRequest

	replicas         int
	isDelete         bool
	desiredName      string
	desiredTemplate  *unstructured.Unstructured
	desiredResources []*unstructured.Unstructured
	explicitDeletes  []*unstructured.Unstructured

	err error
}

func (r *CreateOrDeleteBuilder) init() {
	if r.Request.Replicas == nil {
		// default to single replica of the desired state
		r.replicas = 1
	} else {
		r.replicas = *r.Request.Replicas
	}
}

func (r *CreateOrDeleteBuilder) setDesiredTemplate() {
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

func (r *CreateOrDeleteBuilder) trySetDeleteFlag() {
	if r.replicas == 0 {
		// when replicas is set to 0, it implies deleting
		// this resource from cluster
		r.isDelete = true
		return
	}
	spec, found, _ := unstructured.NestedFieldNoCopy(
		r.desiredTemplate.Object,
		"spec",
	)
	if found && spec == nil {
		// a nil spec implies a delete operation
		r.isDelete = true
		// hence no need to build the desired states
		return
	}
}

func (r *CreateOrDeleteBuilder) generateDesiredName() {
	// start with generate name if any
	name := r.desiredTemplate.GetGenerateName()
	// **reset** desired state's generate name to empty
	r.desiredTemplate.SetGenerateName("")
	if name == "" {
		// use name from specified desired state
		name = r.desiredTemplate.GetName()
	}
	if name == "" && !r.isDelete {
		r.err = errors.Errorf(
			"Invalid desired state: Missing name: %q",
			r.Request.TaskKey,
		)
		return
	}
	// final desired name
	r.desiredName = name
}

func (r *CreateOrDeleteBuilder) updateDesiredTemplateAnnotations() {
	if r.isDelete {
		// since this is a delete operation
		// no need to update with desired annotations
		return
	}
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

func (r *CreateOrDeleteBuilder) buildStateFromTemplateWithName(
	name string,
) *unstructured.Unstructured {
	obj := r.desiredTemplate.DeepCopy()
	obj.SetName(name)
	return obj
}

func (r *CreateOrDeleteBuilder) buildDesiredStates() {
	if r.isDelete {
		// since this is a delete operation no need
		// to build desired states
		return
	}
	if r.replicas == 1 {
		r.desiredResources = append(
			r.desiredResources,
			r.buildStateFromTemplateWithName(r.desiredName),
		)
		return
	}
	for i := 0; i < r.replicas; i++ {
		name := fmt.Sprintf("%s-%d", r.desiredName, i)
		r.desiredResources = append(
			r.desiredResources,
			r.buildStateFromTemplateWithName(name),
		)
	}
}

func (r *CreateOrDeleteBuilder) markResourcesForExplicitDelete() {
	if !r.isDelete {
		// nothing to do for non delete operation
		// as well as for the cases when resource name
		// is not available
		return
	}
	var watchuid = string(r.Request.Watch.GetUID())
	for _, observed := range r.Request.ObservedResources {
		if !strings.HasPrefix(observed.GetName(), r.desiredName) ||
			r.desiredTemplate.GetKind() != observed.GetKind() ||
			r.desiredTemplate.GetAPIVersion() != observed.GetAPIVersion() ||
			r.desiredTemplate.GetNamespace() != observed.GetNamespace() {
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
		}
	}
}

// Build returns the built up desired resource
func (r *CreateOrDeleteBuilder) Build() (CreateOrDeleteResponse, error) {
	if r.Request.TaskKey == "" {
		return CreateOrDeleteResponse{}, errors.Errorf(
			"Can't build desired state: Empty task key",
		)
	}
	if r.Request.Run == nil {
		return CreateOrDeleteResponse{}, errors.Errorf(
			"Can't build desired state: Nil run: %q",
			r.Request.TaskKey,
		)
	}
	if r.Request.Watch == nil {
		return CreateOrDeleteResponse{}, errors.Errorf(
			"Can't build desired state: Nil watch: %q",
			r.Request.TaskKey,
		)
	}
	if len(r.Request.Apply) == 0 {
		return CreateOrDeleteResponse{}, errors.Errorf(
			"Can't build desired state: Nil desired state found: %q",
			r.Request.TaskKey,
		)
	}
	fns := []func(){
		r.init,
		r.setDesiredTemplate,
		r.trySetDeleteFlag,
		r.generateDesiredName,
		r.updateDesiredTemplateAnnotations,
		r.buildDesiredStates,
		r.markResourcesForExplicitDelete,
	}
	for _, fn := range fns {
		fn()
		if r.err != nil {
			return CreateOrDeleteResponse{}, r.err
		}
	}
	return CreateOrDeleteResponse{
		ExplicitDeletes:  r.explicitDeletes,
		DesiredResources: r.desiredResources,
		Phase:            types.TaskResultPhaseOnline,
		Message: fmt.Sprintf(
			"Create/Delete was successful: Desired resources %d: Explicit deletes %d",
			len(r.desiredResources),
			len(r.explicitDeletes),
		),
	}, nil
}

// BuildCreateOrDeleteStates returns the desired resources
// that need to either applied or deleted from the k8s cluster
func BuildCreateOrDeleteStates(
	request CreateOrDeleteRequest,
) (CreateOrDeleteResponse, error) {
	r := &CreateOrDeleteBuilder{
		Request: request,
	}
	return r.Build()
}
