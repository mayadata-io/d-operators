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

package recipe

import (
	"testing"

	"k8s.io/client-go/rest"
)

func TestNewFixture(t *testing.T) {
	var tests = map[string]struct {
		kubeConfig *rest.Config
		isErr      bool
	}{
		"nil kubeconfig": {
			isErr: true,
		},
		"empty kubeconfig": {
			kubeConfig: &rest.Config{},
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			_, err := NewFixture(FixtureConfig{
				KubeConfig: mock.kubeConfig,
			})
			if mock.isErr && err == nil {
				t.Fatal("Expected error got none")
			}
			if !mock.isErr && err != nil {
				t.Fatalf("Expected no error got %s", err.Error())
			}
		})
	}
}
