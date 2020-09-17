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

// PathCheckOperator defines the check that needs to be
// done against the resource's field based on this field's
// path
type PathCheckOperator string

const (
	// PathCheckOperatorExists verifies if expected field
	// is found in the observed resource found in the cluster
	//
	// NOTE:
	// 	This is the **default** if nothing is specified
	//
	// NOTE:
	//	This is a path only check operation
	PathCheckOperatorExists PathCheckOperator = "Exists"

	// PathCheckOperatorNotExists verifies if expected field
	// is not found in the observed resource found in the cluster
	//
	// NOTE:
	//	This is a path only check operation
	PathCheckOperatorNotExists PathCheckOperator = "NotExists"

	// PathCheckOperatorEquals verifies if expected field
	// value matches the field value of the observed resource
	// found in the cluster
	//
	// NOTE:
	//	This is a path as well as value based check operation
	PathCheckOperatorEquals PathCheckOperator = "Equals"

	// PathCheckOperatorNotEquals verifies if expected field
	// value does not match the observed resource's field value
	// found in the cluster
	//
	// NOTE:
	//	This is a path as well as value based check operation
	PathCheckOperatorNotEquals PathCheckOperator = "NotEquals"

	// PathCheckOperatorGTE verifies if expected field value
	// is greater than or equal to the field value of the
	// observed resource found in the cluster
	//
	// NOTE:
	//	This is a path as well as value based check operation
	PathCheckOperatorGTE PathCheckOperator = "GTE"

	// PathCheckOperatorLTE verifies if expected field value
	// is less than or equal to the field value of the
	// observed resource found in the cluster
	//
	// NOTE:
	//	This is a path as well as value based check operation
	PathCheckOperatorLTE PathCheckOperator = "LTE"
)

// PathValueDataType defines the expected data type of the value
// set against the path
type PathValueDataType string

const (
	// PathValueDataTypeInt64 expects path's value with int64
	// as its data type
	PathValueDataTypeInt64 PathValueDataType = "int64"

	// PathValueDataTypeFloat64 expects path's value with float64
	// as its data type
	PathValueDataTypeFloat64 PathValueDataType = "float64"

	// PathValueDataTypeString expects path's value with string
	// as its data type
	PathValueDataTypeString PathValueDataType = "string"
)

// PathCheck verifies expected field value against
// the field value of the observed resource found in the
// cluster
type PathCheck struct {
	// Check operation performed between the expected field
	// value and the field value of the observed resource
	// found in the cluster
	Operator PathCheckOperator `json:"pathCheckOperator,omitempty"`

	// Nested path of the field found in the resource
	//
	// NOTE:
	//	This is a mandatory field
	Path string `json:"path"`

	// Expected value that gets verified against the observed
	// value based on the path & operator
	Value interface{} `json:"value,omitempty"`

	// Data type of the value e.g. int64 or float64 etc
	DataType PathValueDataType `json:"dataType,omitempty"`
}

// PathCheckResultPhase defines the result of PathCheck operation
type PathCheckResultPhase string

const (
	// PathCheckResultPassed defines a successful PathCheckResult
	PathCheckResultPassed PathCheckResultPhase = "PathCheckPassed"

	// PathCheckResultWarning defines a PathCheckResult that has warnings
	PathCheckResultWarning PathCheckResultPhase = "PathCheckWarning"

	// PathCheckResultFailed defines an un-successful PathCheckResult
	PathCheckResultFailed PathCheckResultPhase = "PathCheckFailed"
)

// ToAssertResultPhase transforms StateCheckResultPhase to AssertResultPhase
func (phase PathCheckResultPhase) ToAssertResultPhase() AssertStatusPhase {
	switch phase {
	case PathCheckResultPassed:
		return AssertResultPassed
	case PathCheckResultFailed:
		return AssertResultFailed
	case PathCheckResultWarning:
		return AssertResultWarning
	default:
		return ""
	}
}

// PathCheckResult holds the result of PathCheck operation
type PathCheckResult struct {
	Phase   PathCheckResultPhase `json:"phase"`
	Message string               `json:"message,omitempty"`
	Verbose string               `json:"verbose,omitempty"`
	Warning string               `json:"warning,omitempty"`
	Timeout string               `json:"timeout,omitempty"`
}
