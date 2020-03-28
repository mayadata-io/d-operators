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

import (
	"reflect"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// List is a custom datatype representing a list of
// unstructured instances
type List []*unstructured.Unstructured

// ContainsByIdentity returns true if provided target is available
// by its name, uid & other metadata fields
func (s List) ContainsByIdentity(target *unstructured.Unstructured) bool {
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

// IdentifiesAll returns true if each item in the provided
// targets is available & match by their identity
func (s List) IdentifiesAll(targets []*unstructured.Unstructured) bool {
	if len(s) == len(targets) && len(s) == 0 {
		return true
	}
	if len(s) != len(targets) {
		return false
	}
	for _, t := range targets {
		if !s.ContainsByIdentity(t) {
			// return false if any item does not match
			return false
		}
	}
	return true
}

// ContainsByEquality does a field to field match of provided target
// against the corresponding object present in this list
func (s List) ContainsByEquality(target *unstructured.Unstructured) bool {
	if target == nil || target.Object == nil {
		// we can't match a nil target
		return false
	}
	for _, src := range s {
		if src == nil || src.Object == nil {
			continue
		}
		// use meta fields as much as possible to verify
		// if target & src do not match
		if src.GetName() != target.GetName() ||
			src.GetNamespace() != target.GetNamespace() ||
			src.GetUID() != target.GetUID() ||
			src.GetKind() != target.GetKind() ||
			src.GetAPIVersion() != target.GetAPIVersion() ||
			len(src.GetAnnotations()) != len(target.GetAnnotations()) ||
			len(src.GetLabels()) != len(target.GetLabels()) ||
			len(src.GetOwnerReferences()) != len(target.GetOwnerReferences()) ||
			len(src.GetFinalizers()) != len(target.GetFinalizers()) {
			// continue since target does not match src
			continue
		}
		// Since target matches with this src based on meta
		// information we need to **verify further** by running
		// reflect based match
		return reflect.DeepEqual(target, src)
	}
	return false
}

// EqualsAll does a field to field match of each target against
// the corresponding object present in this list
func (s List) EqualsAll(targets []*unstructured.Unstructured) bool {
	if len(s) == len(targets) && len(s) == 0 {
		return true
	}
	if len(s) != len(targets) {
		return false
	}
	for _, t := range targets {
		if !s.ContainsByEquality(t) {
			return false
		}
	}
	return true
}
