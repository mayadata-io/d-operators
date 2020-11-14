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
	"openebs.io/metac/dynamic/clientset"
)

// CreatableConfig helps in creating new instance of Creatable
type CreatableConfig struct {
	BaseRunner
	Create *types.Create
}

// Creatable helps creating desired state(s) against the cluster
type Creatable struct {
	BaseRunner
	Create *types.Create

	result *types.CreateResult
	err    error
}

func (c *Creatable) String() string {
	if c.Create == nil {
		return ""
	}
	return fmt.Sprintf(
		"Create action: Resource %s %s: GVK %s: TaskName %s",
		c.Create.State.GetNamespace(),
		c.Create.State.GetName(),
		c.Create.State.GroupVersionKind(),
		c.TaskName,
	)
}

// NewCreator returns a new instance of Creatable
func NewCreator(config CreatableConfig) *Creatable {
	return &Creatable{
		BaseRunner: config.BaseRunner,
		Create:     config.Create,
		result:     &types.CreateResult{},
	}
}

func (c *Creatable) postCreateCRDV1(
	crd *v1.CustomResourceDefinition,
) error {
	if len(crd.Spec.Versions) == 0 {
		return errors.Errorf(
			"Invalid CRD spec: Missing spec.versions",
		)
	}
	var versionToVerify = crd.Spec.Versions[0].Name
	message := fmt.Sprintf(
		"PostCreate CRD: Kind %s: APIVersion %s: TaskName %s",
		crd.Spec.Names.Singular,
		crd.Spec.Group+"/"+versionToVerify,
		c.TaskName,
	)
	// discover custom resource API
	return c.Retry.Waitf(
		func() (bool, error) {
			api := c.GetAPIForAPIVersionAndResource(
				crd.Spec.Group+"/"+versionToVerify,
				crd.Spec.Names.Plural,
			)
			if api == nil {
				return c.IsFailFastOnDiscoveryError(),
					errors.Errorf(
						"Failed to discover: Kind %s: APIVersion %s",
						crd.Spec.Names.Singular,
						crd.Spec.Group+"/"+versionToVerify,
					)
			}
			// fetch dynamic client for the custom resource
			// corresponding to this CRD
			customResourceClient, err := c.GetClientForAPIVersionAndResource(
				crd.Spec.Group+"/"+versionToVerify,
				crd.Spec.Names.Plural,
			)
			if err != nil {
				return c.IsFailFastOnDiscoveryError(), err
			}
			_, err = customResourceClient.List(metav1.ListOptions{})
			if err != nil {
				return false, err
			}
			return true, nil
		},
		message,
	)
}

func (c *Creatable) postCreateCRD(
	crd *v1beta1.CustomResourceDefinition,
) error {
	message := fmt.Sprintf(
		"PostCreate CRD: Kind %s: APIVersion %s: TaskName %s",
		crd.Spec.Names.Singular,
		crd.Spec.Group+"/"+crd.Spec.Version,
		c.TaskName,
	)
	// discover custom resource API
	return c.Retry.Waitf(
		func() (bool, error) {
			api := c.GetAPIForAPIVersionAndResource(
				crd.Spec.Group+"/"+crd.Spec.Version,
				crd.Spec.Names.Plural,
			)
			if api == nil {
				return c.IsFailFastOnDiscoveryError(),
					errors.Errorf(
						"Failed to discover: Kind %s: APIVersion %s",
						crd.Spec.Names.Singular,
						crd.Spec.Group+"/"+crd.Spec.Version,
					)
			}
			// fetch dynamic client for the custom resource
			// corresponding to this CRD
			customResourceClient, err := c.GetClientForAPIVersionAndResource(
				crd.Spec.Group+"/"+crd.Spec.Version,
				crd.Spec.Names.Plural,
			)
			if err != nil {
				return c.IsFailFastOnDiscoveryError(), err
			}
			_, err = customResourceClient.List(metav1.ListOptions{})
			if err != nil {
				return false, err
			}
			return true, nil
		},
		message,
	)
}

func (c *Creatable) createCRDV1() (*types.CreateResult, error) {
	var crd *v1.CustomResourceDefinition
	err := UnstructToTyped(c.Create.State, &crd)
	if err != nil {
		return nil, err
	}
	// use crd client to create crd
	crd, err = c.crdClientV1.
		CustomResourceDefinitions().
		Create(crd)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"%s",
			c,
		)
	}
	// add to teardown functions
	c.AddToTeardown(func() error {
		_, err := c.crdClientV1.
			CustomResourceDefinitions().
			Get(
				crd.GetName(),
				metav1.GetOptions{},
			)
		if err != nil && apierrors.IsNotFound(err) {
			// nothing to do
			return nil
		}
		return c.crdClientV1.
			CustomResourceDefinitions().
			Delete(
				crd.Name,
				nil,
			)
	})
	// run an additional step to wait till this CRD
	// is discovered at apiserver
	err = c.postCreateCRDV1(crd)
	if err != nil {
		return nil, err
	}
	return &types.CreateResult{
		Phase: types.CreateStatusPassed,
		Message: fmt.Sprintf(
			"Create CRD: Kind %s: APIVersion %s",
			crd.Spec.Names.Singular,
			c.Create.State.GetAPIVersion(),
		),
	}, nil
}

func (c *Creatable) createCRD() (*types.CreateResult, error) {
	ver := c.Create.State.GetAPIVersion()
	if strings.HasSuffix(ver, "/v1") {
		return c.createCRDV1()
	}

	var crd *v1beta1.CustomResourceDefinition
	err := UnstructToTyped(c.Create.State, &crd)
	if err != nil {
		return nil, err
	}
	// use crd client to create crd
	crd, err = c.crdClient.
		CustomResourceDefinitions().
		Create(crd)
	if err != nil {
		return nil, errors.Wrapf(err, "%s", c)
	}
	// add to teardown functions
	c.AddToTeardown(func() error {
		_, err := c.crdClient.
			CustomResourceDefinitions().
			Get(
				crd.GetName(),
				metav1.GetOptions{},
			)
		if err != nil && apierrors.IsNotFound(err) {
			// nothing to do
			return nil
		}
		return c.crdClient.
			CustomResourceDefinitions().
			Delete(
				crd.Name,
				nil,
			)
	})
	if !c.Create.IgnoreDiscovery {
		// run an additional step to wait till this CRD
		// is discovered at apiserver
		err = c.postCreateCRD(crd)
		if err != nil {
			return nil, err
		}
	}
	return &types.CreateResult{
		Phase: types.CreateStatusPassed,
		Message: fmt.Sprintf(
			"Create CRD: Kind %s: APIVersion %s",
			crd.Spec.Names.Singular,
			crd.Spec.Group+"/"+crd.Spec.Version,
		),
	}, nil
}

func (c *Creatable) createResource(
	obj *unstructured.Unstructured,
	client *clientset.ResourceClient,
) error {
	_, err := client.
		Namespace(obj.GetNamespace()).
		Create(
			obj,
			metav1.CreateOptions{},
		)
	if err != nil {
		return err
	}
	c.AddToTeardown(func() error {
		_, err := client.
			Namespace(obj.GetNamespace()).
			Get(
				obj.GetName(),
				metav1.GetOptions{},
			)
		if err != nil && apierrors.IsNotFound(err) {
			// nothing to do since resource is already deleted
			return nil
		}
		return client.
			Namespace(obj.GetNamespace()).
			Delete(
				obj.GetName(),
				&metav1.DeleteOptions{},
			)
	})
	return nil
}

func buildNamesFromGivenState(
	obj *unstructured.Unstructured,
	replicas int,
) ([]string, error) {
	var name string
	name = obj.GetGenerateName()
	if name == "" {
		name = obj.GetName()
	}
	if name == "" {
		return nil, errors.Errorf(
			"Failed to generate names: Either name or generateName required",
		)
	}
	if replicas == 1 {
		return []string{name}, nil
	}
	var out []string
	for i := 0; i < replicas; i++ {
		out = append(out, fmt.Sprintf("%s-%d", name, i))
	}
	return out, nil
}

func (c *Creatable) createResourceReplicas() (*types.CreateResult, error) {
	replicas := 1
	if c.Create.Replicas != nil {
		replicas = *c.Create.Replicas
	}
	if replicas <= 0 {
		return nil, errors.Errorf(
			"Failed to create: Invalid replicas %d: %s",
			replicas,
			c,
		)
	}
	client, err := c.GetClientForAPIVersionAndKind(
		c.Create.State.GetAPIVersion(),
		c.Create.State.GetKind(),
	)
	if err != nil {
		return nil, err
	}
	names, err := buildNamesFromGivenState(c.Create.State, replicas)
	if err != nil {
		return nil, errors.Wrapf(err, "%s", c)
	}
	for _, name := range names {
		obj := &unstructured.Unstructured{
			Object: c.Create.State.Object,
		}
		obj.SetName(name)
		err = c.createResource(obj, client)
		if err != nil {
			return nil, errors.Wrapf(err, "%s", c)
		}
	}
	return &types.CreateResult{
		Phase:   types.CreateStatusPassed,
		Message: c.String(),
	}, nil
}

// Run creates the desired state against the cluster
func (c *Creatable) Run() (*types.CreateResult, error) {
	if c.Create.State.GetKind() == "CustomResourceDefinition" {
		// create CRD
		return c.createCRD()
	}
	return c.createResourceReplicas()
}
