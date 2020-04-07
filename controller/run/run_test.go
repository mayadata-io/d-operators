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
	"mayadata.io/d-operators/common/pointer"
	types "mayadata.io/d-operators/types/run"
	"openebs.io/metac/apis/metacontroller/v1alpha1"
)

func TestRunnableValidateArgs(t *testing.T) {
	var tests = map[string]struct {
		Run      *unstructured.Unstructured
		Watch    *unstructured.Unstructured
		Tasks    []types.Task
		Response *Response
		isErr    bool
	}{
		"nil run": {
			isErr: true,
		},
		"nil watch": {
			Run: &unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
			isErr: true,
		},
		"nil response": {
			Run: &unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
			Watch: &unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
			isErr: true,
		},
		"nil tasks": {
			Run: &unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
			Watch: &unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
			Response: &Response{},
			isErr:    true,
		},
		"all ok": {
			Run: &unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
			Watch: &unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
			Response: &Response{
				RunStatus: &types.RunStatus{},
			},
			Tasks: []types.Task{
				types.Task{},
			},
			isErr: false,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			r := &Runnable{
				Request: Request{
					Run:   mock.Run,
					Watch: mock.Watch,
					Tasks: mock.Tasks,
				},
				Response: mock.Response,
			}
			err := r.validateArgs()
			if mock.isErr && err == nil {
				t.Fatalf("Expected error got none")
			}
			if !mock.isErr && err != nil {
				t.Fatalf("Expected no error got [%+v]", err)
			}
		})
	}
}

func TestRunnableRunIfCond(t *testing.T) {
	var tests = map[string]struct {
		RunCond       *types.If
		Resources     []*unstructured.Unstructured
		isRunCondPass bool
		isErr         bool
	}{
		"nil run cond - assert pass": {
			RunCond:       nil,
			isRunCondPass: true,
		},
		"run cond - assert pass": {
			RunCond: &types.If{
				IfConditions: []types.IfCondition{
					types.IfCondition{
						ResourceSelector: v1alpha1.ResourceSelector{
							SelectorTerms: []*v1alpha1.SelectorTerm{
								&v1alpha1.SelectorTerm{
									MatchAnnotations: map[string]string{
										"app": "test",
									},
								},
							},
						},
					},
				},
			},
			Resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"annotations": map[string]interface{}{
								"app": "test",
							},
						},
					},
				},
			},
			isRunCondPass: true,
		},
		"run cond + nil resources - error": {
			RunCond: &types.If{
				IfConditions: []types.IfCondition{
					types.IfCondition{
						ResourceSelector: v1alpha1.ResourceSelector{
							SelectorTerms: []*v1alpha1.SelectorTerm{
								&v1alpha1.SelectorTerm{
									MatchAnnotations: map[string]string{
										"app": "test",
									},
								},
							},
						},
					},
				},
			},
			isErr: true,
		},
		"run cond - assert fail": {
			RunCond: &types.If{
				IfConditions: []types.IfCondition{
					types.IfCondition{
						ResourceSelector: v1alpha1.ResourceSelector{
							SelectorTerms: []*v1alpha1.SelectorTerm{
								&v1alpha1.SelectorTerm{
									MatchAnnotations: map[string]string{
										"app": "test",
									},
								},
							},
						},
					},
				},
			},
			Resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"annotations": map[string]interface{}{
								"app": "prod",
							},
						},
					},
				},
			},
			isRunCondPass: false,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			r := &Runnable{
				Request: Request{
					RunCond:           mock.RunCond,
					ObservedResources: mock.Resources,
				},
				Response: &Response{
					RunStatus: &types.RunStatus{},
				},
			}
			r.runIfCondition()
			if mock.isErr && r.err == nil {
				t.Fatalf("Expected error got none")
			}
			if !mock.isErr && r.err != nil {
				t.Fatalf("Expected no error got [%+v]", r.err)
			}
			if mock.isErr {
				return
			}
			if mock.isRunCondPass != r.isRunCondSuccess {
				t.Fatalf(
					"Expected if cond %t got %t",
					mock.isRunCondPass,
					r.isRunCondSuccess,
				)
			}
		})
	}
}

func TestRunnableRunAllTasks(t *testing.T) {
	var tests = map[string]struct {
		Tasks                       []types.Task
		Resources                   []*unstructured.Unstructured
		expectedErrCount            int
		expectedDesiredCount        int
		expectedExplicitUpdateCount int
		expectedExplicitDeleteCount int
		expectedAssertTaskCount     int
		expectedUpdateTaskCount     int
		expectedCreateTaskCount     int
		expectedDeleteTaskCount     int
		expectedIfCondTaskCount     int
		expectedSkippedTaskCount    int
		expectedTaskResultCount     int
		isErr                       bool
	}{
		"create pod + 1 task": {
			Tasks: []types.Task{
				types.Task{
					Key: "create-a-pod",
					Apply: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
						},
					},
				},
			},
			expectedDesiredCount:    1,
			expectedCreateTaskCount: 1,
			expectedTaskResultCount: 1,
		},
		"create 2 pods + 1 task": {
			Tasks: []types.Task{
				types.Task{
					Key: "create-a-pod",
					Apply: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
						},
					},
					Replicas: pointer.Int(2),
				},
			},
			expectedDesiredCount:    2,
			expectedCreateTaskCount: 1,
			expectedTaskResultCount: 1,
		},
		"create 2 pods + 1 pod per task": {
			Tasks: []types.Task{
				types.Task{
					Key: "create-a-pod",
					Apply: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
						},
					},
				},
				types.Task{
					Key: "create-a-pod-2",
					Apply: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod-2",
						},
					},
				},
			},
			expectedDesiredCount:    2,
			expectedCreateTaskCount: 2,
			expectedTaskResultCount: 2,
		},
		"delete all pods": {
			Tasks: []types.Task{
				types.Task{
					Key: "delete-a-pod",
					Apply: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
						},
						"spec": nil, // implies delete
					},
				},
			},
			Resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod-2",
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
						},
					},
				},
			},
			expectedExplicitDeleteCount: 2,
			expectedDeleteTaskCount:     1,
			expectedTaskResultCount:     1,
		},
		"delete specific pod": {
			Tasks: []types.Task{
				types.Task{
					Key: "delete-a-pod",
					Apply: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod-2",
						},
						"spec": nil, // implies delete
					},
				},
			},
			Resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod-2",
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
						},
					},
				},
			},
			expectedExplicitDeleteCount: 1,
			expectedDeleteTaskCount:     1,
			expectedTaskResultCount:     1,
		},
		"delete no pods due to mismatch": {
			Tasks: []types.Task{
				types.Task{
					Key: "delete-a-pod",
					Apply: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod-10",
						},
						"spec": nil, // implies delete
					},
				},
			},
			Resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod-2",
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod-1",
						},
					},
				},
			},
			expectedDeleteTaskCount: 1,
			expectedTaskResultCount: 1,
		},
		"delete no pods due to failing if cond": {
			Tasks: []types.Task{
				types.Task{
					Key: "delete-a-pod",
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
							"name": "my-pod",
						},
						"spec": nil, // implies delete
					},
				},
			},
			Resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod-2",
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod-1",
						},
					},
				},
			},
			expectedExplicitDeleteCount: 0,
			expectedSkippedTaskCount:    1,
			expectedIfCondTaskCount:     1,
			expectedTaskResultCount:     1,
		},
		"delete all pods with passing if cond": {
			Tasks: []types.Task{
				types.Task{
					Key: "delete-all-pods",
					If: &types.If{
						IfConditions: []types.IfCondition{
							types.IfCondition{
								ResourceSelector: v1alpha1.ResourceSelector{
									SelectorTerms: []*v1alpha1.SelectorTerm{
										&v1alpha1.SelectorTerm{
											MatchFields: map[string]string{
												"kind":       "Pod",
												"apiVersion": "v1",
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
							"name": "my-pod",
						},
						"spec": nil, // implies delete
					},
				},
			},
			Resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod-2",
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod-1",
						},
					},
				},
			},
			expectedExplicitDeleteCount: 2,
			expectedDeleteTaskCount:     1,
			expectedIfCondTaskCount:     1,
			expectedTaskResultCount:     1,
		},
		"create 5 pods with passing if cond": {
			Tasks: []types.Task{
				types.Task{
					Key: "create-5-pods",
					If: &types.If{
						IfConditions: []types.IfCondition{
							types.IfCondition{
								ResourceSelector: v1alpha1.ResourceSelector{
									SelectorTerms: []*v1alpha1.SelectorTerm{
										&v1alpha1.SelectorTerm{
											MatchFields: map[string]string{
												"kind":       "Service",
												"apiVersion": "v1",
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
							"name": "my-pod",
						},
					},
					Replicas: pointer.Int(5),
				},
			},
			Resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Service",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-presence-is-imp",
						},
					},
				},
			},
			expectedDesiredCount:    5,
			expectedCreateTaskCount: 1,
			expectedIfCondTaskCount: 1,
			expectedTaskResultCount: 1,
		},
		"create no pods due to failing if cond": {
			Tasks: []types.Task{
				types.Task{
					Key: "can-not-create-pods",
					If: &types.If{
						IfConditions: []types.IfCondition{
							types.IfCondition{
								ResourceSelector: v1alpha1.ResourceSelector{
									SelectorTerms: []*v1alpha1.SelectorTerm{
										&v1alpha1.SelectorTerm{
											MatchFields: map[string]string{
												"kind":       "Deployment",
												"apiVersion": "apps/v1",
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
							"name": "my-pod",
						},
					},
					Replicas: pointer.Int(5),
				},
			},
			Resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Service",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-presence-is-imp",
						},
					},
				},
			},
			expectedTaskResultCount:  1,
			expectedIfCondTaskCount:  1,
			expectedSkippedTaskCount: 1,
		},
		"update anns of all pods with specific labels": {
			Tasks: []types.Task{
				types.Task{
					Key: "update-anns-of-all-pods-with-specific-labels",
					Apply: map[string]interface{}{
						"metadata": map[string]interface{}{
							"annotations": map[string]interface{}{
								"app": "test",
							},
						},
					},
					TargetSelector: types.TargetSelector{
						ResourceSelector: v1alpha1.ResourceSelector{
							SelectorTerms: []*v1alpha1.SelectorTerm{
								&v1alpha1.SelectorTerm{
									MatchLabels: map[string]string{
										"app": "test",
									},
									MatchFields: map[string]string{
										"kind":       "Pod",
										"apiVersion": "v1",
									},
								},
							},
						},
					},
				},
			},
			Resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod-1",
							"labels": map[string]interface{}{
								"app": "test",
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
							"labels": map[string]interface{}{
								"app": "test",
							},
						},
					},
				},
			},
			expectedExplicitUpdateCount: 2,
			expectedUpdateTaskCount:     1,
			expectedTaskResultCount:     1,
		},
		"cannot update anns of pods due to failing target selector": {
			Tasks: []types.Task{
				types.Task{
					Key: "update-anns-of-all-pods-with-specific-labels",
					Apply: map[string]interface{}{
						"metadata": map[string]interface{}{
							"annotations": map[string]interface{}{
								"app": "test",
							},
						},
					},
					TargetSelector: types.TargetSelector{
						ResourceSelector: v1alpha1.ResourceSelector{
							SelectorTerms: []*v1alpha1.SelectorTerm{
								&v1alpha1.SelectorTerm{
									MatchLabels: map[string]string{
										"app": "no-test",
									},
									MatchFields: map[string]string{
										"kind":       "Pod",
										"apiVersion": "v1",
									},
								},
							},
						},
					},
				},
			},
			Resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod-1",
							"labels": map[string]interface{}{
								"app": "test",
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
							"labels": map[string]interface{}{
								"app": "test",
							},
						},
					},
				},
			},
			expectedUpdateTaskCount: 1,
			expectedTaskResultCount: 1,
		},
		"cannot update anns of pods due to failing if cond": {
			Tasks: []types.Task{
				types.Task{
					Key: "update-anns-of-all-pods-with-specific-labels",
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
						"metadata": map[string]interface{}{
							"annotations": map[string]interface{}{
								"app": "test",
							},
						},
					},
					TargetSelector: types.TargetSelector{
						ResourceSelector: v1alpha1.ResourceSelector{
							SelectorTerms: []*v1alpha1.SelectorTerm{
								&v1alpha1.SelectorTerm{
									MatchFields: map[string]string{
										"kind":       "Pod",
										"apiVersion": "v1",
									},
								},
							},
						},
					},
				},
			},
			Resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod-1",
							"labels": map[string]interface{}{
								"app": "test",
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
							"labels": map[string]interface{}{
								"app": "test",
							},
						},
					},
				},
			},
			expectedIfCondTaskCount:  1,
			expectedSkippedTaskCount: 1,
			expectedTaskResultCount:  1,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			r := &Runnable{
				Request: Request{
					IncludeInfo: map[types.IncludeInfoKey]bool{
						types.IncludeAllInfo: true,
					},
					ObservedResources: mock.Resources,
					Tasks:             mock.Tasks,
					Run: &unstructured.Unstructured{
						Object: map[string]interface{}{},
					},
					Watch: &unstructured.Unstructured{
						Object: map[string]interface{}{},
					},
				},
				Response: &Response{
					RunStatus: &types.RunStatus{
						TaskResultList: map[string]types.TaskResult{},
					},
				},
			}
			r.runAllTasks()
			if mock.isErr && r.err == nil {
				t.Fatalf("Expected error got none")
			}
			if !mock.isErr && r.err != nil {
				t.Fatalf(
					"Expected no error got [%+v]: \n%s",
					r.err,
					r.Response.RunStatus,
				)
			}
			if mock.isErr && mock.expectedErrCount != len(r.Response.RunStatus.Errors) {
				t.Fatalf(
					"Expected error count %d got %d: %s",
					mock.expectedErrCount,
					len(r.Response.RunStatus.Errors),
					r.Response.RunStatus,
				)
			}
			if mock.expectedDesiredCount != len(r.Response.DesiredResources) {
				t.Fatalf(
					"Expected desired resource count %d got %d: %s",
					mock.expectedDesiredCount,
					len(r.Response.DesiredResources),
					r.Response.RunStatus,
				)
			}
			if mock.expectedExplicitDeleteCount != len(r.Response.ExplicitDeletes) {
				t.Fatalf(
					"Expected explicit delete resource count %d got %d: %s",
					mock.expectedExplicitDeleteCount,
					len(r.Response.ExplicitDeletes),
					r.Response.RunStatus,
				)
			}
			if mock.expectedExplicitUpdateCount != len(r.Response.ExplicitUpdates) {
				t.Fatalf(
					"Expected explicit update resource count %d got %d: %s",
					mock.expectedExplicitUpdateCount,
					len(r.Response.ExplicitUpdates),
					r.Response.RunStatus,
				)
			}
			if mock.expectedTaskResultCount != len(r.Response.RunStatus.TaskResultList) {
				t.Fatalf(
					"Expected task result count %d got %d: %s",
					mock.expectedTaskResultCount,
					len(r.Response.RunStatus.TaskResultList),
					r.Response.RunStatus,
				)
			}
			if mock.expectedSkippedTaskCount !=
				r.Response.RunStatus.TaskResultList.SkipTaskCount() {
				t.Fatalf(
					"Expected skipped task count %d got %d: %s",
					mock.expectedSkippedTaskCount,
					r.Response.RunStatus.TaskResultList.SkipTaskCount(),
					r.Response.RunStatus,
				)
			}
			if mock.expectedAssertTaskCount !=
				r.Response.RunStatus.TaskResultList.AssertTaskCount() {
				t.Fatalf(
					"Expected assert task count %d got %d: %s",
					mock.expectedAssertTaskCount,
					r.Response.RunStatus.TaskResultList.AssertTaskCount(),
					r.Response.RunStatus,
				)
			}
			if mock.expectedCreateTaskCount !=
				r.Response.RunStatus.TaskResultList.CreateTaskCount() {
				t.Fatalf(
					"Expected create task count %d got %d: %s",
					mock.expectedCreateTaskCount,
					r.Response.RunStatus.TaskResultList.CreateTaskCount(),
					r.Response.RunStatus,
				)
			}
			if mock.expectedDeleteTaskCount !=
				r.Response.RunStatus.TaskResultList.DeleteTaskCount() {
				t.Fatalf(
					"Expected delete task count %d got %d: %s",
					mock.expectedDeleteTaskCount,
					r.Response.RunStatus.TaskResultList.DeleteTaskCount(),
					r.Response.RunStatus,
				)
			}
			if mock.expectedUpdateTaskCount !=
				r.Response.RunStatus.TaskResultList.UpdateTaskCount() {
				t.Fatalf(
					"Expected update task count %d got %d: %s",
					mock.expectedUpdateTaskCount,
					r.Response.RunStatus.TaskResultList.UpdateTaskCount(),
					r.Response.RunStatus,
				)
			}
			if mock.expectedIfCondTaskCount !=
				r.Response.RunStatus.TaskResultList.IfCondTaskCount() {
				t.Fatalf(
					"Expected if-cond task count %d got %d: %s",
					mock.expectedIfCondTaskCount,
					r.Response.RunStatus.TaskResultList.IfCondTaskCount(),
					r.Response.RunStatus,
				)
			}
		})
	}
}
