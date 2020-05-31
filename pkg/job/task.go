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

package job

import (
	"fmt"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	types "mayadata.io/d-operators/types/job"
	"openebs.io/metac/dynamic/clientset"
)

// TaskRunner executes a Task
type TaskRunner struct {
	BaseRunner
	Task types.Task
}

func (r *TaskRunner) isDeleteFromApply() (bool, error) {
	if r.Task.Apply == nil || r.Task.Apply.State == nil {
		return false, nil
	}
	if r.Task.Apply.State != nil &&
		r.Task.Apply.Replicas != nil &&
		*r.Task.Apply.Replicas == 0 {
		// if replicas is 0 then this task is a delete action
		return true, nil
	}
	spec, found, err := unstructured.NestedFieldNoCopy(
		r.Task.Apply.State.UnstructuredContent(),
		"spec",
	)
	if err != nil {
		return false, err
	}
	if found && spec == nil {
		return true, nil
	}
	return false, nil
}

func (r *TaskRunner) delete() (*types.TaskStatus, error) {
	var message = fmt.Sprintf(
		"Delete: Resource %s %s: GVK %s",
		r.Task.Delete.State.GetNamespace(),
		r.Task.Delete.State.GetName(),
		r.Task.Delete.State.GroupVersionKind(),
	)
	var client *clientset.ResourceClient
	var err error
	err = r.Retry.Waitf(
		func() (bool, error) {
			client, err = r.GetClientForAPIVersionAndKind(
				r.Task.Delete.State.GetAPIVersion(),
				r.Task.Delete.State.GetKind(),
			)
			if err != nil {
				return r.IsFailFastOnDiscoveryError(), err
			}
			return true, nil
		},
		message,
	)
	if err != nil {
		return nil, err
	}
	err = client.
		Namespace(r.Task.Delete.State.GetNamespace()).
		Delete(
			r.Task.Delete.State.GetName(),
			&metav1.DeleteOptions{},
		)
	if err != nil {
		return nil, err
	}
	return &types.TaskStatus{
		Phase:   types.TaskStatusPassed,
		Message: message,
	}, nil
}

func (r *TaskRunner) create() (*types.TaskStatus, error) {
	c := NewCreator(CreatableConfig{
		BaseRunner: r.BaseRunner,
		Create:     r.Task.Create,
	})
	got, err := c.Run()
	if err != nil {
		return nil, err
	}
	return &types.TaskStatus{
		Phase:   got.Phase.ToTaskStatusPhase(),
		Message: got.Message,
		Verbose: got.Verbose,
		Warning: got.Warning,
	}, nil
}

func (r *TaskRunner) deleteFromApply() (*types.TaskStatus, error) {
	var message = fmt.Sprintf(
		"Apply based delete: Resource %s %s: GVK %s",
		r.Task.Apply.State.GetNamespace(),
		r.Task.Apply.State.GetName(),
		r.Task.Apply.State.GroupVersionKind(),
	)
	var client *clientset.ResourceClient
	var err error
	err = r.Retry.Waitf(
		func() (bool, error) {
			client, err = r.GetClientForAPIVersionAndKind(
				r.Task.Apply.State.GetAPIVersion(),
				r.Task.Apply.State.GetKind(),
			)
			if err != nil {
				return r.IsFailFastOnDiscoveryError(), err
			}
			return true, nil
		},
		message,
	)
	if err != nil {
		return nil, err
	}
	_, err = client.
		Namespace(r.Task.Apply.State.GetNamespace()).
		Get(
			r.Task.Apply.State.GetName(),
			metav1.GetOptions{},
		)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// return pass since the observed state is not available
			return &types.TaskStatus{
				Phase:   types.TaskStatusPassed,
				Message: message,
			}, nil
		}
		return nil, err
	}
	err = client.
		Namespace(r.Task.Apply.State.GetNamespace()).
		Delete(
			r.Task.Apply.State.GetName(),
			&metav1.DeleteOptions{},
		)
	if err != nil {
		return nil, err
	}
	return &types.TaskStatus{
		Phase:   types.TaskStatusPassed,
		Message: message,
	}, nil
}

func (r *TaskRunner) assert() (*types.TaskStatus, error) {
	a := NewAsserter(AssertableConfig{
		BaseRunner: r.BaseRunner,
		Assert:     r.Task.Assert,
	})
	got, err := a.Run()
	if err != nil {
		return nil, err
	}
	return &types.TaskStatus{
		Phase:   got.Phase.ToTaskStatusPhase(),
		Message: got.Message,
		Verbose: got.Verbose,
		Warning: got.Warning,
		Timeout: got.Timeout,
	}, nil
}

func (r *TaskRunner) apply() (*types.TaskStatus, error) {
	a := NewApplier(
		ApplyableConfig{
			BaseRunner: r.BaseRunner,
			Apply:      r.Task.Apply,
		},
	)
	got, err := a.Run()
	if err != nil {
		return nil, err
	}
	return &types.TaskStatus{
		Phase:   got.Phase.ToTaskStatusPhase(),
		Message: got.Message,
		Warning: got.Warning,
	}, nil
}

func (r *TaskRunner) tryRunAssert() (*types.TaskStatus, bool, error) {
	if r.Task.Assert == nil {
		return nil, false, nil
	}
	got, err := r.assert()
	return got, true, err
}

func (r *TaskRunner) tryRunDelete() (*types.TaskStatus, bool, error) {
	if r.Task.Delete == nil || r.Task.Delete.State == nil {
		return nil, false, nil
	}
	// delete from Delete action
	got, err := r.delete()
	return got, true, err

}

func (r *TaskRunner) tryRunApply() (*types.TaskStatus, bool, error) {
	// check if this is delete from Apply action
	isDel, err := r.isDeleteFromApply()
	if err != nil {
		return nil, false, err
	}
	if isDel {
		got, err := r.deleteFromApply()
		return got, true, err
	}
	if r.Task.Apply == nil || r.Task.Apply.State == nil {
		return nil, false, nil
	}
	got, err := r.apply()
	return got, true, err
}

func (r *TaskRunner) tryRunCreate() (*types.TaskStatus, bool, error) {
	if r.Task.Create == nil || r.Task.Create.State == nil {
		return nil, false, nil
	}
	got, err := r.create()
	return got, true, err
}

// Run executes the test step
func (r *TaskRunner) Run() (types.TaskStatus, error) {
	// only one of the probables will run
	var probables = []func() (*types.TaskStatus, bool, error){
		r.tryRunCreate,
		r.tryRunAssert,
		r.tryRunDelete,
		r.tryRunApply,
	}
	for _, fn := range probables {
		got, hasRun, err := fn()
		if err != nil {
			if r.Task.IgnoreErrorRule == types.IgnoreErrorAsWarning {
				// treat error as warning & continue
				return types.TaskStatus{
					Step:    r.TaskIndex,
					Phase:   types.TaskStatusWarning,
					Warning: err.Error(),
				}, nil
			}
			return types.TaskStatus{}, err
		}
		if !hasRun {
			continue
		}
		got.Step = r.TaskIndex
		return *got, nil
	}
	return types.TaskStatus{}, errors.Errorf(
		"Invalid task: Can't determine action",
	)
}
