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
	"k8s.io/apimachinery/pkg/util/yaml"
	ptr "mayadata.io/d-operators/common/pointer"
	"mayadata.io/d-operators/common/unstruct"
)

func TestInit(t *testing.T) {
	var tests = map[string]struct {
		config         CreateOrDeleteRequest
		expectReplicas int
	}{
		"empty config": {
			config:         CreateOrDeleteRequest{},
			expectReplicas: 1,
		},
		"nil action": {
			config: CreateOrDeleteRequest{
				Replicas: nil,
			},
			expectReplicas: 1,
		},
		"nil replicas": {
			config: CreateOrDeleteRequest{
				Replicas: nil,
			},
			expectReplicas: 1,
		},
		"0 replicas": {
			config: CreateOrDeleteRequest{
				Replicas: ptr.Int(0),
			},
			expectReplicas: 0,
		},
		"1 replicas": {
			config: CreateOrDeleteRequest{
				Replicas: ptr.Int(1),
			},
			expectReplicas: 1,
		},
		"2 replicas": {
			config: CreateOrDeleteRequest{
				Replicas: ptr.Int(2),
			},
			expectReplicas: 2,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			b := CreateOrDeleteBuilder{
				Request: mock.config,
			}
			b.init()
			if mock.expectReplicas != b.replicas {
				t.Fatalf(
					"Expected replicas %d got %d",
					mock.expectReplicas,
					b.replicas,
				)
			}
		})
	}
}

func TestTrySetDeleteFlag(t *testing.T) {
	var tests = map[string]struct {
		template *unstructured.Unstructured
		replicas int
		isDelete bool
	}{
		"0 replicas": {
			replicas: 0,
			isDelete: true,
		},
		"1 replicas + empty template": {
			template: &unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
			replicas: 1,
			isDelete: false,
		},
		"1 replicas + template with spec not found": {
			template: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Test",
				},
			},
			replicas: 1,
			isDelete: false,
		},
		"1 replicas + template with spec != nil": {
			template: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Test",
					"spec": map[string]interface{}{},
				},
			},
			replicas: 1,
			isDelete: false,
		},
		"1 replicas + template with spec == nil": {
			template: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Test",
					"spec": nil,
				},
			},
			replicas: 1,
			isDelete: true,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			b := CreateOrDeleteBuilder{
				desiredTemplate: mock.template,
				replicas:        mock.replicas,
			}
			b.isDeleteAction()
			if mock.isDelete != b.isDelete {
				t.Fatalf(
					"Expected delete %t got %t",
					mock.isDelete,
					b.isDelete,
				)
			}
		})
	}
}

func TestTrySetDeleteFlagJson(t *testing.T) {
	var tests = map[string]struct {
		template string
		replicas int
		isDelete bool
	}{
		"1 replicas + template with spec not found": {
			template: `{
				"kind": "Pod"
			}`,
			replicas: 1,
			isDelete: false,
		},
		"1 replicas + template with spec != nil": {
			template: `{
				"kind": "Pod",
				"spec": {
					"hi": "hello"
				}
			}`,
			replicas: 1,
			isDelete: false,
		},
		"1 replicas + template with spec = null": {
			template: `{
				"kind": "Pod",
				"spec": null
			}`,
			replicas: 1,
			isDelete: true,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			obj := unstructured.Unstructured{}
			err := obj.UnmarshalJSON([]byte(mock.template))
			if err != nil {
				t.Fatalf("Can't unmarshal [%+v]", err)
			}
			b := CreateOrDeleteBuilder{
				desiredTemplate: &obj,
				replicas:        mock.replicas,
			}
			b.isDeleteAction()
			if mock.isDelete != b.isDelete {
				t.Fatalf(
					"Expected delete %t got %t",
					mock.isDelete,
					b.isDelete,
				)
			}
		})
	}
}

func TestTrySetDeleteFlagYAML(t *testing.T) {
	var tests = map[string]struct {
		template string
		replicas int
		isDelete bool
	}{
		"1 replicas + template with spec not found": {
			template: `
kind: Pod
`,
			replicas: 1,
			isDelete: false,
		},
		"1 replicas + template with spec != nil": {
			template: `
kind: Pod
spec: 
  hi: hello	
`,
			replicas: 1,
			isDelete: false,
		},
		"1 replicas + template with spec == empty": {
			template: `
kind: Pod
spec: ""
`,
			replicas: 1,
			isDelete: false,
		},
		"1 replicas + template with spec == null": {
			template: `
kind: Pod
spec: null
`,
			replicas: 1,
			isDelete: true,
		},
		"1 replicas + template with spec == ": {
			template: `
kind: Pod
spec:
`,
			replicas: 1,
			isDelete: true,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			// step 1 - yaml to json
			raw, err := yaml.ToJSON([]byte(mock.template))
			if err != nil {
				t.Fatalf("Can't convert to json [%+v]", err)
			}
			// step 2 - json to unstructured
			obj := unstructured.Unstructured{}
			err = obj.UnmarshalJSON(raw)
			if err != nil {
				t.Fatalf("Can't unmarshal [%+v]", err)
			}
			b := CreateOrDeleteBuilder{
				desiredTemplate: &obj,
				replicas:        mock.replicas,
			}
			b.isDeleteAction()
			if mock.isDelete != b.isDelete {
				t.Fatalf(
					"Expected delete %t got %t",
					mock.isDelete,
					b.isDelete,
				)
			}
		})
	}
}

func TestSetDesiredTemplate(t *testing.T) {
	var tests = map[string]struct {
		config CreateOrDeleteRequest
		isErr  bool
	}{
		"nil desired state": {
			config: CreateOrDeleteRequest{},
			isErr:  true,
		},
		"empty desired state": {
			config: CreateOrDeleteRequest{
				Apply: map[string]interface{}{},
			},
			isErr: true,
		},
		"missing apiVersion": {
			config: CreateOrDeleteRequest{
				Apply: map[string]interface{}{
					"kind": "Pod",
				},
			},
			isErr: true,
		},
		"missing kind": {
			config: CreateOrDeleteRequest{
				Apply: map[string]interface{}{
					"apiVersion": "v1",
				},
			},
			isErr: true,
		},
		"with kind & apiVersion": {
			config: CreateOrDeleteRequest{
				Apply: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Pod",
				},
			},
			isErr: false,
		},
		"desired state as a custom resource": {
			config: CreateOrDeleteRequest{
				Apply: map[string]interface{}{
					"apiVersion": "dao.mayadata.io/v1alpha1",
					"kind":       "MyCustom",
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"app": "custom",
						},
						"annotations": map[string]interface{}{
							"app": "custom",
						},
					},
					"spec": map[string]interface{}{
						"replicas": 3,
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
			b := CreateOrDeleteBuilder{
				Request: mock.config,
			}
			b.evalDesiredTemplate()
			if mock.isErr && b.err == nil {
				t.Fatalf("Expected error got none")
			}
			if !mock.isErr && b.err != nil {
				t.Fatalf("Expected no error got [%+v]", b.err)
			}
		})
	}
}

func TestBuildCreateOrDeleteStates(t *testing.T) {
	var tests = map[string]struct {
		config        CreateOrDeleteRequest
		expectStates  []*unstructured.Unstructured
		expectDeletes []*unstructured.Unstructured
		isErr         bool
	}{
		"empty config": {
			config: CreateOrDeleteRequest{},
			isErr:  true,
		},
		"run + no desired state + nil action": {
			config: CreateOrDeleteRequest{
				TaskKey: "test.101",
				Run: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "test",
							"namespace": "dope",
						},
					},
				},
				Watch: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "test",
							"namespace": "dope",
						},
					},
				},
			},
			isErr: true,
		},
		"run + desired state + nil action": {
			config: CreateOrDeleteRequest{
				TaskKey: "test.101",
				Run: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Deployment",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "test",
							"namespace": "dope",
						},
					},
				},
				Watch: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Deployment",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "test",
							"namespace": "dope",
						},
					},
				},
				Apply: map[string]interface{}{
					"kind":       "Pod",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name":      "test",
						"namespace": "dope",
					},
				},
			},
			expectStates: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "test",
							"namespace": "dope",
						},
					},
				},
			},
			isErr: false,
		},
		"run + desired state + empty action": {
			config: CreateOrDeleteRequest{
				TaskKey: "test.101",
				Run: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Deployment",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "test",
							"namespace": "dope",
						},
					},
				},
				Watch: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Deployment",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "test",
							"namespace": "dope",
						},
					},
				},
				Replicas: nil,
				Apply: map[string]interface{}{
					"kind":       "Pod",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name":      "test",
						"namespace": "dope",
					},
				},
			},
			expectStates: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "test",
							"namespace": "dope",
						},
					},
				},
			},
			isErr: false,
		},
		"run + desired state + 0 replicas": {
			config: CreateOrDeleteRequest{
				TaskKey: "test.101",
				Run: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Deployment",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "test",
							"namespace": "dope",
						},
					},
				},
				Watch: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Deployment",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "test",
							"namespace": "dope",
						},
					},
				},
				Replicas: ptr.Int(0),
				Apply: map[string]interface{}{
					"kind":       "Pod",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name":      "test",
						"namespace": "dope",
					},
				},
			},
			expectStates: nil,
			isErr:        false,
		},
		"run + desired state + 1 replicas": {
			config: CreateOrDeleteRequest{
				TaskKey: "test.101",
				Run: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Deployment",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "test",
							"namespace": "dope",
						},
					},
				},
				Watch: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Deployment",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "test",
							"namespace": "dope",
						},
					},
				},
				Replicas: ptr.Int(1),
				Apply: map[string]interface{}{
					"kind":       "Pod",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name":      "test",
						"namespace": "dope",
					},
				},
			},
			expectStates: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "test",
							"namespace": "dope",
						},
					},
				},
			},
			isErr: false,
		},
		"run + desired state + 2 replicas": {
			config: CreateOrDeleteRequest{
				TaskKey: "test.101",
				Run: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Deployment",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "test",
							"namespace": "dope",
						},
					},
				},
				Watch: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Deployment",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "test",
							"namespace": "dope",
						},
					},
				},
				Replicas: ptr.Int(2),
				Apply: map[string]interface{}{
					"kind":       "Pod",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name":      "test",
						"namespace": "dope",
					},
				},
			},
			expectStates: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "test-0",
							"namespace": "dope",
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "test-1",
							"namespace": "dope",
						},
					},
				},
			},
			isErr: false,
		},
		"run + multiple explicit deletes": {
			config: CreateOrDeleteRequest{
				TaskKey: "test.101",
				Run: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Deployment",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "test",
							"namespace": "dope",
							"uid":       "test-101",
						},
					},
				},
				Watch: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Deployment",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "test",
							"namespace": "dope",
							"uid":       "test-101",
						},
					},
				},
				ObservedResources: []*unstructured.Unstructured{
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name":      "test-0",
								"namespace": "dope",
							},
						},
					},
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name":      "test-1",
								"namespace": "dope",
							},
						},
					},
				},
				Replicas: ptr.Int(2), // replicas really dont matter
				Apply: map[string]interface{}{
					"kind":       "Pod",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name":      "test",
						"namespace": "dope",
					},
					"spec": nil, // i.e. delete this resource
				},
			},
			expectDeletes: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "test-0",
							"namespace": "dope",
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "test-1",
							"namespace": "dope",
						},
					},
				},
			},
			isErr: false,
		},
		"run + single explicit delete": {
			config: CreateOrDeleteRequest{
				TaskKey: "test.101",
				Run: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Deployment",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "test",
							"namespace": "dope",
							"uid":       "test-101",
						},
					},
				},
				Watch: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Deployment",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "test",
							"namespace": "dope",
							"uid":       "test-101",
						},
					},
				},
				ObservedResources: []*unstructured.Unstructured{
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name":      "test-0",
								"namespace": "dope",
							},
						},
					},
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name":      "test-1",
								"namespace": "junk",
							},
						},
					},
				},
				Replicas: ptr.Int(2), // replicas really dont matter
				Apply: map[string]interface{}{
					"kind":       "Pod",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name":      "test",
						"namespace": "dope",
					},
					"spec": nil, // i.e. delete this resource
				},
			},
			expectDeletes: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "test-0",
							"namespace": "dope",
						},
					},
				},
			},
			isErr: false,
		},
		"run + single explicit delete + replicas = 0": {
			config: CreateOrDeleteRequest{
				TaskKey: "test.101",
				Run: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Deployment",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "test",
							"namespace": "dope",
							"uid":       "test-101",
						},
					},
				},
				Watch: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Deployment",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "test",
							"namespace": "dope",
							"uid":       "test-101",
						},
					},
				},
				ObservedResources: []*unstructured.Unstructured{
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name":      "test-0",
								"namespace": "dope",
							},
						},
					},
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name":      "test-1",
								"namespace": "junk",
							},
						},
					},
				},
				Replicas: ptr.Int(0), // implies delete
				Apply: map[string]interface{}{
					"kind":       "Pod",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name":      "test",
						"namespace": "dope",
					},
				},
			},
			expectDeletes: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "test-0",
							"namespace": "dope",
						},
					},
				},
			},
			isErr: false,
		},
		"run + multiple explicit delete + replicas = 0": {
			config: CreateOrDeleteRequest{
				TaskKey: "test.101",
				Run: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Deployment",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "test",
							"namespace": "dope",
							"uid":       "test-101",
						},
					},
				},
				Watch: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Deployment",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "test",
							"namespace": "dope",
							"uid":       "test-101",
						},
					},
				},
				ObservedResources: []*unstructured.Unstructured{
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name":      "test-0",
								"namespace": "dope",
							},
						},
					},
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name":      "test-1",
								"namespace": "dope",
							},
						},
					},
				},
				Replicas: ptr.Int(0), // implies delete
				Apply: map[string]interface{}{
					"kind":       "Pod",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name":      "test",
						"namespace": "dope",
					},
				},
			},
			expectDeletes: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "test-0",
							"namespace": "dope",
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "test-1",
							"namespace": "dope",
						},
					},
				},
			},
			isErr: false,
		},
		"run + 2/3 explicit deletes + replicas = 0": {
			config: CreateOrDeleteRequest{
				TaskKey: "test.101",
				Run: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Deployment",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "test",
							"namespace": "dope",
							"uid":       "test-101",
						},
					},
				},
				Watch: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Deployment",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "test",
							"namespace": "dope",
							"uid":       "test-101",
						},
					},
				},
				ObservedResources: []*unstructured.Unstructured{
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name":      "test-0",
								"namespace": "dope",
							},
						},
					},
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name":      "test-1",
								"namespace": "dope",
							},
						},
					},
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"metadata": map[string]interface{}{
								"name":      "test-2",
								"namespace": "dopes", // mismatch
							},
						},
					},
				},
				Replicas: ptr.Int(0), // implies delete
				Apply: map[string]interface{}{
					"kind":       "Pod",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name":      "test",
						"namespace": "dope",
					},
				},
			},
			expectDeletes: []*unstructured.Unstructured{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "test-0",
							"namespace": "dope",
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":      "test-1",
							"namespace": "dope",
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
			got, err := BuildCreateOrDeleteStates(mock.config)
			if mock.isErr && err == nil {
				t.Fatalf("Expected error got none")
			}
			if !mock.isErr && err != nil {
				t.Fatalf("Expected no error got [%+v]", err)
			}
			if mock.isErr {
				return
			}
			if len(got.DesiredResources) != len(mock.expectStates) {
				t.Fatalf(
					"Expected state count %d got %d",
					len(mock.expectStates),
					len(got.DesiredResources),
				)
			}
			if !unstruct.List(got.DesiredResources).IdentifiesAll(mock.expectStates) {
				t.Fatalf(
					"Expected no diff in states got\n%s",
					cmp.Diff(got.DesiredResources, mock.expectStates),
				)
			}
			if len(got.ExplicitDeletes) != len(mock.expectDeletes) {
				t.Fatalf(
					"Expected delete count %d got %d",
					len(mock.expectDeletes),
					len(got.ExplicitDeletes),
				)
			}
			if !unstruct.List(got.ExplicitDeletes).IdentifiesAll(mock.expectDeletes) {
				t.Fatalf(
					"Expected no diff in deletes got\n%s",
					cmp.Diff(got.ExplicitDeletes, mock.expectDeletes),
				)
			}
		})
	}
}
