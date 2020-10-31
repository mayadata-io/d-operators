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

// Listable helps listing desired state from the cluster
type Listable struct {
	BaseRunner
	List *types.List

	result *types.ListResult
	err    error
}

// ListableConfig helps in creating new instance of Listable
type ListableConfig struct {
	BaseRunner
	List *types.List
}

// NewLister returns a new instance of Listable
func NewLister(config ListableConfig) *Listable {
	return &Listable{
		BaseRunner: config.BaseRunner,
		List:       config.List,
		result:     &types.ListResult{},
	}
}

func (l *Listable) listCRDs() (*types.ListResult, error) {
	var crd *v1beta1.CustomResourceDefinition
	err := UnstructToTyped(l.List.State, &crd)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Failed to transform unstruct instance to crd equivalent",
		)
	}
	// use crd client to list crds
	items, err := l.crdClient.
		CustomResourceDefinitions().
		List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Failed to list crds",
		)
	}
	return &types.ListResult{
		Phase: types.ListStatusPassed,
		Message: fmt.Sprintf(
			"List CRD: Kind %s: APIVersion %s",
			crd.Spec.Names.Singular,
			crd.Spec.Group+"/"+crd.Spec.Version,
		),
		V1Beta1CRDItems: items,
	}, nil
}

func (l *Listable) listResources() (*types.ListResult, error) {
	var message = fmt.Sprintf(
		"List resources with %s / %s: GVK %s",
		l.List.State.GetNamespace(),
		l.List.State.GetName(),
		l.List.State.GroupVersionKind(),
	)
	client, err := l.GetClientForAPIVersionAndKind(
		l.List.State.GetAPIVersion(),
		l.List.State.GetKind(),
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Failed to get resource client",
		)
	}
	items, err := client.
		Namespace(l.List.State.GetNamespace()).
		List(metav1.ListOptions{}) // TODO add label selector
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Failed to list resources",
		)
	}
	return &types.ListResult{
		Phase:   types.ListStatusPassed,
		Message: message,
		Items:   items,
	}, nil
}

// Run executes applying the desired state against the
// cluster
func (l *Listable) Run() (*types.ListResult, error) {
	if l.List.State.GetKind() == "CustomResourceDefinition" {
		// list CRDs
		return l.listCRDs()
	}
	return l.listResources()
}
