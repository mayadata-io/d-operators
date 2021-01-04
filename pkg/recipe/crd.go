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
	"strings"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	types "mayadata.io/d-operators/types/recipe"
	dynamicapply "openebs.io/metac/dynamic/apply"
)

// ExecutableCRD helps to apply or create desired CRD state
// against the cluster
type ExecutableCRD struct {
	BaseRunner
	IgnoreDiscovery  bool
	State            *unstructured.Unstructured
	DesiredCRVersion string

	// These are custom resource specific & are not
	// related to CRD
	CRResource   string
	CRAPIVersion string
}

// IsCRDVersion returns true if provided version is
// set as a suffix of APIVersion
func IsCRDVersion(apiVersion, version string) bool {
	return strings.HasSuffix(apiVersion, "/"+version)
}

func (e *ExecutableCRD) postCreate() error {
	message := fmt.Sprintf(
		"PostCreate CRD: Resource %s: APIVersion %s",
		e.CRResource,
		e.CRAPIVersion,
	)
	// discover custom resource API
	err := e.Retry.Waitf(
		func() (bool, error) {
			got := e.GetAPIForAPIVersionAndResource(
				e.CRAPIVersion,
				e.CRResource,
			)
			if got == nil {
				return e.IsFailFastOnDiscoveryError(),
					errors.Errorf(
						"Failed to discover CRD: Resource %s: APIVersion %s",
						e.CRResource,
						e.CRAPIVersion,
					)
			}
			// fetch dynamic client for the custom resource
			// corresponding to this CRD
			cli, err := e.GetClientForAPIVersionAndResource(
				e.CRAPIVersion,
				e.CRResource,
			)
			if err != nil {
				return e.IsFailFastOnDiscoveryError(), err
			}
			// A successful list implies CRD is registered & is discovered
			// in this binary
			_, err = cli.List(metav1.ListOptions{})
			if err != nil {
				return false, err
			}
			return true, nil
		},
		message,
	)
	return err
}

// Create creates the CRD in Kubernetes cluster
func (e *ExecutableCRD) Create() (*types.CreateResult, error) {
	cli, err := e.dynamicClientset.GetClientForAPIVersionAndKind(
		e.State.GetAPIVersion(), // CRD APIVersion & not CR APIVersion
		e.State.GetKind(),       // CRD Kind & not CR Kind
	)
	if err != nil {
		return nil, err
	}

	_, err = cli.Create(e.State, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	// add to teardown functions
	e.AddToTeardown(func() error {
		_, err := cli.Get(
			e.State.GetName(),
			metav1.GetOptions{},
		)
		if err != nil && apierrors.IsNotFound(err) {
			// nothing to do
			return nil
		}
		return cli.Delete(
			e.State.GetName(),
			nil,
		)
	})
	if !e.IgnoreDiscovery {
		// run an additional step to wait till this CRD
		// is discovered at apiserver
		err = e.postCreate()
		if err != nil {
			return nil, err
		}
	}
	return &types.CreateResult{
		Phase: types.CreateStatusPassed,
		Message: fmt.Sprintf(
			"Create CRD: Resource %s: APIVersion %s",
			e.CRResource,
			e.CRAPIVersion,
		),
	}, nil
}

// Update performs a 3-way merge of the CRD in the Kubernetes cluster
func (e *ExecutableCRD) Update() (*types.ApplyResult, error) {
	cli, err := e.dynamicClientset.GetClientForAPIVersionAndKind(
		e.State.GetAPIVersion(), // CRD APIVersion & not CR APIVersion
		e.State.GetKind(),       // CRD Kind & not CR Kind
	)
	if err != nil {
		return nil, err
	}

	// Get the CRD observed at the cluster
	target, err := cli.Get(
		e.State.GetName(),
		metav1.GetOptions{},
	)
	if err != nil {
		return nil, err
	}

	merged := &unstructured.Unstructured{}
	// 3-way merge
	merged.Object, err = dynamicapply.Merge(
		target.UnstructuredContent(),  // observed
		e.State.UnstructuredContent(), // last applied
		e.State.UnstructuredContent(), // desired
	)
	if err != nil {
		return nil, err
	}

	// Update the final merged state of CRD
	//
	// NOTE:
	//	This is server side apply
	_, err = cli.Update(
		merged,
		metav1.UpdateOptions{},
	)
	if err != nil {
		return nil, err
	}

	return &types.ApplyResult{
		Phase: types.ApplyStatusPassed,
		Message: fmt.Sprintf(
			"Update CRD: Resource %s: APIVersion %s",
			e.CRResource,
			e.CRAPIVersion,
		),
	}, nil
}

// Apply either creates or performs a 3-way merge of the CRD in
// the Kubernetes cluster
func (e *ExecutableCRD) Apply() (*types.ApplyResult, error) {
	// following code belongs to v1beta1 version
	message := fmt.Sprintf(
		"Apply CRD: Resource %s: APIVersion %s",
		e.CRResource,
		e.CRAPIVersion,
	)
	// use crd client to get crd
	err := e.Retry.Waitf(
		func() (bool, error) {
			cli, err :=
				e.dynamicClientset.GetClientForAPIVersionAndKind(
					e.State.GetAPIVersion(), // CRD APIVersion & not CR APIVersion
					e.State.GetKind(),       // CRD Kind & not CR Kind
				)
			if err != nil {
				return false, err
			}
			// Get the CRD observed at the cluster
			_, err = cli.Get(
				e.State.GetName(),
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
			got, err := e.Create()
			if err != nil {
				return nil, err
			}
			return &types.ApplyResult{
				Phase:   got.Phase.ToApplyStatusPhase(),
				Message: got.Message,
				Verbose: got.Verbose,
				Warning: got.Warning,
			}, nil
		}
		return nil, err
	}
	// this is an **update** operation
	return e.Update()
}
