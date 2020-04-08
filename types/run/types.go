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

// TODO (@amitkumardas): draft specs
//
// UseCase: Delete a resource
//
// kind: Run
// spec:
//   tasks:
//   - if:
//     apply:
//     replicas: 0   # delete since replicas = 0
//
// UseCase: Create a resource
//
// kind: Run
// spec:
//   tasks:
//   - if:           # optional, create only if condition passes
//     apply:        # a single replica resource gets created
//
// UseCase: Delete a resource only once
//
// kind: Run
// spec:
//   tasks:
//   - apply:        # resource to be deleted
//     replicas: 0
//     once:         # run this task only once
//
// UseCase: Update a resource
//
// kind: Run
// spec:
//   tasks:
//   - if:           # if condition
//     apply:        # desired state to update
//     target:       # if (cond) then (update target)
//
// UseCase: Assert presence of a resource
//
// kind: Run
// spec:
//   tasks:
//   - assert:
//
// UseCase: Assert absence of a resource
//
// kind: Run
// spec:
//   tasks:
//   - assert:

// TODO (@amitkumardas): draft design
//
// NOTE: Run resource is the declarative way to code a controller
//
// - When any controller wants to use Run specs
//   - Run specs should be mounted into doperator binary
//   - Run name will be annotated at corresponding GenericController
//   - GenericController's sync will invoke run.SyncDelegate func
//   - GenericController's finalize will invoke run.FinalizeDelegate func
//	 - run.SyncDelegate will invoke run.Sync
//   - run.FinalizeDelegate will invoke run.Finalize
// - Run can be applied as CR as well
//	 - GenericController's sync will invoke run.Sync
//	 - GenericController's finalize will invoke run.Finalize

// TODO (@amitkumardas):
// Refactor create & delete into separate files & structures
// Add more informative messages to create & delete actions
// Add AssertResponse to assert action

// TODO (@amitkumardas):
// Ensure a single Watch is able to make use of more than one Run
// resources.

// TODO (@amitkumardas):
// annotations:
//   run.dao.mayadata.io/use-watch-for-result: true
// - If Run is a custom resource then its status is set with task results
// - If Run is not a custom resource then RunResult CR is set with task results
//   - RunResult namespace is set to namespace of watch if watch is namespaced
//   - RunResult namespace is set to namespace of operator if watch is cluster scoped

const (
	// AnnotationKeyMetacCreatedDueToWatch is the annotation key
	// found in GenericController attachments that were created
	// by the GenericController
	AnnotationKeyMetacCreatedDueToWatch string = "metac.openebs.io/created-due-to-watch"
)

const (
	// AnnotationKeyRunUID is the annotation key that holds
	// the uid of the Run resource
	AnnotationKeyRunUID string = "run.dao.mayadata.io/uid"

	// AnnotationKeyRunName is the annotation key that holds
	// the name of the Run resource
	AnnotationKeyRunName string = "run.dao.mayadata.io/name"

	// AnnotationKeyWatchUID is the annotation key that holds
	// the uid of the watch resource
	AnnotationKeyWatchUID string = "run.dao.mayadata.io/watch-uid"

	// AnnotationKeyWatchName is the annotation key that holds
	// the name of the watch resource
	AnnotationKeyWatchName string = "run.dao.mayadata.io/watch-name"

	// AnnotationKeyTaskKey is the annotationn key that holds the
	// taskkey value
	AnnotationKeyTaskKey string = "run.dao.mayadata.io/task-key"
)

// // RunStatusPhase determines the current phase of Run resource
// type RunStatusPhase string

// const (
// 	// RunStatusPhaseError indicates error during Run
// 	RunStatusPhaseError RunStatusPhase = "Error"

// 	// RunStatusPhaseOnline indicates last Run was successful
// 	RunStatusPhaseOnline RunStatusPhase = "Online"

// 	// RunStatusPhaseExited indicates Run's if cond failed
// 	RunStatusPhaseExited RunStatusPhase = "Exited"
// )

// ResultPhase determines the current result of a Task
type ResultPhase string

const (
	// ResultPhaseInProgress indicates task is in progress
	ResultPhaseInProgress ResultPhase = "InProgress"

	// ResultPhaseCompleted indicates task is completed
	ResultPhaseCompleted ResultPhase = "Completed"

	// ResultPhaseError indicates error in Task execution
	ResultPhaseError ResultPhase = "Error"

	// ResultPhaseOnline indicates Task executed without any errors
	ResultPhaseOnline ResultPhase = "Online"

	// ResultPhaseSkipped indicates Task was skipped
	//
	// NOTE:
	//  This can happen if condition to run this task was not met
	ResultPhaseSkipped ResultPhase = "Skipped"

	// ResultPhaseAssertFailed indicates assertion failed
	ResultPhaseAssertFailed ResultPhase = "AssertFailed"

	// ResultPhaseAssertPassed indicates assertion passed
	ResultPhaseAssertPassed ResultPhase = "AssertPassed"
)

// ResourceOperator is a typed definition of operator
type ResourceOperator string

const (
	// ResourceOperatorExists verifies if the expected resource exists
	//
	// Is the default ResourceOperator
	ResourceOperatorExists ResourceOperator = "Exists"

	// ResourceOperatorNotExist verifies if the expected resource does not
	// exist
	ResourceOperatorNotExist ResourceOperator = "NotExist"

	// ResourceOperatorEqualsCount matches actual resource count with expected
	// resource count
	ResourceOperatorEqualsCount ResourceOperator = "EqualsCount"

	// ResourceOperatorGTE verifies if actual resource count is greater than
	// or equal to expected resource count
	ResourceOperatorGTE ResourceOperator = "GTE"

	// ResourceOperatorLTE verifies if actual resource count is lesser than
	// or equal to expected resource count
	ResourceOperatorLTE ResourceOperator = "LTE"
)

// IfOperator defines the operator that needs to be applied
// against a list of Item(s)
type IfOperator string

const (
	// IfOperatorAND does an AND operation amongst the
	// list of Item(s)
	IfOperatorAND IfOperator = "AND"

	// IfOperatorOR does an **OR** operation amongst the
	// list of Item(s)
	IfOperatorOR IfOperator = "OR"
)

// IncludeInfoKey determines the information type that gets
// added to Run status
type IncludeInfoKey string

const (
	// IncludeAllInfo if true will include info on all
	// resources participating in Run controller
	IncludeAllInfo IncludeInfoKey = "*"

	// IncludeSkippedInfo includes info on skipped resources
	IncludeSkippedInfo IncludeInfoKey = "skipped-resources"

	// IncludeDesiredInfo includes info on desired resources
	IncludeDesiredInfo IncludeInfoKey = "desired-resources"

	// IncludeExplicitInfo includes info on resources that are
	// handled explicitly i.e. explicit delete or explicit update
	IncludeExplicitInfo IncludeInfoKey = "explicit-resources"

	// IncludeWarnInfo includes warnings
	IncludeWarnInfo IncludeInfoKey = "warnings"
)

// TargetResourceType is a typed constant used to indicate the type
// of resource to be updated
type TargetResourceType string

const (
	// TargetResourceTypeObserved indicates updates to be applied
	// against this Run's observed resources
	//
	// This is the default setting
	TargetResourceTypeObserved TargetResourceType = "Observed"

	// TargetResourceTypeDesired indicates update to be applied against
	// this Run's desired resources
	TargetResourceTypeDesired TargetResourceType = "Desired"
)

// Run is a Kubernetes custom resource that defines
// the specifications to operate on various Kubernetes
// resources
type Run struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   RunSpec   `json:"spec"`
	Status RunStatus `json:"status,omitempty"`
}

// RunSpec defines the configuration required
// to operate against one or more Kubernetes resources
type RunSpec struct {
	// Proceed with Run only if this condition succeeds
	//
	// NOTE:
	// 	RunIf is optional
	RunIf *If `json:"runIf,omitempty"`

	// Tasks represents a set of tasks that are executed
	// in a level triggered reconciliation loop
	//
	// Tasks is used to achieve the desired state(s) of
	// this Run spec
	Tasks []Task `json:"tasks"`

	// If set then details on resources participating in
	// Run controller get published in the status
	//
	// NOTE:
	//	This is useful for debugging purposes
	IncludeInfo map[IncludeInfoKey]bool `json:"includeInfoOn,omitempty"`
}

// Task represents the unit of execution for the Run resource
type Task struct {
	// Key to uniquely identify this task in this Run spec
	Key string `json:"key"`

	// A short or verbose description of this task
	Desc string `json:"desc,omitempty"`

	// Proceed either with Create, Delete, Update or Assert
	// if this condition succeeds
	//
	// NOTE:
	// 	If is optional
	If *If `json:"if,omitempty"`

	// Apply defines the desired state that needs to be
	// applied against the Kubernetes cluster
	//
	// Entire resource yaml _(native or custom)_ is embedded
	// here
	//
	// NOTE:
	// 	Apply is optional
	Apply map[string]interface{} `json:"desired,omitempty"`

	// Replicas when set to 0 implies **deletion** of resource
	// at the cluster. Similarly, when set to some value that is
	// greater than 1, implies applying multiple copies of the
	// resource specified in **state** field.
	//
	// Default value is 1
	//
	// Replicas is optional
	Replicas *int `json:"replicas,omitempty"`

	// The target(s) to update. State found in Apply will be
	// applied against the resources matched by this selector
	//
	// NOTE:
	//	Create, Delete & Update cannot be used together in a
	// single task
	//
	// NOTE:
	//	TargetSelector is optional
	TargetSelector TargetSelector `json:"targetSelector,omitempty"`

	// Assert verifies the presence of, absence of one or more
	// resources in the cluster.
	//
	// NOTE:
	// 	One should not try to create, update or delete along
	// with assert in a single task
	//
	// NOTE:
	// 	Assert is optional
	Assert *Assert `json:"assert,omitempty"`
}

// TargetSelector selects one or more resources that need to be
// updated
type TargetSelector struct {
	metac.ResourceSelector

	// ResourceType to be considered for update
	//
	// Defaults to TargetResourceTypeObserved
	ResourceType TargetResourceType `json:"resourceType,omitempty"`
}

// Assert any condition or state of resource
type Assert struct {
	// State of any resource is asserted as a whole
	//
	// This must have the kind & apiVersion as its
	// identifying fields
	State map[string]interface{} `json:"state,omitempty"`

	// Embed If structure which in turn plays the
	// role of assertion condition(s)
	If
}

// If defines the conditions reqired to verify the
// presence of, absence of, equals to & other checks
// against one or more resources observed in the cluster
type If struct {
	// OR-ing or AND-ing of conditions
	IfOperator IfOperator `json:"operator,omitempty"`

	// List of conditions that get executed as part of
	// if operation
	IfConditions []IfCondition `json:"conditions,omitempty"`
}

// IfCondition to match, filter, verify a kubernetes resource.
type IfCondition struct {
	// Selector to filter one or more resources that are expected
	// to be present in the cluster
	ResourceSelector metac.ResourceSelector `json:"resourceSelector,omitempty"`

	// ResourceOperator refers to the operation that gets executed to
	// the selected resources
	//
	// Defaults to 'Exists'
	ResourceOperator ResourceOperator `json:"operator,omitempty"`

	// Count comes into effect when operator is related to count
	// e.g. EqualsCount, GreaterThanEqualTo, LessThanEqualTo.
	Count *int `json:"count,omitempty"`
}

// RunStatus has the operational state the Run resource
type RunStatus struct {
	Result

	// TaskResultList provides status of each task
	TaskResultList TaskResultList `json:"taskResults,omitempty"`
}

// String implements Stringer interface
func (r RunStatus) String() string {
	raw, err := json.MarshalIndent(r, "", ".")
	if err != nil {
		panic(err)
	}
	return string(raw)
}

// TaskResultList holds the results of all tasks in the Run resource
type TaskResultList map[string]TaskResult

// TaskResult provide details of a task execution
//
// NOTE:
//	One of the result(s) should get filled once the task
// gets executed
type TaskResult struct {
	SkipResult   *SkipResult `json:"skipResult,omitempty"`
	IfCondResult *Result     `json:"ifCondResult,omitempty"`
	AssertResult *Result     `json:"assertResult,omitempty"`
	UpdateResult *Result     `json:"updateResult,omitempty"`
	CreateResult *Result     `json:"createResult,omitempty"`
	DeleteResult *Result     `json:"deleteResult,omitempty"`
}

// String implements Stringer interface
func (r TaskResult) String() string {
	raw, err := json.MarshalIndent(r, "", ".")
	if err != nil {
		panic(err)
	}
	return string(raw)
}

// Result provides details of a task execution. For
// example a assert task will have assert results filled up.
type Result struct {
	Phase                 ResultPhase `json:"phase,omitempty"`
	Message               string      `json:"message,omitempty"`
	Errors                []string    `json:"errors,omitempty"`
	Warns                 []string    `json:"warns,omitempty"`
	ExplicitResourcesInfo []string    `json:"explicitResourcesInfo,omitempty"`
	DesiredResourcesInfo  []string    `json:"desiredResourcesInfo,omitempty"`
	SkippedResourcesInfo  []string    `json:"skippedResourcesInfo,omitempty"`
	HasRunOnce            *bool       `json:"hasRunOnce,omitempty"`
	HasSkippedOnce        *bool       `json:"hasSkippedOnce,omitempty"`
}

// SkipResult provides details of the action which was not executed
type SkipResult struct {
	Phase   ResultPhase `json:"phase,omitempty"`
	Message string      `json:"message,omitempty"`
}

// HasSkipTask returns true if any task was skipped due to
// failing if condition
func (l TaskResultList) HasSkipTask() bool {
	for _, result := range l {
		if result.SkipResult != nil {
			return true
		}
	}
	return false
}

// SkipTaskCount returns the number of tasks in the Run resource
// that were skipped due to failing if condition
func (l TaskResultList) SkipTaskCount() int {
	var count int
	for _, result := range l {
		if result.SkipResult != nil {
			count++
		}
	}
	return count
}

// IfCondTaskCount returns the number of tasks in the Run resource
// that were executed as if cond action
func (l TaskResultList) IfCondTaskCount() int {
	var count int
	for _, result := range l {
		if result.IfCondResult != nil {
			count++
		}
	}
	return count
}

// AssertTaskCount returns the number of tasks in the Run resource
// that were executed as assert action
func (l TaskResultList) AssertTaskCount() int {
	var count int
	for _, result := range l {
		if result.AssertResult != nil {
			count++
		}
	}
	return count
}

// FailedAssertTaskCount returns the number of tasks in the
// Run resource whose assert executions failed
func (l TaskResultList) FailedAssertTaskCount() int {
	var count int
	for _, result := range l {
		if result.AssertResult != nil &&
			result.AssertResult.Phase == ResultPhaseAssertFailed {
			count++
		}
	}
	return count
}

// PassedAssertTaskCount returns the number of tasks in the
// Run resource whose assert executions passed
func (l TaskResultList) PassedAssertTaskCount() int {
	var count int
	for _, result := range l {
		if result.AssertResult != nil &&
			result.AssertResult.Phase == ResultPhaseAssertPassed {
			count++
		}
	}
	return count
}

// CreateTaskCount returns the number of tasks in the Run resource
// that were executed as create action
func (l TaskResultList) CreateTaskCount() int {
	var count int
	for _, result := range l {
		if result.CreateResult != nil {
			count++
		}
	}
	return count
}

// DeleteTaskCount returns the number of tasks in the Run resource
// that were executed as delete action
func (l TaskResultList) DeleteTaskCount() int {
	var count int
	for _, result := range l {
		if result.DeleteResult != nil {
			count++
		}
	}
	return count
}

// UpdateTaskCount returns the number of tasks in the Run resource
// that were executed as update action
func (l TaskResultList) UpdateTaskCount() int {
	var count int
	for _, result := range l {
		if result.UpdateResult != nil {
			count++
		}
	}
	return count
}
