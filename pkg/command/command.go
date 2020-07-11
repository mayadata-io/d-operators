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

package command

import (
	"errors"
	"fmt"

	types "mayadata.io/d-operators/types/command"
)

// CommandableConfig helps constructing new Commandable instances
type CommandableConfig struct {
	Command *types.Command
}

// Commandable helps executing one or more shell commands
type Commandable struct {
	Command types.Command
	Status  *types.CommandStatus

	// err as value
	err error
}

// NewCommander returns a new instance of Commandable
func NewCommander(config CommandableConfig) (*Commandable, error) {
	if config.Command == nil {
		return nil, errors.New("Failed to initialise: Nil command")
	}
	var enabled types.When = types.Once // default
	if config.Command.Spec.Enabled != nil &&
		config.Command.Spec.Enabled.When != "" {
		enabled = config.Command.Spec.Enabled.When
	}
	config.Command.Spec.Enabled = &types.Enabled{
		When: enabled,
	}
	return &Commandable{
		Command: *config.Command,
		Status:  &types.CommandStatus{},
	}, nil
}

func (c *Commandable) setStatus(out map[types.CommandName]types.CommandOutput) {
	var totalTimetaken float64
	var errCount int
	var warnCount int
	var timedoutCount int
	var isTimedout bool
	for _, op := range out {
		totalTimetaken = totalTimetaken + op.TimetakenInSeconds
		if op.Error != nil {
			errCount++
		}
		if op.Warning != "" {
			warnCount++
		}
		if op.Timedout {
			isTimedout = true
			timedoutCount++
		}
	}
	switch c.Command.Spec.Enabled.When {
	case types.Once:
		c.Status.Phase = types.CommandStatusPhaseCompleted
	default:
		c.Status.Phase = types.CommandStatusPhaseRunning
	}
	if isTimedout {
		c.Status.Timedout = true
		c.Status.Timeout = fmt.Sprintf(
			"Command(s) timed out: timedout-count %d",
			timedoutCount,
		)
	}
	if errCount > 0 {
		c.Status.Phase = types.CommandStatusPhaseFailed
		c.Status.Error = fmt.Sprintf(
			"Error(s) were found while running the command(s): error-count %d",
			errCount,
		)
	}
	if warnCount > 0 {
		c.Status.Warning = fmt.Sprintf(
			"Warnings(s) were found while running the command(s): warn-count %d",
			warnCount,
		)
	}
	c.Status.TimetakenInSeconds = &totalTimetaken
	c.Status.Outputs = out
}

// Run executes the commands in a sequential order
func (c *Commandable) Run() (status *types.CommandStatus, err error) {
	runcmdlist, err := NewRunCmdListFromSpec(c.Command.Spec)
	if err != nil {
		return nil, err
	}
	out := runcmdlist.Run()
	c.setStatus(out)
	return c.Status, nil
}
