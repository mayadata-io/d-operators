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

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	types "mayadata.io/d-operators/types/recipe"
	dynamicapply "openebs.io/metac/dynamic/apply"
)

// Applyable helps applying desired state(s) against the cluster
type Applyable struct {
	BaseRunner
	Apply *types.Apply

	result *types.ApplyResult
	err    error
}

// ApplyableConfig helps in creating new instance of Applyable
type ApplyableConfig struct {
	BaseRunner
	Apply *types.Apply
}

// NewApplier returns a new instance of Applyable
func NewApplier(config ApplyableConfig) *Applyable {
	return &Applyable{
		BaseRunner: config.BaseRunner,
		Apply:      config.Apply,
		result:     &types.ApplyResult{},
	}
}

func (a *Applyable) applyCRD() (*types.ApplyResult, error) {
	if IsCRDVersion(a.Apply.State.GetAPIVersion(), "v1") {
		e, err := NewCRDV1Executor(ExecutableCRDV1Config{
			BaseRunner:      a.BaseRunner,
			IgnoreDiscovery: a.Apply.IgnoreDiscovery,
			State:           a.Apply.State,
		})
		if err != nil {
			return nil, err
		}
		return e.Apply()
	} else if IsCRDVersion(a.Apply.State.GetAPIVersion(), "v1beta1") {
		e, err := NewCRDV1Beta1Executor(ExecutableCRDV1Beta1Config{
			BaseRunner:      a.BaseRunner,
			IgnoreDiscovery: a.Apply.IgnoreDiscovery,
			State:           a.Apply.State,
		})
		if err != nil {
			return nil, err
		}
		return e.Apply()
	} else {
		return nil, errors.Errorf(
			"Unsupported CRD API version %q",
			a.Apply.State.GetAPIVersion(),
		)
	}
}

func (a *Applyable) createResource() (*types.ApplyResult, error) {
	var message = fmt.Sprintf(
		"Create resource %s %s: GVK %s",
		a.Apply.State.GetNamespace(),
		a.Apply.State.GetName(),
		a.Apply.State.GroupVersionKind(),
	)
	client, err := a.GetClientForAPIVersionAndKind(
		a.Apply.State.GetAPIVersion(),
		a.Apply.State.GetKind(),
	)
	if err != nil {
		return nil, err
	}
	_, err = client.
		Namespace(a.Apply.State.GetNamespace()).
		Create(
			a.Apply.State,
			metav1.CreateOptions{},
		)
	if err != nil {
		return nil, err
	}
	a.AddToTeardown(func() error {
		_, err := client.
			Namespace(a.Apply.State.GetNamespace()).
			Get(
				a.Apply.State.GetName(),
				metav1.GetOptions{},
			)
		if err != nil && apierrors.IsNotFound(err) {
			// nothing to do since resource is already deleted
			return nil
		}
		return client.
			Namespace(a.Apply.State.GetNamespace()).
			Delete(
				a.Apply.State.GetName(),
				&metav1.DeleteOptions{},
			)
	})
	return &types.ApplyResult{
		Phase:   types.ApplyStatusPassed,
		Message: message,
	}, nil
}

func (a *Applyable) updateResource() (*types.ApplyResult, error) {
	var message = fmt.Sprintf(
		"Update resource %s %s: GVK %s",
		a.Apply.State.GetNamespace(),
		a.Apply.State.GetName(),
		a.Apply.State.GroupVersionKind(),
	)
	err := a.Retry.Waitf(
		func() (bool, error) {
			// get appropriate dynamic client
			client, err := a.GetClientForAPIVersionAndKind(
				a.Apply.State.GetAPIVersion(),
				a.Apply.State.GetKind(),
			)
			if err != nil {
				return a.IsFailFastOnDiscoveryError(), err
			}
			// get the resource from cluster to update
			target, err := client.
				Namespace(a.Apply.State.GetNamespace()).
				Get(
					a.Apply.State.GetName(),
					metav1.GetOptions{},
				)
			if err != nil {
				return false, err
			}
			merged := &unstructured.Unstructured{}
			// 3-way merge
			merged.Object, err = dynamicapply.Merge(
				target.UnstructuredContent(),        // observed
				a.Apply.State.UnstructuredContent(), // last applied
				a.Apply.State.UnstructuredContent(), // desired
			)
			if err != nil {
				return false, err
			}
			// update the final merged state
			//
			// NOTE:
			//	At this point we are performing a server
			// side apply against the resource
			_, err = client.
				Namespace(a.Apply.State.GetNamespace()).
				Update(
					merged,
					metav1.UpdateOptions{},
				)
			if err != nil {
				return false, err
			}
			return true, nil
		},
		message,
	)
	if err != nil {
		return nil, err
	}
	return &types.ApplyResult{
		Phase:   types.ApplyStatusPassed,
		Message: message,
	}, nil
}

func (a *Applyable) applyResource() (*types.ApplyResult, error) {
	message := fmt.Sprintf(
		"Apply resource %s %s: GVK %s",
		a.Apply.State.GetNamespace(),
		a.Apply.State.GetName(),
		a.Apply.State.GroupVersionKind(),
	)
	err := a.Retry.Waitf(
		func() (bool, error) {
			var err error
			client, err := a.GetClientForAPIVersionAndKind(
				a.Apply.State.GetAPIVersion(),
				a.Apply.State.GetKind(),
			)
			if err != nil {
				return a.IsFailFastOnDiscoveryError(), err
			}
			_, err = client.
				Namespace(a.Apply.State.GetNamespace()).
				Get(
					a.Apply.State.GetName(),
					metav1.GetOptions{},
				)
			if err != nil {
				if apierrors.IsNotFound(err) {
					// condition exits since this is valid
					return true, err
				}
				return false, err
			}
			return true, nil
		},
		message,
	)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// this is a **create** operation
			return a.createResource()
		}
		return nil, err
	}
	// this is an **update** operation
	return a.updateResource()
}

// Run executes applying the desired state against the
// cluster
func (a *Applyable) Run() (*types.ApplyResult, error) {
	if a.Apply.State.GetKind() == "CustomResourceDefinition" {
		// swtich to applying CRD
		return a.applyCRD()
	}
	return a.applyResource()
}
