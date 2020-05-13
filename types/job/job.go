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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metac "openebs.io/metac/apis/metacontroller/v1alpha1"
)

// When defines when an action should take place
type When string

const (
	// Always marks an action to get executed always
	Always When = "Always"

	// Never marks an action to never get executed
	Never When = "Never"

	// Once marks an action to get executed only once
	Once When = "Once"

	// Exists marks an action to get executed when corresponding
	// reference (kubernetes resource, etc) exists
	Exists When = "Exists"

	// NotExists marks an action to get executed when corresponding
	// reference (kubernetes resource, etc) does not exist
	NotExists When = "NotExists"

	// ListCountEquals marks an action to get executed when corresponding
	// list of references (kubernetes resource, etc) equals the given count
	ListCountEquals When = "ListCountEquals"

	// ListCountNotEquals marks an action to get executed when corresponding
	// list of references (kubernetes resource, etc) does not equal the given count
	ListCountNotEquals When = "ListCountNotEquals"
)

// Job is a kubernetes custom resource that defines
// the specifications to invoke kubernetes operations
// against any kubernetes custom resource
type Job struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	JobSpec   JobSpec   `json:"spec"`
	JobStatus JobStatus `json:"status"`
}

// JobSpec defines the tasks that get executed as part of
// executing this Job
type JobSpec struct {
	Teardown *bool    `json:"teardown,omitempty"`
	Enabled  *Enabled `json:"enabled,omitempty"`
	Tasks    []Task   `json:"tasks"`
}

// Enabled defines if the job is enabled to be executed
// or not
type Enabled struct {
	If    metac.ResourceSelector `json:"if,omitempty"`
	When  When                   `json:"when,omitempty"`
	Count *int                   `json:"count,omitempty"`
}

// JobStatusPhase is a typed definition to determine the
// result of executing a Job
type JobStatusPhase string

const (
	// JobStatusLocked implies a locked Job
	JobStatusLocked JobStatusPhase = "Locked"

	// JobStatusDisabled implies a disabled Job
	JobStatusDisabled JobStatusPhase = "Disabled"

	// JobStatusPassed implies a passed Job
	JobStatusPassed JobStatusPhase = "Passed"

	// JobStatusCompleted implies a successfully completed Job
	JobStatusCompleted JobStatusPhase = "Completed"

	// JobStatusFailed implies a failed Job
	JobStatusFailed JobStatusPhase = "Failed"

	// JobStatusWarning implies a Job with warnings
	JobStatusWarning JobStatusPhase = "Warning"
)

// JobStatus holds the results of all tasks specified
// in a Job
type JobStatus struct {
	Phase           JobStatusPhase        `json:"phase"`
	Reason          string                `json:"reason,omitempty"`
	Message         string                `json:"message,omitempty"`
	FailedTaskCount int                   `json:"failedTaskCount"`
	TaskCount       int                   `json:"taskCount"`
	TaskListStatus  map[string]TaskStatus `json:"taskListStatus"`
}

// String implements the Stringer interface
func (jr JobStatus) String() string {
	raw, err := json.MarshalIndent(
		jr,
		" ",
		".",
	)
	if err != nil {
		panic(err)
	}
	return string(raw)
}
