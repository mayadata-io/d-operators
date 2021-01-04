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

func TestCRDApply(t *testing.T) {
	state := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "CustomResourceDefinition",
			"metadata": map[string]interface{}{
				"name": "mayastorpools.openebs.io",
			},
			"spec": map[string]interface{}{
				"group": "openebs.io",
				"scope": "Namespaced",
				"names": map[string]interface{}{
					"kind":     "MayastorPool",
					"listKind": "MayastorPoolList",
					"plural":   "mayastorpools",
					"singular": "mayastorpool",
					"shortNames": []interface{}{
						"msp",
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
	}

	br, err := NewDefaultBaseRunnerWithTeardown("apply crd testing")
	if err != nil {
		t.Fatalf(
			"Failed to create kubernetes base runner: %v",
			err,
		)
	}
	e, err := NewCRDV1Executor(ExecutableCRDV1Config{
		BaseRunner: *br,
		State:      state,
	})
	if err != nil {
		t.Fatalf(
			"Failed to construct crd executor: %v",
			err,
		)
	}

	result, err := e.Apply()
	if err != nil {
		t.Fatalf(
			"Error while testing create via apply: %v: %s",
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
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "CustomResourceDefinition",
			"metadata": map[string]interface{}{
				"name": "mayastorpools.openebs.io",
			},
			"spec": map[string]interface{}{
				"group": "openebs.io",
				"names": map[string]interface{}{
					"plural": "mayastorpools",
					"shortNames": []interface{}{
						"msp",
						"mayasp", // new addition
					},
				},
				"versions": []interface{}{
					map[string]interface{}{
						"name": "v1alpha1",
					},
				},
			},
		},
	}
	e, err = NewCRDV1Executor(ExecutableCRDV1Config{
		BaseRunner: *br,
		State:      update,
	})
	if err != nil {
		t.Fatalf(
			"Failed to construct crd executor: %v",
			err,
		)
	}

	result, err = e.Apply()
	if err != nil {
		t.Fatalf(
			"Error while testing update via apply: %v: %s",
			err,
			result,
		)
	}
	if result.Phase != types.ApplyStatusPassed {
		t.Fatalf("Test failed while updating CRD via apply: %s", result)
	}
}
