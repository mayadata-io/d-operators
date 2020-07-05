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

import "encoding/json"

// FailFastRule defines the condition that leads to fail fast
type FailFastRule string

const (
	// FailFastOnDiscoveryError defines a fail fast based on
	// DiscoveryError
	FailFastOnDiscoveryError FailFastRule = "OnDiscoveryError"
)

// IgnoreErrorRule defines the rule to ignore an error
type IgnoreErrorRule string

const (
	// IgnoreErrorAsPassed defines the rule to ignore error
	// and treat it as passed
	IgnoreErrorAsPassed IgnoreErrorRule = "AsPassed"

	// IgnoreErrorAsWarning defines the rule to ignore error
	// and treat it as a warning
	IgnoreErrorAsWarning IgnoreErrorRule = "AsWarning"
)

// FailFast holds the condition that determines if an error
// should not result in retries and instead be allowed to fail
// immediately
type FailFast struct {
	When FailFastRule `json:"when,omitempty"`
}

// Task that needs to be executed as part of a Job
//
// Task forms the fundamental unit of execution within a
// Job
type Task struct {
	Name            string          `json:"name"`
	Assert          *Assert         `json:"assert,omitempty"`
	Apply           *Apply          `json:"apply,omitempty"`
	Delete          *Delete         `json:"delete,omitempty"`
	Create          *Create         `json:"create,omitempty"`
	IgnoreErrorRule IgnoreErrorRule `json:"ignoreError,omitempty"`
	FailFast        *FailFast       `json:"failFast,omitempty"`
}

// String implements the Stringer interface
func (t Task) String() string {
	raw, err := json.MarshalIndent(
		t,
		" ",
		".",
	)
	if err != nil {
		panic(err)
	}
	return string(raw)
}

// TaskStatusPhase defines the task execution status
type TaskStatusPhase string

const (
	// TaskStatusPassed implies a passed task
	TaskStatusPassed TaskStatusPhase = "Passed"

	// TaskStatusFailed implies a failed task
	TaskStatusFailed TaskStatusPhase = "Failed"

	// TaskStatusWarning implies a failed task
	TaskStatusWarning TaskStatusPhase = "Warning"
)

// TaskStatus holds task execution details
type TaskStatus struct {
	Step                 int             `json:"step"`
	Phase                TaskStatusPhase `json:"phase"`
	ElapsedTimeInSeconds *float64        `json:"elapsedTimeInSeconds,omitempty"`
	Internal             *bool           `json:"internal,omitempty"`
	Message              string          `json:"message,omitempty"`
	Verbose              string          `json:"verbose,omitempty"`
	Warning              string          `json:"warning,omitempty"`
	Timeout              string          `json:"timeout,omitempty"`
}
