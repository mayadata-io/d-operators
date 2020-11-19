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
	metac "openebs.io/metac/apis/metacontroller/v1alpha1"
)

// Apply represents the desired state that needs to
// be applied against the cluster
type Apply struct {
	// Desired state that needs to be created or
	// updated or deleted. Resource gets created if
	// this state is not observed in the cluster.
	// However, if this state is found in the cluster,
	// then the corresponding resource gets updated
	// via a 3-way merge.
	State *unstructured.Unstructured `json:"state"`

	// Desired count that needs to be created
	//
	// NOTE:
	//	If value is 0 then this state needs to be
	// deleted
	Replicas *int `json:"replicas,omitempty"`

	// Resources that needs to be **updated** with above
	// desired state
	//
	// NOTE:
	//	Presence of Targets implies an update operation
	Targets metac.ResourceSelector `json:"targets,omitempty"`

	// IgnoreDiscovery if set to true will not retry till
	// resource gets discovered
	//
	// NOTE:
	//	This is only applicable for kind: CustomResourceDefinition
	IgnoreDiscovery bool `json:"ignoreDiscovery"`
}

// String implements the Stringer interface
func (a Apply) String() string {
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

// ApplyStatusPhase is a typed definition to determine the
// result of executing an apply
type ApplyStatusPhase string

const (
	// ApplyStatusPassed defines a successful apply
	ApplyStatusPassed ApplyStatusPhase = "Passed"

	// ApplyStatusWarning defines an apply that resulted in warnings
	ApplyStatusWarning ApplyStatusPhase = "Warning"

	// ApplyStatusFailed defines a failed apply
	ApplyStatusFailed ApplyStatusPhase = "Failed"
)

// ToTaskStatusPhase transforms ApplyStatusPhase to TestResultPhase
func (phase ApplyStatusPhase) ToTaskStatusPhase() TaskStatusPhase {
	switch phase {
	case ApplyStatusPassed:
		return TaskStatusPassed
	case ApplyStatusFailed:
		return TaskStatusFailed
	case ApplyStatusWarning:
		return TaskStatusWarning
	default:
		return ""
	}
}

// ApplyResult holds the result of the apply operation
type ApplyResult struct {
	Phase   ApplyStatusPhase `json:"phase"`
	Message string           `json:"message,omitempty"`
	Verbose string           `json:"verbose,omitempty"`
	Warning string           `json:"warning,omitempty"`
}
