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
	stringutil "mayadata.io/d-operators/common/string"
	types "mayadata.io/d-operators/types/run"
)

func TestCreateBuilderUpdateDesiredTemplateAnnotations(t *testing.T) {
	var tests = map[string]struct {
		run               *unstructured.Unstructured
		template          *unstructured.Unstructured
		expectAnnotations map[string]string
	}{
		"nil annotations": {
			run: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name": "run",
						"uid":  "run-101",
					},
				},
			},
			template: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{},
				},
			},
			expectAnnotations: map[string]string{
				types.AnnotationKeyRunUID:    "run-101",
				types.AnnotationKeyRunName:   "run",
				types.AnnotationKeyWatchUID:  "run-101",
				types.AnnotationKeyWatchName: "run",
				types.AnnotationKeyTaskKey:   "",
			},
		},
		"empty annotations": {
			run: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name": "run",
						"uid":  "run-101",
					},
				},
			},
			template: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"annotations": map[string]interface{}{},
					},
				},
			},
			expectAnnotations: map[string]string{
				types.AnnotationKeyRunUID:    "run-101",
				types.AnnotationKeyRunName:   "run",
				types.AnnotationKeyWatchUID:  "run-101",
				types.AnnotationKeyWatchName: "run",
				types.AnnotationKeyTaskKey:   "",
			},
		},
		"with annotations": {
			run: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name": "run",
						"uid":  "run-101",
					},
				},
			},
			template: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"annotations": map[string]interface{}{
							"app": "dope",
						},
					},
				},
			},
			expectAnnotations: map[string]string{
				"app":                        "dope",
				types.AnnotationKeyRunUID:    "run-101",
				types.AnnotationKeyRunName:   "run",
				types.AnnotationKeyWatchUID:  "run-101",
				types.AnnotationKeyWatchName: "run",
				types.AnnotationKeyTaskKey:   "",
			},
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			b := CreateStatesBuilder{
				Request: CreateRequest{
					Run:   mock.run,
					Watch: mock.run, // watch & run are considered same
				},
				desiredTemplate: mock.template,
			}
			b.updateDesiredTemplateAnnotations()
			if len(mock.expectAnnotations) != len(b.desiredTemplate.GetAnnotations()) {
				t.Fatalf(
					"Expected anns count %d got %d",
					len(mock.expectAnnotations),
					len(b.desiredTemplate.GetAnnotations()),
				)
			}
			for key, value := range mock.expectAnnotations {
				if value != b.desiredTemplate.GetAnnotations()[key] {
					t.Fatalf(
						"Expected value %q got %q for key %q",
						value,
						b.desiredTemplate.GetAnnotations()[key],
						key,
					)
				}
			}
		})
	}
}

func TestEvalDesiredName(t *testing.T) {
	var tests = map[string]struct {
		template   *unstructured.Unstructured
		expectName string
		isErr      bool
	}{
		"no names": {
			template: &unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
			isErr: true,
		},
		"with name": {
			template: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name": "test",
					},
				},
			},
			expectName: "test",
			isErr:      false,
		},
		"with generateName": {
			template: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"generateName": "test",
					},
				},
			},
			expectName: "test",
			isErr:      false,
		},
		"with name & generateName": {
			template: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name":         "test",
						"generateName": "gtest",
					},
				},
			},
			expectName: "gtest",
			isErr:      false,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			b := CreateStatesBuilder{
				desiredTemplate: mock.template,
				Request:         CreateRequest{},
			}
			b.evalDesiredName()
			if mock.isErr && b.err == nil {
				t.Fatalf("Expected error got none")
			}
			if !mock.isErr && b.err != nil {
				t.Fatalf("Expected no error got [%+v]", b.err)
			}
			if mock.isErr {
				return
			}
			if mock.expectName != b.desiredName {
				t.Fatalf(
					"Expected name %q got %q",
					mock.expectName,
					b.desiredName,
				)
			}
		})
	}
}

func TestCreateBuilderBuildDesiredStates(t *testing.T) {
	var tests = map[string]struct {
		name             string
		template         *unstructured.Unstructured
		replicas         int
		expectStateCount int
		expectNames      []string
	}{
		"empty template + 1 replica": {
			name:             "test",
			template:         &unstructured.Unstructured{},
			replicas:         1,
			expectStateCount: 1,
			expectNames:      []string{"test"},
		},
		"empty template + 2 replicas": {
			name:             "test",
			template:         &unstructured.Unstructured{},
			replicas:         2,
			expectStateCount: 2,
			expectNames:      []string{"test-0", "test-1"},
		},
		"empty template.Object + 1 replica": {
			name: "test",
			template: &unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
			replicas:         1,
			expectStateCount: 1,
			expectNames:      []string{"test"},
		},
		"empty template.Object + 2 replicas": {
			name: "test",
			template: &unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
			replicas:         2,
			expectStateCount: 2,
			expectNames:      []string{"test-0", "test-1"},
		},
		"valid template + 2 replicas": {
			name: "verify",
			template: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"app": "verify",
						},
						"annotations": map[string]interface{}{
							"app": "verify",
						},
					},
				},
			},
			replicas:         2,
			expectStateCount: 2,
			expectNames:      []string{"verify-0", "verify-1"},
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			b := CreateStatesBuilder{
				Request: CreateRequest{
					Replicas: mock.replicas,
				},
				desiredName:     mock.name,
				desiredTemplate: mock.template,
			}
			b.buildDesiredStates()
			if mock.expectStateCount != len(b.desiredResources) {
				t.Fatalf(
					"Expected states %d got %d",
					mock.expectStateCount,
					len(b.desiredResources),
				)
			}
			var gotNames []string
			for _, obj := range b.desiredResources {
				gotNames = append(
					gotNames,
					obj.GetName(),
				)
			}
			e := stringutil.NewEquality(
				mock.expectNames,
				gotNames,
			)
			if e.IsDiff() {
				t.Fatalf(
					"Expected names [%+v] got [%+v]",
					mock.expectNames,
					gotNames,
				)
			}
		})
	}
}

func TestCreateBuilderIncludeDesiredInfoIfEnabled(t *testing.T) {
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
			b := &CreateStatesBuilder{
				Request: CreateRequest{
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
			if mock.expectedDesiredCount != len(b.Result.DesiredInfo) {
				t.Fatalf(
					"Expected desired count %d got %d",
					mock.expectedDesiredCount,
					len(b.Result.DesiredInfo),
				)
			}
		})
	}
}

func TestCreateBuilderBuild(t *testing.T) {
	var tests = map[string]struct {
		Apply                    map[string]interface{}
		DesiredTemplate          *unstructured.Unstructured
		Replicas                 int
		expectedDesiredCount     int
		expectedDesiredInfoCount int
		expectedSkippedInfoCount int
		isErr                    bool
	}{
		"0 replicas is error": {
			Apply: map[string]interface{}{
				"kind":       "Pod",
				"apiVersion": "v1",
				"metadata": map[string]interface{}{
					"name":      "my-pod",
					"namespace": "cool",
					"labels": map[string]interface{}{
						"app": "pod",
					},
					"annotations": map[string]interface{}{
						"app": "pod",
					},
				},
				"spec": map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"image": "hello",
						},
					},
				},
			},
			Replicas: 0,
			isErr:    true,
		},
		"Apply a Pod": {
			Apply: map[string]interface{}{
				"kind":       "Pod",
				"apiVersion": "v1",
				"metadata": map[string]interface{}{
					"name":      "my-pod",
					"namespace": "cool",
					"labels": map[string]interface{}{
						"app": "pod",
					},
					"annotations": map[string]interface{}{
						"app": "pod",
					},
				},
				"spec": map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"image": "hello",
						},
					},
				},
			},
			Replicas:                 1,
			expectedDesiredCount:     1,
			expectedDesiredInfoCount: 1,
		},
		"1 Pod as desired template": {
			DesiredTemplate: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":       "Pod",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name":      "my-pod",
						"namespace": "cool",
						"labels": map[string]interface{}{
							"app": "pod",
						},
						"annotations": map[string]interface{}{
							"app": "pod",
						},
					},
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"image": "hello",
							},
						},
					},
				},
			},
			Replicas:                 1,
			expectedDesiredCount:     1,
			expectedDesiredInfoCount: 1,
		},
		"2 Pods as desired template": {
			DesiredTemplate: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":       "Pod",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name":      "my-pod",
						"namespace": "cool",
						"labels": map[string]interface{}{
							"app": "pod",
						},
						"annotations": map[string]interface{}{
							"app": "pod",
						},
					},
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"image": "hello",
							},
						},
					},
				},
			},
			Replicas:                 2,
			expectedDesiredCount:     2,
			expectedDesiredInfoCount: 2,
		},
		"Apply 2 Pods": {
			Apply: map[string]interface{}{
				"kind":       "Pod",
				"apiVersion": "v1",
				"metadata": map[string]interface{}{
					"name":      "my-pod",
					"namespace": "cool",
					"labels": map[string]interface{}{
						"app": "pod",
					},
					"annotations": map[string]interface{}{
						"app": "pod",
					},
				},
				"spec": map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"image": "hello",
						},
					},
				},
			},
			Replicas:                 2,
			expectedDesiredCount:     2,
			expectedDesiredInfoCount: 2,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			r := &CreateStatesBuilder{
				Request: CreateRequest{
					IncludeInfo: map[types.IncludeInfoKey]bool{
						types.IncludeAllInfo: true,
					},
					Apply:           mock.Apply,
					DesiredTemplate: mock.DesiredTemplate,
					Replicas:        mock.Replicas,
					Run: &unstructured.Unstructured{
						Object: map[string]interface{}{},
					},
					Watch: &unstructured.Unstructured{
						Object: map[string]interface{}{},
					},
					TaskKey: "try-create",
				},
				Result: &types.Result{}, // initialize
			}
			got, err := r.Build()
			if mock.isErr && err == nil {
				t.Fatalf("Expected error got none")
			}
			if !mock.isErr && err != nil {
				t.Fatalf("Expected no error got [%+v]", err)
			}
			if mock.isErr {
				return
			}
			if mock.expectedDesiredCount != len(got.DesiredResources) {
				t.Fatalf(
					"Expected desired count %d got %d",
					mock.expectedDesiredCount,
					len(got.DesiredResources),
				)
			}
			if mock.expectedDesiredInfoCount !=
				len(got.Result.DesiredInfo) {
				t.Fatalf(
					"Expected desired info count %d got %d",
					mock.expectedDesiredInfoCount,
					len(got.Result.DesiredInfo),
				)
			}
			if mock.expectedSkippedInfoCount !=
				len(got.Result.SkippedInfo) {
				t.Fatalf(
					"Expected skipped info count %d got %d",
					mock.expectedSkippedInfoCount,
					len(got.Result.SkippedInfo),
				)
			}
		})
	}
}

func TestCreateBuilderBuildNegative(t *testing.T) {
	var tests = map[string]struct {
		TaskKey         string
		Apply           map[string]interface{}
		DesiredTemplate *unstructured.Unstructured
		Replicas        int
		Watch           *unstructured.Unstructured
		Run             *unstructured.Unstructured
		isErr           bool
	}{
		"empty task key": {
			isErr: true,
		},
		"nil run": {
			TaskKey: "create-it",
			isErr:   true,
		},
		"nil watch": {
			TaskKey: "create-it",
			Run: &unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
			isErr: true,
		},
		"nil apply && nil desired template": {
			TaskKey: "create-it",
			Run: &unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
			Watch: &unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
			isErr: true,
		},
		"0 replicas": {
			TaskKey: "create-it",
			Run: &unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
			Watch: &unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
			Apply: map[string]interface{}{
				"kind": "Service",
			},
			isErr: true,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			r := &CreateStatesBuilder{
				Request: CreateRequest{
					IncludeInfo: map[types.IncludeInfoKey]bool{
						types.IncludeAllInfo: true,
					},
					Apply:           mock.Apply,
					DesiredTemplate: mock.DesiredTemplate,
					Replicas:        mock.Replicas,
					Run:             mock.Run,
					Watch:           mock.Watch,
					TaskKey:         mock.TaskKey,
				},
				Result: &types.Result{}, // initialize
			}
			_, err := r.Build()
			if mock.isErr && err == nil {
				t.Fatalf("Expected error got none")
			}
			if !mock.isErr && err != nil {
				t.Fatalf("Expected no error got [%+v]", err)
			}
		})
	}
}
