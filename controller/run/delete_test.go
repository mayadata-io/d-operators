// +build !integration

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
	types "mayadata.io/d-operators/types/run"
)

func TestMarkResourcesForExplicitDelete(t *testing.T) {
	var tests = map[string]struct {
		watch             *unstructured.Unstructured
		observed          []*unstructured.Unstructured
		template          *unstructured.Unstructured
		resourceName      string
		expectDeleteCount int
		isErr             bool
	}{
		"watch didn't create observed + observed == template + empty resource name": {
			watch: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Run",
					"metadata": map[string]interface{}{
						"uid": "w-101",
					},
				},
			},
			observed: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"name": "my-pod",
						},
					},
				},
			},
			template: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":     "Pod",
					"metadata": map[string]interface{}{}, // spec is nil; hence delete
				},
			},
			expectDeleteCount: 1,
		},
		"watch didn't create observed + observed != template + empty reource name": {
			watch: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Run",
					"metadata": map[string]interface{}{
						"uid": "w-101",
					},
				},
			},
			observed: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":     "Pod",
						"metadata": map[string]interface{}{},
					},
				},
			},
			template: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":     "StatefulSet",
					"metadata": map[string]interface{}{},
				},
			},
			expectDeleteCount: 0,
		},
		"watch created observed + observed != template + empty resource name": {
			watch: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Run",
					"metadata": map[string]interface{}{
						"uid": "w-101",
					},
				},
			},
			observed: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"annotations": map[string]interface{}{
								"metac.openebs.io/created-due-to-watch": "w-101",
							},
						},
					},
				},
			},
			template: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":     "StatefulSet",
					"metadata": map[string]interface{}{},
				},
			},
			expectDeleteCount: 0,
		},
		"watch created observed + observed == template + observed name == desired name": {
			watch: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"uid": "w-101",
					},
				},
			},
			observed: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"name": "hello",
							"annotations": map[string]interface{}{
								"metac.openebs.io/created-due-to-watch": "w-101",
							},
						},
					},
				},
			},
			template: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":     "Pod",
					"metadata": map[string]interface{}{},
				},
			},
			resourceName:      "hello",
			expectDeleteCount: 0,
		},
		"watch didn't create observed + observed == template + observed name == desired name": {
			watch: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"uid": "w-101",
					},
				},
			},
			observed: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"name": "hello",
						},
					},
				},
			},
			template: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":     "Pod",
					"metadata": map[string]interface{}{},
				},
			},
			resourceName:      "hello",
			expectDeleteCount: 1,
		},
		"watch didn't create observed + observed == template + observed name prefixes desired name": {
			watch: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"uid": "w-101",
					},
				},
			},
			observed: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"name": "hello-0",
						},
					},
				},
			},
			template: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":     "Pod",
					"metadata": map[string]interface{}{},
				},
			},
			resourceName:      "hello",
			expectDeleteCount: 1,
		},
		"watch didn't create observed + observed == template + observed name != desired name": {
			watch: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"uid": "w-101",
					},
				},
			},
			observed: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"name": "hello",
						},
					},
				},
			},
			template: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":     "Pod",
					"metadata": map[string]interface{}{},
				},
			},
			resourceName:      "hi",
			expectDeleteCount: 0,
		},
		"watch uid != observed uid + observed == template + observed name == desired name": {
			watch: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"uid": "w-101",
					},
				},
			},
			observed: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"name": "hello",
							"annotations": map[string]interface{}{
								"metac.openebs.io/created-due-to-watch": "w-102",
							},
						},
					},
				},
			},
			template: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":     "Pod",
					"metadata": map[string]interface{}{},
				},
			},
			resourceName:      "hello",
			expectDeleteCount: 1,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			b := &DeleteStatesBuilder{
				Request: DeleteRequest{
					Watch:             mock.watch,
					ObservedResources: mock.observed,
				},
				deleteResourceName: mock.resourceName,
				deleteTemplate:     mock.template,
			}
			b.markResourcesForExplicitDelete()
			if mock.isErr && b.err == nil {
				t.Fatalf(
					"Expected error got none",
				)
			}
			if !mock.isErr && b.err != nil {
				t.Fatalf(
					"Expected no error got [%+v]",
					b.err,
				)
			}
			if mock.isErr {
				return
			}
			if mock.expectDeleteCount != len(b.explicitDeletes) {
				t.Fatalf(
					"Expected deletes %d got %d",
					mock.expectDeleteCount,
					len(b.explicitDeletes),
				)
			}
		})
	}
}

func TestDeleteBuilderIncludeDesiredInfoIfEnabled(t *testing.T) {
	var tests = map[string]struct {
		IncludeInfo          map[types.IncludeInfoKey]bool
		expectedDesiredCount int
	}{
		"include all": {
			IncludeInfo: map[types.IncludeInfoKey]bool{
				types.IncludeAllInfo: true,
			},
			expectedDesiredCount: 1,
		},
		"exclude all": {
			IncludeInfo: map[types.IncludeInfoKey]bool{
				types.IncludeAllInfo: false,
			},
			expectedDesiredCount: 0,
		},
		"include *": {
			IncludeInfo: map[types.IncludeInfoKey]bool{
				"*": true,
			},
			expectedDesiredCount: 1,
		},
		"exclude *": {
			IncludeInfo: map[types.IncludeInfoKey]bool{
				"*": false,
			},
			expectedDesiredCount: 0,
		},
		"include none": {
			IncludeInfo:          map[types.IncludeInfoKey]bool{},
			expectedDesiredCount: 0,
		},
		"include nil": {
			expectedDesiredCount: 0,
		},
		"include desired": {
			IncludeInfo: map[types.IncludeInfoKey]bool{
				types.IncludeDesiredInfo: true,
			},
			expectedDesiredCount: 1,
		},
		"exclude desired": {
			IncludeInfo: map[types.IncludeInfoKey]bool{
				types.IncludeDesiredInfo: false,
			},
			expectedDesiredCount: 0,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			b := &DeleteStatesBuilder{
				Request: DeleteRequest{
					IncludeInfo: mock.IncludeInfo,
				},
				Result: &types.Result{},
			}
			b.includeDesiredInfoIfEnabled(&unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":       "Pod",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name":      "my-pod",
						"namespace": "my-ns",
					},
				},
			}, "Desired info")
			if mock.expectedDesiredCount != len(b.Result.DesiredResourcesInfo) {
				t.Fatalf(
					"Expected desired count %d got %d",
					mock.expectedDesiredCount,
					len(b.Result.DesiredResourcesInfo),
				)
			}
		})
	}
}

func TestDeleteBuilderIncludeExplicitInfoIfEnabled(t *testing.T) {
	var tests = map[string]struct {
		IncludeInfo           map[types.IncludeInfoKey]bool
		expectedExplicitCount int
	}{
		"include all": {
			IncludeInfo: map[types.IncludeInfoKey]bool{
				types.IncludeAllInfo: true,
			},
			expectedExplicitCount: 1,
		},
		"exclude all": {
			IncludeInfo: map[types.IncludeInfoKey]bool{
				types.IncludeAllInfo: false,
			},
			expectedExplicitCount: 0,
		},
		"include *": {
			IncludeInfo: map[types.IncludeInfoKey]bool{
				"*": true,
			},
			expectedExplicitCount: 1,
		},
		"exclude *": {
			IncludeInfo: map[types.IncludeInfoKey]bool{
				"*": false,
			},
			expectedExplicitCount: 0,
		},
		"include none": {
			IncludeInfo:           map[types.IncludeInfoKey]bool{},
			expectedExplicitCount: 0,
		},
		"include nil": {
			expectedExplicitCount: 0,
		},
		"include explicit": {
			IncludeInfo: map[types.IncludeInfoKey]bool{
				types.IncludeExplicitInfo: true,
			},
			expectedExplicitCount: 1,
		},
		"exclude explicit": {
			IncludeInfo: map[types.IncludeInfoKey]bool{
				types.IncludeExplicitInfo: false,
			},
			expectedExplicitCount: 0,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			b := &DeleteStatesBuilder{
				Request: DeleteRequest{
					IncludeInfo: mock.IncludeInfo,
				},
				Result: &types.Result{},
			}
			b.includeExplicitInfoIfEnabled(&unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":       "Pod",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name":      "my-pod",
						"namespace": "my-ns",
					},
				},
			}, "Explicit info")
			if mock.expectedExplicitCount != len(b.Result.ExplicitResourcesInfo) {
				t.Fatalf(
					"Expected explicit count %d got %d",
					mock.expectedExplicitCount,
					len(b.Result.ExplicitResourcesInfo),
				)
			}
		})
	}
}

func TestDeleteBuilderIncludeSkippedInfoIfEnabled(t *testing.T) {
	var tests = map[string]struct {
		IncludeInfo          map[types.IncludeInfoKey]bool
		expectedSkippedCount int
	}{
		"include all": {
			IncludeInfo: map[types.IncludeInfoKey]bool{
				types.IncludeAllInfo: true,
			},
			expectedSkippedCount: 1,
		},
		"exclude all": {
			IncludeInfo: map[types.IncludeInfoKey]bool{
				types.IncludeAllInfo: false,
			},
			expectedSkippedCount: 0,
		},
		"include *": {
			IncludeInfo: map[types.IncludeInfoKey]bool{
				"*": true,
			},
			expectedSkippedCount: 1,
		},
		"exclude *": {
			IncludeInfo: map[types.IncludeInfoKey]bool{
				"*": false,
			},
			expectedSkippedCount: 0,
		},
		"include none": {
			IncludeInfo:          map[types.IncludeInfoKey]bool{},
			expectedSkippedCount: 0,
		},
		"include nil": {
			expectedSkippedCount: 0,
		},
		"include skipped": {
			IncludeInfo: map[types.IncludeInfoKey]bool{
				types.IncludeSkippedInfo: true,
			},
			expectedSkippedCount: 1,
		},
		"exclude skipped": {
			IncludeInfo: map[types.IncludeInfoKey]bool{
				types.IncludeSkippedInfo: false,
			},
			expectedSkippedCount: 0,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			b := &DeleteStatesBuilder{
				Request: DeleteRequest{
					IncludeInfo: mock.IncludeInfo,
				},
				Result: &types.Result{},
			}
			b.includeSkippedInfoIfEnabled(&unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":       "Pod",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name":      "my-pod",
						"namespace": "my-ns",
					},
				},
			}, "skipped info")
			if mock.expectedSkippedCount != len(b.Result.SkippedResourcesInfo) {
				t.Fatalf(
					"Expected skipped count %d got %d",
					mock.expectedSkippedCount,
					len(b.Result.SkippedResourcesInfo),
				)
			}
		})
	}
}

func TestDeleteBuilderBuild(t *testing.T) {
	var tests = map[string]struct {
		Apply                     map[string]interface{}
		DeleteTemplate            *unstructured.Unstructured
		Resources                 []*unstructured.Unstructured
		expectedExplicitInfoCount int
		expectedDesiredInfoCount  int
		expectedSkippedInfoCount  int
		isErr                     bool
	}{
		"explicit delete 1 Pod out of 1 Pod via apply": {
			Apply: map[string]interface{}{
				"kind":       "Pod",
				"apiVersion": "v1",
				"metadata": map[string]interface{}{
					"name": "my-pod",
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
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"image": "nginx",
								},
							},
						},
					},
				},
			},
			expectedExplicitInfoCount: 1,
		},
		"explicit delete 1 Pod out of 1 Pod via delete template": {
			DeleteTemplate: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":       "Pod",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name": "my-pod",
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
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"image": "nginx",
								},
							},
						},
					},
				},
			},
			expectedExplicitInfoCount: 1,
		},
		"skip 1 Pod out of 1 Pod via apply": {
			Apply: map[string]interface{}{
				"kind":       "Pod",
				"apiVersion": "v1",
				"metadata": map[string]interface{}{
					"name": "my-pod-1",
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
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"image": "nginx",
								},
							},
						},
					},
				},
			},
			expectedSkippedInfoCount: 1,
		},
		"skip 1 Pod out of 1 Pod via delete template": {
			DeleteTemplate: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":       "Pod",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name": "my-pod-1",
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
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"image": "nginx",
								},
							},
						},
					},
				},
			},
			expectedSkippedInfoCount: 1,
		},
		"desired delete 1 Pod out of 1 Pod via apply": {
			Apply: map[string]interface{}{
				"kind":       "Pod",
				"apiVersion": "v1",
				"metadata": map[string]interface{}{
					"name": "my-pod",
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
								types.AnnotationKeyMetacCreatedDueToWatch: "watch-101",
							},
						},
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"image": "nginx",
								},
							},
						},
					},
				},
			},
			expectedDesiredInfoCount: 1,
		},
		"desired delete 1 Pod out of 1 Pod via delete template": {
			DeleteTemplate: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":       "Pod",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name": "my-pod",
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
								types.AnnotationKeyMetacCreatedDueToWatch: "watch-101",
							},
						},
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"image": "nginx",
								},
							},
						},
					},
				},
			},
			expectedDesiredInfoCount: 1,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			b := &DeleteStatesBuilder{
				Request: DeleteRequest{
					IncludeInfo: map[types.IncludeInfoKey]bool{
						types.IncludeAllInfo: true,
					},
					Apply:             mock.Apply,
					DeleteTemplate:    mock.DeleteTemplate,
					ObservedResources: mock.Resources,
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
					TaskKey: "delete-it",
				},
				Result: &types.Result{},
			}
			got, err := b.Build()
			if mock.isErr && err == nil {
				t.Fatalf("Expected error got none")
			}
			if !mock.isErr && err != nil {
				t.Fatalf("Expected no error got [%+v]", err)
			}
			if mock.isErr {
				return
			}
			if mock.expectedExplicitInfoCount !=
				len(got.Result.ExplicitResourcesInfo) {
				t.Fatalf(
					"Expected explicit delete count %d got %d",
					mock.expectedExplicitInfoCount,
					len(got.Result.ExplicitResourcesInfo),
				)
			}
			if mock.expectedDesiredInfoCount !=
				len(got.Result.DesiredResourcesInfo) {
				t.Fatalf(
					"Expected desired delete count %d got %d",
					mock.expectedDesiredInfoCount,
					len(got.Result.DesiredResourcesInfo),
				)
			}
			if mock.expectedSkippedInfoCount !=
				len(got.Result.SkippedResourcesInfo) {
				t.Fatalf(
					"Expected skipped count %d got %d",
					mock.expectedSkippedInfoCount,
					len(got.Result.SkippedResourcesInfo),
				)
			}
		})
	}
}

func TestDeleteBuilderBuildNegative(t *testing.T) {
	var tests = map[string]struct {
		Apply          map[string]interface{}
		DeleteTemplate *unstructured.Unstructured
		Resources      []*unstructured.Unstructured
		Run            *unstructured.Unstructured
		Watch          *unstructured.Unstructured
		TaskKey        string
		isErr          bool
	}{
		"no taskkey": {
			isErr: true,
		},
		"nil run": {
			TaskKey: "delete-it",
			isErr:   true,
		},
		"nil watch": {
			TaskKey: "delete-it",
			Run: &unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
			isErr: true,
		},
		"nil apply & delete template": {
			TaskKey: "delete-it",
			Run: &unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
			Watch: &unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
			isErr: true,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			b := &DeleteStatesBuilder{
				Request: DeleteRequest{
					Apply:             mock.Apply,
					DeleteTemplate:    mock.DeleteTemplate,
					ObservedResources: mock.Resources,
					Run:               mock.Run,
					Watch:             mock.Watch,
					TaskKey:           mock.TaskKey,
				},
				Result: &types.Result{},
			}
			_, err := b.Build()
			if mock.isErr && err == nil {
				t.Fatalf("Expected error got none")
			}
			if !mock.isErr && err != nil {
				t.Fatalf("Expected no error got [%+v]", err)
			}
		})
	}
}
