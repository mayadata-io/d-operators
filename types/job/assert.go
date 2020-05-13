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

// Assert handles assertion of desired state against
// the observed state found in the cluster
type Assert struct {
	// Desired state(s) that is asserted against the observed
	// state(s)
	State *unstructured.Unstructured `json:"state"`

	// StateCheck has assertions related to state of resources
	StateCheck *StateCheck `json:"stateCheck,omitempty"`

	// PathCheck has assertions related to resource paths
	PathCheck *PathCheck `json:"pathCheck,omitempty"`
}

// String implements the Stringer interface
func (a Assert) String() string {
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

// AssertStatusPhase defines the status of executing an assertion
type AssertStatusPhase string

const (
	// AssertResultPassed defines a successful assertion
	AssertResultPassed AssertStatusPhase = "AssertPassed"

	// AssertResultWarning defines an assertion that resulted in warning
	AssertResultWarning AssertStatusPhase = "AssertWarning"

	// AssertResultFailed defines a failed assertion
	AssertResultFailed AssertStatusPhase = "AssertFailed"
)

// ToTaskStatusPhase transforms AssertResultPhase to TaskResultPhase
func (phase AssertStatusPhase) ToTaskStatusPhase() TaskStatusPhase {
	switch phase {
	case AssertResultPassed:
		return TaskStatusPassed
	case AssertResultFailed:
		return TaskStatusFailed
	case AssertResultWarning:
		return TaskStatusWarning
	default:
		return ""
	}
}

// AssertStatus holds the result of assertion
type AssertStatus struct {
	Phase   AssertStatusPhase `json:"phase"`
	Message string            `json:"message,omitempty"`
	Verbose string            `json:"verbose,omitempty"`
	Warning string            `json:"warning,omitempty"`
}

// AssertCheckType defines the type of assert check
type AssertCheckType int

const (
	// AssertCheckTypeState defines a state check based assertion
	AssertCheckTypeState AssertCheckType = iota

	// AssertCheckTypePath defines a path check based assertion
	AssertCheckTypePath
)
