// +build integration

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
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	types "mayadata.io/d-operators/types/recipe"
)

func TestCRDV1Beta1Apply(t *testing.T) {
	state := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1beta1",
			"kind":       "CustomResourceDefinition",
			"metadata": map[string]interface{}{
				"name": "somethings.openebs.io",
			},
			"spec": map[string]interface{}{
				"group": "openebs.io",
				"scope": "Namespaced",
				"names": map[string]interface{}{
					"kind":     "SomeThing",
					"listKind": "SomeThingList",
					"plural":   "somethings",
					"singular": "something",
					"shortNames": []interface{}{
						"sme",
					},
				},
				"version": "v1alpha1",
				"versions": []interface{}{
					map[string]interface{}{
						"name":    "v1alpha1",
						"served":  true,
						"storage": true,
					},
				},
			},
		},
	}

	br, err := NewDefaultBaseRunnerWithTeardown("apply crd testing")
	if err != nil {
		t.Fatalf(
			"Failed to create kubernetes base runner: %v",
			err,
		)
	}
	e, err := NewCRDV1Beta1Executor(ExecutableCRDV1Beta1Config{
		BaseRunner: *br,
		State:      state,
	})
	if err != nil {
		t.Fatalf(
			"Failed to construct crd v1beta1 executor: %v",
			err,
		)
	}

	result, err := e.Apply()
	if err != nil {
		t.Fatalf(
			"Error while testing create CRD via apply: %v: %s",
			err,
			result,
		)
	}
	if result.Phase != types.ApplyStatusPassed {
		t.Fatalf("Test failed while creating CRD via apply: %s", result)
	}

	// ---------------
	// UPDATE i.e. 3-WAY MERGE
	// ---------------
	update := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1beta1",
			"kind":       "CustomResourceDefinition",
			"metadata": map[string]interface{}{
				"name": "somethings.openebs.io",
			},
			"spec": map[string]interface{}{
				"group": "openebs.io",
				"names": map[string]interface{}{
					"plural": "somethings",
					"shortNames": []interface{}{
						"sme",
						"smethng",
					},
				},
				"version": "v1alpha1",
			},
		},
	}
	e, err = NewCRDV1Beta1Executor(ExecutableCRDV1Beta1Config{
		BaseRunner: *br,
		State:      update,
	})
	if err != nil {
		t.Fatalf(
			"Failed to construct crd v1beta1 executor: %v",
			err,
		)
	}

	result, err = e.Apply()
	if err != nil {
		t.Fatalf(
			"Error while testing update CRD via apply: %v: %s",
			err,
			result,
		)
	}
	if result.Phase != types.ApplyStatusPassed {
		t.Fatalf("Test failed while updating CRD via apply: %s", result)
	}
}
