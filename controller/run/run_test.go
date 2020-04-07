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
