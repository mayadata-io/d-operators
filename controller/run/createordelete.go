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
)

// CreateOrDeleteRequest provides various data required to
// build the desired state(s)
type CreateOrDeleteRequest struct {
	IncludeInfo       map[types.IncludeInfoKey]bool
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
	CreateResult     *types.TaskActionResult
	DeleteResult     *types.TaskActionResult
}

// CreateOrDeleteBuilder builds the desired resource(s)
type CreateOrDeleteBuilder struct {
	Request CreateOrDeleteRequest

	replicas        int
	isDelete        bool
	desiredTemplate *unstructured.Unstructured

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

func (r *CreateOrDeleteBuilder) evalDesiredTemplate() {
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
			"Invalid desired state: Nil Object: %q",
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

func (r *CreateOrDeleteBuilder) isDeleteAction() {
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
		// a nil spec implies a delete operation as well
		r.isDelete = true
	}
}

// Build returns the built up desired resource
func (r *CreateOrDeleteBuilder) Build() (*CreateOrDeleteResponse, error) {
	if r.Request.TaskKey == "" {
		return nil, errors.Errorf(
			"Can't build desired state: Empty task key",
		)
	}
	if r.Request.Run == nil {
		return nil, errors.Errorf(
			"Can't build desired state: Nil run: %q",
			r.Request.TaskKey,
		)
	}
	if r.Request.Watch == nil {
		return nil, errors.Errorf(
			"Can't build desired state: Nil watch: %q",
			r.Request.TaskKey,
		)
	}
	if len(r.Request.Apply) == 0 {
		return nil, errors.Errorf(
			"Can't build desired state: Nil desired state found: %q",
			r.Request.TaskKey,
		)
	}
	fns := []func(){
		r.init,
		r.evalDesiredTemplate,
		r.isDeleteAction,
	}
	for _, fn := range fns {
		fn()
		if r.err != nil {
			return nil, r.err
		}
	}
	if r.isDelete {
		// delete action
		resp, err := BuildDeleteStates(
			DeleteRequest{
				IncludeInfo:       r.Request.IncludeInfo,
				ObservedResources: r.Request.ObservedResources,
				TaskKey:           r.Request.TaskKey,
				Run:               r.Request.Run,
				Watch:             r.Request.Watch,
				DeleteTemplate:    r.desiredTemplate,
			},
		)
		if err != nil {
			return nil, err
		}
		return &CreateOrDeleteResponse{
			ExplicitDeletes: resp.ExplicitDeletes,
			DeleteResult:    resp.Result,
		}, nil
	}
	// create action
	resp, err := BuildCreateStates(
		CreateRequest{
			IncludeInfo:       r.Request.IncludeInfo,
			ObservedResources: r.Request.ObservedResources,
			TaskKey:           r.Request.TaskKey,
			Run:               r.Request.Run,
			Watch:             r.Request.Watch,
			DesiredTemplate:   r.desiredTemplate,
			Replicas:          r.replicas,
		},
	)
	if err != nil {
		return nil, err
	}
	return &CreateOrDeleteResponse{
		DesiredResources: resp.DesiredResources,
		CreateResult:     resp.Result,
	}, nil
}

// BuildCreateOrDeleteStates returns the desired resources
// that need to either applied or deleted from the k8s cluster
func BuildCreateOrDeleteStates(
	request CreateOrDeleteRequest,
) (*CreateOrDeleteResponse, error) {
	r := &CreateOrDeleteBuilder{
		Request: request,
	}
	return r.Build()
}
