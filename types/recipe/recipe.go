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

// EnabledRule defines when & how often a Recipe should get executed
type EnabledRule string

const (
	// EnabledRuleAlways enables the recipe to get executed as
	// many times this recipe resource is reconciled
	EnabledRuleAlways EnabledRule = "Always"

	// EnabledRuleNever disables the recipe execution forever
	//
	// NOTE:
	//	This is as good as disabling execution
	EnabledRuleNever EnabledRule = "Never"

	// EnabledRuleOnce enables the recipe to get executed only once
	// in its lifetime
	//
	// NOTE:
	//	This is the default mode of execution
	EnabledRuleOnce EnabledRule = "Once"
)

// EligibleItemRule defines the eligibility criteria to grant a Recipe to get executed
type EligibleItemRule string

const (
	// EligibleItemRuleExists allows Recipe execution if desired resources exist
	EligibleItemRuleExists EligibleItemRule = "Exists"

	// EligibleItemRuleNotFound allows Recipe execution if desired resources
	// are not found
	EligibleItemRuleNotFound EligibleItemRule = "NotFound"

	// EligibleItemRuleListCountEquals allows Recipe execution if desired resources
	// count match the provided count
	EligibleItemRuleListCountEquals EligibleItemRule = "ListCountEquals"

	// EligibleItemRuleListCountNotEquals allows Recipe execution if desired resources
	// count do not match the provided count
	EligibleItemRuleListCountNotEquals EligibleItemRule = "ListCountNotEquals"

	// EligibleItemRuleListCountGTE allows Recipe execution if desired resources
	// count is greater than or equal to the provided count
	EligibleItemRuleListCountGTE EligibleItemRule = "ListCountGreaterThanEquals"

	// EligibleItemRuleListCountLTE allows Recipe execution if desired resources
	// count is less than or equal to the provided count
	EligibleItemRuleListCountLTE EligibleItemRule = "ListCountLessThanEquals"
)

// EligibleRule defines the eligibility criteria to grant a Recipe to get executed
type EligibleRule string

const (
	// EligibleRuleAllChecksPass allows Recipe execution if all
	// specified checks passes
	EligibleRuleAllChecksPass EligibleRule = "AllChecksPass"

	// EligibleRuleAnyCheckPass allows Recipe execution if any
	// specified checks pass
	EligibleRuleAnyCheckPass EligibleRule = "AnyCheckPass"
)

// Recipe is a kubernetes custom resource that defines
// the specifications to invoke kubernetes operations
// against any kubernetes custom resource
type Recipe struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   RecipeSpec   `json:"spec"`
	Status RecipeStatus `json:"status"`
}

// RecipeSpec defines the tasks that get executed as part of
// executing this Recipe
type RecipeSpec struct {
	Teardown           *bool     `json:"teardown,omitempty"`
	ThinkTimeInSeconds *float64  `json:"thinkTimeInSeconds,omitempty"`
	Enabled            *Enabled  `json:"enabled,omitempty"`
	Eligible           *Eligible `json:"eligible,omitempty"`
	Refresh            Refresh   `json:"refresh,omitempty"`
	Tasks              []Task    `json:"tasks"`
}

// Refresh options to reconcile Recipe
type Refresh struct {
	ResyncAfterSeconds        *float64 `json:"resyncAfterSeconds,omitempty"`
	OnErrorResyncAfterSeconds *float64 `json:"onErrorResyncAfterSeconds,omitempty"`
}

// Enabled defines if the recipe is enabled to be executed
// or not
type Enabled struct {
	// Condition to enable or disable this Recipe
	When EnabledRule `json:"when,omitempty"`
}

// Eligible defines the eligibility criteria to grant a Recipe to get
// executed
type Eligible struct {
	Checks []EligibleItem `json:"checks"`
	When   EligibleRule   `json:"when,omitempty"`
}

// EligibleItem defines the eligibility criteria to grant a Recipe to get
// executed
type EligibleItem struct {
	ID            string               `json:"id,omitempty"`
	APIVersion    string               `json:"apiVersion,omitempty"`
	Kind          string               `json:"kind,omitempty"`
	LabelSelector metav1.LabelSelector `json:"labelSelector,omitempty"`
	When          EligibleItemRule     `json:"when,omitempty"`
	Count         *int                 `json:"count,omitempty"`
}

// RecipeStatusPhase is a typed definition to determine the
// result of executing a Recipe
type RecipeStatusPhase string

const (
	// RecipeStatusLocked implies a locked Recipe
	RecipeStatusLocked RecipeStatusPhase = "Locked"

	// RecipeStatusDisabled implies a disabled Recipe
	RecipeStatusDisabled RecipeStatusPhase = "Disabled"

	// RecipeStatusNotEligible implies a Recipe that is not
	// eligible for execution
	//
	// NOTE:
	//	This might be a temporary phase. In other words,
	// Recipe might be eligible to be executed in subsequent
	// reconcile attempts.
	RecipeStatusNotEligible RecipeStatusPhase = "NotEligible"

	// RecipeStatusPassed implies a passed Recipe
	RecipeStatusPassed RecipeStatusPhase = "Passed"

	// RecipeStatusCompleted implies a successfully completed Recipe
	RecipeStatusCompleted RecipeStatusPhase = "Completed"

	// RecipeStatusFailed implies a failed Recipe
	RecipeStatusFailed RecipeStatusPhase = "Failed"

	// RecipeStatusWarning implies a Recipe with warnings
	RecipeStatusWarning RecipeStatusPhase = "Warning"
)

// TaskCount holds various counts related to execution of tasks
// specified in the Recipe
type TaskCount struct {
	Failed  int `json:"failed"`  // Number of failed tasks
	Skipped int `json:"skipped"` // Number of skipped tasks
	Total   int `json:"total"`   // Total number of tasks in the Recipe
}

// RecipeStatus holds the results of all tasks specified
// in a Recipe
type RecipeStatus struct {
	// A single word status
	// Can be used to compare, assert, etc
	Phase RecipeStatusPhase `json:"phase"`

	// Short description of the Phase
	Reason string `json:"reason,omitempty"`

	// Long description of the Phase
	// Can be used to provide remedial action if any
	Message string `json:"message,omitempty"`

	// Time taken to execute the Recipe
	ExecutionTimeInSeconds *float64 `json:"executionTimeInSeconds,omitempty"`

	TaskCount      TaskCount             `json:"taskCount"`
	TaskResultList map[string]TaskResult `json:"taskResultList"`
}

// String implements the Stringer interface
func (jr RecipeStatus) String() string {
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
