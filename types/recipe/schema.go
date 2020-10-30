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

// SupportedAbsolutePaths represent the nested field paths that
// are supported by Recipe custom resource schema
//
// NOTE:
//	Nested path is represented by field name(s) joined
// by dots i.e. '.'
//
// NOTE:
//	Field with list-of-map(s) datatype are appended with [*]
//
// NOTE:
//	Whenever the Recipe schema is updated its field path needs to
// be updated at SupportedAbsolutePaths or UserAllowedPathPrefixes
//
// NOTE:
//	Each field path set here should represent its absolute field path
var SupportedAbsolutePaths = []string{
	"apiVersion",
	"kind",
	"spec.teardown",
	"spec.resync.onNotEligibleResyncInSeconds",
	"spec.resync.onErrorResyncInSeconds",
	"spec.resync.intervalInSeconds",
	"spec.eligible.checks.[*].kind",
	"spec.eligible.checks.[*].apiVersion",
	"spec.eligible.checks.[*].count",
	"spec.eligible.checks.[*].when",
	"spec.eligible.checks.[*].labelSelector.matchExpressions.[*].key",
	"spec.eligible.checks.[*].labelSelector.matchExpressions.[*].operator",
	"spec.eligible.checks.[*].labelSelector.matchExpressions.[*].values",
	"spec.enabled.when",
	"spec.tasks.[*].name",
	"spec.tasks.[*].failFast.when",
	"spec.tasks.[*].ignoreError",
	"spec.tasks.[*].create.replicas",
	"spec.tasks.[*].assert.stateCheck.stateCheckOperator",
	"spec.tasks.[*].assert.stateCheck.count",
	"spec.tasks.[*].assert.pathCheck.path",
	"spec.tasks.[*].assert.pathCheck.pathCheckOperator",
	"spec.tasks.[*].assert.pathCheck.value",
	"spec.tasks.[*].assert.pathCheck.dataType",
}

// UserAllowedPathPrefixes represent the nested field paths
// that can have further fields. These fields are not managed
// by Recipe schema & should not be validated as per the schema
// definition.
//
// NOTE:
// 	'metadata' is a native Kubernetes type. It is
// dependent on Kubernetes versions and hence makes little sense
// to validate the fields of metadata. UserAllowedPathPrefixes can
// be used to skip such field path(s).
//
// NOTE:
//	Each prefix set here must end with a dot i.e. `.`
var UserAllowedPathPrefixes = []string{
	"metadata.",                    // K8s controlled
	"status.",                      // dope controlled
	"spec.tasks.[*].apply.state.",  // can be any K8s resource
	"spec.tasks.[*].delete.state.", // can be any K8s resource
	"spec.tasks.[*].create.state.", // can be any K8s resource
	"spec.tasks.[*].assert.state.", // can be any K8s resource
	"spec.eligible.checks.[*].labelSelector.matchLabels.", // can be any label pairs
}

type SchemaStatus string

const (
	// SchemaStatusValid conveys a successful validation
	SchemaStatusValid SchemaStatus = "Valid"

	// SchemaStatusInvalid conveys a failed validation
	SchemaStatusInvalid SchemaStatus = "Invalid"
)

type SchemaFailure struct {
	Error  string `json:"error"`
	Remedy string `json:"remedy,omitempty"`
}

type SchemaResult struct {
	Phase    SchemaStatus    `json:"phase"`
	Failures []SchemaFailure `json:"failures,omitempty"`
	Verbose  []string        `json:"verbose,omitempty"`
}
