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

func TestListCRDsRun(t *testing.T) {
	tasks := []types.Task{
		{
			Name: "create-crd-v1",
			Create: &types.Create{
				State: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "apiextensions.k8s.io/v1",
						"kind":       "CustomResourceDefinition",
						"metadata": map[string]interface{}{
							"name": "vonelists.openebs.io",
							"labels": map[string]interface{}{
								"list-crd-v1-testing": "true",
							},
						},
						"spec": map[string]interface{}{
							"group": "openebs.io",
							"scope": "Namespaced",
							"names": map[string]interface{}{
								"kind":     "VoneList",
								"listKind": "VoneListList",
								"plural":   "vonelists",
								"singular": "vonelist",
								"shortNames": []interface{}{
									"vone",
								},
							},
							"versions": []interface{}{
								map[string]interface{}{
									"name":    "v1alpha1",
									"served":  true,
									"storage": true,
									"subresources": map[string]interface{}{
										"status": map[string]interface{}{},
									},
									"schema": map[string]interface{}{
										"openAPIV3Schema": map[string]interface{}{
											"type": "object",
											"properties": map[string]interface{}{
												"apiVersion": map[string]interface{}{
													"type": "string",
												},
												"kind": map[string]interface{}{
													"type": "string",
												},
												"metadata": map[string]interface{}{
													"type": "object",
												},
												"spec": map[string]interface{}{
													"description": "Specification of the mayastor pool.",
													"type":        "object",
													"required": []interface{}{
														"node",
														"disks",
													},
													"properties": map[string]interface{}{
														"node": map[string]interface{}{
															"description": "Name of the k8s node where the storage pool is located.",
															"type":        "string",
														},
														"disks": map[string]interface{}{
															"description": "Disk devices (paths or URIs) that should be used for the pool.",
															"type":        "array",
															"items": map[string]interface{}{
																"type": "string",
															},
														},
													},
												},
												"status": map[string]interface{}{
													"description": "Status part updated by the pool controller.",
													"type":        "object",
													"properties": map[string]interface{}{
														"state": map[string]interface{}{
															"description": "Pool state.",
															"type":        "string",
														},
														"reason": map[string]interface{}{
															"description": "Reason for the pool state value if applicable.",
															"type":        "string",
														},
														"disks": map[string]interface{}{
															"description": "Disk device URIs that are actually used for the pool.",
															"type":        "array",
															"items": map[string]interface{}{
																"type": "string",
															},
														},
														"capacity": map[string]interface{}{
															"description": "Capacity of the pool in bytes.",
															"type":        "integer",
															"format":      "int64",
															"minimum":     int64(0),
														},
														"used": map[string]interface{}{
															"description": "How many bytes are used in the pool.",
															"type":        "integer",
															"format":      "int64",
															"minimum":     int64(0),
														},
													},
												},
											},
										},
									},
									"additionalPrinterColumns": []interface{}{
										map[string]interface{}{
											"name":        "Node",
											"type":        "string",
											"description": "Node where the storage pool is located",
											"jsonPath":    ".spec.node",
										},
										map[string]interface{}{
											"name":        "State",
											"type":        "string",
											"description": "State of the storage pool",
											"jsonPath":    ".status.state",
										},
										map[string]interface{}{
											"name":     "Age",
											"type":     "date",
											"jsonPath": ".metadata.creationTimestamp",
										},
									},
								},
							},
						},
					},
				},
				IgnoreDiscovery: true,
			},
		},
		{
			Name: "create-crd-v1beta1",
			Create: &types.Create{
				State: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "apiextensions.k8s.io/v1beta1",
						"kind":       "CustomResourceDefinition",
						"metadata": map[string]interface{}{
							"name": "betalists.openebs.io",
							"labels": map[string]interface{}{
								"list-crd-v1beta1-testing": "true",
							},
						},
						"spec": map[string]interface{}{
							"group": "openebs.io",
							"scope": "Namespaced",
							"names": map[string]interface{}{
								"kind":     "BetaList",
								"listKind": "BetaListList",
								"plural":   "betalists",
								"singular": "betalist",
								"shortNames": []interface{}{
									"bl",
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
				},
				IgnoreDiscovery: true,
			},
		},
	}
	recipe := types.Recipe{
		Spec: types.RecipeSpec{
			Tasks: tasks,
		},
	}
	runner, err := NewNonCustomResourceRunnerWithOptions(
		"list-crds-integration-testing",
		recipe,
		NonCustomResourceRunnerOption{
			SingleTry: true,  // Should Work Since Discovery is False
			Teardown:  false, // Should Not Teardown For Further List Ops
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

	// ----
	// Test Lister feature for CRD v1beta
	// ----
	br, err := NewDefaultBaseRunner("list-crds-int-testing")
	if err != nil {
		t.Fatalf("Failed to create base runner: %v", err)
	}
	l1 := NewLister(ListableConfig{
		BaseRunner: *br,
		List: &types.List{
			State: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apiextensions.k8s.io/v1beta1",
					"kind":       "CustomResourceDefinition",
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"list-crd-v1beta1-testing": "true",
						},
					},
				},
			},
		},
	})
	res, err := l1.Run()
	if err != nil {
		t.Fatalf("Failed to execute lister: %v", err)
	}
	if res.Phase != types.ListStatusPassed {
		t.Fatalf("Lister execution resulted in error: %s", res)
	}
	if res.Items == nil || len(res.Items.Items) != 1 {
		t.Fatalf(
			"Invalid list count: Got %d: Want 1", len(res.Items.Items),
		)
	}

	// ----
	// Test Lister Feature For CRD v1beta Instance
	// ----
	l11 := NewLister(ListableConfig{
		BaseRunner: *br,
		List: &types.List{
			State: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "openebs.io/v1alpha1",
					"kind":       "BetaList",
				},
			},
		},
	})
	res, err = l11.Run()
	if err != nil {
		t.Fatalf("Failed to execute lister: %v", err)
	}
	if res.Phase != types.ListStatusPassed {
		t.Fatalf("Lister execution resulted in error: %s", res)
	}
	if res.Items == nil || len(res.Items.Items) != 0 {
		t.Fatalf(
			"Invalid list count: Got %d: Want 0", len(res.Items.Items),
		)
	}

	// ----
	// Test Lister feature for CRD v1
	// ----
	l2 := NewLister(ListableConfig{
		BaseRunner: *br,
		List: &types.List{
			State: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apiextensions.k8s.io/v1",
					"kind":       "CustomResourceDefinition",
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"list-crd-v1-testing": "true",
						},
					},
				},
			},
		},
	})
	res, err = l2.Run()
	if err != nil {
		t.Fatalf("Failed to execute lister: %v", err)
	}
	if res.Phase != types.ListStatusPassed {
		t.Fatalf("Lister execution resulted in error: %s", res)
	}
	if res.Items == nil || len(res.Items.Items) != 1 {
		t.Fatalf(
			"Invalid list count: Got %d: Want 1", len(res.Items.Items),
		)
	}

	// ----
	// Test Lister Feature For CRD v1 Instance
	// ----
	l22 := NewLister(ListableConfig{
		BaseRunner: *br,
		List: &types.List{
			State: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "openebs.io/v1alpha1",
					"kind":       "VoneList",
				},
			},
		},
	})
	res, err = l22.Run()
	if err != nil {
		t.Fatalf("Failed to execute lister: %v", err)
	}
	if res.Phase != types.ListStatusPassed {
		t.Fatalf("Lister execution resulted in error: %s", res)
	}
	if res.Items == nil || len(res.Items.Items) != 0 {
		t.Fatalf(
			"Invalid list count: Got %d: Want 0", len(res.Items.Items),
		)
	}
}
