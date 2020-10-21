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

package pkg

import (
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
)

// ToTyped transforms the provided unstruct instance
// to target type
func ToTyped(src *unstructured.Unstructured, target interface{}) error {
	if src == nil || src.Object == nil {
		return errors.Errorf(
			"Can't transform unstruct to typed: Nil unstruct content",
		)
	}
	if target == nil {
		return errors.Errorf(
			"Can't transform unstruct to typed: Nil target",
		)
	}
	return runtime.DefaultUnstructuredConverter.FromUnstructured(
		src.UnstructuredContent(),
		target,
	)
}

// MarshalThenUnmarshal marshals the provided src and unmarshals
// it back into the dest
func MarshalThenUnmarshal(src interface{}, dest interface{}) error {
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// SetLabels updates the given labels with the ones
// found in the provided unstructured instance
func SetLabels(obj *unstructured.Unstructured, lbls map[string]string) {
	if len(lbls) == 0 {
		return
	}
	if obj == nil || obj.Object == nil {
		return
	}
	got := obj.GetLabels()
	if got == nil {
		got = make(map[string]string)
	}
	for k, v := range lbls {
		// update given label against existing
		got[k] = v
	}
	obj.SetLabels(got)
}
