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

package job

import (
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"mayadata.io/d-operators/common/pointer"
	types "mayadata.io/d-operators/types/job"
)

func TestRunnerRunAllTasks(t *testing.T) {
	var tests = map[string]struct {
		baseFixture         *BaseFixture
		job                 types.Job
		expectedStatusPhase types.JobStatusPhase
		isErr               bool
	}{
		"no tasks": {
			expectedStatusPhase: types.JobStatusCompleted,
		},
		"create a config map": {
			job: types.Job{
				Spec: types.JobSpec{
					Tasks: []types.Task{
						{
							Name: "create-config-map",
							Create: &types.Create{
								State: &unstructured.Unstructured{
									Object: map[string]interface{}{
										"kind":       "ConfigMap",
										"apiVersion": "v1",
										"metadata": map[string]interface{}{
											"name": "cm",
										},
									},
								},
							},
						},
					},
				},
			},
			expectedStatusPhase: types.JobStatusCompleted,
			isErr:               false,
		},
		"create a failfast config map": {
			job: types.Job{
				Spec: types.JobSpec{
					Tasks: []types.Task{
						{
							Name: "create-config-map",
							FailFast: &types.FailFast{
								When: types.FailFastOnDiscoveryError,
							},
							Create: &types.Create{
								State: &unstructured.Unstructured{
									Object: map[string]interface{}{
										"kind":       "ConfigMap",
										"apiVersion": "v1",
										"metadata": map[string]interface{}{
											"name": "cm",
										},
									},
								},
							},
						},
					},
				},
			},
			expectedStatusPhase: types.JobStatusCompleted,
			isErr:               false,
		},
		"invalid assertion of config map": {
			job: types.Job{
				Spec: types.JobSpec{
					Tasks: []types.Task{
						{
							Name: "invalid-assertion-of-config-map",
							FailFast: &types.FailFast{
								When: types.FailFastOnDiscoveryError,
							},
							Assert: &types.Assert{
								State: &unstructured.Unstructured{
									Object: map[string]interface{}{
										"kind":       "ConfigMap",
										"apiVersion": "v1",
										"metadata": map[string]interface{}{
											"name": "cm",
										},
									},
								},
								StateCheck: &types.StateCheck{
									Operator: types.StateCheckOperatorEquals,
									Count:    pointer.Int(1), // invalid
								},
							},
						},
					},
				},
			},
			expectedStatusPhase: types.JobStatusCompleted,
			isErr:               true,
		},
		"assert absence of config map": {
			job: types.Job{
				Spec: types.JobSpec{
					Tasks: []types.Task{
						{
							Name: "assert-absence-of-config-map",
							FailFast: &types.FailFast{
								When: types.FailFastOnDiscoveryError,
							},
							Assert: &types.Assert{
								State: &unstructured.Unstructured{
									Object: map[string]interface{}{
										"kind":       "ConfigMap",
										"apiVersion": "v1",
										"metadata": map[string]interface{}{
											"name": "cm",
										},
									},
								},
								StateCheck: &types.StateCheck{
									Operator: types.StateCheckOperatorNotFound,
								},
							},
						},
					},
				},
			},
			expectedStatusPhase: types.JobStatusCompleted,
			isErr:               false,
		},
		"assert presence of the fake config map": {
			baseFixture: NoopConfigMapFixture,
			job: types.Job{
				Spec: types.JobSpec{
					Tasks: []types.Task{
						{
							Name: "assert-presence-of-fake-config-map",
							FailFast: &types.FailFast{
								When: types.FailFastOnDiscoveryError,
							},
							Assert: &types.Assert{
								State: &unstructured.Unstructured{
									Object: map[string]interface{}{
										"kind":       "ConfigMap",
										"apiVersion": "v1",
										"metadata": map[string]interface{}{
											"name": "cm-1",
										},
									},
								},
							},
						},
					},
				},
			},
			expectedStatusPhase: types.JobStatusCompleted,
			isErr:               false,
		},
		"assert presence of spec in the fake config map": {
			baseFixture: NoopConfigMapFixture,
			job: types.Job{
				Spec: types.JobSpec{
					Tasks: []types.Task{
						{
							Name: "assert-presence-of-spec-in-the-fake-config-map",
							FailFast: &types.FailFast{
								When: types.FailFastOnDiscoveryError,
							},
							Assert: &types.Assert{
								State: &unstructured.Unstructured{
									Object: map[string]interface{}{
										"kind":       "ConfigMap",
										"apiVersion": "v1",
										"metadata": map[string]interface{}{
											"name": "cm-1",
										},
									},
								},
								PathCheck: &types.PathCheck{
									Path:     "spec",
									Operator: types.PathCheckOperatorExists,
								},
							},
						},
					},
				},
			},
			expectedStatusPhase: types.JobStatusCompleted,
			isErr:               false,
		},
		"assert presence of spec in the fake config map - ii": {
			baseFixture: NoopConfigMapFixture,
			job: types.Job{
				Spec: types.JobSpec{
					Tasks: []types.Task{
						{
							Name: "assert-presence-of-spec-in-the-fake-config-map",
							FailFast: &types.FailFast{
								When: types.FailFastOnDiscoveryError,
							},
							Assert: &types.Assert{
								State: &unstructured.Unstructured{
									Object: map[string]interface{}{
										"kind":       "ConfigMap",
										"apiVersion": "v1",
										"metadata": map[string]interface{}{
											"name": "cm-1",
										},
										"spec": nil,
									},
								},
							},
						},
					},
				},
			},
			expectedStatusPhase: types.JobStatusCompleted,
			isErr:               false,
		},
		"assert absence of junk in the fake config map": {
			baseFixture: NoopConfigMapFixture,
			job: types.Job{
				Spec: types.JobSpec{
					Tasks: []types.Task{
						{
							Name: "assert-absence-of-junk-in-the-fake-config-map",
							FailFast: &types.FailFast{
								When: types.FailFastOnDiscoveryError,
							},
							Assert: &types.Assert{
								State: &unstructured.Unstructured{
									Object: map[string]interface{}{
										"kind":       "ConfigMap",
										"apiVersion": "v1",
										"metadata": map[string]interface{}{
											"name": "cm-1",
										},
									},
								},
								PathCheck: &types.PathCheck{
									Path:     "junk",
									Operator: types.PathCheckOperatorNotExists,
								},
							},
						},
					},
				},
			},
			expectedStatusPhase: types.JobStatusCompleted,
			isErr:               false,
		},
		"assert absence of junk in the fake config map - ii": {
			baseFixture: NoopConfigMapFixture,
			job: types.Job{
				Spec: types.JobSpec{
					Tasks: []types.Task{
						{
							Name: "assert-absence-of-junk-in-the-fake-config-map",
							FailFast: &types.FailFast{
								When: types.FailFastOnDiscoveryError,
							},
							Assert: &types.Assert{
								State: &unstructured.Unstructured{
									Object: map[string]interface{}{
										"kind":       "ConfigMap",
										"apiVersion": "v1",
										"metadata": map[string]interface{}{
											"name": "cm-1",
										},
										"junk": nil,
									},
								},
								StateCheck: &types.StateCheck{
									Operator: types.StateCheckOperatorNotEquals,
								},
							},
						},
					},
				},
			},
			expectedStatusPhase: types.JobStatusCompleted,
			isErr:               false,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			f := &Fixture{
				BaseFixture: NoopFixture,
			}
			if mock.baseFixture != nil {
				f.BaseFixture = mock.baseFixture
			}
			timeout := 1 * time.Second // unit test don't need to retry
			r := &Runner{
				Retry: NewRetry(RetryConfig{
					WaitTimeout: &timeout,
				}),
				Job: mock.job,
				JobStatus: &types.JobStatus{
					TaskListStatus: make(map[string]types.TaskStatus),
				},
				fixture: f,
			}
			r.initEnabled()        // init to avoid nil pointers
			got, err := r.runAll() // method under test
			if mock.isErr && err == nil {
				t.Fatal("Expected error got none")
			}
			if !mock.isErr && err != nil {
				t.Fatalf("Expected no error got %s", err.Error())
			}
			if mock.isErr {
				return
			}
			if got.Phase != mock.expectedStatusPhase {
				t.Fatalf(
					"Expected status.phase %q got %q",
					mock.expectedStatusPhase,
					got.Phase,
				)
			}
		})
	}
}
