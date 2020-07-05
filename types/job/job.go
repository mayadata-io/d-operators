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
)

// EnabledRule defines when & how often a Job should get executed
type EnabledRule string

const (
	// EnabledRuleAlways enables the job to get executed as
	// many times this job resource is reconciled
	EnabledRuleAlways EnabledRule = "Always"

	// EnabledRuleNever disables the job execution forever
	//
	// NOTE:
	//	This is as good as disabling execution
	EnabledRuleNever EnabledRule = "Never"

	// EnabledRuleOnce enables the job to get executed only once
	// in its lifetime
	//
	// NOTE:
	//	This is the default mode of execution
	EnabledRuleOnce EnabledRule = "Once"
)

// EligibleItemRule defines the eligibility criteria to grant a Job to get executed
type EligibleItemRule string

const (
	// EligibleItemRuleExists allows Job execution if desired resources exist
	EligibleItemRuleExists EligibleItemRule = "Exists"

	// EligibleItemRuleNotFound allows Job execution if desired resources
	// are not found
	EligibleItemRuleNotFound EligibleItemRule = "NotFound"

	// EligibleItemRuleListCountEquals allows Job execution if desired resources
	// count match the provided count
	EligibleItemRuleListCountEquals EligibleItemRule = "ListCountEquals"

	// EligibleItemRuleListCountNotEquals allows Job execution if desired resources
	// count do not match the provided count
	EligibleItemRuleListCountNotEquals EligibleItemRule = "ListCountNotEquals"

	// EligibleItemRuleListCountGTE allows Job execution if desired resources
	// count is greater than or equal to the provided count
	EligibleItemRuleListCountGTE EligibleItemRule = "ListCountGreaterThanEquals"

	// EligibleItemRuleListCountLTE allows Job execution if desired resources
	// count is less than or equal to the provided count
	EligibleItemRuleListCountLTE EligibleItemRule = "ListCountLessThanEquals"
)

// EligibleRule defines the eligibility criteria to grant a Job to get executed
type EligibleRule string

const (
	// EligibleRuleAllChecksPass allows Job execution if all
	// specified checks passes
	EligibleRuleAllChecksPass EligibleRule = "AllChecksPass"

	// EligibleRuleAnyCheckPass allows Job execution if any
	// specified checks pass
	EligibleRuleAnyCheckPass EligibleRule = "AnyCheckPass"
)

// Job is a kubernetes custom resource that defines
// the specifications to invoke kubernetes operations
// against any kubernetes custom resource
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type Job struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   JobSpec   `json:"spec"`
	Status JobStatus `json:"status"`
}

// JobSpec defines the tasks that get executed as part of
// executing this Job
// +kubebuilder:subresource:status
type JobSpec struct {
	Teardown           *bool     `json:"teardown,omitempty"`
	ThinkTimeInSeconds *float64  `json:"thinkTimeInSeconds,omitempty"`
	Enabled            *Enabled  `json:"enabled,omitempty"`
	Eligible           *Eligible `json:"eligible,omitempty"`
	Refresh            Refresh   `json:"refresh,omitempty"`
	Tasks              []Task    `json:"tasks"`
}

// Refresh options to reconcile Job
// +kubebuilder:subresource:status
type Refresh struct {
	ResyncAfterSeconds        *float64 `json:"resyncAfterSeconds,omitempty"`
	OnErrorResyncAfterSeconds *float64 `json:"onErrorResyncAfterSeconds,omitempty"`
}

// Enabled defines if the job is enabled to be executed
// or not
// +kubebuilder:subresource:status
type Enabled struct {
	// Condition to enable or disable this Job
	When EnabledRule `json:"when,omitempty"`
}

// Eligible defines the eligibility criteria to grant a Job to get
// executed
// +kubebuilder:subresource:status
type Eligible struct {
	Checks []EligibleItem `json:"checks"`
	When   EligibleRule   `json:"when,omitempty"`
}

// EligibleItem defines the eligibility criteria to grant a Job to get
// executed
// +kubebuilder:subresource:status
type EligibleItem struct {
	ID            string               `json:"id,omitempty"`
	APIVersion    string               `json:"apiVersion,omitempty"`
	Kind          string               `json:"kind,omitempty"`
	LabelSelector metav1.LabelSelector `json:"labelSelector,omitempty"`
	When          EligibleItemRule     `json:"when,omitempty"`
	Count         *int                 `json:"count,omitempty"`
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
// +kubebuilder:subresource:status
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
