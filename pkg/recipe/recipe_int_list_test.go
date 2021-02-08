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
	"mayadata.io/d-operators/common/pointer"
	types "mayadata.io/d-operators/types/recipe"
)

func TestListSimpleRun(t *testing.T) {
	tasks := []types.Task{
		{
			Name: "create-ns",
			Create: &types.Create{
				State: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "Namespace",
						"metadata": map[string]interface{}{
							"name": "list-simple-integration-testing",
						},
					},
				},
			},
		},
		{
			Name: "create-configmap",
			Create: &types.Create{
				State: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name":      "cm-one",
							"namespace": "list-simple-integration-testing",
							"labels": map[string]interface{}{
								"ns": "list-simple-integration-testing",
							},
						},
					},
				},
			},
		},
		{
			Name: "create-configmap-2",
			Create: &types.Create{
				State: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name":      "cm-two",
							"namespace": "list-simple-integration-testing",
							"labels": map[string]interface{}{
								"ns": "list-simple-integration-testing",
							},
						},
					},
				},
			},
		},
		{
			Name: "assert-ns",
			Assert: &types.Assert{
				State: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "Namespace",
						"metadata": map[string]interface{}{
							"name": "list-simple-integration-testing",
						},
					},
				},
			},
		},
		{
			Name: "assert-configmap",
			Assert: &types.Assert{
				State: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name":      "cm-one",
							"namespace": "list-simple-integration-testing",
						},
					},
				},
			},
		},
		{
			Name: "assert-configmap-list",
			Assert: &types.Assert{
				State: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"namespace": "list-simple-integration-testing",
							"labels": map[string]interface{}{
								"ns": "list-simple-integration-testing",
							},
						},
					},
				},
				StateCheck: &types.StateCheck{
					Operator: types.StateCheckOperatorListCountEquals,
					Count:    pointer.Int(2),
				},
			},
		},
	}
	recipe := types.Recipe{
		Spec: types.RecipeSpec{
			Tasks: tasks,
		},
	}
	runner, err := NewNonCustomResourceRunnerWithOptions(
		"list-simple-integration-testing",
		recipe,
		NonCustomResourceRunnerOption{
			SingleTry: true,
			Teardown:  true,
		},
	)
	if err != nil {
		t.Fatalf(
			"Failed to create kubernetes runner: %v",
			err,
		)
	}
	result, err := runner.RunWithoutLocking()
	if err != nil {
		t.Fatalf("Error while testing: %v: %s", err, result)
	}
	if !(result.Phase == types.RecipeStatusCompleted ||
		result.Phase == types.RecipeStatusPassed) {
		t.Fatalf("Test failed: %s", result)
	}
}
