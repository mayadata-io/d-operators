/*
Copyright 2020 The MayaData Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package unstruct

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestListIdentifiesAll(t *testing.T) {
	var tests = map[string]struct {
		src          []*unstructured.Unstructured
		target       []*unstructured.Unstructured
		isIdentitied bool
	}{
		"no src + no target": {
			isIdentitied: true,
		},
		"1 src + no target": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
					},
				},
			},
			isIdentitied: false,
		},
		"2 srcs + no target": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Deployment",
					},
				},
			},
			isIdentitied: false,
		},
		"no src + 1 target": {
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
					},
				},
			},
			isIdentitied: false,
		},
		"no src + 2 targets": {
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Deployment",
					},
				},
			},
			isIdentitied: false,
		},
		"1 src + 1 target + src != target + kind": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
					},
				},
			},
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Service",
					},
				},
			},
			isIdentitied: false,
		},
		"1 src + 1 target + src == target + kind": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
					},
				},
			},
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
					},
				},
			},
			isIdentitied: true,
		},
		"1 src + 1 target + src != target + apiVersion": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
					},
				},
			},
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v2",
					},
				},
			},
			isIdentitied: false,
		},
		"1 src + 1 target + src == target + apiVersion": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
					},
				},
			},
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
					},
				},
			},
			isIdentitied: true,
		},
		"1 src + 1 target + src != target + name": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"name": "my-pod",
						},
					},
				},
			},
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"metadata": map[string]interface{}{
								"name": "unknown-pod",
							},
						},
					},
				},
			},
			isIdentitied: false,
		},
		"1 src + 1 target + src == target + name": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"name": "my-pod",
						},
					},
				},
			},
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"name": "my-pod",
						},
					},
				},
			},
			isIdentitied: true,
		},
		"1 src + 1 target + src != target + namespace": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"namespace": "my-default",
						},
					},
				},
			},
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"namespace": "my-non-default",
						},
					},
				},
			},
			isIdentitied: false,
		},
		"1 src + 1 target + src == target + namespace": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"namespace": "my-default",
						},
					},
				},
			},
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"namespace": "my-default",
						},
					},
				},
			},
			isIdentitied: true,
		},
		"1 src + 1 target + src != target + uid": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"uid": "pod-101",
						},
					},
				},
			},
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"uid": "non-pod-101",
						},
					},
				},
			},
			isIdentitied: false,
		},
		"1 src + 1 target + src == target + uid": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"uid": "pod-101",
						},
					},
				},
			},
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"uid": "pod-101",
						},
					},
				},
			},
			isIdentitied: true,
		},
		"2 srcs + 1 target": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"uid": "pod-101",
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Deployment",
						"metadata": map[string]interface{}{
							"uid": "deployment-101",
						},
					},
				},
			},
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"uid": "pod-101",
						},
					},
				},
			},
			isIdentitied: false,
		},
		"2 srcs + 2 target + src == target": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"uid": "pod-101",
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Deployment",
						"metadata": map[string]interface{}{
							"uid": "deployment-101",
						},
					},
				},
			},
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"uid": "pod-101",
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Deployment",
						"metadata": map[string]interface{}{
							"uid": "deployment-101",
						},
					},
				},
			},
			isIdentitied: true,
		},
		"2 srcs + 2 target + src != target": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"uid": "pod-101",
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Deployment",
						"metadata": map[string]interface{}{
							"uid": "deployment-101",
						},
					},
				},
			},
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"uid": "pod-101",
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"uid": "pod-201",
						},
					},
				},
			},
			isIdentitied: false,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			l := List(mock.src)
			got := l.IdentifiesAll(mock.target)
			if got != mock.isIdentitied {
				t.Fatalf(
					"Expected equals %t got %t: Diff \n%s",
					mock.isIdentitied,
					got,
					cmp.Diff(
						mock.src,
						mock.target,
					),
				)
			}
		})
	}
}

func TestListEqualsAll(t *testing.T) {
	var tests = map[string]struct {
		src     []*unstructured.Unstructured
		target  []*unstructured.Unstructured
		isEqual bool
	}{
		"no src + no target": {
			isEqual: true,
		},
		"1 src + no target": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
					},
				},
			},
			isEqual: false,
		},
		"2 srcs + no target": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Deployment",
					},
				},
			},
			isEqual: false,
		},
		"no src + 1 target": {
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
					},
				},
			},
			isEqual: false,
		},
		"no src + 2 targets": {
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Deployment",
					},
				},
			},
			isEqual: false,
		},
		"1 src + 1 target + src != target": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
					},
				},
			},
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Service",
					},
				},
			},
			isEqual: false,
		},
		"1 src + 1 target + src == target + kind": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
					},
				},
			},
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
					},
				},
			},
			isEqual: true,
		},
		"1 src + 1 target + src != target + labels": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app": "cool",
							},
						},
					},
				},
			},
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app": "ice",
							},
						},
					},
				},
			},
			isEqual: false,
		},
		"1 src + 1 target + src == target + labels": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app": "cool",
							},
						},
					},
				},
			},
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app": "cool",
							},
						},
					},
				},
			},
			isEqual: true,
		},
		"1 src + 1 target + src != target + annotations": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app": "cool",
							},
							"annotations": map[string]interface{}{
								"app": "cool",
							},
						},
					},
				},
			},
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app": "cool",
							},
							"annotations": map[string]interface{}{
								"app": "ice",
							},
						},
					},
				},
			},
			isEqual: false,
		},
		"1 src + 1 target + src == target + annotations": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app": "cool",
							},
							"annotations": map[string]interface{}{
								"app": "cool",
							},
						},
					},
				},
			},
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app": "cool",
							},
							"annotations": map[string]interface{}{
								"app": "cool",
							},
						},
					},
				},
			},
			isEqual: true,
		},
		"1 src + 1 target + src != target + finalizers": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app": "cool",
							},
							"annotations": map[string]interface{}{
								"app": "cool",
							},
							"finalizers": []interface{}{
								"protect-1",
								"protect-2",
							},
						},
					},
				},
			},
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app": "cool",
							},
							"annotations": map[string]interface{}{
								"app": "cool",
							},
							"finalizers": []interface{}{
								"protect-1",
								"protect-unknown",
							},
						},
					},
				},
			},
			isEqual: false,
		},
		"1 src + 1 target + src == target + finalizers": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app": "cool",
							},
							"annotations": map[string]interface{}{
								"app": "cool",
							},
							"finalizers": []interface{}{
								"protect-1",
								"protect-2",
							},
						},
					},
				},
			},
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app": "cool",
							},
							"annotations": map[string]interface{}{
								"app": "cool",
							},
							"finalizers": []interface{}{
								"protect-1",
								"protect-2",
							},
						},
					},
				},
			},
			isEqual: true,
		},

		"1 src + 1 target + src != target + ownerReferences": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app": "cool",
							},
							"annotations": map[string]interface{}{
								"app": "cool",
							},
							"finalizers": []interface{}{
								"protect-1",
								"protect-2",
							},
							"ownerReferences": []interface{}{
								map[string]interface{}{
									"apiVersion":         "apps/v1",
									"controller":         true,
									"blockOwnerDeletion": true,
									"kind":               "ReplicaSet",
									"name":               "my-repset",
									"uid":                "d9607e19",
								},
							},
						},
					},
				},
			},
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app": "cool",
							},
							"annotations": map[string]interface{}{
								"app": "cool",
							},
							"finalizers": []interface{}{
								"protect-1",
								"protect-unknown",
							},
							"ownerReferences": []interface{}{
								map[string]interface{}{
									"apiVersion":         "apps/v1",
									"controller":         true,
									"blockOwnerDeletion": true,
								},
							},
						},
					},
				},
			},
			isEqual: false,
		},
		"1 src + 1 target + src == target + ownerReferences": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app": "cool",
							},
							"annotations": map[string]interface{}{
								"app": "cool",
							},
							"finalizers": []interface{}{
								"protect-1",
								"protect-2",
							},
							"ownerReferences": []interface{}{
								map[string]interface{}{
									"apiVersion":         "apps/v1",
									"controller":         true,
									"blockOwnerDeletion": true,
									"kind":               "ReplicaSet",
									"name":               "my-repset",
									"uid":                "d9607e19",
								},
							},
						},
					},
				},
			},
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app": "cool",
							},
							"annotations": map[string]interface{}{
								"app": "cool",
							},
							"finalizers": []interface{}{
								"protect-1",
								"protect-2",
							},
							"ownerReferences": []interface{}{
								map[string]interface{}{
									"apiVersion":         "apps/v1",
									"controller":         true,
									"blockOwnerDeletion": true,
									"kind":               "ReplicaSet",
									"name":               "my-repset",
									"uid":                "d9607e19",
								},
							},
						},
					},
				},
			},
			isEqual: true,
		},
		"1 src + 1 target + src != target + spec.replicas": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":     "Pod",
						"metadata": map[string]interface{}{},
						"spec": map[string]interface{}{
							"replicas": int64(3),
						},
					},
				},
			},
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":     "Pod",
						"metadata": map[string]interface{}{},
						"spec": map[string]interface{}{
							"replicas": int64(0),
						},
					},
				},
			},
			isEqual: false,
		},
		"1 src + 1 target + src == target + spec.replicas": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":     "Pod",
						"metadata": map[string]interface{}{},
						"spec": map[string]interface{}{
							"replicas": int64(3),
						},
					},
				},
			},
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":     "Pod",
						"metadata": map[string]interface{}{},
						"spec": map[string]interface{}{
							"replicas": int64(3),
						},
					},
				},
			},
			isEqual: true,
		},
		"1 src + 1 target + src != target + status.phase": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":     "Pod",
						"metadata": map[string]interface{}{},
						"spec": map[string]interface{}{
							"replicas": int64(3),
						},
						"status": map[string]interface{}{
							"phase": "Online",
						},
					},
				},
			},
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":     "Pod",
						"metadata": map[string]interface{}{},
						"spec": map[string]interface{}{
							"replicas": int64(3),
						},
						"status": map[string]interface{}{
							"phase": "Offline",
						},
					},
				},
			},
			isEqual: false,
		},
		"1 src + 1 target + src == target + status.phase": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":     "Pod",
						"metadata": map[string]interface{}{},
						"spec": map[string]interface{}{
							"replicas": int64(3),
						},
						"status": map[string]interface{}{
							"phase": "Online",
						},
					},
				},
			},
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":     "Pod",
						"metadata": map[string]interface{}{},
						"spec": map[string]interface{}{
							"replicas": int64(3),
						},
						"status": map[string]interface{}{
							"phase": "Online",
						},
					},
				},
			},
			isEqual: true,
		},
		"1 src + 1 target + src != target + spec.items": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":     "Pod",
						"metadata": map[string]interface{}{},
						"spec": map[string]interface{}{
							"items": []interface{}{
								int64(1),
								int64(2),
							},
						},
					},
				},
			},
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":     "Pod",
						"metadata": map[string]interface{}{},
						"spec": map[string]interface{}{
							"items": []interface{}{
								int64(1),
							},
						},
					},
				},
			},
			isEqual: false,
		},
		"1 src + 1 target + src == target + spec.items": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":     "Pod",
						"metadata": map[string]interface{}{},
						"spec": map[string]interface{}{
							"items": []interface{}{
								int64(1),
								int64(2),
							},
						},
					},
				},
			},
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":     "Pod",
						"metadata": map[string]interface{}{},
						"spec": map[string]interface{}{
							"items": []interface{}{
								int64(1),
								int64(2),
							},
						},
					},
				},
			},
			isEqual: true,
		},
		"2 srcs + 1 target + spec.items": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":     "Pod",
						"metadata": map[string]interface{}{},
						"spec": map[string]interface{}{
							"items": []interface{}{
								int64(1),
								int64(2),
							},
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":     "Pod",
						"metadata": map[string]interface{}{},
						"spec": map[string]interface{}{
							"items": []interface{}{
								int64(1),
								int64(2),
							},
						},
					},
				},
			},
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":     "Pod",
						"metadata": map[string]interface{}{},
						"spec": map[string]interface{}{
							"items": []interface{}{
								int64(1),
								int64(2),
							},
						},
					},
				},
			},
			isEqual: false,
		},
		"2 srcs + 2 target + src!=target + spec.items": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":     "Pod",
						"metadata": map[string]interface{}{},
						"spec": map[string]interface{}{
							"items": []interface{}{
								int64(1),
								int64(2),
							},
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":     "Deployment",
						"metadata": map[string]interface{}{},
						"spec": map[string]interface{}{
							"items": []interface{}{
								int64(1),
								int64(2),
								int64(3),
							},
						},
					},
				},
			},
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":     "Pod",
						"metadata": map[string]interface{}{},
						"spec": map[string]interface{}{
							"items": []interface{}{
								int64(1),
								int64(2),
							},
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":     "Deployment",
						"metadata": map[string]interface{}{},
						"spec": map[string]interface{}{
							"items": []interface{}{
								int64(1),
							},
						},
					},
				},
			},
			isEqual: false,
		},
		"2 srcs + 2 target + src==target + spec.items": {
			src: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":     "Pod",
						"metadata": map[string]interface{}{},
						"spec": map[string]interface{}{
							"items": []interface{}{
								int64(1),
								int64(2),
							},
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":     "Deployment",
						"metadata": map[string]interface{}{},
						"spec": map[string]interface{}{
							"items": []interface{}{
								int64(1),
								int64(2),
							},
						},
					},
				},
			},
			target: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":     "Pod",
						"metadata": map[string]interface{}{},
						"spec": map[string]interface{}{
							"items": []interface{}{
								int64(1),
								int64(2),
							},
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":     "Deployment",
						"metadata": map[string]interface{}{},
						"spec": map[string]interface{}{
							"items": []interface{}{
								int64(1),
								int64(2),
							},
						},
					},
				},
			},
			isEqual: true,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			l := List(mock.src)
			got := l.EqualsAll(mock.target)
			if got != mock.isEqual {
				t.Fatalf(
					"Expected equals %t got %t: Diff \n%s",
					mock.isEqual,
					got,
					cmp.Diff(
						mock.src,
						mock.target,
					),
				)
			}
		})
	}
}
