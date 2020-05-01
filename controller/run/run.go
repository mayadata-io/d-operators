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

// Request forms the input required to execute a Run resource
type Request struct {
	IncludeInfo       map[types.IncludeInfoKey]bool
	RunCond           *types.ResourceCheck
	Tasks             []types.Task
	Run               *unstructured.Unstructured
	Watch             *unstructured.Unstructured
	ObservedResources []*unstructured.Unstructured
}

// Response holds the result of executing a Run resource
type Response struct {
	DesiredResources []*unstructured.Unstructured
	ExplicitUpdates  []*unstructured.Unstructured
	ExplicitDeletes  []*unstructured.Unstructured

	RunStatus *types.RunStatus
}

// Runnable forms the unit of execution
type Runnable struct {
	Request  Request
	Response *Response

	taskKeyRegistrar map[string]bool
	isRunCondSuccess bool

	err error
}

func (r *Runnable) validateArgs() error {
	if r.Request.Run == nil || r.Request.Run.Object == nil {
		return errors.New("Invalid run: Nil run resource")
	}
	if r.Request.Watch == nil || r.Request.Watch.Object == nil {
		return errors.New("Invalid run: Nil watch resource")
	}
	if r.Response == nil || r.Response.RunStatus == nil {
		return errors.New("Invalid run: Nil response")
	}
	if len(r.Request.Tasks) == 0 {
		return errors.New("Invalid run: No tasks to run")
	}
	return nil
}

// proceed only when **IF** condition succeeds
func (r *Runnable) runIfCondition() {
	if r.Request.RunCond == nil {
		// defaults to true if no condition is set
		r.isRunCondSuccess = true
		return
	}
	got, err := ExecuteCondition(
		AssertRequest{
			TaskKey: fmt.Sprintf(
				"%s:%s:%s",
				r.Request.Watch.GetNamespace(),
				r.Request.Watch.GetName(),
				r.Request.Watch.GroupVersionKind().String(),
			),
			Assert: &types.Assert{
				ResourceCheck: *r.Request.RunCond,
			},
			Resources: r.Request.ObservedResources,
		},
	)
	if err != nil {
		r.err = err
		return
	}
	if got.AssertResult.Phase == types.ResultPhaseAssertPassed {
		r.isRunCondSuccess = true
	} else {
		r.Response.RunStatus.Result = *got.AssertResult
		msg := got.AssertResult.Message
		if msg == "" {
			msg = "Run was skipped: If cond failed"
		}
		r.Response.RunStatus.Message = msg
	}
}

func (r *Runnable) isDuplicateTaskKey(key string) bool {
	if r.taskKeyRegistrar[key] {
		return true
	}
	// add this key to verify its duplicates in future calls
	r.taskKeyRegistrar[key] = true
	return false
}

func (r *Runnable) runAllTasks() {
	r.taskKeyRegistrar = make(map[string]bool)
	for _, task := range r.Request.Tasks {
		if r.isDuplicateTaskKey(task.Key) {
			r.Response.RunStatus.Errors = append(
				r.Response.RunStatus.Errors,
				fmt.Sprintf(
					"Duplicate task key %q",
					task.Key,
				),
			)
			continue
		}
		resp, err := ExecTask(
			TaskRequest{
				IncludeInfo:       r.Request.IncludeInfo,
				ObservedResources: r.Request.ObservedResources,
				Run:               r.Request.Run,
				Watch:             r.Request.Watch,
				Task:              task,
			},
		)
		if err != nil {
			r.Response.RunStatus.Errors = append(
				r.Response.RunStatus.Errors,
				err.Error(),
			)
			// we aggregate results from remaining tasks
			continue
		}
		r.Response.DesiredResources = append(
			r.Response.DesiredResources,
			resp.DesiredResources...,
		)
		r.Response.ExplicitUpdates = append(
			r.Response.ExplicitUpdates,
			resp.ExplicitUpdates...,
		)
		r.Response.ExplicitDeletes = append(
			r.Response.ExplicitDeletes,
			resp.ExplicitDeletes...,
		)
		r.Response.RunStatus.TaskResultList[task.Key] = *resp.Result
	}
	if len(r.Response.RunStatus.Errors) != 0 {
		r.err = errors.Errorf("One or more tasks have error(s)")
	}
}

// Run executes this task
func (r *Runnable) Run() error {
	err := r.validateArgs()
	if err != nil {
		return err
	}
	fns := []func(){
		r.runIfCondition, // if-cond
		r.runAllTasks,    // if (cond) then (run tasks)
	}
	// above functions are invoked here
	for _, fn := range fns {
		fn()
		if r.err != nil {
			return r.err
		}
		// run can be executed only when **IF** condition succeeds
		if !r.isRunCondSuccess {
			return nil
		}
	}
	return nil
}

// ExecRun executes the run resource
func ExecRun(req Request) (*Response, error) {
	r := &Runnable{
		Request: req,
		Response: &Response{
			RunStatus: &types.RunStatus{
				TaskResultList: map[string]types.TaskResult{},
			},
		},
	}
	err := r.Run()
	return r.Response, err
}
