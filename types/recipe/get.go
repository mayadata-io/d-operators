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

// Get represents the desired state that needs to
// be fetched from the cluster
type Get struct {
	// Desired state that needs to be fetched from the
	// Kubernetes cluster
	State *unstructured.Unstructured `json:"state"`
}

// String implements the Stringer interface
func (g Get) String() string {
	raw, err := json.MarshalIndent(
		g,
		" ",
		".",
	)
	if err != nil {
		panic(err)
	}
	return string(raw)
}

// GetStatusPhase is a typed definition to determine the
// result of executing a get invocation
type GetStatusPhase string

const (
	// GetStatusPassed defines a successful get
	GetStatusPassed GetStatusPhase = "Passed"

	// GetStatusWarning defines a get that resulted in warnings
	GetStatusWarning GetStatusPhase = "Warning"

	// GetStatusFailed defines a failed get
	GetStatusFailed GetStatusPhase = "Failed"
)

// ToTaskStatusPhase transforms GetStatusPhase to TaskStatusPhase
func (phase GetStatusPhase) ToTaskStatusPhase() TaskStatusPhase {
	switch phase {
	case GetStatusPassed:
		return TaskStatusPassed
	case GetStatusFailed:
		return TaskStatusFailed
	case GetStatusWarning:
		return TaskStatusWarning
	default:
		return ""
	}
}

// GetResult holds the result of the get operation
type GetResult struct {
	Phase      GetStatusPhase                    `json:"phase"`
	Message    string                            `json:"message,omitempty"`
	Verbose    string                            `json:"verbose,omitempty"`
	Warning    string                            `json:"warning,omitempty"`
	V1Beta1CRD *v1beta1.CustomResourceDefinition `json:"v1b1CRD,omitempty"`
	Object     *unstructured.Unstructured        `json:"object,omitempty"`
}
