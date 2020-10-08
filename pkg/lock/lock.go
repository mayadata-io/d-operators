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

package lock

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	"mayadata.io/d-operators/pkg/kubernetes"
	dynamicdiscovery "openebs.io/metac/dynamic/discovery"
)

// Result after executing the lock
type Result struct {
	Phase   string
	Message string
}

type LockingConfig struct {
	KubeConfig    *rest.Config
	APIDiscovery  *dynamicdiscovery.APIResourceDiscovery
	LockingObj    *unstructured.Unstructured
	IsLockForever bool
}

// Locking places a cluster wide lock
type Locking struct {
	*kubernetes.Utility
	LockingObj    *unstructured.Unstructured
	IsLockForever bool

	// error as value
	err error
}

// NewLocker returns a new Locking instance
func NewLocker(config LockingConfig) (*Locking, error) {
	k8s, err := kubernetes.NewUtility(kubernetes.UtilityConfig{
		KubeConfig:   config.KubeConfig,
		APIDiscovery: config.APIDiscovery,
	})
	if err != nil {
		return nil, err
	}
	return &Locking{
		Utility:       k8s,
		LockingObj:    config.LockingObj,
		IsLockForever: config.IsLockForever,
	}, nil
}

func (r *Locking) delete() (Result, error) {
	var message = fmt.Sprintf(
		"Delete: Lock %s %s: GVK %s",
		r.LockingObj.GetNamespace(),
		r.LockingObj.GetName(),
		r.LockingObj.GroupVersionKind(),
	)
	client, err := r.GetClientForAPIVersionAndKind(
		r.LockingObj.GetAPIVersion(),
		r.LockingObj.GetKind(),
	)
	if err != nil {
		return Result{}, err
	}
	err = client.
		Namespace(r.LockingObj.GetNamespace()).
		Delete(
			r.LockingObj.GetName(),
			&metav1.DeleteOptions{},
		)
	if err != nil {
		return Result{}, err
	}
	klog.V(3).Infof(
		"Lock deleted successfully: Name %q %q",
		r.LockingObj.GetNamespace(),
		r.LockingObj.GetName(),
	)
	return Result{
		Phase:   "Passed",
		Message: message,
	}, nil
}

func (r *Locking) create() (Result, error) {
	var message = fmt.Sprintf(
		"Create: Lock %s %s: GVK %s",
		r.LockingObj.GetNamespace(),
		r.LockingObj.GetName(),
		r.LockingObj.GroupVersionKind(),
	)
	client, err := r.GetClientForAPIVersionAndKind(
		r.LockingObj.GetAPIVersion(),
		r.LockingObj.GetKind(),
	)
	if err != nil {
		return Result{}, err
	}
	_, err = client.
		Namespace(r.LockingObj.GetNamespace()).
		Create(
			r.LockingObj,
			metav1.CreateOptions{},
		)
	if err != nil {
		return Result{}, err
	}
	klog.V(3).Infof(
		"Lock created successfully: Name %q %q",
		r.LockingObj.GetNamespace(),
		r.LockingObj.GetName(),
	)
	return Result{
		Phase:   "Passed",
		Message: message,
	}, nil
}

// Lock acquires the lock and returns unlock
func (r *Locking) Lock() (Result, func() (Result, error), error) {
	lockstatus, err := r.create()
	if err != nil {
		return Result{}, nil, err
	}
	// build the unlock logic
	var unlock func() (Result, error)
	if r.IsLockForever {
		unlock = func() (Result, error) {
			// this is a noop if this lock is meant
			// to be present forever
			return Result{
				Phase:   "Passed",
				Message: "Will not unlock: Locked forever",
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
func (r *Locking) MustUnlock() (Result, error) {
	return r.delete()
}

// IsLocked returns true if lock was taken previously
func (r *Locking) IsLocked() (bool, error) {
	client, err := r.GetClientForAPIVersionAndKind(
		r.LockingObj.GetAPIVersion(),
		r.LockingObj.GetKind(),
	)
	if err != nil {
		return false, err
	}
	got, err := client.
		Namespace(r.LockingObj.GetNamespace()).
		Get(
			r.LockingObj.GetName(),
			metav1.GetOptions{},
		)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
	}
	klog.V(3).Infof(
		"Lock %q %q: Exists=%t",
		r.LockingObj.GetNamespace(),
		r.LockingObj.GetName(),
		got != nil,
	)
	return got != nil, err
}
