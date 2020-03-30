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
	Task              types.Task
	Run               *unstructured.Unstructured
	Watch             *unstructured.Unstructured
	ObservedResources []*unstructured.Unstructured
}

// TaskResponse forms the execution output of
// the task
type TaskResponse struct {
	DesiredResources []*unstructured.Unstructured
	DesiredUpdates   []*unstructured.Unstructured
	DesiredDeletes   []*unstructured.Unstructured
	Message          string
}

// RunnableTask forms the unit of execution
type RunnableTask struct {
	Request  TaskRequest
	Response *TaskResponse

	isNilApply  bool
	isNilUpdate bool
	isNilAction bool
	isNilAssert bool

	isIfCondSuccess bool
	isAssertSuccess bool

	err error
}

func (r *RunnableTask) validate() {
	if len(r.Request.Task.Target.SelectorTerms) == 0 {
		r.isNilUpdate = true
	}
	if len(r.Request.Task.Apply) == 0 {
		r.isNilApply = true
	}
	if r.Request.Task.Action == nil {
		r.isNilAction = true
	}
	if r.Request.Task.Assert == nil ||
		len(r.Request.Task.Assert.IfConditions) == 0 {
		r.isNilAssert = true
	}
	if r.isNilAssert && r.isNilApply {
		r.err = errors.Errorf(
			"Both Assert & Apply can't be nil: %q",
			r.Request.Task.Key,
		)
		return
	}
	if !r.isNilAssert && !r.isNilApply {
		r.err = errors.Errorf(
			"Both Assert & Apply can't be used together: %q",
			r.Request.Task.Key,
		)
		return
	}
	if !r.isNilUpdate && r.isNilApply {
		r.err = errors.Errorf(
			"Apply can't be nil with update action: %q",
			r.Request.Task.Key,
		)
		return
	}
	if !r.isNilAssert && !r.isNilUpdate {
		r.err = errors.Errorf(
			"Both Assert & Update can't be used together: %q",
			r.Request.Task.Key,
		)
		return
	}
}

// execute condition needed to run further action
func (r *RunnableTask) runIfCondition() {
	if r.Request.Task.If == nil {
		r.isIfCondSuccess = true
		return
	}
	r.isIfCondSuccess, r.err = ExecuteAssert(
		AssertRequest{
			Assert:    r.Request.Task.Assert,
			Resources: r.Request.ObservedResources,
		},
	)
}

// update the desired resource(s)
func (r *RunnableTask) runUpdate() {
	if r.isNilUpdate || r.isNilApply {
		// nothing to be updated
		return
	}
	if !r.isIfCondSuccess {
		return
	}
	resp, err := BuildUpdateStates(UpdateRequest{
		Run:               r.Request.Run,
		Watch:             r.Request.Watch,
		Apply:             r.Request.Task.Apply,
		Target:            r.Request.Task.Target,
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
	r.Response.DesiredDeletes = append(
		r.Response.DesiredUpdates,
		resp.ExplicitUpdates...,
	)
}

// create or delete the desired resource(s)
func (r *RunnableTask) runCreateOrDelete() {
	if r.isNilApply {
		// nothing to create or delete
		return
	}
	if !r.isIfCondSuccess {
		return
	}
	resp, err := BuildCreateOrDeleteStates(CreateOrDeleteRequest{
		Run:               r.Request.Run,
		Watch:             r.Request.Watch,
		Action:            r.Request.Task.Action,
		Apply:             r.Request.Task.Apply,
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
	r.Response.DesiredDeletes = append(
		r.Response.DesiredDeletes,
		resp.DesiredDeletes...,
	)
}

// execute task as an assertion
func (r *RunnableTask) runAssert() {
	if r.Request.Task.Assert == nil {
		// nothing to be done
		return
	}
	r.isAssertSuccess, r.err = ExecuteAssert(
		AssertRequest{
			Assert:    r.Request.Task.Assert,
			Resources: r.Request.ObservedResources,
		},
	)
}

func (r *RunnableTask) finalMessage() {
	var msg string
	if !r.isNilAssert {
		if r.isAssertSuccess {
			msg = "Assert was successful"
		} else {
			msg = "Assert failed"
		}
	}
	if !r.isNilApply {
		if r.isIfCondSuccess {
			msg = fmt.Sprintf(
				"Apply was successful: Desired resources %d: Explicit deletes %d",
				len(r.Response.DesiredResources),
				len(r.Response.DesiredDeletes),
			)
		} else {
			msg = "Apply did't run: If cond failed"
		}
	}
	if !r.isNilUpdate {
		if r.isIfCondSuccess {
			msg = fmt.Sprintf(
				"Update was successful: Updated resources %d",
				len(r.Response.DesiredResources),
			)
		} else {
			msg = "Update didn't run: If cond failed"
		}
	}
	r.Response.Message = msg
}

// Run executes this task
func (r *RunnableTask) Run() error {
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
	fns := []func(){
		r.validate,
		r.runIfCondition,    // verify if condition passes
		r.runUpdate,         // if (cond) then (update)
		r.runCreateOrDelete, // if (cond) then (create or delete)
		r.runAssert,         // assert in itself is a complete task
		r.finalMessage,      // this should be the last function
	}
	// above functions are invoked here
	for _, fn := range fns {
		fn()
		if r.err != nil {
			return r.err
		}
	}
	return nil
}

// ExecTask executes the task based on the provided request
// and response
func ExecTask(req TaskRequest) (TaskResponse, error) {
	r := &RunnableTask{
		Request:  req,
		Response: &TaskResponse{},
	}
	err := r.Run()
	if err != nil {
		return TaskResponse{}, err
	}
	return *r.Response, nil
}
