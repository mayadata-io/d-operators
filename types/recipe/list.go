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

package types

import (
	"encoding/json"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// List represents the desired state that needs to
// be listed from the cluster
type List struct {
	// Desired state that needs to be listed from the
	// Kubernetes cluster
	State *unstructured.Unstructured `json:"state"`
}

// String implements the Stringer interface
func (l List) String() string {
	raw, err := json.MarshalIndent(
		l,
		" ",
		".",
	)
	if err != nil {
		panic(err)
	}
	return string(raw)
}

// ListStatusPhase is a typed definition to determine the
// result of executing a list invocation
type ListStatusPhase string

const (
	// ListStatusPassed defines a successful list
	ListStatusPassed ListStatusPhase = "Passed"

	// ListStatusWarning defines a list that resulted in warnings
	ListStatusWarning ListStatusPhase = "Warning"

	// ListStatusFailed defines a failed list
	ListStatusFailed ListStatusPhase = "Failed"
)

// ToTaskStatusPhase transforms ListStatusPhase to TaskStatusPhase
func (phase ListStatusPhase) ToTaskStatusPhase() TaskStatusPhase {
	switch phase {
	case ListStatusPassed:
		return TaskStatusPassed
	case ListStatusFailed:
		return TaskStatusFailed
	case ListStatusWarning:
		return TaskStatusWarning
	default:
		return ""
	}
}

// ListResult holds the result of the list operation
type ListResult struct {
	Phase           ListStatusPhase                       `json:"phase"`
	Message         string                                `json:"message,omitempty"`
	Verbose         string                                `json:"verbose,omitempty"`
	Warning         string                                `json:"warning,omitempty"`
	V1Beta1CRDItems *v1beta1.CustomResourceDefinitionList `json:"v1b1CRDs,omitempty"`
	Items           *unstructured.UnstructuredList        `json:"items,omitempty"`
}
