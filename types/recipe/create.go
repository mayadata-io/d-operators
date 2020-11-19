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

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Create creates the state found in the cluster
type Create struct {
	// Desired state that needs to be created
	State *unstructured.Unstructured `json:"state"`

	// Desired count that needs to be created
	Replicas *int `json:"replicas,omitempty"`

	// IgnoreDiscovery if set to true will not retry till
	// resource gets discovered
	//
	// NOTE:
	//	This is only applicable for kind: CustomResourceDefinition
	IgnoreDiscovery bool `json:"ignoreDiscovery"`
}

// String implements the Stringer interface
func (a Create) String() string {
	raw, err := json.MarshalIndent(
		a,
		" ",
		".",
	)
	if err != nil {
		panic(err)
	}
	return string(raw)
}

// CreateStatusPhase is a typed definition to determine the
// result of executing a create
type CreateStatusPhase string

const (
	// CreateStatusPassed defines a successful create
	CreateStatusPassed CreateStatusPhase = "Passed"

	// CreateStatusWarning defines a create that resulted in warnings
	CreateStatusWarning CreateStatusPhase = "Warning"

	// CreateStatusFailed defines a failed create
	CreateStatusFailed CreateStatusPhase = "Failed"
)

// ToTaskStatusPhase transforms CreateStatusPhase to TestResultPhase
func (phase CreateStatusPhase) ToTaskStatusPhase() TaskStatusPhase {
	switch phase {
	case CreateStatusPassed:
		return TaskStatusPassed
	case CreateStatusFailed:
		return TaskStatusFailed
	case CreateStatusWarning:
		return TaskStatusWarning
	default:
		return ""
	}
}

// CreateResult holds the result of the create operation
type CreateResult struct {
	Phase   CreateStatusPhase `json:"phase"`
	Message string            `json:"message,omitempty"`
	Verbose string            `json:"verbose,omitempty"`
	Warning string            `json:"warning,omitempty"`
}
