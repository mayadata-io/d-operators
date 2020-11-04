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
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	types "mayadata.io/d-operators/types/recipe"
)

// Gettable helps fetching desired state from the cluster
type Gettable struct {
	BaseRunner
	Get *types.Get

	result *types.GetResult
	err    error
}

// GettableConfig helps in creating new instance of Gettable
type GettableConfig struct {
	BaseRunner
	Get *types.Get
}

// NewGetter returns a new instance of Gettable
func NewGetter(config GettableConfig) *Gettable {
	return &Gettable{
		BaseRunner: config.BaseRunner,
		Get:        config.Get,
		result:     &types.GetResult{},
	}
}

func (g *Gettable) getCRD() (*types.GetResult, error) {
	var crd *v1beta1.CustomResourceDefinition
	err := UnstructToTyped(g.Get.State, &crd)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Failed to transform unstruct instance to crd equivalent",
		)
	}
	// use crd client to get crds
	obj, err := g.crdClient.
		CustomResourceDefinitions().
		Get(g.Get.State.GetName(), metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Failed to get crd %q",
			g.Get.State.GetName(),
		)
	}
	return &types.GetResult{
		Phase: types.GetStatusPassed,
		Message: fmt.Sprintf(
			"Get CRD: Kind %s: APIVersion %s: Name %s",
			crd.Spec.Names.Singular,
			crd.Spec.Group+"/"+crd.Spec.Version,
			g.Get.State.GetName(),
		),
		V1Beta1CRD: obj,
	}, nil
}

func (g *Gettable) getResource() (*types.GetResult, error) {
	var message = fmt.Sprintf(
		"Get resource with %s / %s: GVK %s",
		g.Get.State.GetNamespace(),
		g.Get.State.GetName(),
		g.Get.State.GroupVersionKind(),
	)
	client, err := g.GetClientForAPIVersionAndKind(
		g.Get.State.GetAPIVersion(),
		g.Get.State.GetKind(),
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Failed to get resource client",
		)
	}
	obj, err := client.
		Namespace(g.Get.State.GetNamespace()).
		Get(g.Get.State.GetName(), metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Failed to get resource",
		)
	}
	return &types.GetResult{
		Phase:   types.GetStatusPassed,
		Message: message,
		Object:  obj,
	}, nil
}

// Run executes applying the desired state against the
// cluster
func (g *Gettable) Run() (*types.GetResult, error) {
	if g.Get.State.GetKind() == "CustomResourceDefinition" {
		// get CRD
		return g.getCRD()
	}
	return g.getResource()
}
