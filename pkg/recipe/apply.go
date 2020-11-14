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
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
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

func (a *Applyable) postCreateCRDV1(crd *v1.CustomResourceDefinition) error {
	if len(crd.Spec.Versions) == 0 {
		return errors.Errorf(
			"Invalid CRD spec: Missing spec.versions",
		)
	}
	var versionToVerify = crd.Spec.Versions[0].Name
	message := fmt.Sprintf(
		"PostCreate CRD: Kind %s: APIVersion %s",
		crd.Spec.Names.Singular,
		crd.Spec.Group+"/"+versionToVerify,
	)
	// Is custom resource definition discovered &
	// can its resource(s) be listed
	err := a.Retry.Waitf(
		func() (bool, error) {
			got := a.GetAPIForAPIVersionAndResource(
				crd.Spec.Group+"/"+versionToVerify,
				crd.Spec.Names.Plural,
			)
			if got == nil {
				return a.IsFailFastOnDiscoveryError(),
					errors.Errorf(
						"Failed to discover: Kind %s: APIVersion %s",
						crd.Spec.Names.Singular,
						crd.Spec.Group+"/"+versionToVerify,
					)
			}
			// fetch dynamic client for the custom resource
			// corresponding to this CRD
			customResourceClient, err := a.GetClientForAPIVersionAndResource(
				crd.Spec.Group+"/"+versionToVerify,
				crd.Spec.Names.Plural,
			)
			if err != nil {
				return a.IsFailFastOnDiscoveryError(), err
			}
			_, err = customResourceClient.List(metav1.ListOptions{})
			if err != nil {
				return false, err
			}
			return true, nil
		},
		message,
	)
	return err
}

func (a *Applyable) postCreateCRD(
	crd *v1beta1.CustomResourceDefinition,
) error {
	message := fmt.Sprintf(
		"PostCreate CRD: Kind %s: APIVersion %s",
		crd.Spec.Names.Singular,
		crd.Spec.Group+"/"+crd.Spec.Version,
	)
	// discover custom resource API
	err := a.Retry.Waitf(
		func() (bool, error) {
			got := a.GetAPIForAPIVersionAndResource(
				crd.Spec.Group+"/"+crd.Spec.Version,
				crd.Spec.Names.Plural,
			)
			if got == nil {
				return a.IsFailFastOnDiscoveryError(),
					errors.Errorf(
						"Failed to discover: Kind %s: APIVersion %s",
						crd.Spec.Names.Singular,
						crd.Spec.Group+"/"+crd.Spec.Version,
					)
			}
			// fetch dynamic client for the custom resource
			// corresponding to this CRD
			customResourceClient, err := a.GetClientForAPIVersionAndResource(
				crd.Spec.Group+"/"+crd.Spec.Version,
				crd.Spec.Names.Plural,
			)
			if err != nil {
				return a.IsFailFastOnDiscoveryError(), err
			}
			_, err = customResourceClient.List(metav1.ListOptions{})
			if err != nil {
				return false, err
			}
			return true, nil
		},
		message,
	)
	return err
}

func (a *Applyable) createCRDV1() (*types.ApplyResult, error) {
	var crd *v1.CustomResourceDefinition
	err := UnstructToTyped(a.Apply.State, &crd)
	if err != nil {
		return nil, err
	}
	// use crd client to create crd
	crd, err = a.crdClientV1.
		CustomResourceDefinitions().
		Create(crd)
	if err != nil {
		return nil, err
	}
	// add to teardown functions
	a.AddToTeardown(func() error {
		_, err := a.crdClientV1.
			CustomResourceDefinitions().
			Get(
				crd.GetName(),
				metav1.GetOptions{},
			)
		if err != nil && apierrors.IsNotFound(err) {
			// nothing to do
			return nil
		}
		return a.crdClientV1.
			CustomResourceDefinitions().
			Delete(
				crd.Name,
				nil,
			)
	})
	// run an additional step to wait till this CRD
	// is discovered at apiserver
	err = a.postCreateCRDV1(crd)
	if err != nil {
		return nil, err
	}
	return &types.ApplyResult{
		Phase: types.ApplyStatusPassed,
		Message: fmt.Sprintf(
			"Create CRD: Kind %s: APIVersion %s",
			crd.Spec.Names.Singular,
			a.Apply.State.GetAPIVersion(),
		),
	}, nil
}

func (a *Applyable) createCRD() (*types.ApplyResult, error) {
	var crd *v1beta1.CustomResourceDefinition
	err := UnstructToTyped(a.Apply.State, &crd)
	if err != nil {
		return nil, err
	}
	// use crd client to create crd
	crd, err = a.crdClient.
		CustomResourceDefinitions().
		Create(crd)
	if err != nil {
		return nil, err
	}
	// add to teardown functions
	a.AddToTeardown(func() error {
		_, err := a.crdClient.
			CustomResourceDefinitions().
			Get(
				crd.GetName(),
				metav1.GetOptions{},
			)
		if err != nil && apierrors.IsNotFound(err) {
			// nothing to do
			return nil
		}
		return a.crdClient.
			CustomResourceDefinitions().
			Delete(
				crd.Name,
				nil,
			)
	})
	if !a.Apply.IgnoreDiscovery {
		// run an additional step to wait till this CRD
		// is discovered at apiserver
		err = a.postCreateCRD(crd)
		if err != nil {
			return nil, err
		}
	}
	return &types.ApplyResult{
		Phase: types.ApplyStatusPassed,
		Message: fmt.Sprintf(
			"Create CRD: Kind %s: APIVersion %s",
			crd.Spec.Names.Singular,
			crd.Spec.Group+"/"+crd.Spec.Version,
		),
	}, nil
}

func (a *Applyable) updateCRDV1() (*types.ApplyResult, error) {
	var crd *v1.CustomResourceDefinition
	// transform to typed CRD to make use of crd client
	err := UnstructToTyped(a.Apply.State, &crd)
	if err != nil {
		return nil, err
	}
	// get the CRD observed at the cluster
	target, err := a.crdClientV1.
		CustomResourceDefinitions().
		Get(
			a.Apply.State.GetName(),
			metav1.GetOptions{},
		)
	if err != nil {
		return nil, err
	}
	// tansform back to unstruct type to run 3-way merge
	targetAsUnstruct, err := TypedToUnstruct(target)
	if err != nil {
		return nil, err
	}
	merged := &unstructured.Unstructured{}
	// 3-way merge
	merged.Object, err = dynamicapply.Merge(
		targetAsUnstruct.UnstructuredContent(), // observed
		a.Apply.State.UnstructuredContent(),    // last applied
		a.Apply.State.UnstructuredContent(),    // desired
	)
	if err != nil {
		return nil, err
	}
	// transform again to typed CRD to execute update
	err = UnstructToTyped(merged, crd)
	if err != nil {
		return nil, err
	}
	// update the final merged state of CRD
	//
	// NOTE:
	//	At this point we are performing a server side
	// apply against the CRD
	_, err = a.crdClientV1.
		CustomResourceDefinitions().
		Update(
			crd,
		)
	if err != nil {
		return nil, err
	}
	return &types.ApplyResult{
		Phase: types.ApplyStatusPassed,
		Message: fmt.Sprintf(
			"Update CRD: Kind %s: APIVersion %s",
			crd.Spec.Names.Singular,
			a.Apply.State.GetAPIVersion(),
		),
	}, nil
}

func (a *Applyable) updateCRD() (*types.ApplyResult, error) {
	var crd *v1beta1.CustomResourceDefinition
	// transform to typed CRD to make use of crd client
	err := UnstructToTyped(a.Apply.State, &crd)
	if err != nil {
		return nil, err
	}
	// get the CRD observed at the cluster
	target, err := a.crdClient.
		CustomResourceDefinitions().
		Get(
			a.Apply.State.GetName(),
			metav1.GetOptions{},
		)
	if err != nil {
		return nil, err
	}
	// tansform back to unstruct type to run 3-way merge
	targetAsUnstruct, err := TypedToUnstruct(target)
	if err != nil {
		return nil, err
	}
	merged := &unstructured.Unstructured{}
	// 3-way merge
	merged.Object, err = dynamicapply.Merge(
		targetAsUnstruct.UnstructuredContent(), // observed
		a.Apply.State.UnstructuredContent(),    // last applied
		a.Apply.State.UnstructuredContent(),    // desired
	)
	if err != nil {
		return nil, err
	}
	// transform again to typed CRD to execute update
	err = UnstructToTyped(merged, crd)
	if err != nil {
		return nil, err
	}
	// update the final merged state of CRD
	//
	// NOTE:
	//	At this point we are performing a server side
	// apply against the CRD
	_, err = a.crdClient.
		CustomResourceDefinitions().
		Update(
			crd,
		)
	if err != nil {
		return nil, err
	}
	return &types.ApplyResult{
		Phase: types.ApplyStatusPassed,
		Message: fmt.Sprintf(
			"Update CRD: Kind %s: APIVersion %s",
			crd.Spec.Names.Singular,
			crd.Spec.Group+"/"+crd.Spec.Version,
		),
	}, nil
}

func (a *Applyable) applyCRDV1() (*types.ApplyResult, error) {
	var crd *v1.CustomResourceDefinition
	err := UnstructToTyped(a.Apply.State, &crd)
	if err != nil {
		return nil, err
	}
	message := fmt.Sprintf(
		"Apply CRD: Kind %s: APIVersion %s",
		crd.Spec.Names.Singular,
		a.Apply.State.GetAPIVersion(),
	)
	// use crd client to get crd
	err = a.Retry.Waitf(
		func() (bool, error) {
			_, err = a.crdClientV1.
				CustomResourceDefinitions().
				Get(
					crd.GetName(),
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
			return a.createCRDV1()
		}
		return nil, err
	}
	// this is an **update** operation
	return a.updateCRDV1()
}

func (a *Applyable) applyCRD() (*types.ApplyResult, error) {
	ver := a.Apply.State.GetAPIVersion()
	if strings.HasSuffix(ver, "/v1") {
		return a.applyCRDV1()
	}

	var crd *v1beta1.CustomResourceDefinition
	err := UnstructToTyped(a.Apply.State, &crd)
	if err != nil {
		return nil, err
	}

	// following code belongs to v1beta1 version
	message := fmt.Sprintf(
		"Apply CRD: Kind %s: APIVersion %s",
		crd.Spec.Names.Singular,
		crd.Spec.Group+"/"+crd.Spec.Version,
	)
	// use crd client to get crd
	err = a.Retry.Waitf(
		func() (bool, error) {
			_, err = a.crdClient.
				CustomResourceDefinitions().
				Get(
					crd.GetName(),
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
			return a.createCRD()
		}
		return nil, err
	}
	// this is an **update** operation
	return a.updateCRD()
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
