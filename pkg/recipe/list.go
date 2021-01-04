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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"openebs.io/metac/dynamic/clientset"

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

func (l *Listable) listResources() (*types.ListResult, error) {
	var message = fmt.Sprintf(
		"List resources with %s / %s: GVK %s",
		l.List.State.GetNamespace(),
		l.List.State.GetName(),
		l.List.State.GroupVersionKind(),
	)

	var client *clientset.ResourceClient
	var err error

	// ---
	// Retry in-case resource client is not yet
	// discovered
	// ---
	err = l.Retry.Waitf(
		func() (bool, error) {
			client, err = l.GetClientForAPIVersionAndKind(
				l.List.State.GetAPIVersion(),
				l.List.State.GetKind(),
			)
			return err == nil, err
		},
		message,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Failed to get resource client",
		)
	}
	items, err := client.
		Namespace(l.List.State.GetNamespace()).
		List(metav1.ListOptions{
			LabelSelector: labels.Set(
				l.List.State.GetLabels(),
			).String(),
		})
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
	return l.listResources()
}
