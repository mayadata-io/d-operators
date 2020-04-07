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

func TestExecuteAssertByExecTask(t *testing.T) {
	var tests = map[string]struct {
		req           TaskRequest
		expectedPhase types.ResultPhase
		isSkip        bool
		isErr         bool
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
					Assert: &types.Assert{ // Assert State Based Task
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
			expectedPhase: types.ResultPhaseAssertPassed,
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
					Assert: &types.Assert{ // Assert State Based Task
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
			isSkip:        true,
			expectedPhase: types.ResultPhaseSkipped,
		},
		"assert all pods are running - pass": {
			req: TaskRequest{
				Task: types.Task{
					Key: "assert-all-pods-are-running",
					Assert: &types.Assert{ // Assert State Based Task
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
			expectedPhase: types.ResultPhaseAssertPassed,
		},
		"assert pod-1 is running - pass": {
			req: TaskRequest{
				Task: types.Task{
					Key: "assert-all-pods-are-running",
					Assert: &types.Assert{ // Assert State Based Task
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
			expectedPhase: types.ResultPhaseAssertPassed,
		},
		"assert pod-1 is running - fail": {
			req: TaskRequest{
				Task: types.Task{
					Key: "assert-all-pods-are-running",
					Assert: &types.Assert{ // Assert State Based Task
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
			expectedPhase: types.ResultPhaseAssertFailed,
		},
		"assert pod-1 is error - pass": {
			req: TaskRequest{
				Task: types.Task{
					Key: "assert-all-pods-are-running",
					Assert: &types.Assert{ // Assert State Based Task
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
			expectedPhase: types.ResultPhaseAssertPassed,
		},
		"assert pod-1 is error - fail": {
			req: TaskRequest{
				Task: types.Task{
					Key: "assert-all-pods-are-running",
					Assert: &types.Assert{ // Assert State Based Task
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
			expectedPhase: types.ResultPhaseAssertFailed,
		},
		"assert all pods are running - fail": {
			req: TaskRequest{
				Task: types.Task{
					Key: "assert-all-pods-are-running",
					Assert: &types.Assert{ // Assert State Based Task
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
			expectedPhase: types.ResultPhaseAssertFailed,
		},
		"assert all pods are Error - fail": {
			req: TaskRequest{
				Task: types.Task{
					Key: "assert-all-pods-are-error",
					Assert: &types.Assert{ // Assert State Based Task
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
			expectedPhase: types.ResultPhaseAssertFailed,
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
			if mock.isSkip {
				if got.Result.SkipResult.Phase != mock.expectedPhase {
					t.Fatalf(
						"Expected phase %q got %q",
						mock.expectedPhase,
						got.Result.SkipResult.Phase,
					)
				}
				return
			}
			if got.Result.AssertResult.Phase !=
				mock.expectedPhase {
				t.Fatalf(
					"Expected phase %q got %q",
					mock.expectedPhase,
					got.Result.AssertResult.Phase,
				)
			}
		})
	}
}

func TestExecuteCreateOrDeleteTask(t *testing.T) {
	var tests = map[string]struct {
		req             TaskRequest
		expectedPhase   types.ResultPhase
		expectedMessage string
		isErr           bool
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
					Replicas: ptr.Int(5), // Create Task
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
			expectedPhase:   types.ResultPhaseOnline,
			expectedMessage: "Create action was successful for 5 resource(s)",
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
					Replicas: ptr.Int(5), // Create Task
				},
			},
			expectedPhase:   types.ResultPhaseOnline,
			expectedMessage: "Create action was successful for 5 resource(s)",
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
			expectedPhase:   types.ResultPhaseOnline,
			expectedMessage: "Delete action was successful: Desired deletes 0: Explicit deletes 2: Total 2",
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
			expectedPhase:   types.ResultPhaseOnline,
			expectedMessage: "Delete action was successful: Desired deletes 0: Explicit deletes 2: Total 2",
		},
		"delete all owned pods - pass": {
			req: TaskRequest{
				Task: types.Task{
					Key: "delete-all-pods",
					Apply: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"spec":       nil, // Delete Task
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
			expectedPhase:   types.ResultPhaseOnline,
			expectedMessage: "Delete action was successful: Desired deletes 2: Explicit deletes 0: Total 2",
		},
		"delete 1 owned & 1 not owned pod - pass": {
			req: TaskRequest{
				Task: types.Task{
					Key: "delete-all-pods",
					Apply: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"spec":       nil, // Delete Task
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
			expectedPhase:   types.ResultPhaseOnline,
			expectedMessage: "Delete action was successful: Desired deletes 1: Explicit deletes 1: Total 2",
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
			expectedPhase:   types.ResultPhaseOnline,
			expectedMessage: "Delete action was successful: Desired deletes 0: Explicit deletes 0: Total 0",
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
			expectedPhase:   types.ResultPhaseOnline,
			expectedMessage: "Delete action was successful: Desired deletes 0: Explicit deletes 0: Total 2",
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
			r := &RunnableTask{
				Request: newreq,
				Response: &TaskResponse{
					Result: &types.TaskResult{},
				},
				isApplyAction:  true,  // create or delete needs Apply
				isUpdateAction: false, // this is not update testing
			}
			r.runCreateOrDelete()
			if mock.isErr && r.err == nil {
				t.Fatalf("Expected error got none")
			}
			if !mock.isErr && r.err != nil {
				t.Fatalf(
					"Expected no error got %+v: \n%s",
					r.err,
					r.Response.Result,
				)
			}
			if mock.isErr {
				return
			}
			if r.Response.Result.DeleteResult == nil &&
				r.Response.Result.CreateResult == nil {
				t.Fatalf(
					"Expected either delete result or create result got none",
				)
			}
			// delete checks
			if r.Response.Result.DeleteResult != nil {
				if r.Response.Result.DeleteResult.Phase !=
					mock.expectedPhase {
					t.Fatalf(
						"Expected phase %q got %q",
						mock.expectedPhase,
						r.Response.Result.DeleteResult.Phase,
					)
				}
				if r.Response.Result.DeleteResult.Message !=
					mock.expectedMessage {
					t.Fatalf(
						"Expected message %q got %q",
						mock.expectedMessage,
						r.Response.Result.DeleteResult.Message,
					)
				}
			}
			// create checks
			if r.Response.Result.CreateResult != nil {
				if r.Response.Result.CreateResult.Phase !=
					mock.expectedPhase {
					t.Fatalf(
						"Expected phase %q got %q",
						mock.expectedPhase,
						r.Response.Result.CreateResult.Phase,
					)
				}
				if r.Response.Result.CreateResult.Message !=
					mock.expectedMessage {
					t.Fatalf(
						"Expected message %q got %q",
						mock.expectedMessage,
						r.Response.Result.CreateResult.Message,
					)
				}
			}
		})
	}
}

func TestCreateOrDeleteByExecTask(t *testing.T) {
	var tests = map[string]struct {
		req             TaskRequest
		expectedPhase   types.ResultPhase
		expectedMessage string
		isErr           bool
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
					Replicas: ptr.Int(5), // Create Task
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
			expectedPhase:   types.ResultPhaseOnline,
			expectedMessage: "Create action was successful for 5 resource(s)",
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
					Replicas: ptr.Int(5), // Create Task
				},
			},
			expectedPhase:   types.ResultPhaseOnline,
			expectedMessage: "Create action was successful for 5 resource(s)",
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
			expectedPhase:   types.ResultPhaseOnline,
			expectedMessage: "Delete action was successful: Desired deletes 0: Explicit deletes 2: Total 2",
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
			expectedPhase:   types.ResultPhaseOnline,
			expectedMessage: "Delete action was successful: Desired deletes 0: Explicit deletes 2: Total 2",
		},
		"delete all owned pods - pass": {
			req: TaskRequest{
				Task: types.Task{
					Key: "delete-all-pods",
					Apply: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"spec":       nil, // Delete Task
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
			expectedPhase:   types.ResultPhaseOnline,
			expectedMessage: "Delete action was successful: Desired deletes 2: Explicit deletes 0: Total 2",
		},
		"delete 1 owned & 1 not owned pod - pass": {
			req: TaskRequest{
				Task: types.Task{
					Key: "delete-all-pods",
					Apply: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"spec":       nil, // Delete Task
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
			expectedPhase:   types.ResultPhaseOnline,
			expectedMessage: "Delete action was successful: Desired deletes 1: Explicit deletes 1: Total 2",
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
			expectedPhase:   types.ResultPhaseOnline,
			expectedMessage: "Delete action was successful: Desired deletes 0: Explicit deletes 0: Total 0",
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
			expectedPhase:   types.ResultPhaseOnline,
			expectedMessage: "Delete action was successful: Desired deletes 0: Explicit deletes 0: Total 2",
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
			if got.Result.DeleteResult == nil &&
				got.Result.CreateResult == nil {
				t.Fatalf("Expected either delete result or create result got none")
			}
			// delete checks
			if got.Result.DeleteResult != nil {
				if got.Result.DeleteResult.Phase !=
					mock.expectedPhase {
					t.Fatalf(
						"Expected phase %q got %q",
						mock.expectedPhase,
						got.Result.DeleteResult.Phase,
					)
				}
				if got.Result.DeleteResult.Message !=
					mock.expectedMessage {
					t.Fatalf(
						"Expected message %q got %q",
						mock.expectedMessage,
						got.Result.DeleteResult.Message,
					)
				}
			}
			// create checks
			if got.Result.CreateResult != nil {
				if got.Result.CreateResult.Phase !=
					mock.expectedPhase {
					t.Fatalf(
						"Expected phase %q got %q",
						mock.expectedPhase,
						got.Result.CreateResult.Phase,
					)
				}
				if got.Result.CreateResult.Message !=
					mock.expectedMessage {
					t.Fatalf(
						"Expected message %q got %q",
						mock.expectedMessage,
						got.Result.CreateResult.Message,
					)
				}
			}
		})
	}
}
