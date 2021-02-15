// +build integration

/*
Copyright 2021 The MayaData Authors.

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

// import (
// 	"fmt"
// 	"testing"
// 	"time"
//
// 	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
// 	"k8s.io/klog/v2"
// 	"mayadata.io/d-operators/common/pointer"
// 	types "mayadata.io/d-operators/types/recipe"
// )
//
// // TestCommandCreationAndDeletion test command resource behaviour
// // 1. Create new Namespace
// // 2. Create service account
// // 3. Create clusterrole
// // 4. Create clusterrolebinding
// // 5. Create Command resource
// // 6. Make sure k8s ConfigMap is created for above command
// // 6. Make sure k8s Job is created for above command
// // 7. Make sure k8s pod is created for above command
// // 8. Delete Command resource
// // 9. Make sure K8s ConfigMap, Job & Pod related to above
// //    command resource should be deleted
// // 10. Delete clusterrolebinding
// // 11. Delete clusterrole
// // 12. Delete namespace
// func TestCommandCreationAndDeletion(t *testing.T) {
// 	tasks := []types.Task{
// 		{
// 			Name: "create-ns",
// 			Create: &types.Create{
// 				State: &unstructured.Unstructured{
// 					Object: map[string]interface{}{
// 						"apiVersion": "v1",
// 						"kind":       "Namespace",
// 						"metadata": map[string]interface{}{
// 							"name": "recipe-integration-cmd-testing",
// 						},
// 					},
// 				},
// 			},
// 		},
// 		{
// 			Name: "create-service-account",
// 			Create: &types.Create{
// 				State: &unstructured.Unstructured{
// 					Object: map[string]interface{}{
// 						"apiVersion": "v1",
// 						"kind":       "ServiceAccount",
// 						"metadata": map[string]interface{}{
// 							"name":      "recipe-integration-cmd-testing-sa",
// 							"namespace": "recipe-integration-cmd-testing",
// 						},
// 					},
// 				},
// 			},
// 		},
// 		{
// 			Name: "create-rbac-cluster-role",
// 			Create: &types.Create{
// 				State: &unstructured.Unstructured{
// 					Object: map[string]interface{}{
// 						"apiVersion": "rbac.authorization.k8s.io/v1",
// 						"kind":       "ClusterRole",
// 						"metadata": map[string]interface{}{
// 							"name": "recipe-integration-cmd-testing-dope",
// 						},
// 						"rules": []interface{}{
// 							map[string]interface{}{
// 								"apiGroups": []string{
// 									"*",
// 								},
// 								"resources": []string{
// 									"*",
// 								},
// 								"verbs": []string{
// 									"*",
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 		{
// 			Name: "create-rbac-cluster-role-binding",
// 			Create: &types.Create{
// 				State: &unstructured.Unstructured{
// 					Object: map[string]interface{}{
// 						"apiVersion": "rbac.authorization.k8s.io/v1",
// 						"kind":       "ClusterRoleBinding",
// 						"metadata": map[string]interface{}{
// 							"name": "recipe-integration-cmd-testing-dope",
// 						},
// 						"subjects": []interface{}{
// 							map[string]interface{}{
// 								"kind":      "ServiceAccount",
// 								"name":      "recipe-integration-cmd-testing-sa",
// 								"namespace": "recipe-integration-cmd-testing-dope",
// 							},
// 						},
// 						"roleRef": map[string]interface{}{
// 							"kind":     "ClusterRole",
// 							"name":     "recipe-integration-cmd-testing-dope",
// 							"apiGroup": "rbac.authorization.k8s.io",
// 						},
// 					},
// 				},
// 			},
// 		},
// 		{
// 			Name: "create command resource",
// 			Create: &types.Create{
// 				State: &unstructured.Unstructured{
// 					Object: map[string]interface{}{
// 						"apiVersion": "dope.mayadata.io/v1",
// 						"kind":       "Command",
// 						"metadata": map[string]interface{}{
// 							"name":      "testing-command",
// 							"namespace": "recipe-integration-cmd-testing",
// 						},
// 						"spec": map[string]interface{}{
// 							"commands": []interface{}{
// 								map[string]interface{}{
// 									"name":   "Get Node Information",
// 									"script": "sleep 3",
// 								},
// 							},
// 							"template": map[string]interface{}{
// 								"job": map[string]interface{}{
// 									"apiVersion": "batch/v1",
// 									"kind":       "Job",
// 									"spec": map[string]interface{}{
// 										"template": map[string]interface{}{
// 											"spec": map[string]interface{}{
// 												"serviceAccountName": "recipe-integration-cmd-testing-sa",
// 												"containers": []interface{}{
// 													map[string]interface{}{
// 														"command": []string{
// 															"/usr/bin/daction",
// 														},
// 														"image":           "mittachaitu/daction:latest",
// 														"imagePullPolicy": "IfNotPresent",
// 														"name":            "daction",
// 														"args": []interface{}{
// 															"-v=3",
// 															fmt.Sprintf("--command-name=%s", "testing-command"),
// 															fmt.Sprintf("--command-ns=%s", "dope"),
// 														},
// 													},
// 												},
// 											},
// 										},
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 		{
// 			Name: "assert-cm-lock-creation",
// 			Assert: &types.Assert{
// 				State: &unstructured.Unstructured{
// 					Object: map[string]interface{}{
// 						"apiVersion": "v1",
// 						"kind":       "ConfigMap",
// 						"metadata": map[string]interface{}{
// 							"name":      "testing-command-lock",
// 							"namespace": "recipe-integration-cmd-testing",
// 							"labels": map[string]interface{}{
// 								"command.dope.mayadata.io/name": "testing-command",
// 							},
// 						},
// 					},
// 				},
// 				StateCheck: &types.StateCheck{
// 					Operator: types.StateCheckOperatorEquals,
// 				},
// 			},
// 		},
// 		{
// 			Name: "assert-job-creation-via-command",
// 			Assert: &types.Assert{
// 				State: &unstructured.Unstructured{
// 					Object: map[string]interface{}{
// 						"apiVersion": "batch/v1",
// 						"kind":       "Job",
// 						"metadata": map[string]interface{}{
// 							"name":      "testing-command",
// 							"namespace": "recipe-integration-cmd-testing",
// 							"labels": map[string]interface{}{
// 								"command.dope.mayadata.io/controller": "true",
// 							},
// 						},
// 					},
// 				},
// 				StateCheck: &types.StateCheck{
// 					Operator: types.StateCheckOperatorEquals,
// 				},
// 			},
// 		},
// 		{
// 			Name: "assert-pod-creation-via-command-creation",
// 			Assert: &types.Assert{
// 				State: &unstructured.Unstructured{
// 					Object: map[string]interface{}{
// 						"apiVersion": "v1",
// 						"kind":       "Pod",
// 						"metadata": map[string]interface{}{
// 							"namespace": "recipe-integration-cmd-testing",
// 							"labels": map[string]interface{}{
// 								// Job name will be available as label on pod
// 								"job-name": "testing-command",
// 							},
// 						},
// 					},
// 				},
// 				StateCheck: &types.StateCheck{
// 					Operator: types.StateCheckOperatorListCountEquals,
// 					Count:    pointer.Int(1),
// 				},
// 			},
// 		},
// 		{
// 			Name: "delete-command",
// 			Delete: &types.Delete{
// 				State: &unstructured.Unstructured{
// 					Object: map[string]interface{}{
// 						"apiVersion": "dope.mayadata.io/v1",
// 						"kind":       "Command",
// 						"metadata": map[string]interface{}{
// 							"name":      "testing-command",
// 							"namespace": "recipe-integration-cmd-testing",
// 						},
// 					},
// 				},
// 			},
// 		},
// 		{
// 			Name: "assert-pod-deletion-via-command-deletion",
// 			Assert: &types.Assert{
// 				State: &unstructured.Unstructured{
// 					Object: map[string]interface{}{
// 						"apiVersion": "v1",
// 						"kind":       "Pod",
// 						"metadata": map[string]interface{}{
// 							"namespace": "recipe-integration-cmd-testing",
// 							"labels": map[string]interface{}{
// 								// Job name will be available as label on pod
// 								"job-name": "testing-command",
// 							},
// 						},
// 					},
// 				},
// 				StateCheck: &types.StateCheck{
// 					Operator: types.StateCheckOperatorListCountEquals,
// 					Count:    pointer.Int(0),
// 				},
// 			},
// 		},
// 		{
// 			Name: "assert-job-deletion-via-command-deletion",
// 			Assert: &types.Assert{
// 				State: &unstructured.Unstructured{
// 					Object: map[string]interface{}{
// 						"apiVersion": "batch/v1",
// 						"kind":       "Job",
// 						"metadata": map[string]interface{}{
// 							"name":      "testing-command",
// 							"namespace": "recipe-integration-cmd-testing",
// 							"labels": map[string]interface{}{
// 								"command.dope.mayadata.io/controller": "true",
// 							},
// 						},
// 					},
// 				},
// 				StateCheck: &types.StateCheck{
// 					Operator: types.StateCheckOperatorNotFound,
// 				},
// 			},
// 		},
// 		{
// 			Name: "assert-cm-lock-deletion-via-command-deletion",
// 			Assert: &types.Assert{
// 				State: &unstructured.Unstructured{
// 					Object: map[string]interface{}{
// 						"apiVersion": "v1",
// 						"kind":       "ConfigMap",
// 						"metadata": map[string]interface{}{
// 							"name":      "testing-command-lock",
// 							"namespace": "recipe-integration-cmd-testing",
// 							"labels": map[string]interface{}{
// 								"command.dope.mayadata.io/name": "testing-command",
// 							},
// 						},
// 					},
// 				},
// 				StateCheck: &types.StateCheck{
// 					Operator: types.StateCheckOperatorNotFound,
// 				},
// 			},
// 		},
// 		{
// 			Name: "delete-clusterrole",
// 			Delete: &types.Delete{
// 				State: &unstructured.Unstructured{
// 					Object: map[string]interface{}{
// 						"apiVersion": "rbac.authorization.k8s.io/v1",
// 						"kind":       "ClusterRole",
// 						"metadata": map[string]interface{}{
// 							"name": "recipe-integration-cmd-testing-dope",
// 						},
// 					},
// 				},
// 			},
// 		},
// 		{
// 			Name: "delete-clusterrolebinding",
// 			Delete: &types.Delete{
// 				State: &unstructured.Unstructured{
// 					Object: map[string]interface{}{
// 						"apiVersion": "rbac.authorization.k8s.io/v1",
// 						"kind":       "ClusterRoleBinding",
// 						"metadata": map[string]interface{}{
// 							"name": "recipe-integration-cmd-testing-dope",
// 						},
// 					},
// 				},
// 			},
// 		},
// 		{
// 			Name: "delete-ns",
// 			Delete: &types.Delete{
// 				State: &unstructured.Unstructured{
// 					Object: map[string]interface{}{
// 						"apiVersion": "v1",
// 						"kind":       "Namespace",
// 						"metadata": map[string]interface{}{
// 							"name": "recipe-integration-cmd-testing",
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}
// 	recipe := types.Recipe{
// 		Spec: types.RecipeSpec{
// 			Tasks: tasks,
// 		},
// 	}
// 	runner, err := NewNonCustomResourceRunnerWithOptions(
// 		"integration-testing-simple-command",
// 		recipe,
// 		NonCustomResourceRunnerOption{
// 			SingleTry: false,
// 			Teardown:  false,
// 		},
// 	)
// 	if err != nil {
// 		t.Fatalf(
// 			"Failed to create kubernetes runner: %v",
// 			err,
// 		)
// 	}
// 	result, err := runner.RunWithoutLocking()
// 	if err != nil {
// 		t.Fatalf("Error while testing: %v: %s", err, result)
// 	}
// 	if !(result.Phase == types.RecipeStatusCompleted ||
// 		result.Phase == types.RecipeStatusPassed) {
// 		t.Fatalf("Test failed: %s", result)
// 	}
// }
//
// // TestCommandRunAlways test behaviour of command when configured for run always
// // 1. Create new Namespace
// // 2. Create service account
// // 3. Create clusterrole
// // 4. Create clusterrolebinding
// // 5. Create Command resource
// // 7. Make sure k8s pod is created for above command
// // 8. Make sure phase update on command resource
// // 9. Wait for resync time period i.e 10seconds
// // 10. Make sure new job is created
// // 11. Delete Command resource
// // 12. Delete clusterrolebinding
// // 13. Delete clusterrole
// // 14. Delete namespace
// func TestCommandRunAlways(t *testing.T) {
// 	tasks := []types.Task{
// 		{
// 			Name: "create-ns",
// 			Create: &types.Create{
// 				State: &unstructured.Unstructured{
// 					Object: map[string]interface{}{
// 						"apiVersion": "v1",
// 						"kind":       "Namespace",
// 						"metadata": map[string]interface{}{
// 							"name": "recipe-integration-cmd-always",
// 						},
// 					},
// 				},
// 			},
// 		},
// 		{
// 			Name: "create-service-account",
// 			Create: &types.Create{
// 				State: &unstructured.Unstructured{
// 					Object: map[string]interface{}{
// 						"apiVersion": "v1",
// 						"kind":       "ServiceAccount",
// 						"metadata": map[string]interface{}{
// 							"name":      "recipe-integration-cmd-testing-sa",
// 							"namespace": "recipe-integration-cmd-always",
// 						},
// 					},
// 				},
// 			},
// 		},
// 		{
// 			Name: "create-rbac-cluster-role",
// 			Create: &types.Create{
// 				State: &unstructured.Unstructured{
// 					Object: map[string]interface{}{
// 						"apiVersion": "rbac.authorization.k8s.io/v1",
// 						"kind":       "ClusterRole",
// 						"metadata": map[string]interface{}{
// 							"name": "recipe-integration-cmd-always-dope",
// 						},
// 						"rules": []interface{}{
// 							map[string]interface{}{
// 								"apiGroups": []string{
// 									"*",
// 								},
// 								"resources": []string{
// 									"*",
// 								},
// 								"verbs": []string{
// 									"*",
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 		{
// 			Name: "create-rbac-cluster-role-binding",
// 			Create: &types.Create{
// 				State: &unstructured.Unstructured{
// 					Object: map[string]interface{}{
// 						"apiVersion": "rbac.authorization.k8s.io/v1",
// 						"kind":       "ClusterRoleBinding",
// 						"metadata": map[string]interface{}{
// 							"name": "recipe-integration-cmd-always-dope",
// 						},
// 						"subjects": []interface{}{
// 							map[string]interface{}{
// 								"kind":      "ServiceAccount",
// 								"name":      "recipe-integration-cmd-testing-sa",
// 								"namespace": "recipe-integration-cmd-always",
// 							},
// 						},
// 						"roleRef": map[string]interface{}{
// 							"kind":     "ClusterRole",
// 							"name":     "recipe-integration-cmd-always-dope",
// 							"apiGroup": "rbac.authorization.k8s.io",
// 						},
// 					},
// 				},
// 			},
// 		},
// 		{
// 			Name: "create command which will run Always",
// 			Create: &types.Create{
// 				State: &unstructured.Unstructured{
// 					Object: map[string]interface{}{
// 						"apiVersion": "dope.mayadata.io/v1",
// 						"kind":       "Command",
// 						"metadata": map[string]interface{}{
// 							"name":      "testing-command-run-always",
// 							"namespace": "recipe-integration-cmd-always",
// 						},
// 						"spec": map[string]interface{}{
// 							"commands": []interface{}{
// 								map[string]interface{}{
// 									"name":   "Get Node Information",
// 									"script": "sleep 3",
// 								},
// 							},
// 							"resync": map[string]interface{}{
// 								"intervalInSeconds": 10,
// 							},
// 							"enabled": map[string]interface{}{
// 								"when": "Always",
// 							},
// 							"template": map[string]interface{}{
// 								"job": map[string]interface{}{
// 									"apiVersion": "batch/v1",
// 									"kind":       "Job",
// 									"spec": map[string]interface{}{
// 										"template": map[string]interface{}{
// 											"spec": map[string]interface{}{
// 												"serviceAccountName": "recipe-integration-cmd-testing-sa",
// 												"containers": []interface{}{
// 													map[string]interface{}{
// 														"command": []string{
// 															"/usr/bin/daction",
// 														},
// 														"image":           "mittachaitu/daction:ci",
// 														"imagePullPolicy": "IfNotPresent",
// 														"name":            "daction",
// 														"args": []interface{}{
// 															"-v=3",
// 															fmt.Sprintf("--command-name=%s", "testing-command-run-always"),
// 															fmt.Sprintf("--command-ns=%s", "recipe-integration-cmd-always"),
// 														},
// 													},
// 												},
// 											},
// 										},
// 									}}}}}}},
// 		},
// 		{
// 			Name: "assert-pod-creation-via-command-creation",
// 			Assert: &types.Assert{
// 				State: &unstructured.Unstructured{
// 					Object: map[string]interface{}{
// 						"apiVersion": "v1",
// 						"kind":       "Pod",
// 						"metadata": map[string]interface{}{
// 							"namespace": "recipe-integration-cmd-always",
// 							"labels": map[string]interface{}{
// 								// Job name will be available as label on pod
// 								"job-name": "testing-command-run-always",
// 							},
// 						},
// 					},
// 				},
// 				StateCheck: &types.StateCheck{
// 					Operator: types.StateCheckOperatorListCountEquals,
// 					Count:    pointer.Int(1),
// 				},
// 			},
// 		},
// 		// Make sure command phase is updated once the pod is completed the task
// 		{
// 			Name: "assert-command-running",
// 			Assert: &types.Assert{
// 				State: &unstructured.Unstructured{
// 					Object: map[string]interface{}{
// 						"apiVersion": "dope.mayadata.io/v1",
// 						"kind":       "Command",
// 						"metadata": map[string]interface{}{
// 							"name":      "testing-command-run-always",
// 							"namespace": "recipe-integration-cmd-always",
// 						},
// 						"status": map[string]interface{}{
// 							"phase": "Running",
// 						},
// 					},
// 				},
// 				StateCheck: &types.StateCheck{
// 					Operator: types.StateCheckOperatorEquals,
// 				},
// 			},
// 		},
// 	}
// 	recipe := types.Recipe{
// 		Spec: types.RecipeSpec{
// 			Tasks: tasks,
// 		},
// 	}
// 	runner, err := NewNonCustomResourceRunnerWithOptions(
// 		"integration-testing-simple-command",
// 		recipe,
// 		NonCustomResourceRunnerOption{
// 			SingleTry: false,
// 			Teardown:  false,
// 		},
// 	)
// 	if err != nil {
// 		t.Fatalf(
// 			"Failed to create kubernetes runner: %v",
// 			err,
// 		)
// 	}
// 	result, err := runner.RunWithoutLocking()
// 	if err != nil {
// 		t.Fatalf("Error while testing: %v: %s", err, result)
// 	}
// 	if !(result.Phase == types.RecipeStatusCompleted ||
// 		result.Phase == types.RecipeStatusPassed) {
// 		t.Fatalf("Test failed: %s", result)
// 	}
// 	// Since command is configured to run periodically for every 10 seconds
// 	// Job should be created for every 10 seconds
// 	klog.Infof("Waiting for New Job Creation")
// 	time.Sleep(12)
//
// 	tasks = []types.Task{
// 		{
// 			Name: "assert-new-pod-creation-via-command-run-always",
// 			Assert: &types.Assert{
// 				State: &unstructured.Unstructured{
// 					Object: map[string]interface{}{
// 						"apiVersion": "v1",
// 						"kind":       "Pod",
// 						"metadata": map[string]interface{}{
// 							"namespace": "recipe-integration-cmd-always",
// 							"labels": map[string]interface{}{
// 								// Job name will be available as label on pod
// 								"job-name": "testing-command-run-always",
// 							},
// 						},
// 					},
// 				},
// 				StateCheck: &types.StateCheck{
// 					Operator: types.StateCheckOperatorListCountEquals,
// 					Count:    pointer.Int(1),
// 				},
// 			},
// 		},
// 		{
// 			Name: "delete-command",
// 			Delete: &types.Delete{
// 				State: &unstructured.Unstructured{
// 					Object: map[string]interface{}{
// 						"apiVersion": "dope.mayadata.io/v1",
// 						"kind":       "Command",
// 						"metadata": map[string]interface{}{
// 							"name":      "testing-command-run-always",
// 							"namespace": "recipe-integration-cmd-always",
// 						},
// 					},
// 				},
// 			},
// 		},
// 		{
// 			Name: "delete-clusterrole",
// 			Delete: &types.Delete{
// 				State: &unstructured.Unstructured{
// 					Object: map[string]interface{}{
// 						"apiVersion": "rbac.authorization.k8s.io/v1",
// 						"kind":       "ClusterRole",
// 						"metadata": map[string]interface{}{
// 							"name": "recipe-integration-cmd-always-dope",
// 						},
// 					},
// 				},
// 			},
// 		},
// 		{
// 			Name: "delete-clusterrolebinding",
// 			Delete: &types.Delete{
// 				State: &unstructured.Unstructured{
// 					Object: map[string]interface{}{
// 						"apiVersion": "rbac.authorization.k8s.io/v1",
// 						"kind":       "ClusterRoleBinding",
// 						"metadata": map[string]interface{}{
// 							"name": "recipe-integration-cmd-always-dope",
// 						},
// 					},
// 				},
// 			},
// 		},
// 		{
// 			Name: "delete-ns",
// 			Delete: &types.Delete{
// 				State: &unstructured.Unstructured{
// 					Object: map[string]interface{}{
// 						"apiVersion": "v1",
// 						"kind":       "Namespace",
// 						"metadata": map[string]interface{}{
// 							"name": "recipe-integration-cmd-always",
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}
// 	recipe = types.Recipe{
// 		Spec: types.RecipeSpec{
// 			Tasks: tasks,
// 		},
// 	}
// 	runner.Recipe = recipe
// 	result, err = runner.RunWithoutLocking()
// 	if err != nil {
// 		t.Fatalf("Error while testing: %v: %s", err, result)
// 	}
// 	if !(result.Phase == types.RecipeStatusCompleted ||
// 		result.Phase == types.RecipeStatusPassed) {
// 		t.Fatalf("Test failed: %s", result)
// 	}
// }
