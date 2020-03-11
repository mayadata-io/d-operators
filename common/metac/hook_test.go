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

package metac

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"openebs.io/metac/controller/common"
	"openebs.io/metac/controller/generic"
)

func TestGetDetailsFromRequest(t *testing.T) {
	var tests = map[string]struct {
		request  *generic.SyncHookRequest
		isExpect bool
	}{
		"nil request": {
			isExpect: false,
		},
		"nil watch & attachments": {
			request:  &generic.SyncHookRequest{},
			isExpect: true,
		},
		"not nil watch & nil attachments": {
			request: &generic.SyncHookRequest{
				Watch: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Some",
						"metadata": map[string]interface{}{
							"name":      "test",
							"namespace": "test",
						},
					},
				},
			},
			isExpect: true,
		},
		"not nil watch & not nil attachments": {
			request: &generic.SyncHookRequest{
				Watch: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Some",
						"metadata": map[string]interface{}{
							"name":      "test",
							"namespace": "test",
						},
					},
				},
				Attachments: common.AnyUnstructRegistry(
					map[string]map[string]*unstructured.Unstructured{
						"gvk": map[string]*unstructured.Unstructured{
							"nsname1": &unstructured.Unstructured{
								Object: map[string]interface{}{
									"kind": "Some1",
								},
							},
							"nsname2": &unstructured.Unstructured{
								Object: map[string]interface{}{
									"kind": "Some2",
								},
							},
						},
					},
				),
			},
			isExpect: true,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			got := GetDetailsFromRequest(mock.request)
			if len(got) != 0 != mock.isExpect {
				t.Fatalf(
					"Expected response %t got %s",
					mock.isExpect,
					got,
				)
			}
		})
	}
}

func TestGetDetailsFromResponse(t *testing.T) {
	var tests = map[string]struct {
		response *generic.SyncHookResponse
		isExpect bool
	}{
		"nil response": {},
		"nil response attachments": {
			response: &generic.SyncHookResponse{},
		},
		"1 response attachment": {
			response: &generic.SyncHookResponse{
				Attachments: []*unstructured.Unstructured{
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"kind": "Some",
						},
					},
				},
			},
			isExpect: true,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			got := GetDetailsFromResponse(mock.response)
			if len(got) != 0 != mock.isExpect {
				t.Fatalf(
					"Expected response %t got %s",
					mock.isExpect,
					got,
				)
			}
		})
	}
}
