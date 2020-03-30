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

// Pair represents the labels
type Pair map[string]string

// New returns a new Pair type
func New(labels map[string]string) Pair {
	return Pair(labels)
}

// Has returns true if all the given labels are
// available
func (p Pair) Has(given map[string]string) bool {
	if len(given) == len(p) && len(p) == 0 {
		return true
	}
	if len(p) == 0 {
		return false
	}
	for k, v := range given {
		if p[k] != v {
			return false
		}
	}
	return true
}
