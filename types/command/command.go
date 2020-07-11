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
type When string

const (
	// Never disables the Command resource from being executed
	Never When = "Never"

	// Always enables the Command resource to get executed periodically
	Always When = "Always"

	// Once enables the Command resource to get executed only once
	Once When = "Once"
)

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
	Enabled          *Enabled          `json:"enabled,omitempty"`
	Template         *Template         `json:"template,omitempty"`
	Env              map[string]string `json:"env,omitempty"`
	ContinueOnError  *bool             `json:"continueOnError,omitempty"`
	TimeoutInSeconds *float64          `json:"timeoutInSeconds,omitempty"`
	Refresh          Refresh           `json:"refresh,omitempty"`
	Commands         []CommandInfo     `json:"commands"`
}

// Enabled determines if the Command is eligible to be executed
// and how often it should get executed
type Enabled struct {
	When When `json:"when,omitempty"`
}

// Refresh options to reconcile Command
type Refresh struct {
	ResyncAfterSeconds *float64 `json:"resyncAfterSeconds,omitempty"`
}

// Template to be used to host the binary that in turn is supposed
// to execute the commands specified in this Command resource
type Template struct {
	Job       *unstructured.Unstructured `json:"job,omitempty"`
	DaemonJob *unstructured.Unstructured `json:"daemonJob,omitempty"`
}

// CommandName is the unique command name
type CommandName string

// CommandInfo has the details of the command that needs to be run
// along with its arguments
//
// NOTE:
//	If a command needs to be executed as a shell script then
// `Sh` field should be populated
type CommandInfo struct {
	Name             CommandName `json:"name"`
	Desc             string      `json:"desc,omitempty"`
	Cmd              []string    `json:"cmd,omitempty"` // either command or script
	Sh               string      `json:"sh,omitempty"`  // either command or script
	TimeoutInSeconds *float64    `json:"timeoutInSeconds,omitempty"`
}

// CommandStatusPhase defines the phase of the command after
// its execution
type CommandStatusPhase string

const (
	// CommandStatusPhaseFailed defines the failed command
	CommandStatusPhaseFailed CommandStatusPhase = "Failed"

	// CommandStatusPhaseCompleted defines the completed command
	CommandStatusPhaseCompleted CommandStatusPhase = "Completed"

	// CommandStatusPhaseRunning defines a running command
	CommandStatusPhaseRunning CommandStatusPhase = "Running"
)

// CommandStatus holds the status of executing the Command resource
type CommandStatus struct {
	Phase              CommandStatusPhase            `json:"phase"`
	Error              string                        `json:"error,omitempty"`
	Warning            string                        `json:"warning,omitempty"`
	Message            string                        `json:"message,omitempty"`
	Timedout           bool                          `json:"timedout"`
	Timeout            string                        `json:"timeout,omitempty"`
	TimetakenInSeconds *float64                      `json:"timetakenInSeconds,omitempty"`
	Outputs            map[CommandName]CommandOutput `json:"outputs,omitempty"`
}

type CommandOutput struct {
	Cmd                string  `json:"cmd"`
	PID                int     `json:"pid"`
	Completed          bool    `json:"completed"`          // false if stopped or signaled
	Timedout           bool    `json:"timedout"`           // true if command timed out
	Exit               int     `json:"exit"`               // exit code of process
	Error              error   `json:"error,omitempty"`    // error during execution if any
	TimetakenInSeconds float64 `json:"timetakenInSeconds"` // zero if Cmd not started
	Stdout             string  `json:"stdout"`             // streamed STDOUT
	Stderr             string  `json:"stderr"`             // streamed STDERR
	Warning            string  `json:"warning,omitempty"`  // warnings if any
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
