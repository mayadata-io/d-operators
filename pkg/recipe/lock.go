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

package recipe

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"mayadata.io/d-operators/common/pointer"
	types "mayadata.io/d-operators/types/recipe"
	"openebs.io/metac/dynamic/clientset"
)

// LockRunner executes a lock task
type LockRunner struct {
	BaseRunner
	LockForever bool
	Task        types.Task

	// Number of tasks that are scoped in this lock
	ProtectedTaskCount int
}

func (r *LockRunner) delete() (types.TaskResult, error) {
	var message = fmt.Sprintf(
		"Delete: Lock %s %s: GVK %s",
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
		return types.TaskResult{}, err
	}
	err = client.
		Namespace(r.Task.Apply.State.GetNamespace()).
		Delete(
			r.Task.Apply.State.GetName(),
			&metav1.DeleteOptions{},
		)
	if err != nil {
		return types.TaskResult{}, err
	}
	klog.V(3).Infof(
		"Lock deleted successfully: Name %q %q",
		r.Task.Apply.State.GetNamespace(),
		r.Task.Apply.State.GetName(),
	)
	return types.TaskResult{
		// last step is always the unlock
		Step:     r.ProtectedTaskCount + 1,
		Internal: pointer.Bool(true),
		Phase:    types.TaskStatusPassed,
		Message:  message,
	}, nil
}

func (r *LockRunner) create() (types.TaskResult, error) {
	var message = fmt.Sprintf(
		"Create: Lock %s %s: GVK %s",
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
		return types.TaskResult{}, err
	}
	_, err = client.
		Namespace(r.Task.Apply.State.GetNamespace()).
		Create(
			r.Task.Apply.State,
			metav1.CreateOptions{},
		)
	if err != nil {
		return types.TaskResult{}, err
	}
	klog.V(3).Infof(
		"Lock created successfully: Name %q %q",
		r.Task.Apply.State.GetNamespace(),
		r.Task.Apply.State.GetName(),
	)
	return types.TaskResult{
		Step:     0, // 0 is reserved for lock
		Internal: pointer.Bool(true),
		Phase:    types.TaskStatusPassed,
		Message:  message,
	}, nil
}

// Lock acquires the lock and returns unlock
func (r *LockRunner) Lock() (
	types.TaskResult,
	func() (types.TaskResult, error),
	error,
) {
	lockstatus, err := r.create()
	if err != nil {
		return types.TaskResult{}, nil, err
	}
	// build the unlock logic
	var unlock func() (types.TaskResult, error)
	if r.LockForever {
		unlock = func() (types.TaskResult, error) {
			// this is a noop if this lock is meant
			// to be present forever
			return types.TaskResult{
				// last step is always the unlock
				Step:     r.ProtectedTaskCount + 1,
				Internal: pointer.Bool(true),
				Phase:    types.TaskStatusPassed,
				Message:  "Will not unlock: Locked forever",
			}, nil
		}
	} else {
		// this is a one time lock that should be removed
		unlock = r.delete
	}
	return lockstatus, unlock, nil
}

// MustUnlock executes unlock logic without considering
// at any criteria
func (r *LockRunner) MustUnlock() (types.TaskResult, error) {
	return r.delete()
}

// IsLocked returns true if lock was taken previously
func (r *LockRunner) IsLocked() (bool, error) {
	client, err := r.GetClientForAPIVersionAndKind(
		r.Task.Apply.State.GetAPIVersion(),
		r.Task.Apply.State.GetKind(),
	)
	if err != nil {
		return false, err
	}
	got, err := client.
		Namespace(r.Task.Apply.State.GetNamespace()).
		Get(
			r.Task.Apply.State.GetName(),
			metav1.GetOptions{},
		)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
	}
	klog.V(3).Infof(
		"Lock %q %q: Exists=%t",
		r.Task.Apply.State.GetNamespace(),
		r.Task.Apply.State.GetName(),
		got != nil,
	)
	return got != nil, err
}
