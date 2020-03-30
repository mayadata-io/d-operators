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
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	types "mayadata.io/d-operators/types/run"
	metac "openebs.io/metac/apis/metacontroller/v1alpha1"
	"openebs.io/metac/controller/common/selector"
	"openebs.io/metac/dynamic/apply"
)

// UpdateRequest provides various data required to
// build the update state(s)
type UpdateRequest struct {
	Run               *unstructured.Unstructured
	Watch             *unstructured.Unstructured
	TaskKey           string
	Apply             map[string]interface{}
	Target            metac.ResourceSelector
	ObservedResources []*unstructured.Unstructured
}

// UpdateResponse holds the updated states that
// needs to be applied against the cluster
type UpdateResponse struct {
	DesiredUpdates  []*unstructured.Unstructured
	ExplicitUpdates []*unstructured.Unstructured
	Message         string
	Phase           types.TaskStatusPhase
}

// UpdateBuilder builds the desired resource(s) to
// be updated at the cluster
type UpdateBuilder struct {
	Request UpdateRequest

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

func (r *UpdateBuilder) filterResources() {
	for _, observed := range r.Request.ObservedResources {
		if observed == nil || observed.Object == nil {
			r.err = errors.Errorf(
				"Can't filter resources: Nil observed object",
			)
			return
		}
		e := selector.Evaluation{
			Terms:     r.Request.Target.SelectorTerms,
			Target:    observed,
			Reference: r.Request.Watch,
		}
		isMatch, err := e.RunMatch()
		if err != nil {
			r.err = err
			return
		}
		if !isMatch {
			continue
		}
		r.filteredResources = append(
			r.filteredResources,
			observed,
		)
	}
}

func (r *UpdateBuilder) isSkipUpdate() {
	if len(r.filteredResources) == 0 {
		r.isSkip = true
	}
}

func (r *UpdateBuilder) groupResourcesByUpdateType() {
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
			continue
		}
		// This observed resource was not created by this controller.
		// Hence, this resource needs to be **updated explicitly** by metac.
		r.markedForExplicitUpdates = append(
			r.markedForExplicitUpdates,
			filtered,
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
func (r *UpdateBuilder) runApplyForDesiredUpdates() {
	for _, obj := range r.markedForDesiredUpdates {
		final, err := apply.Merge(
			obj.UnstructuredContent(), // observed
			r.Request.Apply,           // last applied
			r.Request.Apply,           // desired update content
		)
		if err != nil {
			r.err = errors.Wrapf(
				err,
				"Can't update: %s: %q / %q",
				obj.GroupVersionKind().String(),
				obj.GetNamespace(),
				obj.GetName(),
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
func (r *UpdateBuilder) runApplyForExplicitUpdates() {
	for _, obj := range r.markedForExplicitUpdates {
		final, err := apply.Merge(
			obj.UnstructuredContent(), // observed
			r.Request.Apply,           // last applied
			r.Request.Apply,           // desired update content
		)
		if err != nil {
			r.err = errors.Wrapf(
				err,
				"Can't update: %s: %q / %q",
				obj.GroupVersionKind().String(),
				obj.GetNamespace(),
				obj.GetName(),
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

// Build returns the built up desired resource
func (r *UpdateBuilder) Build() (UpdateResponse, error) {
	if r.Request.TaskKey == "" {
		return UpdateResponse{}, errors.Errorf(
			"Can't update: Missing task key",
		)
	}
	if r.Request.Run == nil {
		return UpdateResponse{}, errors.Errorf(
			"Can't update: Missing run: %q",
			r.Request.TaskKey,
		)
	}
	if r.Request.Watch == nil {
		return UpdateResponse{}, errors.Errorf(
			"Can't update: Missing watch: %q",
			r.Request.TaskKey,
		)
	}
	if len(r.Request.Apply) == 0 {
		return UpdateResponse{}, errors.Errorf(
			"Can't update: Missing update state: %q",
			r.Request.TaskKey,
		)
	}
	if len(r.Request.Target.SelectorTerms) == 0 {
		return UpdateResponse{}, errors.Errorf(
			"Can't update: Missing target: %q",
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
			return UpdateResponse{}, r.err
		}
		if r.isSkip {
			return UpdateResponse{
				Phase:   types.TaskStatusPhaseSkipped,
				Message: "No eligible resources to update",
			}, nil
		}
	}
	return UpdateResponse{
		ExplicitUpdates: r.explicitUpdates,
		DesiredUpdates:  r.desiredUpdates,
		Phase:           types.TaskStatusPhaseOnline,
	}, nil
}

// BuildUpdateStates returns the desired resources
// that need to be updated at the k8s cluster
func BuildUpdateStates(
	request UpdateRequest,
) (UpdateResponse, error) {
	r := &UpdateBuilder{
		Request: request,
	}
	return r.Build()
}
