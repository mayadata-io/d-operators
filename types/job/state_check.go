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

// StateCheckOperator defines the check that needs to be
// done against the resource's state
type StateCheckOperator string

const (
	// StateCheckOperatorEquals verifies if expected state
	// matches the observed state found in the cluster
	StateCheckOperatorEquals StateCheckOperator = "Equals"

	// StateCheckOperatorNotEquals verifies if expected state
	// does not match the observed state found in the cluster
	StateCheckOperatorNotEquals StateCheckOperator = "NotEquals"

	// StateCheckOperatorNotFound verifies if expected state
	// is not found in the cluster
	StateCheckOperatorNotFound StateCheckOperator = "NotFound"

	// StateCheckOperatorListCountEquals verifies if expected
	// states matches the observed states found in the cluster
	StateCheckOperatorListCountEquals StateCheckOperator = "ListCountEquals"

	// StateCheckOperatorListCountNotEquals verifies if count of
	// expected states does not match the count of observed states
	// found in the cluster
	StateCheckOperatorListCountNotEquals StateCheckOperator = "ListCountNotEquals"
)

// StateCheck verifies expected resource state against
// the observed state found in the cluster
type StateCheck struct {
	// Check operation performed between the expected state
	// and the observed state
	Operator StateCheckOperator `json:"stateCheckOperator,omitempty"`

	// Count defines the expected number of observed states
	Count *int `json:"count,omitempty"`
}

// StateCheckResultPhase defines the result of StateCheck operation
type StateCheckResultPhase string

const (
	// StateCheckResultPassed defines a successful StateCheckResult
	StateCheckResultPassed StateCheckResultPhase = "StateCheckPassed"

	// StateCheckResultWarning defines a StateCheckResult that has warnings
	StateCheckResultWarning StateCheckResultPhase = "StateCheckResultWarning"

	// StateCheckResultFailed defines an un-successful StateCheckResult
	StateCheckResultFailed StateCheckResultPhase = "StateCheckResultFailed"
)

// ToAssertResultPhase transforms StateCheckResultPhase to AssertResultPhase
func (phase StateCheckResultPhase) ToAssertResultPhase() AssertStatusPhase {
	switch phase {
	case StateCheckResultPassed:
		return AssertResultPassed
	case StateCheckResultFailed:
		return AssertResultFailed
	case StateCheckResultWarning:
		return AssertResultWarning
	default:
		return ""
	}
}

// StateCheckResult holds the result of StateCheck operation
type StateCheckResult struct {
	Phase   StateCheckResultPhase `json:"phase"`
	Message string                `json:"message,omitempty"`
	Verbose string                `json:"verbose,omitempty"`
	Warning string                `json:"warning,omitempty"`
}
