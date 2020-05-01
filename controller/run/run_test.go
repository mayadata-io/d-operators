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

func TestRunnableExecRun(t *testing.T) {
	var tests = map[string]struct {
		RunCond   *types.ResourceCheck
		Resources []*unstructured.Unstructured
		Tasks     []types.Task
		isErr     bool
	}{
		"duplicate task key error": {
			Tasks: []types.Task{
				types.Task{
					Key: "duplicate",
				},
				types.Task{
					Key: "duplicate",
				},
			},
			isErr: true,
		},
		"simple assert task": {
			Tasks: []types.Task{
				types.Task{
					Key: "simple-assert",
					Assert: &types.Assert{
						State: map[string]interface{}{
							"kind": "Pod",
						},
					},
				},
			},
			Resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
					},
				},
			},
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			_, err := ExecRun(
				Request{
					IncludeInfo: map[types.IncludeInfoKey]bool{
						types.IncludeAllInfo: true,
					},
					ObservedResources: mock.Resources,
					Run: &unstructured.Unstructured{
						Object: map[string]interface{}{},
					},
					Watch: &unstructured.Unstructured{
						Object: map[string]interface{}{},
					},
					RunCond: mock.RunCond,
					Tasks:   mock.Tasks,
				},
			)
			if mock.isErr && err == nil {
				t.Fatalf("Expected error got none")
			}
			if !mock.isErr && err != nil {
				t.Fatalf("Expected no error got [%+v]", err)
			}
		})
	}
}

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
		RunCond       *types.ResourceCheck
		Resources     []*unstructured.Unstructured
		isRunCondPass bool
		isErr         bool
	}{
		"nil run cond - assert pass": {
			RunCond:       nil,
			isRunCondPass: true,
		},
		"run cond - assert pass": {
			RunCond: &types.ResourceCheck{
				SelectChecks: []types.ResourceSelectCheck{
					types.ResourceSelectCheck{
						Selector: v1alpha1.ResourceSelector{
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
			RunCond: &types.ResourceCheck{
				SelectChecks: []types.ResourceSelectCheck{
					types.ResourceSelectCheck{
						Selector: v1alpha1.ResourceSelector{
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
			RunCond: &types.ResourceCheck{
				SelectChecks: []types.ResourceSelectCheck{
					types.ResourceSelectCheck{
						Selector: v1alpha1.ResourceSelector{
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
					Watch: &unstructured.Unstructured{
						Object: map[string]interface{}{},
					},
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
		Tasks                         []types.Task
		Resources                     []*unstructured.Unstructured
		expectedErrCount              int
		expectedDesiredCount          int
		expectedExplicitUpdateCount   int
		expectedExplicitDeleteCount   int
		expectedAssertTaskCount       int
		expectedFailedAssertTaskCount int
		expectedPassedAssertTaskCount int
		expectedUpdateTaskCount       int
		expectedCreateTaskCount       int
		expectedDeleteTaskCount       int
		expectedIfCondTaskCount       int
		expectedSkippedTaskCount      int
		expectedTaskResultCount       int
		isErr                         bool
	}{
		// ------------------------------------------------------
		//  RUN as DELETE TASK(s)
		// ------------------------------------------------------
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
					Enabled: &types.ResourceCheck{
						SelectChecks: []types.ResourceSelectCheck{
							types.ResourceSelectCheck{
								Selector: v1alpha1.ResourceSelector{
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
					Enabled: &types.ResourceCheck{
						SelectChecks: []types.ResourceSelectCheck{
							types.ResourceSelectCheck{
								Selector: v1alpha1.ResourceSelector{
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
		// ------------------------------------------------------
		//  RUN as CREATE TASK(s)
		// ------------------------------------------------------
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
		"create 5 pods with passing if cond": {
			Tasks: []types.Task{
				types.Task{
					Key: "create-5-pods",
					Enabled: &types.ResourceCheck{
						SelectChecks: []types.ResourceSelectCheck{
							types.ResourceSelectCheck{
								Selector: v1alpha1.ResourceSelector{
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
					Enabled: &types.ResourceCheck{
						SelectChecks: []types.ResourceSelectCheck{
							types.ResourceSelectCheck{
								Selector: v1alpha1.ResourceSelector{
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
		// ------------------------------------------------------
		//  RUN as UPDATE TASK(s)
		// ------------------------------------------------------
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
					Enabled: &types.ResourceCheck{
						SelectChecks: []types.ResourceSelectCheck{
							types.ResourceSelectCheck{
								Selector: v1alpha1.ResourceSelector{
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
		// ------------------------------------------------------
		//  RUN as ASSERT STATE TASK(s)
		// ------------------------------------------------------
		"cannot assert state of pod due to failing if cond": {
			Tasks: []types.Task{
				types.Task{
					Key: "cannot-assert-state-of-pod-due-to-failing-if-cond",
					Enabled: &types.ResourceCheck{
						SelectChecks: []types.ResourceSelectCheck{
							types.ResourceSelectCheck{
								Selector: v1alpha1.ResourceSelector{
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
					Assert: &types.Assert{
						State: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name": "my-pod-1",
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
					},
				},
			},
			expectedIfCondTaskCount:  1,
			expectedSkippedTaskCount: 1,
			expectedTaskResultCount:  1,
		},
		"assert state of pod meta fields with passing if cond": {
			Tasks: []types.Task{
				types.Task{
					Key: "assert-state-of-pod-meta-fields-with-passing-if-cond",
					Enabled: &types.ResourceCheck{
						SelectChecks: []types.ResourceSelectCheck{
							types.ResourceSelectCheck{
								Selector: v1alpha1.ResourceSelector{
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
					Assert: &types.Assert{
						State: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"labels": map[string]interface{}{
									"app": "test",
								},
								"annotations": map[string]interface{}{
									"app": "test",
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
							"annotations": map[string]interface{}{
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
							"annotations": map[string]interface{}{
								"app": "test",
							},
						},
					},
				},
			},
			expectedIfCondTaskCount:       1,
			expectedAssertTaskCount:       1,
			expectedPassedAssertTaskCount: 1,
			expectedTaskResultCount:       1,
		},
		"assert state of pod name match by prefix": {
			Tasks: []types.Task{
				types.Task{
					Key: "assert-state-of-pod-name-match-by-prefix",
					Assert: &types.Assert{
						State: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name": "my-pod",
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
							"name": "my-pod-0",
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
			expectedPassedAssertTaskCount: 1,
			expectedAssertTaskCount:       1,
			expectedTaskResultCount:       1,
		},
		"cannot assert state of pod due to name mismatch": {
			Tasks: []types.Task{
				types.Task{
					Key: "cannot-assert-state-of-pod-due-to-name-mismatch",
					Assert: &types.Assert{
						State: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name": "my-no-pod",
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
							"name": "my-pod-0",
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
			expectedAssertTaskCount:       1,
			expectedFailedAssertTaskCount: 1,
			expectedTaskResultCount:       1,
		},
		// ------------------------------------------------------
		//  RUN as ASSERT COND TASK(s)
		// ------------------------------------------------------
		"if cond assert pod with match fields": {
			Tasks: []types.Task{
				types.Task{
					Key: "if-cond-assert-pod-with-match-fields",
					Assert: &types.Assert{
						ResourceCheck: types.ResourceCheck{
							SelectChecks: []types.ResourceSelectCheck{
								types.ResourceSelectCheck{
									Selector: v1alpha1.ResourceSelector{
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
					},
				},
			},
			Resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod-0",
						},
					},
				},
			},
			expectedAssertTaskCount:       1,
			expectedPassedAssertTaskCount: 1,
			expectedTaskResultCount:       1,
		},
		"if cond assert pod with match fields & labels": {
			Tasks: []types.Task{
				types.Task{
					Key: "if-cond-assert-pod-with-match-fields-&-labels",
					Assert: &types.Assert{
						ResourceCheck: types.ResourceCheck{
							SelectChecks: []types.ResourceSelectCheck{
								types.ResourceSelectCheck{
									Selector: v1alpha1.ResourceSelector{
										SelectorTerms: []*v1alpha1.SelectorTerm{
											&v1alpha1.SelectorTerm{
												MatchFields: map[string]string{
													"kind":       "Pod",
													"apiVersion": "v1",
												},
												MatchLabels: map[string]string{
													"app": "test",
												},
											},
										},
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
							"name": "my-pod-0",
							"labels": map[string]interface{}{
								"app": "test",
							},
						},
					},
				},
			},
			expectedAssertTaskCount:       1,
			expectedPassedAssertTaskCount: 1,
			expectedTaskResultCount:       1,
		},
		"failed if cond assert pod with match fields & labels": {
			Tasks: []types.Task{
				types.Task{
					Key: "failed-if-cond-assert-pod-with-match-fields-&-labels",
					Assert: &types.Assert{
						ResourceCheck: types.ResourceCheck{
							SelectChecks: []types.ResourceSelectCheck{
								types.ResourceSelectCheck{
									Selector: v1alpha1.ResourceSelector{
										SelectorTerms: []*v1alpha1.SelectorTerm{
											&v1alpha1.SelectorTerm{
												MatchFields: map[string]string{
													"kind":       "Pod",
													"apiVersion": "v1",
												},
												MatchLabels: map[string]string{
													"app": "prod",
												},
											},
										},
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
							"name": "my-pod-0",
							"labels": map[string]interface{}{
								"app": "test",
							},
						},
					},
				},
			},
			expectedAssertTaskCount:       1,
			expectedFailedAssertTaskCount: 1,
			expectedTaskResultCount:       1,
		},
		"if cond assert pod with match fields error out due to no resources": {
			Tasks: []types.Task{
				types.Task{
					Key: "if-cond-assert-pod-with-match-fields-err-out-due-to-no-resources",
					Assert: &types.Assert{
						ResourceCheck: types.ResourceCheck{
							SelectChecks: []types.ResourceSelectCheck{
								types.ResourceSelectCheck{
									Selector: v1alpha1.ResourceSelector{
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
					},
				},
			},
			Resources:        []*unstructured.Unstructured{},
			isErr:            true,
			expectedErrCount: 1,
		},
		"if cond assert pod with match fields fail due to failing if cond": {
			Tasks: []types.Task{
				types.Task{
					Key: "if-cond-assert-pod-with-match-fields-fail-due-to-failing-if-cond",
					Enabled: &types.ResourceCheck{
						SelectChecks: []types.ResourceSelectCheck{
							types.ResourceSelectCheck{
								Selector: v1alpha1.ResourceSelector{
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
					Assert: &types.Assert{
						ResourceCheck: types.ResourceCheck{
							SelectChecks: []types.ResourceSelectCheck{
								types.ResourceSelectCheck{
									Selector: v1alpha1.ResourceSelector{
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
					},
				},
			},
			Resources: []*unstructured.Unstructured{
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
			expectedSkippedTaskCount: 1,
			expectedIfCondTaskCount:  1,
			expectedTaskResultCount:  1,
		},
		"if cond assert errors due to duplicate task key": {
			Tasks: []types.Task{
				types.Task{
					Key: "duplicate",
					Assert: &types.Assert{
						ResourceCheck: types.ResourceCheck{
							SelectChecks: []types.ResourceSelectCheck{
								types.ResourceSelectCheck{
									Selector: v1alpha1.ResourceSelector{
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
					},
				},
				types.Task{
					Key: "duplicate",
					Assert: &types.Assert{
						ResourceCheck: types.ResourceCheck{
							SelectChecks: []types.ResourceSelectCheck{
								types.ResourceSelectCheck{
									Selector: v1alpha1.ResourceSelector{
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
					},
				},
			},
			Resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
							"annotations": map[string]interface{}{
								"app": "test",
							},
						},
					},
				},
			},
			expectedErrCount:              1,
			isErr:                         true,
			expectedAssertTaskCount:       1,
			expectedPassedAssertTaskCount: 1,
			expectedTaskResultCount:       1,
		},
		// --------------------------------------------------------------
		// Possible Errors or Failed Cases w.r.t Assert As If Condition
		// --------------------------------------------------------------
		"if cond assert fails due to missing field path": {
			Tasks: []types.Task{
				types.Task{
					Key: "if-cond-assert-fails-due-to-missing-field-path",
					Assert: &types.Assert{
						ResourceCheck: types.ResourceCheck{
							SelectChecks: []types.ResourceSelectCheck{
								types.ResourceSelectCheck{
									Selector: v1alpha1.ResourceSelector{
										SelectorTerms: []*v1alpha1.SelectorTerm{
											&v1alpha1.SelectorTerm{
												MatchFields: map[string]string{
													"kind":      "Pod",
													"dontExist": "v1", // fails here
												},
											},
										},
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
							"name": "my-pod",
							"annotations": map[string]interface{}{
								"app": "test",
							},
						},
					},
				},
			},
			expectedErrCount:              0,
			isErr:                         false,
			expectedFailedAssertTaskCount: 1,
			expectedAssertTaskCount:       1,
			expectedTaskResultCount:       1,
		},
		"if cond assert fails due to mismatch filtered resource count": {
			Tasks: []types.Task{
				types.Task{
					Key: "if-cond-assert-fails-due-to-mismatch-filtered-resource-count",
					Assert: &types.Assert{
						ResourceCheck: types.ResourceCheck{
							SelectChecks: []types.ResourceSelectCheck{
								types.ResourceSelectCheck{
									Selector: v1alpha1.ResourceSelector{
										SelectorTerms: []*v1alpha1.SelectorTerm{
											&v1alpha1.SelectorTerm{
												MatchFields: map[string]string{
													"kind":       "Pod",
													"apiVersion": "v1",
												},
											},
										},
									},
									// fails since operator is not set
									Count: pointer.Int(2),
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
							"name": "my-pod",
							"annotations": map[string]interface{}{
								"app": "test",
							},
						},
					},
				},
			},
			expectedErrCount: 1,
			isErr:            true,
		},
		"if cond assert fails due to missing count": {
			Tasks: []types.Task{
				types.Task{
					Key: "if-cond-assert-fails-due-to-missing-resource-count",
					Assert: &types.Assert{
						ResourceCheck: types.ResourceCheck{
							SelectChecks: []types.ResourceSelectCheck{
								types.ResourceSelectCheck{
									Selector: v1alpha1.ResourceSelector{
										SelectorTerms: []*v1alpha1.SelectorTerm{
											&v1alpha1.SelectorTerm{
												MatchFields: map[string]string{
													"kind":       "Pod",
													"apiVersion": "v1",
												},
											},
										},
									},
									// fails here
									Operator: types.ResourceSelectOperatorEqualsCount,
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
							"name": "my-pod",
							"annotations": map[string]interface{}{
								"app": "test",
							},
						},
					},
				},
			},
			expectedErrCount: 1,
			isErr:            true,
		},
		"if cond assert fails due to invalid operator": {
			Tasks: []types.Task{
				types.Task{
					Key: "if-cond-assert-fails-due-to-invalid-operator",
					Assert: &types.Assert{
						ResourceCheck: types.ResourceCheck{
							SelectChecks: []types.ResourceSelectCheck{
								types.ResourceSelectCheck{
									Selector: v1alpha1.ResourceSelector{
										SelectorTerms: []*v1alpha1.SelectorTerm{
											&v1alpha1.SelectorTerm{
												MatchFields: map[string]string{
													"kind":       "Pod",
													"apiVersion": "v1",
												},
											},
										},
									},
									Operator: "MyInvalidOp", // fails here
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
							"name": "my-pod",
						},
					},
				},
			},
			expectedErrCount: 1,
			isErr:            true,
		},
		"if cond assert fails due to failed match via OR operator": {
			Tasks: []types.Task{
				types.Task{
					Key: "if-cond-assert-fails-due-to-failed-match-via-OR-operator",
					Assert: &types.Assert{
						ResourceCheck: types.ResourceCheck{
							SelectChecks: []types.ResourceSelectCheck{
								types.ResourceSelectCheck{
									Selector: v1alpha1.ResourceSelector{
										SelectorTerms: []*v1alpha1.SelectorTerm{
											&v1alpha1.SelectorTerm{
												MatchFields: map[string]string{
													"kind": "Service",
												},
											},
										},
									},
								},
								types.ResourceSelectCheck{
									Selector: v1alpha1.ResourceSelector{
										SelectorTerms: []*v1alpha1.SelectorTerm{
											&v1alpha1.SelectorTerm{
												MatchFields: map[string]string{
													"kind": "Deployment",
												},
											},
										},
									},
								},
							},
							CheckOperator: types.ResourceCheckOperatorOR,
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
							"name": "my-pod",
						},
					},
				},
			},
			expectedFailedAssertTaskCount: 1,
			expectedAssertTaskCount:       1,
			expectedTaskResultCount:       1,
		},
		"if cond assert fails due to failed match via AND operator": {
			Tasks: []types.Task{
				types.Task{
					Key: "if-cond-assert-fails-due-to-failed-match-via-AND-operator",
					Assert: &types.Assert{
						ResourceCheck: types.ResourceCheck{
							SelectChecks: []types.ResourceSelectCheck{
								types.ResourceSelectCheck{
									Selector: v1alpha1.ResourceSelector{
										SelectorTerms: []*v1alpha1.SelectorTerm{
											&v1alpha1.SelectorTerm{
												MatchFields: map[string]string{
													"kind":       "Pod",
													"apiVersion": "v2",
												},
											},
										},
									},
								},
							},
							CheckOperator: types.ResourceCheckOperatorAND,
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
							"name": "my-pod",
						},
					},
				},
			},
			expectedFailedAssertTaskCount: 1,
			expectedAssertTaskCount:       1,
			expectedTaskResultCount:       1,
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
			if mock.expectedPassedAssertTaskCount !=
				r.Response.RunStatus.TaskResultList.PassedAssertTaskCount() {
				t.Fatalf(
					"Expected passed assert task count %d got %d: %s",
					mock.expectedPassedAssertTaskCount,
					r.Response.RunStatus.TaskResultList.PassedAssertTaskCount(),
					r.Response.RunStatus,
				)
			}
			if mock.expectedFailedAssertTaskCount !=
				r.Response.RunStatus.TaskResultList.FailedAssertTaskCount() {
				t.Fatalf(
					"Expected failed assert task count %d got %d: %s",
					mock.expectedFailedAssertTaskCount,
					r.Response.RunStatus.TaskResultList.FailedAssertTaskCount(),
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
