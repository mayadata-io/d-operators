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

// Label represents the label apply operation against
// one or more desired resources
type Label struct {
	// Desired state i.e. resources that needs to be
	// labeled
	State *unstructured.Unstructured `json:"state"`

	// Filter the resources by these names
	//
	// Optional
	FilterByNames []string `json:"filterByNames,omitempty"`

	// ApplyLabels represents the labels that need to be
	// applied
	ApplyLabels map[string]string `json:"applyLabels"`

	// AutoUnset removes the labels from the resources if
	// they were applied earlier and these resources are
	// no longer elgible to be applied with these labels
	//
	// Defaults to false
	AutoUnset bool `json:"autoUnset"`
}

// String implements the Stringer interface
func (l Label) String() string {
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

// LabelStatusPhase is a typed definition to determine the
// result of executing the label operation
type LabelStatusPhase string

const (
	// LabelStatusPassed defines a successful labeling
	LabelStatusPassed LabelStatusPhase = "Passed"

	// LabelStatusWarning defines the label operation
	// that resulted in warnings
	LabelStatusWarning LabelStatusPhase = "Warning"

	// LabelStatusFailed defines a failed labeling
	LabelStatusFailed LabelStatusPhase = "Failed"
)

// ToTaskStatusPhase transforms ApplyStatusPhase to TestResultPhase
func (phase LabelStatusPhase) ToTaskStatusPhase() TaskStatusPhase {
	switch phase {
	case LabelStatusPassed:
		return TaskStatusPassed
	case LabelStatusFailed:
		return TaskStatusFailed
	case LabelStatusWarning:
		return TaskStatusWarning
	default:
		return ""
	}
}

// LabelResult holds the result of labeling operation
type LabelResult struct {
	Phase   LabelStatusPhase `json:"phase"`
	Message string           `json:"message,omitempty"`
	Verbose string           `json:"verbose,omitempty"`
	Warning string           `json:"warning,omitempty"`
}
