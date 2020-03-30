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

package labels

import "testing"

func TestLabelsHas(t *testing.T) {
	var tests = map[string]struct {
		src        map[string]string
		target     map[string]string
		isContains bool
	}{
		"no src no target": {
			isContains: true,
		},
		"no src": {
			target: map[string]string{
				"hi": "there",
			},
			isContains: false,
		},
		"no target": {
			src: map[string]string{
				"hi": "there",
			},
			isContains: true,
		},
		"src has target": {
			src: map[string]string{
				"hi":    "there",
				"hello": "world",
			},
			target: map[string]string{
				"hello": "world",
			},
			isContains: true,
		},
		"src does not have target": {
			src: map[string]string{
				"hi":    "there",
				"hello": "world",
			},
			target: map[string]string{
				"hello": "earth",
			},
			isContains: false,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			p := New(mock.src)
			got := p.Has(mock.target)
			if got != mock.isContains {
				t.Fatalf(
					"Expected isContains %t got %t",
					mock.isContains,
					got,
				)
			}
		})
	}
}
