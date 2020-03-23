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
	For               metac.ResourceSelector
	ObservedResources []*unstructured.Unstructured
}

// UpdateResponse holds the updated states that
// needs to be applied against the cluster
type UpdateResponse struct {
	DesiredResources []*unstructured.Unstructured
	DesiredUpdates   []*unstructured.Unstructured
	Message          string
	Phase            types.TaskStatusPhase
}

// UpdateBuilder builds the desired resource(s) to
// be updated at the cluster
type UpdateBuilder struct {
	Request UpdateRequest

	// filtered resources that need to be updated
	filteredResources []*unstructured.Unstructured

	// resources created due to this controller & are
	// marked to be updated
	markForDesired []*unstructured.Unstructured

	// resources that are not created by this controller
	// & are marked to be updated
	markForUpdates []*unstructured.Unstructured

	desiredResources []*unstructured.Unstructured
	desiredUpdates   []*unstructured.Unstructured

	// flags if the update process should be skipped
	isSkip bool

	err error
}

func (r *UpdateBuilder) filterResources() {
	for _, observed := range r.Request.ObservedResources {
		if observed == nil || observed.Object == nil {
			r.err = errors.Errorf(
				"Can't filter resources: Nil object found",
			)
			return
		}
		e := selector.Evaluation{
			Terms:     r.Request.For.SelectorTerms,
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

func (r *UpdateBuilder) markResources() {
	var watchuid = string(r.Request.Watch.GetUID())
	for _, observed := range r.filteredResources {
		observedAnns := observed.GetAnnotations()
		if len(observedAnns) != 0 &&
			observedAnns["metac.openebs.io/created-due-to-watch"] == watchuid {
			// This observed resource was created by this controller
			// Hence this can be updated by metac without any special needs
			r.markForDesired = append(
				r.markForDesired,
				observed,
			)
			continue
		}
		// This observed resource was not created by this controller.
		// Hence, this resource needs to be updated explicitly by metac.
		r.markForUpdates = append(
			r.markForUpdates,
			observed,
		)
	}
}

// A 3-way merge is performed against the resource to arrive at
// the final resource state. This final state gets updated against
// the cluster.
//
// NOTE:
//	Last applied state & desired state are applied against the
// observed state to derive the final state.
//
// NOTE:
//	Last applied state and desired state are same here
func (r *UpdateBuilder) runApply() {
	var desired []map[string]interface{}
	var updates []map[string]interface{}
	for _, observed := range r.markForDesired {
		final, err := apply.Merge(
			observed.UnstructuredContent(), // observed
			r.Request.Apply,                // last applied
			r.Request.Apply,                // desired
		)
		if err != nil {
			r.err = err
			return
		}
		desired = append(desired, final)
	}
	for _, observed := range r.markForUpdates {
		final, err := apply.Merge(
			observed.UnstructuredContent(), // observed
			r.Request.Apply,                // last applied
			r.Request.Apply,                // desired
		)
		if err != nil {
			r.err = err
			return
		}
		updates = append(updates, final)
	}
	for _, d := range desired {
		r.desiredResources = append(
			r.desiredResources,
			&unstructured.Unstructured{
				Object: d,
			},
		)
	}
	for _, u := range updates {
		r.desiredUpdates = append(
			r.desiredUpdates,
			&unstructured.Unstructured{
				Object: u,
			},
		)
	}
}

// Build returns the built up desired resource
func (r *UpdateBuilder) Build() (UpdateResponse, error) {
	if r.Request.TaskKey == "" {
		return UpdateResponse{}, errors.Errorf(
			"Can't build desired state: Empty task key",
		)
	}
	if r.Request.Run == nil {
		return UpdateResponse{}, errors.Errorf(
			"Can't build desired state: Nil run: %q",
			r.Request.TaskKey,
		)
	}
	if r.Request.Watch == nil {
		return UpdateResponse{}, errors.Errorf(
			"Can't build desired state: Nil watch: %q",
			r.Request.TaskKey,
		)
	}
	if len(r.Request.Apply) == 0 {
		return UpdateResponse{}, errors.Errorf(
			"Can't build desired state: No desired state found: %q",
			r.Request.TaskKey,
		)
	}
	fns := []func(){
		r.filterResources,
		r.isSkipUpdate,
		r.markResources,
		r.runApply,
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
		DesiredUpdates:   r.desiredUpdates,
		DesiredResources: r.desiredResources,
		Phase:            types.TaskStatusPhaseOnline,
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
