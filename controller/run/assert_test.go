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

func TestNewResourceListCondition(t *testing.T) {
	var tests = map[string]struct {
		condition types.IfCondition
		resources []*unstructured.Unstructured
		expect    types.IfCondition
		isErr     bool
	}{
		"empty condition": {
			condition: types.IfCondition{},
			isErr:     true,
		},
		"condition without count & EqualsCount operator": {
			condition: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{},
					},
				},
				ResourceOperator: types.ResourceOperatorEqualsCount,
			},
			isErr: true,
		},
		"condition without count & GTE operator": {
			condition: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{},
					},
				},
				ResourceOperator: types.ResourceOperatorGTE,
			},
			isErr: true,
		},
		"condition without count & LTE operator": {
			condition: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{},
					},
				},
				ResourceOperator: types.ResourceOperatorLTE,
			},
			isErr: true,
		},
		"condition with count & LTE operator": {
			condition: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{},
					},
				},
				ResourceOperator: types.ResourceOperatorLTE,
				Count:            ptr.Int(1),
			},
			resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{},
				},
			},
			expect: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{},
					},
				},
				ResourceOperator: types.ResourceOperatorLTE,
				Count:            ptr.Int(1),
			},
			isErr: false,
		},
		"condition with count & GTE operator": {
			condition: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{},
					},
				},
				ResourceOperator: types.ResourceOperatorGTE,
				Count:            ptr.Int(3),
			},
			resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{},
				},
			},
			expect: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{},
					},
				},
				ResourceOperator: types.ResourceOperatorGTE,
				Count:            ptr.Int(3),
			},
			isErr: false,
		},
		"condition with count & EqualsCount operator": {
			condition: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{},
					},
				},
				ResourceOperator: types.ResourceOperatorEqualsCount,
				Count:            ptr.Int(2),
			},
			resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{},
				},
			},
			expect: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{},
					},
				},
				ResourceOperator: types.ResourceOperatorEqualsCount,
				Count:            ptr.Int(2),
			},
			isErr: false,
		},
		"condition without resource operator": {
			condition: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{},
					},
				},
			},
			resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{},
				},
			},
			expect: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{},
					},
				},
				ResourceOperator: types.ResourceOperatorExists,
			},
			isErr: false,
		},
		"condition with NotExist resource operator": {
			condition: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{},
					},
				},
				ResourceOperator: types.ResourceOperatorNotExist,
			},
			resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{},
				},
			},
			expect: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{},
					},
				},
				ResourceOperator: types.ResourceOperatorNotExist,
			},
			isErr: false,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			rc := NewResourceListCondition(
				mock.condition,
				mock.resources,
			)
			if mock.isErr && rc.err == nil {
				t.Fatalf("Expected error got none")
			}
			if !mock.isErr && rc.err != nil {
				t.Fatalf(
					"Expected no error got [%+v]",
					rc.err,
				)
			}
			if mock.isErr {
				return
			}
			if (mock.expect.Count != nil && rc.Condition.Count == nil) ||
				(mock.expect.Count == nil && rc.Condition.Count != nil) {
				t.Fatalf("Expected count does not match actual count ")
			}
			if rc.Condition.Count != nil {
				if *mock.expect.Count != *rc.Condition.Count {
					t.Fatalf(
						"Expected count %d got %d",
						*mock.expect.Count,
						*rc.Condition.Count,
					)
				}
			}
			if mock.expect.ResourceOperator != rc.Condition.ResourceOperator {
				t.Fatalf(
					"Expected operator %q got %q",
					mock.expect.ResourceOperator,
					rc.Condition.ResourceOperator,
				)
			}
			if len(mock.expect.ResourceSelector.SelectorTerms) !=
				len(rc.Condition.ResourceSelector.SelectorTerms) {
				t.Fatalf(
					"Expected select term count %d got %d",
					len(mock.expect.ResourceSelector.SelectorTerms),
					len(rc.Condition.ResourceSelector.SelectorTerms),
				)
			}
		})
	}
}

func TestResourceListConditionTryMatchAndRegister(t *testing.T) {
	var tests = map[string]struct {
		condition          types.IfCondition
		resource           *unstructured.Unstructured
		expectSuccessCount int // 1 implies match & 0 means no match
		isErr              bool
	}{
		"select pod + match fields + kind": {
			condition: types.IfCondition{
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
			resource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Pod",
				},
			},
			expectSuccessCount: 1,
		},
		"can not select pod + match fields + kind": {
			condition: types.IfCondition{
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
			resource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "StatefulSet",
				},
			},
			expectSuccessCount: 0,
		},
		"select STS + match fields + kind + apiVersion": {
			condition: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind":       "StatefulSet",
								"apiVersion": "apps/v1",
							},
						},
					},
				},
			},
			resource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":       "StatefulSet",
					"apiVersion": "apps/v1",
				},
			},
			isErr:              false,
			expectSuccessCount: 1,
		},
		"select STS + match fields + kind + spec.replicas": {
			condition: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind":          "StatefulSet",
								"spec.replicas": "1",
							},
						},
					},
				},
			},
			resource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "StatefulSet",
					"spec": map[string]interface{}{
						"replicas": "1",
					},
				},
			},
			expectSuccessCount: 1,
		},
		"can not select STS + match fields + kind + spec.replicas": {
			condition: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind":          "StatefulSet",
								"spec.replicas": "2",
							},
						},
					},
				},
			},
			resource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "StatefulSet",
					"spec": map[string]interface{}{
						"replicas": "1",
					},
				},
			},
			expectSuccessCount: 0,
		},
		"select Deployment + match labels + match fields": {
			condition: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind":          "Deployment",
								"spec.replicas": "2",
							},
							MatchLabels: map[string]string{
								"dao.mayadata.io/name": "d-operator",
							},
						},
					},
				},
			},
			resource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Deployment",
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"dao.mayadata.io/name": "d-operator",
							"i.am.running":         "tests",
						},
					},
					"spec": map[string]interface{}{
						"replicas": "2",
					},
				},
			},
			expectSuccessCount: 1,
		},
		"can not select Deployment + match labels + match fields": {
			condition: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind":          "Deployment",
								"spec.replicas": "2",
							},
							MatchLabels: map[string]string{
								"name": "d-operator",
							},
						},
					},
				},
			},
			resource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Deployment",
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"dao.mayadata.io/name": "d-operator",
						},
					},
					"spec": map[string]interface{}{
						"replicas": "2",
					},
				},
			},
			expectSuccessCount: 0,
		},
		"select Deployment + from multiple select terms": {
			condition: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind": "Deployment",
							},
						},
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind": "StatefulSet",
							},
						},
					},
				},
			},
			resource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Deployment",
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"dao.mayadata.io/name": "d-operator",
						},
					},
					"spec": map[string]interface{}{
						"replicas": "2",
					},
				},
			},
			expectSuccessCount: 1,
		},
		"select StatefulSet + from multiple select terms": {
			condition: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind": "Deployment",
							},
						},
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind": "StatefulSet",
							},
						},
					},
				},
			},
			resource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "StatefulSet",
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"dao.mayadata.io/name": "d-operator",
						},
					},
					"spec": map[string]interface{}{
						"replicas": "2",
					},
				},
			},
			expectSuccessCount: 1,
		},
		"can not select StatefulSet + from multiple select terms": {
			condition: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind": "Deployment",
							},
						},
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind": "PersistentVolume",
							},
						},
					},
				},
			},
			resource: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "StatefulSet",
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"dao.mayadata.io/name": "d-operator",
						},
					},
				},
			},
			expectSuccessCount: 0,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			rc := NewResourceListCondition(
				mock.condition,
				[]*unstructured.Unstructured{
					mock.resource,
				},
			)
			rc.tryMatchAndRegisterFor(mock.resource)
			if mock.isErr && rc.err == nil {
				t.Fatalf("Expected error got none")
			}
			if !mock.isErr && rc.err != nil {
				t.Fatalf(
					"Expected no error got [%+v]",
					rc.err,
				)
			}
			if mock.isErr {
				return
			}
			if mock.expectSuccessCount != len(rc.successfulMatches) {
				t.Fatalf(
					"Expected success %d got %d",
					mock.expectSuccessCount,
					len(rc.successfulMatches),
				)
			}
		})
	}
}

func TestResourceListConditionIsSuccess(t *testing.T) {
	var tests = map[string]struct {
		condition types.IfCondition
		resources []*unstructured.Unstructured
		isSuccess bool
		isErr     bool
	}{
		"select pod + match fields + kind": {
			condition: types.IfCondition{
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
			resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
					},
				},
			},
			isSuccess: true,
		},
		"can not select pod + match fields + kind": {
			condition: types.IfCondition{
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
			resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "StatefulSet",
					},
				},
			},
			isSuccess: false,
		},
		"select STS + match fields + kind + apiVersion": {
			condition: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind":       "StatefulSet",
								"apiVersion": "apps/v1",
							},
						},
					},
				},
			},
			resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "StatefulSet",
						"apiVersion": "apps/v1",
					},
				},
			},
			isSuccess: true,
		},
		"select STS + match fields + kind + spec.replicas": {
			condition: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind":          "StatefulSet",
								"spec.replicas": "1",
							},
						},
					},
				},
			},
			resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "StatefulSet",
						"spec": map[string]interface{}{
							"replicas": "1",
						},
					},
				},
			},
			isSuccess: true,
		},
		"can not select STS + match fields + kind + spec.replicas": {
			condition: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind":          "StatefulSet",
								"spec.replicas": "2",
							},
						},
					},
				},
			},
			resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "StatefulSet",
						"spec": map[string]interface{}{
							"replicas": "1",
						},
					},
				},
			},
			isSuccess: false,
		},
		"select Deployment + match labels + match fields": {
			condition: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind":          "Deployment",
								"spec.replicas": "2",
							},
							MatchLabels: map[string]string{
								"dao.mayadata.io/name": "d-operator",
							},
						},
					},
				},
			},
			resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Deployment",
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"dao.mayadata.io/name": "d-operator",
								"i.am.running":         "tests",
							},
						},
						"spec": map[string]interface{}{
							"replicas": "2",
						},
					},
				},
			},
			isSuccess: true,
		},
		"can not select Deployment + match labels + match fields": {
			condition: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind":          "Deployment",
								"spec.replicas": "2",
							},
							MatchLabels: map[string]string{
								"name": "d-operator",
							},
						},
					},
				},
			},
			resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Deployment",
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"dao.mayadata.io/name": "d-operator",
							},
						},
						"spec": map[string]interface{}{
							"replicas": "2",
						},
					},
				},
			},
			isSuccess: false,
		},
		"select Deployment + from multiple select terms": {
			condition: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind": "Deployment",
							},
						},
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind": "StatefulSet",
							},
						},
					},
				},
			},
			resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Deployment",
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"dao.mayadata.io/name": "d-operator",
							},
						},
						"spec": map[string]interface{}{
							"replicas": "2",
						},
					},
				},
			},
			isSuccess: true,
		},
		"select StatefulSet + from multiple select terms": {
			condition: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind": "Deployment",
							},
						},
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind": "StatefulSet",
							},
						},
					},
				},
			},
			resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "StatefulSet",
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"dao.mayadata.io/name": "d-operator",
							},
						},
						"spec": map[string]interface{}{
							"replicas": "2",
						},
					},
				},
			},
			isSuccess: true,
		},
		"can not select StatefulSet + from multiple select terms": {
			condition: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind": "Deployment",
							},
						},
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind": "PersistentVolume",
							},
						},
					},
				},
			},
			resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "StatefulSet",
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"dao.mayadata.io/name": "d-operator",
							},
						},
					},
				},
			},
			isSuccess: false,
		},
		"can not select StatefulSet + from multiple select terms + count=2": {
			condition: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind": "Deployment",
							},
						},
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind": "StatefulSet",
							},
						},
					},
				},
				ResourceOperator: types.ResourceOperatorEqualsCount,
				Count:            ptr.Int(2),
			},
			resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "StatefulSet",
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"dao.mayadata.io/name": "d-operator",
							},
						},
					},
				},
			},
			isSuccess: false,
		},
		"select both StatefulSet & Deployment + count=2": {
			condition: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind": "Deployment",
							},
						},
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind": "StatefulSet",
							},
						},
					},
				},
				ResourceOperator: types.ResourceOperatorEqualsCount,
				Count:            ptr.Int(2),
			},
			resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "StatefulSet",
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Deployment",
					},
				},
			},
			isSuccess: true,
		},
		"select both StatefulSet & Deployment + spec.replicas as int + count 2": {
			condition: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind":          "Deployment",
								"spec.replicas": "3",
							},
						},
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind":          "StatefulSet",
								"spec.replicas": "3",
							},
						},
					},
				},
				ResourceOperator: types.ResourceOperatorEqualsCount,
				Count:            ptr.Int(2), // 1 Deployment + 1 StatefulSet
			},
			resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "StatefulSet",
						"spec": map[string]interface{}{
							"replicas": 3,
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Deployment",
						"spec": map[string]interface{}{
							"replicas": 3,
						},
					},
				},
			},
			isErr: true, // bug in metac refer https://github.com/AmitKumarDas/metac/issues/112
		},
		"select both StatefulSet & Deployment + count 2": {
			condition: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind":          "Deployment",
								"spec.replicas": "3",
							},
						},
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind":          "StatefulSet",
								"spec.replicas": "3",
							},
						},
					},
				},
				ResourceOperator: types.ResourceOperatorEqualsCount,
				Count:            ptr.Int(2), // 1 Deployment + 1 StatefulSet
			},
			resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "StatefulSet",
						"spec": map[string]interface{}{
							"replicas": "3",
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Deployment",
						"spec": map[string]interface{}{
							"replicas": "3",
						},
					},
				},
			},
			isSuccess: true,
		},
		"select either StatefulSet or Deployment + gte 2": {
			condition: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind":          "Deployment",
								"spec.replicas": "3",
							},
						},
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind":          "StatefulSet",
								"spec.replicas": "3",
							},
						},
					},
				},
				ResourceOperator: types.ResourceOperatorGTE,
				Count:            ptr.Int(2), // 1 Deployment + 1 StatefulSet
			},
			resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "StatefulSet",
						"spec": map[string]interface{}{
							"replicas": "3",
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Deployment",
						"spec": map[string]interface{}{
							"replicas": "3",
						},
					},
				},
			},
			isSuccess: true,
		},
		"select either StatefulSet or Deployment + lte 2": {
			condition: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind":          "Deployment",
								"spec.replicas": "3",
							},
						},
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind":          "StatefulSet",
								"spec.replicas": "3",
							},
						},
					},
				},
				ResourceOperator: types.ResourceOperatorLTE,
				Count:            ptr.Int(2), // 1 Deployment + 1 StatefulSet
			},
			resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "StatefulSet",
						"spec": map[string]interface{}{
							"replicas": "3",
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Deployment",
						"spec": map[string]interface{}{
							"replicas": "3",
						},
					},
				},
			},
			isSuccess: true,
		},
		"select either StatefulSet or Deployment + lte 3": {
			condition: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind":          "Deployment",
								"spec.replicas": "3",
							},
						},
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind":          "StatefulSet",
								"spec.replicas": "3",
							},
						},
					},
				},
				ResourceOperator: types.ResourceOperatorLTE,
				Count:            ptr.Int(3), // (2 <= 3) == true
			},
			resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "StatefulSet",
						"spec": map[string]interface{}{
							"replicas": "3",
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Deployment",
						"spec": map[string]interface{}{
							"replicas": "3",
						},
					},
				},
			},
			isSuccess: true,
		},
		"can't select + 1 StatefulSet & 1 Deployment + gte 3": {
			condition: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind":          "Deployment",
								"spec.replicas": "many",
							},
						},
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind":          "StatefulSet",
								"spec.replicas": "many",
							},
						},
					},
				},
				ResourceOperator: types.ResourceOperatorGTE,
				Count:            ptr.Int(3), // can't match since (2 >= 3) == false
			},
			resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "StatefulSet",
						"spec": map[string]interface{}{
							"replicas": "many",
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Deployment",
						"spec": map[string]interface{}{
							"replicas": "many",
						},
					},
				},
			},
			isSuccess: false,
		},
		"can't select + 1 StatefulSet & 1 Deployment + lte 1": {
			condition: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind":          "Deployment",
								"spec.replicas": "many",
							},
						},
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind":          "StatefulSet",
								"spec.replicas": "many",
							},
						},
					},
				},
				ResourceOperator: types.ResourceOperatorLTE,
				Count:            ptr.Int(1), // can't match since (2 <= 1) == false
			},
			resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "StatefulSet",
						"spec": map[string]interface{}{
							"replicas": "many",
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Deployment",
						"spec": map[string]interface{}{
							"replicas": "many",
						},
					},
				},
			},
			isSuccess: false,
		},
		"successful assert + 1 StatefulSet & 1 Deployment + not exists": {
			condition: types.IfCondition{
				ResourceSelector: v1alpha1.ResourceSelector{
					SelectorTerms: []*v1alpha1.SelectorTerm{
						&v1alpha1.SelectorTerm{
							MatchFields: map[string]string{
								"kind":          "Pod",
								"spec.replicas": "many",
							},
						},
					},
				},
				ResourceOperator: types.ResourceOperatorNotExist,
			},
			resources: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "StatefulSet",
						"spec": map[string]interface{}{
							"replicas": "many",
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Deployment",
						"spec": map[string]interface{}{
							"replicas": "many",
						},
					},
				},
			},
			isSuccess: true,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			rc := NewResourceListCondition(
				mock.condition,
				mock.resources,
			)
			got, err := rc.IsSuccess()
			if mock.isErr && err == nil {
				t.Fatalf("Expected error got none")
			}
			if !mock.isErr && err != nil {
				t.Fatalf(
					"Expected no error got [%+v]",
					err,
				)
			}
			if mock.isErr {
				return
			}
			if mock.isSuccess != got {
				t.Fatalf(
					"Expected success %t got %t",
					mock.isSuccess,
					got,
				)
			}
		})
	}
}

func TestExecuteAssert(t *testing.T) {
	var tests = map[string]struct {
		request   AssertRequest
		isSuccess bool
		isErr     bool
	}{
		"successful match + OR operator + 1 condition + matchlabels": {
			request: AssertRequest{
				Assert: &types.Assert{
					If: types.If{
						IfConditions: []types.IfCondition{
							types.IfCondition{
								ResourceSelector: v1alpha1.ResourceSelector{
									SelectorTerms: []*v1alpha1.SelectorTerm{
										&v1alpha1.SelectorTerm{
											MatchLabels: map[string]string{
												"app": "d-operator",
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
							"metadata": map[string]interface{}{
								"labels": map[string]interface{}{
									"app": "d-operator",
								},
							},
						},
					},
				},
			},
			isSuccess: true,
		},
		"successful match + OR operator + 2 conditions + matchlabels": {
			request: AssertRequest{
				Assert: &types.Assert{
					If: types.If{
						IfConditions: []types.IfCondition{
							types.IfCondition{
								ResourceSelector: v1alpha1.ResourceSelector{
									SelectorTerms: []*v1alpha1.SelectorTerm{
										&v1alpha1.SelectorTerm{
											MatchLabels: map[string]string{
												"app": "d-junk",
											},
										},
									},
								},
							},
							types.IfCondition{
								ResourceSelector: v1alpha1.ResourceSelector{
									SelectorTerms: []*v1alpha1.SelectorTerm{
										&v1alpha1.SelectorTerm{
											MatchLabels: map[string]string{
												"app": "d-operator",
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
							"metadata": map[string]interface{}{
								"labels": map[string]interface{}{
									"app": "d-operator",
								},
							},
						},
					},
				},
			},
			isSuccess: true,
		},
		"successful match + AND operator + 1 condition + matchlabels": {
			request: AssertRequest{
				Assert: &types.Assert{
					If: types.If{
						IfConditions: []types.IfCondition{
							types.IfCondition{
								ResourceSelector: v1alpha1.ResourceSelector{
									SelectorTerms: []*v1alpha1.SelectorTerm{
										&v1alpha1.SelectorTerm{
											MatchLabels: map[string]string{
												"app": "d-operator",
											},
										},
									},
								},
							},
						},
						IfOperator: types.IfOperatorAND,
					},
				},
				Resources: []*unstructured.Unstructured{
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"metadata": map[string]interface{}{
								"labels": map[string]interface{}{
									"app": "d-operator",
								},
							},
						},
					},
				},
			},
			isSuccess: true,
		},
		"successful match + AND operator + 2 conditions + matchlabels": {
			request: AssertRequest{
				Assert: &types.Assert{
					If: types.If{
						IfConditions: []types.IfCondition{
							types.IfCondition{
								ResourceSelector: v1alpha1.ResourceSelector{
									SelectorTerms: []*v1alpha1.SelectorTerm{
										&v1alpha1.SelectorTerm{
											MatchLabels: map[string]string{
												"app": "d-operator",
											},
										},
									},
								},
							},
							types.IfCondition{
								ResourceSelector: v1alpha1.ResourceSelector{
									SelectorTerms: []*v1alpha1.SelectorTerm{
										&v1alpha1.SelectorTerm{
											MatchLabels: map[string]string{
												"name": "test",
											},
										},
									},
								},
							},
						},
						IfOperator: types.IfOperatorAND,
					},
				},
				Resources: []*unstructured.Unstructured{
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"metadata": map[string]interface{}{
								"labels": map[string]interface{}{
									"app":  "d-operator",
									"name": "test",
								},
							},
						},
					},
				},
			},
			isSuccess: true,
		},
		"failed match + OR operator + 1 condition + matchlabels": {
			request: AssertRequest{
				Assert: &types.Assert{
					If: types.If{
						IfConditions: []types.IfCondition{
							types.IfCondition{
								ResourceSelector: v1alpha1.ResourceSelector{
									SelectorTerms: []*v1alpha1.SelectorTerm{
										&v1alpha1.SelectorTerm{
											MatchLabels: map[string]string{
												"app": "k8s",
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
							"metadata": map[string]interface{}{
								"labels": map[string]interface{}{
									"app": "d-operator",
								},
							},
						},
					},
				},
			},
			isSuccess: false,
		},
		"failed match + AND operator + 2 conditions + matchlabels": {
			request: AssertRequest{
				Assert: &types.Assert{
					If: types.If{
						IfConditions: []types.IfCondition{
							types.IfCondition{
								ResourceSelector: v1alpha1.ResourceSelector{
									SelectorTerms: []*v1alpha1.SelectorTerm{
										&v1alpha1.SelectorTerm{
											MatchLabels: map[string]string{
												"app": "k8s",
											},
										},
									},
								},
							},
							types.IfCondition{
								ResourceSelector: v1alpha1.ResourceSelector{
									SelectorTerms: []*v1alpha1.SelectorTerm{
										&v1alpha1.SelectorTerm{
											MatchLabels: map[string]string{
												"app": "d-operator",
											},
										},
									},
								},
							},
						},
						IfOperator: types.IfOperatorAND,
					},
				},
				Resources: []*unstructured.Unstructured{
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"metadata": map[string]interface{}{
								"labels": map[string]interface{}{
									"app": "d-operator",
								},
							},
						},
					},
				},
			},
			isSuccess: false,
		},
		"failed match + AND operator + 1 condition + matchlabels": {
			request: AssertRequest{
				Assert: &types.Assert{
					If: types.If{
						IfConditions: []types.IfCondition{
							types.IfCondition{
								ResourceSelector: v1alpha1.ResourceSelector{
									SelectorTerms: []*v1alpha1.SelectorTerm{
										&v1alpha1.SelectorTerm{
											MatchLabels: map[string]string{
												"app": "k8s",
											},
										},
									},
								},
							},
						},
						IfOperator: types.IfOperatorAND,
					},
				},
				Resources: []*unstructured.Unstructured{
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"metadata": map[string]interface{}{
								"labels": map[string]interface{}{
									"app": "d-operator",
								},
							},
						},
					},
				},
			},
			isSuccess: false,
		},
		"successful match + OR operator + 2 condition + 2 resources": {
			request: AssertRequest{
				Assert: &types.Assert{
					If: types.If{
						IfConditions: []types.IfCondition{
							types.IfCondition{
								ResourceSelector: v1alpha1.ResourceSelector{
									SelectorTerms: []*v1alpha1.SelectorTerm{
										&v1alpha1.SelectorTerm{
											MatchLabels: map[string]string{
												"app": "k8s-1",
											},
										},
									},
								},
							},
							types.IfCondition{
								ResourceSelector: v1alpha1.ResourceSelector{
									SelectorTerms: []*v1alpha1.SelectorTerm{
										&v1alpha1.SelectorTerm{
											MatchLabels: map[string]string{
												"app": "k8s-2",
											},
										},
									},
								},
							},
						},
						IfOperator: types.IfOperatorOR,
					},
				},
				Resources: []*unstructured.Unstructured{
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"metadata": map[string]interface{}{
								"labels": map[string]interface{}{
									"app": "k8s-1",
								},
							},
						},
					},
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"metadata": map[string]interface{}{
								"labels": map[string]interface{}{
									"app": "k8s-2",
								},
							},
						},
					},
				},
			},
			isSuccess: true,
		},
		"successful match + AND operator + 2 condition + 2 resources": {
			request: AssertRequest{
				Assert: &types.Assert{
					If: types.If{
						IfConditions: []types.IfCondition{
							types.IfCondition{
								ResourceSelector: v1alpha1.ResourceSelector{
									SelectorTerms: []*v1alpha1.SelectorTerm{
										&v1alpha1.SelectorTerm{
											MatchLabels: map[string]string{
												"app": "k8s-1",
											},
										},
									},
								},
							},
							types.IfCondition{
								ResourceSelector: v1alpha1.ResourceSelector{
									SelectorTerms: []*v1alpha1.SelectorTerm{
										&v1alpha1.SelectorTerm{
											MatchLabels: map[string]string{
												"app": "k8s-2",
											},
										},
									},
								},
							},
						},
						IfOperator: types.IfOperatorAND,
					},
				},
				Resources: []*unstructured.Unstructured{
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"metadata": map[string]interface{}{
								"labels": map[string]interface{}{
									"app": "k8s-1",
								},
							},
						},
					},
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"metadata": map[string]interface{}{
								"labels": map[string]interface{}{
									"app": "k8s-2",
								},
							},
						},
					},
				},
			},
			isSuccess: true,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			got, err := ExecuteAssert(mock.request)
			if mock.isErr && err == nil {
				t.Fatalf("Expected error got none")
			}
			if !mock.isErr && err != nil {
				t.Fatalf(
					"Expected no error got [%+v]",
					err,
				)
			}
			if got != mock.isSuccess {
				t.Fatalf(
					"Expected success %t got %t",
					mock.isSuccess,
					got,
				)
			}
		})
	}
}
