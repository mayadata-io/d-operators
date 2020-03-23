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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	metac "openebs.io/metac/apis/metacontroller/v1alpha1"
)

// TODO (@amitkumardas): draft specs
//
// kind: Run
// spec:
//   tasks:
//   - if:
//     apply:
//     replicas: 0   # delete since replicas = 0
//   - if:
//     apply:        # create
//     replicas: 1   # optional; default is 1 replicas
//     once:         # if true then this task is run only once
//   - if:
//     apply:
//     for:          # update
//   - assert:       # entire task is just an assertion

// TODO (@amitkumardas): draft design
//
// NOTE: Run resource is the declarative way to code a controller
//
// - Metac specs will be provided to binary as file(s)
// - Run specs will be provided to binary as file(s)
// - Binary will apply the Run resource from file(s) against k8s cluster
// - Binary will apply its internal CRDs as part of kind: DOperator
// - User will deploy this binary as a StatefulSet
// - User will apply their custom resources
// - User's custom resource will be updated with Run status
//
// - Sample metac.yaml with a custom resource & a Run resource:
//   kind: GenericController
//   spec:
//     watch:
//       kind: user's custom resource
//     attachments:
//     - kind: Run
//       select: specific to watch
//     - kind: other1 (as required inside Run)
//     - kind: other2 (as required inside Run)
//     - kind: othern (as required inside Run)
//     inlinehook:
//       # multiple GenericControllers can use this inline func
//       funcname: predefined in the binary
//
// - Sample metac.yaml with Run resource & no custom resource:
//   kind: GenericController
//   spec:
//     watch:
//       kind: Run
//     attachments:
//     - kind: other1 (as required inside Run)
//     - kind: other2 (as required inside Run)
//     - kind: othern (as required inside Run)
//     inlinehook:
//       # multiple GenericControllers can use this inline func
//       funcname: predefined in the binary

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

// RunStatusPhase determines the current phase of Run resource
type RunStatusPhase string

const (
	// RunStatusPhaseError indicates error during Run
	RunStatusPhaseError RunStatusPhase = "Error"

	// RunStatusPhaseOnline indicates last Run was successful
	RunStatusPhaseOnline RunStatusPhase = "Online"

	// RunStatusPhaseExited indicates Run was exited
	RunStatusPhaseExited TaskStatusPhase = "Exited"
)

// TaskStatusPhase determines the current phase of a Task
type TaskStatusPhase string

const (
	// TaskStatusPhaseInProgress indicates task is in progress
	TaskStatusPhaseInProgress TaskStatusPhase = "InProgress"

	// TaskStatusPhaseCompleted indicates task is completed
	TaskStatusPhaseCompleted TaskStatusPhase = "Completed"

	// TaskStatusPhaseError indicates error in Task execution
	TaskStatusPhaseError TaskStatusPhase = "Error"

	// TaskStatusPhaseOnline indicates Task executed without any errors
	TaskStatusPhaseOnline TaskStatusPhase = "Online"

	// TaskStatusPhaseSkipped indicates Task was skipped
	//
	// NOTE:
	//  This can happen if condition to run this task was not met
	TaskStatusPhaseSkipped TaskStatusPhase = "Skipped"

	// TaskStatusPhaseAssertFailed indicates assertion failed
	TaskStatusPhaseAssertFailed TaskStatusPhase = "AssertFailed"

	// TaskStatusPhaseAssertPassed indicates assertion passed
	TaskStatusPhaseAssertPassed TaskStatusPhase = "AssertPassed"
)

// ExecuteStrategy determines if Run tasks need to be executed
// sequentially or without any sequence
type ExecuteStrategy string

const (
	// ExecuteStrategyParallel executes the run tasks in parallel
	//
	// NOTE:
	//	This is the default mode of execution
	ExecuteStrategyParallel ExecuteStrategy = "Parallel"

	// ExecuteStrategySequential executes the run tasks one after
	// the other
	ExecuteStrategySequential ExecuteStrategy = "Sequential"
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

// AssertOperator defines the operator that needs to be applied
// against a list of AssertItem(s)
type AssertOperator string

const (
	// AssertOperatorAND does an AND operation amongst the
	// list of AssertItem(s)
	AssertOperatorAND AssertOperator = "AND"

	// AssertOperatorOR does an OP operation amongst the
	// list of AssertItem(s)
	AssertOperatorOR AssertOperator = "OR"
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
	// Tasks represents a set of tasks that are executed
	// in a level triggered reconciliation loop
	//
	// Tasks is used to achieve the desired state(s) of
	// this Run spec
	Tasks []Task `json:"tasks"`
}

// Task represents the unit of execution for the Run resource
type Task struct {
	// Key to uniquely identify this task in this Run spec
	Key string `json:"key"`

	// A short or verbose description of this task
	Desc string `json:"desc,omitempty"`

	// Proceed with Create or Delete or Update only if this
	// condition succeeds
	//
	// If is optional
	If *Assert `json:"if,omitempty"`

	// Apply defines the desired state that needs to be
	// applied against the Kubernetes cluster
	//
	// Entire resource yaml _(native or custom)_ is embedded
	// here
	//
	// Apply is optional
	Apply map[string]interface{} `json:"desired,omitempty"`

	// Action that needs to be taken against the specified state
	//
	// NOTE:
	// 	Action acts upon the state. Action depends on Assert
	// if set. If Assert fails, then action won't be executed
	// on the state.
	//
	// Action is optional
	Action *Action `json:"action,omitempty"`

	// This implies an update operation. The desired state found
	// in Apply will be applied against the resources selected
	// by this selector
	//
	// NOTE:
	//	One should not try to create or delete along with update
	// in a single task
	For metac.ResourceSelector `json:"for,omitempty"`

	// Assert represents a condition.
	//
	// When used along with state, assert should return a
	// successful match to carry out the state
	//
	// When used without state, assert represents the entire
	// task. In other words, this task becomes a conditional
	// operation.
	//
	// NOTE:
	// 	Assert can not coexist with other fields. In other
	// words if Assert is set then If, Apply & Action should
	// not be set.
	//
	// Assert is optional
	Assert *Assert `json:"assert,omitempty"`
}

// Action to be taken against the resource
type Action struct {
	// Replicas when set to 0 implies **deletion** of resource
	// at the cluster. Similarly, when set to some value that is
	// greater than 1, implies applying multiple copies of the
	// resource specified in **state** field.
	//
	// Default value is 1
	//
	// Replicas is optional
	Replicas *int `json:"replicas,omitempty"`
}

// Assert verifies presence, absence, equals & other checks for
// one or more resources observed in the cluster
type Assert struct {
	// AssertOperator defines the operation that need to be
	// performed against the list of AssertItem
	AssertOperator AssertOperator `json:"assertOperator,omitempty"`

	// Conditions are a list of conditions that gets verified
	// as part of Assert operation
	Conditions []Condition `json:"conditions,omitempty"`
}

// Condition to match, filter, verifying a kubernetes resource.
// When used along with **task.state**, condition should succeed
// to execute action against task.state.
type Condition struct {
	// Selector to filter one or more resources that are expected
	// to be present in the cluster
	ResourceSelector metac.ResourceSelector `json:"resourceSelector,omitempty"`

	// ResourceOperator refers to the operation that gets executed to
	// the selected resources
	//
	// Defaults to 'Exists'
	ResourceOperator ResourceOperator `json:"resourceOperator,omitempty"`

	// Count comes into effect when operator is related to count
	// e.g. EqualsCount, GreaterThanEqualTo, LessThanEqualTo.
	Count *int `json:"count,omitempty"`
}

// OnError provides the details of what needs to be done
// in-case of an error executing Run resource
type OnError struct {
	// Abort execution of Run if this value is true
	Abort *bool `json:"abort,omitempty"`

	// Retry specific details are specified here
	Retry *Retry `json:"retry,omitempty"`
}

// Retry has the details to retry executing a Run resource
type Retry struct {
	// Sleep interval between the retries
	Interval metav1.Duration `json:"interval,omitempty"`

	// Maximum number of times to retry any failed task
	Count int `json:"count,omitempty"`
}

// Execute specifies the way each run tasks should get executed
type Execute struct {
	// Strategy to be followed to execute the tasks
	Stategy ExecuteStrategy `json:"strategy,omitempty"`
}

// RunStatus has the operational state the Run resource
type RunStatus struct {
	// A single word state of Run resource
	Phase string `json:"phase"`

	// A descriptive statement about failure
	Reason string `json:"reason,omitempty"`

	// Warning messages if any
	Warn string `json:"warn,omitempty"`

	// Completion provides current status of each task
	Completion map[string]interface{} `json:"completion"`
}
