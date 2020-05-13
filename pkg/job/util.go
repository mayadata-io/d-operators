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
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// UnstructToTyped transforms the provided unstruct instance
// to target type
func UnstructToTyped(
	src *unstructured.Unstructured,
	target interface{},
) error {
	if src == nil || src.UnstructuredContent() == nil {
		return errors.Errorf(
			"Failed to transform unstruct to typed: Nil unstruct",
		)
	}
	if target == nil {
		return errors.Errorf(
			"Failed to transform unstruct to typed: Nil target",
		)
	}
	return runtime.DefaultUnstructuredConverter.FromUnstructured(
		src.UnstructuredContent(),
		target,
	)
}

// TypedToUnstruct transforms the provided typed instance
// to unstructured instance
func TypedToUnstruct(
	typed interface{},
) (*unstructured.Unstructured, error) {
	if typed == nil {
		return nil, errors.Errorf(
			"Failed to transform typed to unstruct: Nil typed",
		)
	}
	got, err := runtime.DefaultUnstructuredConverter.ToUnstructured(
		typed,
	)
	if err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{
		Object: got,
	}, nil
}
