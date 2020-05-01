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

// TaskRequest forms the input required to execute a
// task
type TaskRequest struct {
	IncludeInfo       map[types.IncludeInfoKey]bool
	Task              types.Task
	Run               *unstructured.Unstructured
	Watch             *unstructured.Unstructured
	ObservedResources []*unstructured.Unstructured
}

// TaskResponse forms the execution output of
// the task
type TaskResponse struct {
	DesiredResources []*unstructured.Unstructured
	ExplicitUpdates  []*unstructured.Unstructured
	ExplicitDeletes  []*unstructured.Unstructured

	Result *types.TaskResult
}

// RunnableTask forms the unit of execution
type RunnableTask struct {
	Request  TaskRequest
	Response *TaskResponse

	isApplyAction  bool
	isUpdateAction bool
	isAssertAction bool

	enabled         bool
	isAssertSuccess bool

	err error
}

func (r *RunnableTask) validateArgs() error {
	if r.Request.Task.Key == "" {
		return errors.New("Invalid run task: Missing task key")
	}
	if r.Request.Run == nil {
		return errors.New("Invalid run task: Nil run resource")
	}
	if r.Request.Watch == nil {
		return errors.New("Invalid run task: Nil watch resource")
	}
	if r.Response == nil {
		return errors.New("Invalid run task: Nil response")
	}
	return nil
}

func (r *RunnableTask) init() error {
	if len(r.Request.Task.TargetSelector.SelectorTerms) != 0 {
		r.isUpdateAction = true
	}
	if len(r.Request.Task.Apply) != 0 {
		// apply action can either be used for create action
		// or update action
		r.isApplyAction = true
	}
	if r.Request.Task.Assert != nil {
		r.isAssertAction = true
	}
	if r.isAssertAction && r.isApplyAction {
		return errors.Errorf(
			"Both Assert & Apply can't be set in a task: %q",
			r.Request.Task.Key,
		)
	}
	if !r.isAssertAction && !r.isApplyAction {
		return errors.Errorf(
			"Both Assert & Apply can't be nil in a task: %q",
			r.Request.Task.Key,
		)
	}
	if r.isAssertAction && r.isUpdateAction {
		return errors.Errorf(
			"Both Assert & Update can't be set in a task: %q",
			r.Request.Task.Key,
		)
	}
	if r.isUpdateAction && !r.isApplyAction {
		return errors.Errorf(
			"Update task needs Apply to be set: %q",
			r.Request.Task.Key,
		)
	}
	return nil
}

// getTaskType returns the action that task is supposed
// to perform
func (r *RunnableTask) getTaskType() string {
	if r.isUpdateAction {
		return "Update"
	}
	if r.isAssertAction {
		return "Assert"
	}
	return "Create/Delete"
}

// execute further action only when this task is enabled
func (r *RunnableTask) isEnabled() {
	if r.Request.Task.Enabled == nil {
		// defaults to true if no condition is set
		r.enabled = true
		return
	}
	got, err := ExecuteCondition(
		AssertRequest{
			IncludeInfo: r.Request.IncludeInfo,
			TaskKey:     r.Request.Task.Key,
			Assert: &types.Assert{
				ResourceCheck: *r.Request.Task.Enabled,
			},
			Resources: r.Request.ObservedResources,
		},
	)
	if err != nil {
		r.err = err
		return
	}
	// save the if-cond result
	r.Response.Result.EnabledResult = got.AssertResult
	// did If condition pass
	if got.AssertResult.Phase == types.ResultPhaseAssertPassed {
		r.enabled = true
	}
}

// update the desired resource(s)
func (r *RunnableTask) runUpdate() {
	if !r.isUpdateAction {
		// not an update task
		return
	}
	resp, err := BuildUpdateStates(UpdateRequest{
		IncludeInfo:       r.Request.IncludeInfo,
		Run:               r.Request.Run,
		Watch:             r.Request.Watch,
		Apply:             r.Request.Task.Apply,
		TargetSelector:    r.Request.Task.TargetSelector,
		ObservedResources: r.Request.ObservedResources,
		TaskKey:           r.Request.Task.Key,
	})
	if err != nil {
		r.err = err
		return
	}
	// add the desired resources to be updated
	//
	// NOTE:
	//	These resources were created by this controller
	r.Response.DesiredResources = append(
		r.Response.DesiredResources,
		resp.DesiredUpdates...,
	)
	// add the resources that need to be updated explicitly
	//
	// NOTE:
	//	These resources were not created by this controller
	r.Response.ExplicitUpdates = append(
		r.Response.ExplicitUpdates,
		resp.ExplicitUpdates...,
	)
	// set the received result
	r.Response.Result.UpdateResult = resp.Result
}

// create or delete the desired resource(s)
func (r *RunnableTask) runCreateOrDelete() {
	if !r.isApplyAction || r.isUpdateAction {
		// not a create or delete task
		return
	}
	resp, err := BuildCreateOrDeleteStates(CreateOrDeleteRequest{
		IncludeInfo:       r.Request.IncludeInfo,
		Run:               r.Request.Run,
		Watch:             r.Request.Watch,
		Apply:             r.Request.Task.Apply,
		Replicas:          r.Request.Task.Replicas,
		ObservedResources: r.Request.ObservedResources,
		TaskKey:           r.Request.Task.Key,
	})
	if err != nil {
		r.err = err
		return
	}
	// add the desired resources i.e. either create or delete
	// the resources due to this task
	r.Response.DesiredResources = append(
		r.Response.DesiredResources,
		resp.DesiredResources...,
	)
	// add the explicit deletes i.e. resources that were not
	// created due to this task
	//
	// NOTE:
	//	Creates are never explicit. The resources that get
	// created by this controller are considered as desired
	// resources which get applied (server side k8s apply)
	// over & over in subsequent reconciliations.
	r.Response.ExplicitDeletes = append(
		r.Response.ExplicitDeletes,
		resp.ExplicitDeletes...,
	)
	// set the received result
	r.Response.Result.CreateResult = resp.CreateResult
	r.Response.Result.DeleteResult = resp.DeleteResult
}

// execute task as an assertion
func (r *RunnableTask) runAssert() {
	if !r.isAssertAction {
		// not an assert task
		return
	}
	got, err := ExecuteCondition(
		AssertRequest{
			IncludeInfo: r.Request.IncludeInfo,
			TaskKey:     r.Request.Task.Key,
			Assert:      r.Request.Task.Assert,
			Resources:   r.Request.ObservedResources,
		},
	)
	if err != nil {
		r.err = err
		return
	}
	r.Response.Result.AssertResult = got.AssertResult
}

// Run executes this task
func (r *RunnableTask) Run() error {
	err := r.validateArgs()
	if err != nil {
		return err
	}
	err = r.init()
	if err != nil {
		return err
	}
	fns := []func(){
		r.isEnabled,         // enabled
		r.runUpdate,         // if (enabled && isupdate) then (update)
		r.runCreateOrDelete, // if (enabled && isdelete) then (create or delete)
		r.runAssert,         // if (enabled && isassert) then (assert)
	}
	// above functions are invoked here
	for _, fn := range fns {
		fn()
		if r.err != nil {
			return r.err
		}
		// task can be executed only when its **IF** condition succeededs
		if !r.enabled {
			r.Response.Result.SkipResult = &types.SkipResult{
				Phase: types.ResultPhaseSkipped,
				Message: fmt.Sprintf(
					"%s task didn't run: enabled=false",
					r.getTaskType(),
				),
			}
			return nil
		}
	}
	return nil
}

// ExecTask executes the task based on the provided request
// and response
func ExecTask(req TaskRequest) (*TaskResponse, error) {
	r := &RunnableTask{
		Request: req,
		Response: &TaskResponse{
			Result: &types.TaskResult{},
		},
	}
	err := r.Run()
	if err != nil {
		return nil, err
	}
	return r.Response, nil
}
