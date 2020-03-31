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

package run

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ptr "mayadata.io/d-operators/common/pointer"
	types "mayadata.io/d-operators/types/run"
	"openebs.io/metac/apis/metacontroller/v1alpha1"
)

func TestExecuteAssertTask(t *testing.T) {
	var tests = map[string]struct {
		req          TaskRequest
		expectedResp TaskResponse
		isErr        bool
	}{
		"assert all pods are running if pods exist - pass": {
			req: TaskRequest{
				Task: types.Task{
					Key: "assert-all-pods-are-running-if-pod-exist",
					If: &types.If{
						IfConditions: []types.IfCondition{
							types.IfCondition{
								ResourceSelector: v1alpha1.ResourceSelector{
									SelectorTerms: []*v1alpha1.SelectorTerm{
										&v1alpha1.SelectorTerm{
											MatchFields: map[string]string{
												"kind": "Pod",
											},
										},
									},
								},
							},
						},
					},
					Assert: &types.Assert{
						State: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"status": map[string]interface{}{
								"phase": "Running",
							},
						},
					},
				},
				ObservedResources: []*unstructured.Unstructured{
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name": "pod-1",
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{},
							},
							"status": map[string]interface{}{
								"phase": "Running",
							},
						},
					},
				},
			},
			expectedResp: TaskResponse{
				Results: map[string]interface{}{
					"assert": map[string]interface{}{
						"message": "",
						"phase":   types.TaskResultPhaseAssertPassed,
					},
				},
			},
		},
		"assert all pods are running if pods exist - skipped": {
			req: TaskRequest{
				Task: types.Task{
					Key: "assert-all-pods-are-running-if-pod-exist",
					If: &types.If{
						IfConditions: []types.IfCondition{
							types.IfCondition{
								ResourceSelector: v1alpha1.ResourceSelector{
									SelectorTerms: []*v1alpha1.SelectorTerm{
										&v1alpha1.SelectorTerm{
											MatchFields: map[string]string{
												"kind": "Pod",
											},
										},
									},
								},
							},
						},
					},
					Assert: &types.Assert{
						State: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"status": map[string]interface{}{
								"phase": "Running",
							},
						},
					},
				},
				ObservedResources: []*unstructured.Unstructured{
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Service",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name": "svc-1",
							},
							"status": map[string]interface{}{
								"phase": "Online",
							},
						},
					},
				},
			},
			expectedResp: TaskResponse{
				Results: map[string]interface{}{
					"assert": map[string]interface{}{
						"message": "Task didn't run: If cond failed",
						"phase":   types.TaskResultPhaseSkipped,
					},
				},
			},
		},
		"assert all pods are running - pass": {
			req: TaskRequest{
				Task: types.Task{
					Key: "assert-all-pods-are-running",
					Assert: &types.Assert{
						State: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"status": map[string]interface{}{
								"phase": "Running",
							},
						},
					},
				},
				ObservedResources: []*unstructured.Unstructured{
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name": "pod-1",
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{},
							},
							"status": map[string]interface{}{
								"phase": "Running",
							},
						},
					},
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name": "pod-2",
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{},
							},
							"status": map[string]interface{}{
								"phase": "Running",
							},
						},
					},
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Service",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name": "svc-1",
							},
							"status": map[string]interface{}{
								"phase": "Online",
							},
						},
					},
				},
			},
			expectedResp: TaskResponse{
				Results: map[string]interface{}{
					"assert": map[string]interface{}{
						"message": "",
						"phase":   types.TaskResultPhaseAssertPassed,
					},
				},
			},
		},
		"assert pod-1 is running - pass": {
			req: TaskRequest{
				Task: types.Task{
					Key: "assert-all-pods-are-running",
					Assert: &types.Assert{
						State: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name": "pod-1",
							},
							"status": map[string]interface{}{
								"phase": "Running",
							},
						},
					},
				},
				ObservedResources: []*unstructured.Unstructured{
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name": "pod-1",
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{},
							},
							"status": map[string]interface{}{
								"phase": "Running",
							},
						},
					},
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name": "pod-2",
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{},
							},
							"status": map[string]interface{}{
								"phase": "Running",
							},
						},
					},
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Service",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name": "svc-1",
							},
							"status": map[string]interface{}{
								"phase": "Online",
							},
						},
					},
				},
			},
			expectedResp: TaskResponse{
				Results: map[string]interface{}{
					"assert": map[string]interface{}{
						"message": "",
						"phase":   types.TaskResultPhaseAssertPassed,
					},
				},
			},
		},
		"assert pod-1 is running - fail": {
			req: TaskRequest{
				Task: types.Task{
					Key: "assert-all-pods-are-running",
					Assert: &types.Assert{
						State: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name": "pod-1",
							},
							"status": map[string]interface{}{
								"phase": "Running",
							},
						},
					},
				},
				ObservedResources: []*unstructured.Unstructured{
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name": "pod-1",
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{},
							},
							"status": map[string]interface{}{
								"phase": "Error",
							},
						},
					},
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name": "pod-2",
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{},
							},
							"status": map[string]interface{}{
								"phase": "Running",
							},
						},
					},
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Service",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name": "svc-1",
							},
							"status": map[string]interface{}{
								"phase": "Online",
							},
						},
					},
				},
			},
			expectedResp: TaskResponse{
				Results: map[string]interface{}{
					"assert": map[string]interface{}{
						"message": "",
						"phase":   types.TaskResultPhaseAssertFailed,
					},
				},
			},
		},
		"assert pod-1 is error - pass": {
			req: TaskRequest{
				Task: types.Task{
					Key: "assert-all-pods-are-running",
					Assert: &types.Assert{
						State: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name": "pod-1",
							},
							"status": map[string]interface{}{
								"phase": "Error",
							},
						},
					},
				},
				ObservedResources: []*unstructured.Unstructured{
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name": "pod-1",
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{},
							},
							"status": map[string]interface{}{
								"phase": "Error",
							},
						},
					},
				},
			},
			expectedResp: TaskResponse{
				Results: map[string]interface{}{
					"assert": map[string]interface{}{
						"message": "",
						"phase":   types.TaskResultPhaseAssertPassed,
					},
				},
			},
		},
		"assert pod-1 is error - fail": {
			req: TaskRequest{
				Task: types.Task{
					Key: "assert-all-pods-are-running",
					Assert: &types.Assert{
						State: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name": "pod-1",
							},
							"status": map[string]interface{}{
								"phase": "Error",
							},
						},
					},
				},
				ObservedResources: []*unstructured.Unstructured{
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name": "pod-1",
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{},
							},
							"status": map[string]interface{}{
								"phase": "Running",
							},
						},
					},
				},
			},
			expectedResp: TaskResponse{
				Results: map[string]interface{}{
					"assert": map[string]interface{}{
						"message": "",
						"phase":   types.TaskResultPhaseAssertFailed,
					},
				},
			},
		},
		"assert all pods are running - fail": {
			req: TaskRequest{
				Task: types.Task{
					Key: "assert-all-pods-are-running",
					Assert: &types.Assert{
						State: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"status": map[string]interface{}{
								"phase": "Running",
							},
						},
					},
				},
				ObservedResources: []*unstructured.Unstructured{
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Service",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name": "svc-1",
							},
							"status": map[string]interface{}{
								"phase": "Online",
							},
						},
					},
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Service",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name": "svc-2",
							},
							"status": map[string]interface{}{
								"phase": "Online",
							},
						},
					},
				},
			},
			expectedResp: TaskResponse{
				Results: map[string]interface{}{
					"assert": map[string]interface{}{
						"message": "",
						"phase":   types.TaskResultPhaseAssertFailed,
					},
				},
			},
		},
		"assert all pods are Error - fail": {
			req: TaskRequest{
				Task: types.Task{
					Key: "assert-all-pods-are-error",
					Assert: &types.Assert{
						State: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"status": map[string]interface{}{
								"phase": "Error",
							},
						},
					},
				},
				ObservedResources: []*unstructured.Unstructured{
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name": "pod-1",
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{},
							},
							"status": map[string]interface{}{
								"phase": "Running",
							},
						},
					},
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name": "pod-2",
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{},
							},
							"status": map[string]interface{}{
								"phase": "Running",
							},
						},
					},
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Service",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name": "svc-1",
							},
							"status": map[string]interface{}{
								"phase": "Online",
							},
						},
					},
				},
			},
			expectedResp: TaskResponse{
				Results: map[string]interface{}{
					"assert": map[string]interface{}{
						"message": "",
						"phase":   types.TaskResultPhaseAssertFailed,
					},
				},
			},
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			newreq := TaskRequest{
				Run: &unstructured.Unstructured{
					Object: map[string]interface{}{},
				},
				Watch: &unstructured.Unstructured{
					Object: map[string]interface{}{},
				},
				Task:              mock.req.Task,
				ObservedResources: mock.req.ObservedResources,
			}
			got, err := ExecTask(newreq)
			if mock.isErr && err == nil {
				t.Fatalf("Expected error got none")
			}
			if !mock.isErr && err != nil {
				t.Fatalf("Expected no error got [%+v]", err)
			}
			if mock.isErr {
				return
			}
			expected := mock.expectedResp.Results["assert"].(map[string]interface{})
			actual := got.Results["assert"].(map[string]interface{})
			// extract phase
			expectedphase := expected["phase"].(types.TaskResultPhase)
			actualphase := actual["phase"].(types.TaskResultPhase)
			if actualphase != expectedphase {
				t.Fatalf(
					"Expected phase %q got %q",
					expectedphase,
					actualphase,
				)
			}
			// extract message
			expectedmsg := expected["message"].(string)
			actualmsg := actual["message"].(string)
			if actualmsg != expectedmsg {
				t.Fatalf(
					"Expected message %q got %q",
					expectedmsg,
					actualmsg,
				)
			}
		})
	}
}

func TestExecuteCreateTask(t *testing.T) {
	var tests = map[string]struct {
		req          TaskRequest
		expectedResp TaskResponse
		isErr        bool
	}{
		"create 5 pods if service exist - pass": {
			req: TaskRequest{
				Task: types.Task{
					Key: "create-5-pods-if-service-exist",
					If: &types.If{
						IfConditions: []types.IfCondition{
							types.IfCondition{
								ResourceSelector: v1alpha1.ResourceSelector{
									SelectorTerms: []*v1alpha1.SelectorTerm{
										&v1alpha1.SelectorTerm{
											MatchFields: map[string]string{
												"kind": "Service",
											},
										},
									},
								},
							},
						},
					},
					Apply: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "my-pod",
							"namespace": "dope",
						},
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "nginx",
									"image": "nginx",
								},
							},
						},
					},
					Replicas: ptr.Int(5),
				},
				ObservedResources: []*unstructured.Unstructured{
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Service",
							"apiVersion": "v1",
						},
					},
				},
			},
			expectedResp: TaskResponse{
				Results: map[string]interface{}{
					"createOrDelete": map[string]interface{}{
						"message": "Create/Delete was successful: Desired resources 5: Explicit deletes 0",
						"phase":   types.TaskResultPhaseOnline,
					},
				},
			},
		},
		"create 5 pods - pass": {
			req: TaskRequest{
				Task: types.Task{
					Key: "create-5-pods",
					Apply: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "my-pod",
							"namespace": "dope",
						},
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "nginx",
									"image": "nginx",
								},
							},
						},
					},
					Replicas: ptr.Int(5),
				},
			},
			expectedResp: TaskResponse{
				Results: map[string]interface{}{
					"createOrDelete": map[string]interface{}{
						"message": "Create/Delete was successful: Desired resources 5: Explicit deletes 0",
						"phase":   types.TaskResultPhaseOnline,
					},
				},
			},
		},
		"delete all pods by setting spec to nil - pass": {
			req: TaskRequest{
				Task: types.Task{
					Key: "delete-all-pods",
					Apply: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"spec":       nil, // this implies delete
					},
				},
				ObservedResources: []*unstructured.Unstructured{
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name": "my-pod-1",
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "nginx",
										"image": "nginx",
									},
								},
							},
						},
					},
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name": "my-pod-2",
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "nginx-2",
										"image": "nginx:latest",
									},
								},
							},
						},
					},
				},
			},
			expectedResp: TaskResponse{
				Results: map[string]interface{}{
					"createOrDelete": map[string]interface{}{
						"message": "Create/Delete was successful: Desired resources 0: Explicit deletes 2",
						"phase":   types.TaskResultPhaseOnline,
					},
				},
			},
		},
		"delete all pods by setting replicas to 0 - pass": {
			req: TaskRequest{
				Task: types.Task{
					Key: "delete-all-pods",
					Apply: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
					},
					Replicas: ptr.Int(0), // 0 implies delete
				},
				ObservedResources: []*unstructured.Unstructured{
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name": "my-pod-1",
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "nginx",
										"image": "nginx",
									},
								},
							},
						},
					},
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name": "my-pod-2",
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "nginx-2",
										"image": "nginx:latest",
									},
								},
							},
						},
					},
				},
			},
			expectedResp: TaskResponse{
				Results: map[string]interface{}{
					"createOrDelete": map[string]interface{}{
						"message": "Create/Delete was successful: Desired resources 0: Explicit deletes 2",
						"phase":   types.TaskResultPhaseOnline,
					},
				},
			},
		},
		"delete all owned pods - pass": {
			req: TaskRequest{
				Task: types.Task{
					Key: "delete-all-pods",
					Apply: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"spec":       nil,
					},
				},
				ObservedResources: []*unstructured.Unstructured{
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name": "my-pod-1",
								"annotations": map[string]interface{}{
									types.AnnotationKeyMetacCreatedDueToWatch: "watch-101",
								},
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "nginx",
										"image": "nginx",
									},
								},
							},
						},
					},
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name": "my-pod-2",
								"annotations": map[string]interface{}{
									types.AnnotationKeyMetacCreatedDueToWatch: "watch-101",
								},
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "nginx-2",
										"image": "nginx:latest",
									},
								},
							},
						},
					},
				},
			},
			expectedResp: TaskResponse{
				Results: map[string]interface{}{
					"createOrDelete": map[string]interface{}{
						"message": "Create/Delete was successful: Desired resources 0: Explicit deletes 0",
						"phase":   types.TaskResultPhaseOnline,
					},
				},
			},
		},
		"delete 1 owned & 1 not owned pod - pass": {
			req: TaskRequest{
				Task: types.Task{
					Key: "delete-all-pods",
					Apply: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"spec":       nil,
					},
				},
				ObservedResources: []*unstructured.Unstructured{
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name": "my-pod-1",
								"annotations": map[string]interface{}{
									types.AnnotationKeyMetacCreatedDueToWatch: "watch-101",
								},
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "nginx",
										"image": "nginx",
									},
								},
							},
						},
					},
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name": "my-pod-2",
								"annotations": map[string]interface{}{
									types.AnnotationKeyMetacCreatedDueToWatch: "watch-102",
								},
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "nginx-2",
										"image": "nginx:latest",
									},
								},
							},
						},
					},
				},
			},
			expectedResp: TaskResponse{
				Results: map[string]interface{}{
					"createOrDelete": map[string]interface{}{
						"message": "Create/Delete was successful: Desired resources 0: Explicit deletes 1",
						"phase":   types.TaskResultPhaseOnline,
					},
				},
			},
		},
		"delete all pods from none by setting replicas to 0 - pass": {
			req: TaskRequest{
				Task: types.Task{
					Key: "delete-all-pods",
					Apply: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
					},
					Replicas: ptr.Int(0), // 0 implies delete
				},
				ObservedResources: []*unstructured.Unstructured{},
			},
			expectedResp: TaskResponse{
				Results: map[string]interface{}{
					"createOrDelete": map[string]interface{}{
						"message": "Create/Delete was successful: Desired resources 0: Explicit deletes 0",
						"phase":   types.TaskResultPhaseOnline,
					},
				},
			},
		},
		"delete no pods due to mismatch - pass": {
			req: TaskRequest{
				Task: types.Task{
					Key: "delete-all-pods",
					Apply: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v2", // mismatch
						"spec":       nil,
					},
				},
				ObservedResources: []*unstructured.Unstructured{
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name": "my-pod-1",
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "nginx",
										"image": "nginx",
									},
								},
							},
						},
					},
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name": "my-pod-2",
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "nginx-2",
										"image": "nginx:latest",
									},
								},
							},
						},
					},
				},
			},
			expectedResp: TaskResponse{
				Results: map[string]interface{}{
					"createOrDelete": map[string]interface{}{
						"message": "Create/Delete was successful: Desired resources 0: Explicit deletes 0",
						"phase":   types.TaskResultPhaseOnline,
					},
				},
			},
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			newreq := TaskRequest{
				Run: &unstructured.Unstructured{
					Object: map[string]interface{}{},
				},
				Watch: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"metadata": map[string]interface{}{
							"uid": "watch-101",
						},
					},
				},
				Task:              mock.req.Task,
				ObservedResources: mock.req.ObservedResources,
			}
			got, err := ExecTask(newreq)
			if mock.isErr && err == nil {
				t.Fatalf("Expected error got none")
			}
			if !mock.isErr && err != nil {
				t.Fatalf("Expected no error got [%+v]", err)
			}
			if mock.isErr {
				return
			}
			expected := mock.expectedResp.Results["createOrDelete"].(map[string]interface{})
			actual := got.Results["createOrDelete"].(map[string]interface{})
			// extract phase
			expectedphase := expected["phase"].(types.TaskResultPhase)
			actualphase := actual["phase"].(types.TaskResultPhase)
			if actualphase != expectedphase {
				t.Fatalf(
					"Expected phase %q got %q",
					expectedphase,
					actualphase,
				)
			}
			// extract message
			expectedmsg := expected["message"].(string)
			actualmsg := actual["message"].(string)
			if actualmsg != expectedmsg {
				t.Fatalf(
					"Expected message %q got %q",
					expectedmsg,
					actualmsg,
				)
			}
		})
	}
}
