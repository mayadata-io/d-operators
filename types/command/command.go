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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// When defines if the command is enabled to run & how often
// it is eligible to run
type EnabledWhen string

const (
	// EnabledNever disables the Command resource from being executed
	EnabledNever EnabledWhen = "Never"

	// EnabledAlways enables the Command resource to get executed periodically
	EnabledAlways EnabledWhen = "Always"

	// EnabledOnce enables the Command resource to get executed only once
	EnabledOnce EnabledWhen = "Once"
)

// EnabledWhenPtr returns the reference of the provided
// enum value
// func EnabledWhenPtr(w EnabledWhen) *EnabledWhen {
// 	return &w
// }

// Enabled determines if the Command is eligible to be executed
// and how often it should get executed
type Enabled struct {
	When EnabledWhen `json:"when,omitempty"`
}

// CommandPhase defines the phase of the command after
// its execution
type CommandPhase string

const (
	// CommandPhaseTimedOut defines the command which timed out
	CommandPhaseTimedOut CommandPhase = "TimedOut"

	// CommandPhaseError defines a failed command
	CommandPhaseError CommandPhase = "Error"

	// CommandPhaseCompleted defines a completed command
	CommandPhaseCompleted CommandPhase = "Completed"

	// CommandPhaseLocked defines a command that has run
	// earlier or is currently being run
	CommandPhaseLocked CommandPhase = "Locked"

	// CommandPhaseSkipped defines a skipped command
	CommandPhaseSkipped CommandPhase = "Skipped"

	// CommandPhaseRunning defines a running command
	CommandPhaseRunning CommandPhase = "Running"

	// CommandPhaseJobCreated defines the command phase which created its
	// child Job
	CommandPhaseJobCreated CommandPhase = "JobCreated"

	// CommandPhaseJobDeleted defines the command phase which deleted its
	// child Job
	CommandPhaseJobDeleted CommandPhase = "JobDeleted"

	// CommandPhaseInProgress defines a in-progress command
	CommandPhaseInProgress CommandPhase = "InProgress"
)

const (
	// JobPhaseCompleted represents a Kubernetes Job with
	// Completed status
	JobPhaseCompleted string = "Completed"

	// KindJob defines the constant to represent the Job's kind
	KindJob string = "Job"

	// JobAPIVersion defines the constant to represent the Job's APIVersion
	JobAPIVersion string = "batch/v1"
)

const (
	// LblKeyIsCommandLock is the label key to determine if the
	// resource is used to lock the reconciliation of Command
	// resource.
	//
	// NOTE:
	// 	This is used to execute reconciliation by only one
	// controller goroutine at a time.
	//
	// NOTE:
	//	A ConfigMap is used as a lock to achieve above behaviour.
	// This ConfigMap will have its labels set with this label key.
	LblKeyIsCommandLock string = "command.dope.mayadata.io/lock"

	// LblKeyCommandName is the label key that identifies the name
	// of the Command that the current resource is associated to
	LblKeyCommandName string = "command.dope.mayadata.io/name"

	// LblKeyCommandUID is the label key that identifies the uid
	// of the Command that the current resource is associated to
	LblKeyCommandUID string = "command.dope.mayadata.io/uid"

	// LblKeyCommandIsController is the label key that identifies
	// if resource is controlled by Command controller
	LblKeyCommandIsController string = "command.dope.mayadata.io/controller"

	// LblKeyCommandPhase is the label key that identifies the phase
	// of the Command. This offers an additional way to determine
	// the phase of the Command apart from Command's status.phase
	// field.
	LblKeyCommandPhase string = "command.dope.mayadata.io/phase"
)

// Resync options to continously reconcile the Command instance
type Resync struct {
	// IntervalInSeconds triggers the next reconciliation of this
	// Command based on this interval
	IntervalInSeconds *int64 `json:"intervalInSeconds,omitempty"`

	// OnErrorResyncInSeconds triggers the next reconciliation of
	// the Command based on this interval if Command's status.phase
	// was set to Error
	OnErrorResyncInSeconds *int64 `json:"onErrorResyncInSeconds,omitempty"`
}

// RetryWhen defines the condition to retry reconciliation of the
// Command resource
type RetryWhen string

const (
	// RetryOnError enables retrying the Command resource if it
	// results in Error
	RetryOnError RetryWhen = "OnError"

	// RetryOnTimeout enables retrying the Command resource if it
	// results in Timeout
	RetryOnTimeout RetryWhen = "OnTimeout"
)

// Retry enables retrying reconciliation of Command resource
type Retry struct {
	When *RetryWhen `json:"when,omitempty"`
}

// Template to be used to host the binary that in turn is supposed
// to execute the commands specified in this Command resource
type Template struct {
	Job *unstructured.Unstructured `json:"job,omitempty"`
}

// Command is a kubernetes custom resource that defines
// the specifications to run one or more commands from
// inside a container
type Command struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   CommandSpec   `json:"spec"`
	Status CommandStatus `json:"status"`
}

// CommandSpec is the specifications of commands that are
// run from inside the container
type CommandSpec struct {
	Enabled  Enabled           `json:"enabled,omitempty"`
	Template Template          `json:"template,omitempty"`
	Env      map[string]string `json:"env,omitempty"`

	// Should all / subsequent commands get executed
	// even in case of errors or timeouts executing
	// current command
	MustRunAllCommands *bool `json:"mustRunAllCommands,omitempty"`

	// Timeout applicable for all commands
	TimeoutInSeconds *int64 `json:"timeoutInSeconds,omitempty"`
	Resync           Resync `json:"resync,omitempty"`
	Retry            Retry  `json:"retry,omitempty"`

	Commands []CommandInfo `json:"commands"`
}

// CommandInfo has the details of the command that needs to be
// run
//
// NOTE:
//	If a command needs to be executed as a shell script then
// `script` field should be populated
type CommandInfo struct {
	Name             string   `json:"name"`
	Desc             string   `json:"desc,omitempty"`
	CMD              []string `json:"cmd,omitempty"`    // shell command
	Script           string   `json:"script,omitempty"` // shell script
	TimeoutInSeconds *int64   `json:"timeoutInSeconds,omitempty"`
}

// CommandStatus holds the status of executing the Command resource
type CommandStatus struct {
	Phase   CommandPhase `json:"phase"`
	Reason  string       `json:"reason,omitempty"`
	Message string       `json:"message,omitempty"`
	// Warning       string                   `json:"warning,omitempty"`
	Timedout bool `json:"timedout"`
	// Timeout       string                   `json:"timeout,omitempty"`
	Counter       Counter                  `json:"counter"`
	ExecutionTime ExecutionTime            `json:"timetakenInSeconds"`
	Outputs       map[string]CommandOutput `json:"outputs,omitempty"`
}

// Counter holds the count of commands that resulted in
// warnings, errors, etc.
type Counter struct {
	WarnCount    int `json:"warnCount"`
	ErrorCount   int `json:"errorCount"`
	TimeoutCount int `json:"timeoutCount"`
}

// ExecutionTime represents the time taken to execute a command
type ExecutionTime struct {
	ValueInSeconds float64 `json:"valueInSeconds"` // zero if CMD did not start
	ReadableValue  string  `json:"readableValue"`
}

// CommandOutput is the result after running the Command
type CommandOutput struct {
	CMD           string        `json:"cmd"`
	PID           int           `json:"pid"`
	Completed     bool          `json:"completed"`       // false if stopped or signaled
	Timedout      bool          `json:"timedout"`        // true if command timed out
	Exit          int           `json:"exit"`            // exit code of process
	Error         string        `json:"error,omitempty"` // error during execution if any
	ExecutionTime ExecutionTime `json:"executionTime"`
	Stdout        string        `json:"stdout"`            // streamed STDOUT
	Stderr        string        `json:"stderr"`            // streamed STDERR
	Warning       string        `json:"warning,omitempty"` // warnings if any
}

// String implements the Stringer interface
func (jr CommandStatus) String() string {
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
