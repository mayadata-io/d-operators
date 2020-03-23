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
	metac "openebs.io/metac/apis/metacontroller/v1alpha1"
)

func TestUpdateBuilderFilterResources(t *testing.T) {
	var tests = map[string]struct {
		observed  []*unstructured.Unstructured
		forselect metac.ResourceSelector
		expected  []*unstructured.Unstructured
		isErr     bool
	}{
		"no observed list": {},
		"nil observed list": {
			observed: nil,
		},
		"1 observed + no select": {
			observed: []*unstructured.Unstructured{
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
		"1 observed + mismatch select": {
			observed: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
					},
				},
			},
			forselect: metac.ResourceSelector{
				SelectorTerms: []*metac.SelectorTerm{
					&metac.SelectorTerm{
						MatchFields: map[string]string{
							"kind": "Service",
						},
					},
				},
			},
		},
		"1 observed + matching select": {
			observed: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
					},
				},
			},
			forselect: metac.ResourceSelector{
				SelectorTerms: []*metac.SelectorTerm{
					&metac.SelectorTerm{
						MatchFields: map[string]string{
							"kind": "Pod",
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
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			b := &UpdateBuilder{
				Request: UpdateRequest{
					ObservedResources: mock.observed,
					For:               mock.forselect,
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
			if !unstruct.List(b.filteredResources).ContainsAll(mock.expected) {
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
