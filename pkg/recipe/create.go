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

func (c *Creatable) createCRD() (*types.CreateResult, error) {
	if IsCRDVersion(c.Create.State.GetAPIVersion(), "v1") {
		e, err := NewCRDV1Executor(ExecutableCRDV1Config{
			BaseRunner:      c.BaseRunner,
			IgnoreDiscovery: c.Create.IgnoreDiscovery,
			State:           c.Create.State,
		})
		if err != nil {
			return nil, err
		}
		return e.Create()
	} else if IsCRDVersion(c.Create.State.GetAPIVersion(), "v1beta1") {
		e, err := NewCRDV1Beta1Executor(ExecutableCRDV1Beta1Config{
			BaseRunner:      c.BaseRunner,
			IgnoreDiscovery: c.Create.IgnoreDiscovery,
			State:           c.Create.State,
		})
		if err != nil {
			return nil, err
		}
		return e.Create()
	} else {
		return nil, errors.Errorf(
			"Unsupported CRD API version %q",
			c.Create.State.GetAPIVersion(),
		)
	}
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
