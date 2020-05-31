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

package job

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// NamespaceName is a utility to hold namespace & name
// of any object
type NamespaceName struct {
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name"`
}

// String implements stringer interface
func (nn NamespaceName) String() string {
	return fmt.Sprintf("NS=%s Name=%s", nn.Namespace, nn.Name)
}

// NewNamespaceName returns a new NamespaceName from the provided
// unstructured object
func NewNamespaceName(u unstructured.Unstructured) NamespaceName {
	return NamespaceName{
		Namespace: u.GetNamespace(),
		Name:      u.GetName(),
	}
}

// NewNamespaceNameList returns a new list of NamespaceName
func NewNamespaceNameList(ul []unstructured.Unstructured) (out []NamespaceName) {
	for _, u := range ul {
		out = append(out, NewNamespaceName(u))
	}
	return
}

// ResourceMappedNamespaceNames returns a map of resource to
// corresponding list of NamespaceNames
func ResourceMappedNamespaceNames(
	given map[string][]unstructured.Unstructured,
) map[string][]NamespaceName {
	var out = make(map[string][]NamespaceName)
	for resource, list := range given {
		var nsNames []NamespaceName
		for _, unstruct := range list {
			nsNames = append(
				nsNames,
				NewNamespaceName(unstruct),
			)
		}
		out[resource] = nsNames
	}
	return out
}
