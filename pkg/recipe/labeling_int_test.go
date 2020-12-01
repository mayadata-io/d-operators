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

func TestLabelRun(t *testing.T) {
	tasks := []types.Task{
		{
			Name: "create-ns-lbl-integration-testing-ignore",
			Create: &types.Create{
				State: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "Namespace",
						"metadata": map[string]interface{}{
							"name": "lbl-integration-testing-ignore",
						},
					},
				},
			},
		},
		{
			Name: "create-config-cm-one-in-ignore-ns",
			Create: &types.Create{
				State: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name":      "cm-one",
							"namespace": "lbl-integration-testing-ignore",
							"labels": map[string]interface{}{
								"common": "true",
								"cm-one": "true",
							},
						},
					},
				},
			},
		},
		{
			Name: "create-ns-lbl-integration-testing",
			Create: &types.Create{
				State: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "Namespace",
						"metadata": map[string]interface{}{
							"name": "lbl-integration-testing",
						},
					},
				},
			},
		},
		{
			Name: "create-config-cm-one",
			Create: &types.Create{
				State: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name":      "cm-one",
							"namespace": "lbl-integration-testing",
							"labels": map[string]interface{}{
								"common": "true",
								"cm-one": "true",
							},
						},
					},
				},
			},
		},
		{
			Name: "create-config-cm-two",
			Create: &types.Create{
				State: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name":      "cm-two",
							"namespace": "lbl-integration-testing",
							"labels": map[string]interface{}{
								"common": "true",
								"cm-two": "true",
							},
						},
					},
				},
			},
		},
		{
			Name: "set-labels-to-configs-in-ns-lbl-integration-testing",
			Label: &types.Label{
				State: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"namespace": "lbl-integration-testing",
						},
					},
				},
				ApplyLabels: map[string]string{
					"my-new-lbl": "true",
				},
			},
		},
		{
			Name: "assert-label-changes-to-config-cm-one",
			Assert: &types.Assert{
				State: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name":      "cm-one",
							"namespace": "lbl-integration-testing",
							"labels": map[string]interface{}{
								"common":     "true",
								"cm-one":     "true",
								"my-new-lbl": "true",
							},
						},
					},
				},
			},
		},
		{
			Name: "assert-label-changes-to-config-cm-two",
			Assert: &types.Assert{
				State: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name":      "cm-two",
							"namespace": "lbl-integration-testing",
							"labels": map[string]interface{}{
								"common":     "true",
								"cm-two":     "true",
								"my-new-lbl": "true",
							},
						},
					},
				},
			},
		},
		{
			Name: "assert-no-change-to-config-cm-one-in-ignore-ns",
			Assert: &types.Assert{
				State: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name":      "cm-one",
							"namespace": "lbl-integration-testing-ignore",
							"labels": map[string]interface{}{
								"common":     "true",
								"cm-one":     "true",
								"my-new-lbl": "true",
							},
						},
					},
				},
				StateCheck: &types.StateCheck{
					Operator: types.StateCheckOperatorNotEquals,
				},
			},
		},
		{
			Name: "set-labels-to-cm-one-in-ns-lbl-integration-testing",
			Label: &types.Label{
				State: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"namespace": "lbl-integration-testing",
						},
					},
				},
				IncludeByNames: []string{
					"cm-one",
				},
				ApplyLabels: map[string]string{
					"my-specific-lbl": "abcd",
				},
			},
		},
		{
			Name: "assert-specific-label-changes-to-config-cm-one",
			Assert: &types.Assert{
				State: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name":      "cm-one",
							"namespace": "lbl-integration-testing",
							"labels": map[string]interface{}{
								"common":          "true",
								"cm-one":          "true",
								"my-new-lbl":      "true",
								"my-specific-lbl": "abcd",
							},
						},
					},
				},
			},
		},
		{
			Name: "assert-no-specific-label-changes-to-config-cm-two",
			Assert: &types.Assert{
				State: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name":      "cm-two",
							"namespace": "lbl-integration-testing",
							"labels": map[string]interface{}{
								"my-specific-lbl": "abcd",
							},
						},
					},
				},
				StateCheck: &types.StateCheck{
					Operator: types.StateCheckOperatorNotEquals,
				},
			},
		},
		{
			Name: "set-labels-to-cm-one-in-ns-lbl-integration-testing-ignore",
			Label: &types.Label{
				State: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"namespace": "lbl-integration-testing-ignore",
						},
					},
				},
				IncludeByNames: []string{
					"cm-one",
				},
				ApplyLabels: map[string]string{
					"my-specific-lbl-ignore": "abcd",
				},
			},
		},
		{
			Name: "assert-specific-label-change-to-config-cm-one-in-ignore-ns",
			Assert: &types.Assert{
				State: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name":      "cm-one",
							"namespace": "lbl-integration-testing-ignore",
							"labels": map[string]interface{}{
								"common":                 "true",
								"cm-one":                 "true",
								"my-specific-lbl-ignore": "abcd",
							},
						},
					},
				},
			},
		},
		{
			Name: "assert-no-change-to-config-cm-one-due-to-changes-at-ignore",
			Assert: &types.Assert{
				State: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name":      "cm-one",
							"namespace": "lbl-integration-testing",
							"labels": map[string]interface{}{
								"my-specific-lbl-ignore": "abcd",
							},
						},
					},
				},
				StateCheck: &types.StateCheck{
					Operator: types.StateCheckOperatorNotEquals,
				},
			},
		},
		{
			Name: "assert-no-change-to-config-cm-two-due-to-changes-at-ignore",
			Assert: &types.Assert{
				State: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name":      "cm-two",
							"namespace": "lbl-integration-testing",
							"labels": map[string]interface{}{
								"my-specific-lbl-ignore": "abcd",
							},
						},
					},
				},
				StateCheck: &types.StateCheck{
					Operator: types.StateCheckOperatorNotEquals,
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
		"integration-testing-for-labels",
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
