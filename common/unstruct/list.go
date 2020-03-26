/*
Copyright 2020 The MayaData Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package unstruct

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

// List is a custom datatype representing a list of
// unstructured instances
type List []*unstructured.Unstructured

// Contains returns true if provided name && uid is
// is available in this List.
func (s List) Contains(target *unstructured.Unstructured) bool {
	if target == nil || target.Object == nil {
		// we don't know how to compare against a nil
		return false
	}
	for _, obj := range s {
		if obj == nil || obj.Object == nil {
			continue
		}
		if obj.GetName() == target.GetName() &&
			obj.GetNamespace() == target.GetNamespace() &&
			obj.GetUID() == target.GetUID() &&
			obj.GetKind() == target.GetKind() &&
			obj.GetAPIVersion() == target.GetAPIVersion() {
			return true
		}
	}
	return false
}

// ContainsAll returns true if each item in the provided targets
// is available in this List.
func (s List) ContainsAll(targets []*unstructured.Unstructured) bool {
	if len(s) == len(targets) && len(s) == 0 {
		return true
	}
	if len(s) < len(targets) {
		return false
	}
	for _, t := range targets {
		if !s.Contains(t) {
			// return false if any item does not match
			return false
		}
	}
	return true
}
