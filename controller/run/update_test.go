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

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"mayadata.io/d-operators/common/unstruct"
	types "mayadata.io/d-operators/types/run"
	metac "openebs.io/metac/apis/metacontroller/v1alpha1"
)

func TestUpdateBuilderFilterResources(t *testing.T) {
	var tests = map[string]struct {
		Observed       []*unstructured.Unstructured
		TargetSelector types.TargetSelector
		expected       []*unstructured.Unstructured
		isErr          bool
	}{
		"no observed list": {},
		"nil observed list": {
			Observed: nil,
		},
		"1 observed + no select + 1 expected": {
			Observed: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
					},
				},
			},
			expected: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
					},
				},
			},
		},
		"1 observed + 1 select + 0 expected": {
			Observed: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
					},
				},
			},
			TargetSelector: types.TargetSelector{
				ResourceSelector: metac.ResourceSelector{
					SelectorTerms: []*metac.SelectorTerm{
						&metac.SelectorTerm{
							MatchFields: map[string]string{
								"kind": "Service",
							},
						},
					},
				},
			},
		},
		"1 observed + 1 select + 1 expected": {
			Observed: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
					},
				},
			},
			TargetSelector: types.TargetSelector{
				ResourceSelector: metac.ResourceSelector{
					SelectorTerms: []*metac.SelectorTerm{
						&metac.SelectorTerm{
							MatchFields: map[string]string{
								"kind": "Pod",
							},
						},
					},
				},
			},
			expected: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
					},
				},
			},
		},
		"2 observed + 1 match + 1 expected": {
			Observed: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "ReplicaSet",
					},
				},
			},
			TargetSelector: types.TargetSelector{
				ResourceSelector: metac.ResourceSelector{
					SelectorTerms: []*metac.SelectorTerm{
						&metac.SelectorTerm{
							MatchFields: map[string]string{
								"kind": "Pod",
							},
						},
						// OR matching
						&metac.SelectorTerm{
							MatchFields: map[string]string{
								"kind": "Deployment",
							},
						},
					},
				},
			},
			expected: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
					},
				},
			},
		},
		"2 observed + 2 selects + OR operator + 2 expected": {
			Observed: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "ReplicaSet",
					},
				},
			},
			TargetSelector: types.TargetSelector{
				ResourceSelector: metac.ResourceSelector{
					SelectorTerms: []*metac.SelectorTerm{
						&metac.SelectorTerm{
							MatchFields: map[string]string{
								"kind": "Pod",
							},
						},
						// OR operator
						&metac.SelectorTerm{
							MatchFields: map[string]string{
								"kind": "ReplicaSet",
							},
						},
					},
				},
			},
			expected: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "ReplicaSet",
					},
				},
			},
		},
		"2 observed + 1 select + AND operator + 2 expected": {
			Observed: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app": "cool",
							},
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "ReplicaSet",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app": "cool",
							},
						},
					},
				},
			},
			TargetSelector: types.TargetSelector{
				ResourceSelector: metac.ResourceSelector{
					SelectorTerms: []*metac.SelectorTerm{
						&metac.SelectorTerm{
							MatchFields: map[string]string{
								"apiVersion":          "v1",
								"metadata.labels.app": "cool", // AND operator
							},
						},
					},
				},
			},
			expected: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "ReplicaSet",
						"apiVersion": "v1",
					},
				},
			},
		},
		"3 observed + 1 select + AND operator + 2 expected": {
			Observed: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app": "cool",
							},
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Deployment",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app": "ice",
							},
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "ReplicaSet",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app": "cool",
							},
						},
					},
				},
			},
			TargetSelector: types.TargetSelector{
				ResourceSelector: metac.ResourceSelector{
					SelectorTerms: []*metac.SelectorTerm{
						&metac.SelectorTerm{
							MatchFields: map[string]string{
								"apiVersion":          "v1",
								"metadata.labels.app": "cool", // AND operator
							},
						},
					},
				},
			},
			expected: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "ReplicaSet",
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
			b := &UpdateStatesBuilder{
				Request: UpdateRequest{
					ObservedResources: mock.Observed,
					TargetSelector:    mock.TargetSelector,
				},
			}
			b.filterResources()
			if mock.isErr && b.err == nil {
				t.Fatalf("Expected error got none")
			}
			if !mock.isErr && b.err != nil {
				t.Fatalf("Expected no error got [%+v]", b.err)
			}
			if mock.isErr {
				return
			}
			if len(mock.expected) != len(b.filteredResources) {
				t.Fatalf(
					"Expected filtered resource count %d got %d",
					len(mock.expected),
					len(b.filteredResources),
				)
			}
			if !unstruct.List(b.filteredResources).IdentifiesAll(mock.expected) {
				t.Fatalf(
					"Expected no diff got \n%s",
					cmp.Diff(
						b.filterResources,
						mock.expected,
					),
				)
			}
		})
	}
}

func TestFilterResourcesWithSkipInfo(t *testing.T) {
	var tests = map[string]struct {
		Resources           []*unstructured.Unstructured
		TargetSelector      types.TargetSelector
		expectSkipCount     int
		expectFilteredCount int
		isErr               bool
	}{
		"nil object in observed": {
			Resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: nil,
				},
			},
			isErr: true,
		},
		"1 observed Pod + Service as target": {
			Resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
					},
				},
			},
			TargetSelector: types.TargetSelector{
				ResourceSelector: metac.ResourceSelector{
					SelectorTerms: []*metac.SelectorTerm{
						&metac.SelectorTerm{
							MatchFields: map[string]string{
								"kind": "Service",
							},
						},
					},
				},
			},
			expectSkipCount: 1,
		},
		"2 observed Pods + Pod + Namespace + Labels + Anns as target - all filtered": {
			Resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "my-pod-1",
							"namespace": "my-ns",
							"labels": map[string]interface{}{
								"app": "pod",
							},
							"annotations": map[string]interface{}{
								"app": "pod",
							},
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "my-pod-2",
							"namespace": "my-ns",
							"labels": map[string]interface{}{
								"app": "pod",
							},
							"annotations": map[string]interface{}{
								"app": "pod",
							},
						},
					},
				},
			},
			TargetSelector: types.TargetSelector{
				ResourceSelector: metac.ResourceSelector{
					SelectorTerms: []*metac.SelectorTerm{
						// all conditions in a SelectorTerm must be satisfied
						&metac.SelectorTerm{
							MatchFields: map[string]string{
								"kind":               "Pod",
								"apiVersion":         "v1",
								"metadata.namespace": "my-ns",
							},
							MatchLabels: map[string]string{
								"app": "pod",
							},
							MatchAnnotations: map[string]string{
								"app": "pod",
							},
						},
					},
				},
			},
			expectFilteredCount: 2,
		},
		"2 observed Pods + Pod + Namespace + Labels + Anns as target - all skipped": {
			Resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "my-pod-1",
							"namespace": "my-ns",
							"labels": map[string]interface{}{
								"app": "pod",
							},
							"annotations": map[string]interface{}{
								"app": "pod",
							},
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "my-pod-2",
							"namespace": "my-ns",
							"labels": map[string]interface{}{
								"app": "pod",
							},
							"annotations": map[string]interface{}{
								"app": "pod",
							},
						},
					},
				},
			},
			TargetSelector: types.TargetSelector{
				ResourceSelector: metac.ResourceSelector{
					SelectorTerms: []*metac.SelectorTerm{
						// all conditions in a SelectorTerm must be satisfied
						&metac.SelectorTerm{
							MatchFields: map[string]string{
								"kind":               "Pod",
								"apiVersion":         "v1",
								"metadata.namespace": "my-ns",
							},
							MatchLabels: map[string]string{
								"app": "no-pod",
							},
							MatchAnnotations: map[string]string{
								"app": "no-pod",
							},
						},
					},
				},
			},
			expectSkipCount: 2,
		},
		"0 observed + Service as target": {
			Resources: []*unstructured.Unstructured{},
			TargetSelector: types.TargetSelector{
				ResourceSelector: metac.ResourceSelector{
					SelectorTerms: []*metac.SelectorTerm{
						&metac.SelectorTerm{
							MatchFields: map[string]string{
								"kind": "Service",
							},
						},
					},
				},
			},
		},
		"0 observed + 0 target": {
			Resources: []*unstructured.Unstructured{},
			TargetSelector: types.TargetSelector{
				ResourceSelector: metac.ResourceSelector{
					SelectorTerms: []*metac.SelectorTerm{},
				},
			},
		},
		"1 observed Pod + Any resource is valid target": {
			Resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
					},
				},
			},
			TargetSelector: types.TargetSelector{
				ResourceSelector: metac.ResourceSelector{
					SelectorTerms: []*metac.SelectorTerm{},
				},
			},
			expectFilteredCount: 1,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			b := &UpdateStatesBuilder{
				Request: UpdateRequest{
					IncludeInfo: map[types.IncludeInfoKey]bool{
						"*": true, // will include skipped info
					},
					TargetSelector:    mock.TargetSelector,
					ObservedResources: mock.Resources,
					Watch: &unstructured.Unstructured{
						Object: map[string]interface{}{},
					},
				},
				Result: &types.Result{},
			}
			b.filterResources()
			if mock.isErr && b.err == nil {
				t.Fatalf("Expected error got none")
			}
			if !mock.isErr && b.err != nil {
				t.Fatalf("Expected no error got [%+v]", b.err)
			}
			if mock.isErr {
				return
			}
			if mock.expectSkipCount != len(b.Result.SkippedInfo) {
				t.Fatalf(
					"Expected skip count %d got %d",
					mock.expectSkipCount,
					len(b.Result.SkippedInfo),
				)
			}
			if mock.expectFilteredCount != len(b.filteredResources) {
				t.Fatalf(
					"Expected filtered count %d got %d",
					mock.expectFilteredCount,
					len(b.filteredResources),
				)
			}
		})
	}
}

func TestIsSkipUpdate(t *testing.T) {
	var tests = map[string]struct {
		filtered []*unstructured.Unstructured
		isSkip   bool
	}{
		"nil filtered": {
			isSkip: true,
		},
		"empty filtered": {
			filtered: []*unstructured.Unstructured{},
			isSkip:   true,
		},
		"1 filtered": {
			filtered: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{},
				},
			},
			isSkip: false,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			b := UpdateStatesBuilder{
				filteredResources: mock.filtered,
			}
			b.isSkipUpdate()
			if mock.isSkip != b.isSkip {
				t.Fatalf(
					"Expected skip %t got %t",
					mock.isSkip,
					b.isSkip,
				)
			}
		})
	}
}

func TestUpdateBuilderGroupResourcesByUpdateType(t *testing.T) {
	var watch = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"uid": "watch-101",
			},
		},
	}
	var tests = map[string]struct {
		filteredResources       []*unstructured.Unstructured
		expectedExplicitUpdates []*unstructured.Unstructured
		expectedDesiredUpdates  []*unstructured.Unstructured
	}{
		"no filtered resources": {},
		"1 managed Pod + 1 un-managed Pod": {
			filteredResources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"name": "pod-1",
							"annotations": map[string]interface{}{
								types.AnnotationKeyMetacCreatedDueToWatch: "watch-101",
							},
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"name": "pod-2",
							"annotations": map[string]interface{}{
								types.AnnotationKeyMetacCreatedDueToWatch: "unknown-101",
							},
						},
					},
				},
			},
			expectedDesiredUpdates: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"name": "pod-1",
						},
					},
				},
			},
			expectedExplicitUpdates: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"name": "pod-2",
						},
					},
				},
			},
		},
		"1 un-managed Pod + 1 un-managed Deployment": {
			filteredResources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Deployment",
						"metadata": map[string]interface{}{
							"name": "deploy-1",
							"annotations": map[string]interface{}{
								types.AnnotationKeyMetacCreatedDueToWatch: "unknown-101",
							},
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"name": "pod-1",
							"annotations": map[string]interface{}{
								types.AnnotationKeyMetacCreatedDueToWatch: "unknown-101",
							},
						},
					},
				},
			},
			expectedExplicitUpdates: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Deployment",
						"metadata": map[string]interface{}{
							"name": "deploy-1",
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"name": "pod-1",
						},
					},
				},
			},
		},
		"1 un-managed Pod + 1 un-managed Deployment + no annotations": {
			filteredResources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Deployment",
						"metadata": map[string]interface{}{
							"name": "deploy-1",
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"name": "pod-1",
						},
					},
				},
			},
			expectedExplicitUpdates: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Deployment",
						"metadata": map[string]interface{}{
							"name": "deploy-1",
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"name": "pod-1",
						},
					},
				},
			},
		},
		"1 managed Pod + 1 managed Deployment": {
			filteredResources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Deployment",
						"metadata": map[string]interface{}{
							"name": "deploy-1",
							"annotations": map[string]interface{}{
								types.AnnotationKeyMetacCreatedDueToWatch: "watch-101",
							},
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"name": "pod-1",
							"annotations": map[string]interface{}{
								types.AnnotationKeyMetacCreatedDueToWatch: "watch-101",
							},
						},
					},
				},
			},
			expectedDesiredUpdates: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Deployment",
						"metadata": map[string]interface{}{
							"name": "deploy-1",
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"name": "pod-1",
						},
					},
				},
			},
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			b := UpdateStatesBuilder{
				filteredResources: mock.filteredResources,
				Request: UpdateRequest{
					Watch: watch,
				},
			}
			b.groupResourcesByUpdateType()
			if len(mock.expectedDesiredUpdates) != len(b.markedForDesiredUpdates) {
				t.Fatalf(
					"Expected desired updates %d got %d",
					len(mock.expectedDesiredUpdates),
					len(b.markedForDesiredUpdates),
				)
			}
			if len(mock.expectedExplicitUpdates) != len(b.markedForExplicitUpdates) {
				t.Fatalf(
					"Expected explicit updates %d got %d",
					len(mock.expectedExplicitUpdates),
					len(b.markedForExplicitUpdates),
				)
			}
			if !unstruct.
				List(b.markedForDesiredUpdates).
				IdentifiesAll(mock.expectedDesiredUpdates) {
				t.Fatalf(
					"Expected no diff in desired updates got\n%s",
					cmp.Diff(
						b.markedForDesiredUpdates,
						mock.expectedDesiredUpdates,
					),
				)
			}
			if !unstruct.
				List(b.markedForExplicitUpdates).
				IdentifiesAll(mock.expectedExplicitUpdates) {
				t.Fatalf(
					"Expected no diff in explicit updates got\n%s",
					cmp.Diff(
						b.markedForExplicitUpdates,
						mock.expectedExplicitUpdates,
					),
				)
			}
		})
	}
}

func TestUpdateBuilderRunApplyForDesiredUpdates(t *testing.T) {
	var tests = map[string]struct {
		original []*unstructured.Unstructured
		apply    map[string]interface{}
		expected []*unstructured.Unstructured
		isErr    bool
	}{
		"no resources": {},
		"1 pod with labels + update labels": {
			original: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
							"labels": map[string]interface{}{
								"app": "cool",
							},
						},
					},
				},
			},
			apply: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"try": "up",
					},
				},
			},
			expected: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
							"labels": map[string]interface{}{
								"app": "cool",
								"try": "up",
							},
						},
					},
				},
			},
			isErr: false,
		},
		"2 pod with labels + update labels": {
			original: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
							"labels": map[string]interface{}{
								"app": "cool",
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
								"app": "ice",
							},
						},
					},
				},
			},
			apply: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"try": "up",
					},
				},
			},
			expected: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
							"labels": map[string]interface{}{
								"app": "cool",
								"try": "up",
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
								"app": "ice",
								"try": "up",
							},
						},
					},
				},
			},
			isErr: false,
		},
		"1 pod & 1 service with labels + update labels": {
			original: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
							"labels": map[string]interface{}{
								"app": "cool",
							},
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Service",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-svc",
							"labels": map[string]interface{}{
								"app": "ice",
							},
						},
					},
				},
			},
			apply: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"try": "up",
					},
				},
			},
			expected: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
							"labels": map[string]interface{}{
								"app": "cool",
								"try": "up",
							},
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Service",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-svc",
							"labels": map[string]interface{}{
								"app": "ice",
								"try": "up",
							},
						},
					},
				},
			},
			isErr: false,
		},
		"1 pod with labels + same labels": {
			original: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
							"labels": map[string]interface{}{
								"app": "cool",
							},
						},
					},
				},
			},
			apply: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "cool",
					},
				},
			},
			expected: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
							"labels": map[string]interface{}{
								"app": "cool",
							},
						},
					},
				},
			},
			isErr: false,
		},
		"2 pod with labels + same labels": {
			original: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
							"labels": map[string]interface{}{
								"app": "cool",
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
								"app": "cool",
							},
						},
					},
				},
			},
			apply: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "cool",
					},
				},
			},
			expected: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
							"labels": map[string]interface{}{
								"app": "cool",
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
								"app": "cool",
							},
						},
					},
				},
			},
			isErr: false,
		},
		"1 pod & 1 Service with labels + same labels": {
			original: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
							"labels": map[string]interface{}{
								"app": "cool",
							},
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Service",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-svc",
							"labels": map[string]interface{}{
								"app": "cool",
							},
						},
					},
				},
			},
			apply: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "cool",
					},
				},
			},
			expected: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
							"labels": map[string]interface{}{
								"app": "cool",
							},
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Service",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-svc",
							"labels": map[string]interface{}{
								"app": "cool",
							},
						},
					},
				},
			},
			isErr: false,
		},
		"1 pod with no labels + new labels": {
			original: []*unstructured.Unstructured{
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
			apply: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "cool",
					},
				},
			},
			expected: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
							"labels": map[string]interface{}{
								"app": "cool",
							},
						},
					},
				},
			},
			isErr: false,
		},
		"2 pods with no labels + new labels": {
			original: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
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
			apply: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "cool",
					},
				},
			},
			expected: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
							"labels": map[string]interface{}{
								"app": "cool",
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
								"app": "cool",
							},
						},
					},
				},
			},
			isErr: false,
		},
		"1 pod & 1 deploy with no labels + new labels": {
			original: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Deployment",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-deploy",
						},
					},
				},
			},
			apply: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "cool",
					},
				},
			},
			expected: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
							"labels": map[string]interface{}{
								"app": "cool",
							},
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Deployment",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-deploy",
							"labels": map[string]interface{}{
								"app": "cool",
							},
						},
					},
				},
			},
			isErr: false,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			b := UpdateStatesBuilder{
				markedForDesiredUpdates: mock.original,
				Request: UpdateRequest{
					Apply: mock.apply,
				},
			}
			b.runApplyForDesiredUpdates()
			if mock.isErr && b.err == nil {
				t.Fatalf("Expected error got none")
			}
			if !mock.isErr && b.err != nil {
				t.Fatalf("Expected no error got [%+v]", b.err)
			}
			if len(mock.expected) != len(b.desiredUpdates) {
				t.Fatalf(
					"Expected updates %d got %d",
					len(mock.expected),
					len(b.desiredUpdates),
				)
			}
			if !unstruct.
				List(b.desiredUpdates).
				EqualsAll(mock.expected) {
				t.Fatalf(
					"Expected no diff got \n%s",
					cmp.Diff(
						b.desiredUpdates,
						mock.expected,
					),
				)
			}
		})
	}
}

func TestUpdateBuilderRunApplyForExplicitUpdates(t *testing.T) {
	var tests = map[string]struct {
		original []*unstructured.Unstructured
		apply    map[string]interface{}
		expected []*unstructured.Unstructured
		isErr    bool
	}{
		"no resources": {},
		"1 pod with labels + update labels": {
			original: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
							"labels": map[string]interface{}{
								"app": "cool",
							},
						},
					},
				},
			},
			apply: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"try": "up",
					},
				},
			},
			expected: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
							"labels": map[string]interface{}{
								"app": "cool",
								"try": "up",
							},
						},
					},
				},
			},
			isErr: false,
		},
		"2 pod with labels + update labels": {
			original: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
							"labels": map[string]interface{}{
								"app": "cool",
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
								"app": "ice",
							},
						},
					},
				},
			},
			apply: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"try": "up",
					},
				},
			},
			expected: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
							"labels": map[string]interface{}{
								"app": "cool",
								"try": "up",
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
								"app": "ice",
								"try": "up",
							},
						},
					},
				},
			},
			isErr: false,
		},
		"1 pod & 1 service with labels + update labels": {
			original: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
							"labels": map[string]interface{}{
								"app": "cool",
							},
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Service",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-svc",
							"labels": map[string]interface{}{
								"app": "ice",
							},
						},
					},
				},
			},
			apply: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"try": "up",
					},
				},
			},
			expected: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
							"labels": map[string]interface{}{
								"app": "cool",
								"try": "up",
							},
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Service",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-svc",
							"labels": map[string]interface{}{
								"app": "ice",
								"try": "up",
							},
						},
					},
				},
			},
			isErr: false,
		},
		"1 pod with labels + same labels": {
			original: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
							"labels": map[string]interface{}{
								"app": "cool",
							},
						},
					},
				},
			},
			apply: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "cool",
					},
				},
			},
			expected: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
							"labels": map[string]interface{}{
								"app": "cool",
							},
						},
					},
				},
			},
			isErr: false,
		},
		"2 pod with labels + same labels": {
			original: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
							"labels": map[string]interface{}{
								"app": "cool",
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
								"app": "cool",
							},
						},
					},
				},
			},
			apply: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "cool",
					},
				},
			},
			expected: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
							"labels": map[string]interface{}{
								"app": "cool",
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
								"app": "cool",
							},
						},
					},
				},
			},
			isErr: false,
		},
		"1 pod & 1 Service with labels + same labels": {
			original: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
							"labels": map[string]interface{}{
								"app": "cool",
							},
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Service",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-svc",
							"labels": map[string]interface{}{
								"app": "cool",
							},
						},
					},
				},
			},
			apply: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "cool",
					},
				},
			},
			expected: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
							"labels": map[string]interface{}{
								"app": "cool",
							},
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Service",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-svc",
							"labels": map[string]interface{}{
								"app": "cool",
							},
						},
					},
				},
			},
			isErr: false,
		},
		"1 pod with no labels + new labels": {
			original: []*unstructured.Unstructured{
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
			apply: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "cool",
					},
				},
			},
			expected: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
							"labels": map[string]interface{}{
								"app": "cool",
							},
						},
					},
				},
			},
			isErr: false,
		},
		"2 pods with no labels + new labels": {
			original: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
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
			apply: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "cool",
					},
				},
			},
			expected: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
							"labels": map[string]interface{}{
								"app": "cool",
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
								"app": "cool",
							},
						},
					},
				},
			},
			isErr: false,
		},
		"1 pod & 1 deploy with no labels + new labels": {
			original: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Deployment",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-deploy",
						},
					},
				},
			},
			apply: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "cool",
					},
				},
			},
			expected: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-pod",
							"labels": map[string]interface{}{
								"app": "cool",
							},
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Deployment",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "my-deploy",
							"labels": map[string]interface{}{
								"app": "cool",
							},
						},
					},
				},
			},
			isErr: false,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			b := UpdateStatesBuilder{
				markedForExplicitUpdates: mock.original,
				Request: UpdateRequest{
					Apply: mock.apply,
				},
			}
			b.runApplyForExplicitUpdates()
			if mock.isErr && b.err == nil {
				t.Fatalf("Expected error got none")
			}
			if !mock.isErr && b.err != nil {
				t.Fatalf("Expected no error got [%+v]", b.err)
			}
			if len(mock.expected) != len(b.explicitUpdates) {
				t.Fatalf(
					"Expected explicit updates %d got %d",
					len(mock.expected),
					len(b.explicitUpdates),
				)
			}
			if !unstruct.
				List(b.explicitUpdates).
				EqualsAll(mock.expected) {
				t.Fatalf(
					"Expected no diff got \n%s",
					cmp.Diff(
						b.explicitUpdates,
						mock.expected,
					),
				)
			}
		})
	}
}

func TestBuildUpdateStates(t *testing.T) {
	var tests = map[string]struct {
		Taskkey         string
		Run             *unstructured.Unstructured
		Watch           *unstructured.Unstructured
		Apply           map[string]interface{}
		TargetSelector  types.TargetSelector
		Observed        []*unstructured.Unstructured
		DesiredUpdates  []*unstructured.Unstructured
		ExplicitUpdates []*unstructured.Unstructured
		isSkip          bool
		isErr           bool
	}{
		"no task key": {
			isErr: true,
		},
		"no run": {
			Taskkey: "task-101",
			isErr:   true,
		},
		"no watch": {
			Taskkey: "task-101",
			Run: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Pod",
				},
			},
			isErr: true,
		},
		"no apply": {
			Taskkey: "task-101",
			Run: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Pod",
					"metadata": map[string]interface{}{
						"uid": "pod-101",
					},
				},
			},
			Watch: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Pod",
					"metadata": map[string]interface{}{
						"uid": "pod-101",
					},
				},
			},
			isErr: true,
		},
		"no for": {
			Taskkey: "task-101",
			Run: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Pod",
					"metadata": map[string]interface{}{
						"uid": "pod-101",
					},
				},
			},
			Watch: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Pod",
					"metadata": map[string]interface{}{
						"uid": "pod-101",
					},
				},
			},
			Apply: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "nocode",
					},
				},
			},
			TargetSelector: types.TargetSelector{
				ResourceSelector: metac.ResourceSelector{
					SelectorTerms: []*metac.SelectorTerm{},
				},
			},
			isErr: true,
		},
		"explicit update pod of Pod & Service": {
			Taskkey: "task-101",
			Run: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Run",
					"metadata": map[string]interface{}{
						"uid": "run-101",
					},
				},
			},
			Watch: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Run",
					"metadata": map[string]interface{}{
						"uid": "run-101",
					},
				},
			},
			Apply: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "nocode",
					},
				},
			},
			TargetSelector: types.TargetSelector{
				ResourceSelector: metac.ResourceSelector{
					SelectorTerms: []*metac.SelectorTerm{
						&metac.SelectorTerm{
							MatchFields: map[string]string{
								"kind": "Pod",
							},
						},
					},
				},
			},
			Observed: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"name": "my-pod",
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Service",
						"metadata": map[string]interface{}{
							"name": "my-svc",
						},
					},
				},
			},
			ExplicitUpdates: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"name": "my-pod",
							"labels": map[string]interface{}{
								"app": "nocode",
							},
						},
					},
				},
			},
			isErr: false,
		},
		"explicit update Service of Pod & Service": {
			Taskkey: "task-101",
			Run: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Run",
					"metadata": map[string]interface{}{
						"uid": "run-101",
					},
				},
			},
			Watch: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Run",
					"metadata": map[string]interface{}{
						"uid": "run-101",
					},
				},
			},
			Apply: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "nocode",
					},
				},
			},
			TargetSelector: types.TargetSelector{
				ResourceSelector: metac.ResourceSelector{
					SelectorTerms: []*metac.SelectorTerm{
						&metac.SelectorTerm{
							MatchFields: map[string]string{
								"kind": "Service",
							},
						},
					},
				},
			},
			Observed: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"name": "my-pod",
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Service",
						"metadata": map[string]interface{}{
							"name": "my-svc",
						},
					},
				},
			},
			ExplicitUpdates: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Service",
						"metadata": map[string]interface{}{
							"name": "my-svc",
							"labels": map[string]interface{}{
								"app": "nocode",
							},
						},
					},
				},
			},
			isErr: false,
		},
		"desired update Service of Pod & Service": {
			Taskkey: "task-101",
			Run: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Run",
					"metadata": map[string]interface{}{
						"uid": "run-101",
					},
				},
			},
			Watch: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Run",
					"metadata": map[string]interface{}{
						"uid": "run-101",
					},
				},
			},
			Apply: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "nocode",
					},
				},
			},
			TargetSelector: types.TargetSelector{
				ResourceSelector: metac.ResourceSelector{
					SelectorTerms: []*metac.SelectorTerm{
						&metac.SelectorTerm{
							MatchFields: map[string]string{
								"kind": "Service",
							},
						},
					},
				},
			},
			Observed: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"name": "my-pod",
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Service",
						"metadata": map[string]interface{}{
							"name": "my-svc",
							"annotations": map[string]interface{}{
								types.AnnotationKeyMetacCreatedDueToWatch: "run-101",
							},
						},
					},
				},
			},
			DesiredUpdates: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Service",
						"metadata": map[string]interface{}{
							"name": "my-svc",
							"annotations": map[string]interface{}{
								types.AnnotationKeyMetacCreatedDueToWatch: "run-101",
							},
							"labels": map[string]interface{}{
								"app": "nocode",
							},
						},
					},
				},
			},
			isErr: false,
		},
		"no updates of Pod & Service": {
			Taskkey: "task-101",
			Run: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Run",
					"metadata": map[string]interface{}{
						"uid": "run-101",
					},
				},
			},
			Watch: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Run",
					"metadata": map[string]interface{}{
						"uid": "run-101",
					},
				},
			},
			Apply: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "nocode",
					},
				},
			},
			TargetSelector: types.TargetSelector{
				ResourceSelector: metac.ResourceSelector{
					SelectorTerms: []*metac.SelectorTerm{
						&metac.SelectorTerm{
							MatchFields: map[string]string{
								"kind": "Deployment",
							},
						},
					},
				},
			},
			Observed: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"name": "my-pod",
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Service",
						"metadata": map[string]interface{}{
							"name": "my-svc",
						},
					},
				},
			},
			isSkip: true,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			got, err := BuildUpdateStates(UpdateRequest{
				Run:               mock.Run,
				Watch:             mock.Watch,
				Apply:             mock.Apply,
				TargetSelector:    mock.TargetSelector,
				ObservedResources: mock.Observed,
				TaskKey:           mock.Taskkey,
			})
			if mock.isErr && err == nil {
				t.Fatalf("Expected error got none")
			}
			if !mock.isErr && err != nil {
				t.Fatalf("Expected no error got [%+v]", err)
			}
			if mock.isErr {
				return
			}
			if mock.isSkip && got.Result.Phase != types.ResultPhaseSkipped {
				t.Fatalf(
					"Expected phase %q got %q",
					types.ResultPhaseSkipped,
					got.Result.Phase,
				)
			}
			if !mock.isSkip && got.Result.Phase == types.ResultPhaseSkipped {
				t.Fatalf(
					"Didn't expect phase %q ",
					types.ResultPhaseSkipped,
				)
			}
			if len(got.ExplicitUpdates) != len(mock.ExplicitUpdates) {
				t.Fatalf(
					"Expected explicit updates %d got %d",
					len(got.ExplicitUpdates),
					len(mock.ExplicitUpdates),
				)
			}
			if !unstruct.List(got.ExplicitUpdates).EqualsAll(mock.ExplicitUpdates) {
				t.Fatalf(
					"Expected explicit updates with no diff got \n%s",
					cmp.Diff(
						got.ExplicitUpdates,
						mock.ExplicitUpdates,
					),
				)
			}
			if len(got.DesiredUpdates) != len(mock.DesiredUpdates) {
				t.Fatalf(
					"Expected updates %d got %d",
					len(got.DesiredUpdates),
					len(mock.DesiredUpdates),
				)
			}
			if !unstruct.List(got.DesiredUpdates).EqualsAll(mock.DesiredUpdates) {
				t.Fatalf(
					"Expected updates with no diff got \n%s",
					cmp.Diff(
						got.DesiredUpdates,
						mock.DesiredUpdates,
					),
				)
			}
		})
	}
}

func TestUpdateBuilderIncludeSkippedInfoIfEnabled(t *testing.T) {
	var tests = map[string]struct {
		IncludeInfo       map[types.IncludeInfoKey]bool
		expectedSkipCount int
	}{
		"include all": {
			IncludeInfo: map[types.IncludeInfoKey]bool{
				types.IncludeAllInfo: true,
			},
			expectedSkipCount: 1,
		},
		"exclude all": {
			IncludeInfo: map[types.IncludeInfoKey]bool{
				types.IncludeAllInfo: false,
			},
			expectedSkipCount: 0,
		},
		"include *": {
			IncludeInfo: map[types.IncludeInfoKey]bool{
				"*": true,
			},
			expectedSkipCount: 1,
		},
		"exclude *": {
			IncludeInfo: map[types.IncludeInfoKey]bool{
				"*": false,
			},
			expectedSkipCount: 0,
		},
		"include none": {
			IncludeInfo:       map[types.IncludeInfoKey]bool{},
			expectedSkipCount: 0,
		},
		"include nil": {
			expectedSkipCount: 0,
		},
		"include skip": {
			IncludeInfo: map[types.IncludeInfoKey]bool{
				types.IncludeSkippedInfo: true,
			},
			expectedSkipCount: 1,
		},
		"exclude skip": {
			IncludeInfo: map[types.IncludeInfoKey]bool{
				types.IncludeSkippedInfo: false,
			},
			expectedSkipCount: 0,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			b := &UpdateStatesBuilder{
				Request: UpdateRequest{
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
			}, "Skip this")
			if mock.expectedSkipCount != len(b.Result.SkippedInfo) {
				t.Fatalf(
					"Expected skip count %d got %d",
					mock.expectedSkipCount,
					len(b.Result.SkippedInfo),
				)
			}
		})
	}
}

func TestUpdateBuilderIncludeDesiredInfoIfEnabled(t *testing.T) {
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
			b := &UpdateStatesBuilder{
				Request: UpdateRequest{
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

func TestUpdateBuilderIncludeExplicitInfoIfEnabled(t *testing.T) {
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
			b := &UpdateStatesBuilder{
				Request: UpdateRequest{
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
			if mock.expectedExplicitCount != len(b.Result.ExplicitInfo) {
				t.Fatalf(
					"Expected explicit count %d got %d",
					mock.expectedExplicitCount,
					len(b.Result.ExplicitInfo),
				)
			}
		})
	}
}
