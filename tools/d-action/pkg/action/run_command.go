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

package action

import (
	"fmt"
	"time"

	types "mayadata.io/d-operators/types/command"
)

// RunnableConfig helps constructing a new instance of Runnable
type RunnableConfig struct {
	Command types.Command
}

// Runnable helps executing one or more commands
// e.g. shell / script commands
type Runnable struct {
	Command types.Command
	Status  *types.CommandStatus

	// determines if Command resource can be reconciled more than once
	enabled types.EnabledWhen

	// err as value
	err error
}

func (r *Runnable) init() {
	// enabled defauls to Once i.e. Command can reconcile only once
	r.enabled = types.EnabledOnce
	// override with user specified value if set
	if r.Command.Spec.Enabled.When != "" {
		r.enabled = r.Command.Spec.Enabled.When
	}
}

// NewRunner returns a new instance of Runnable
func NewRunner(config RunnableConfig) (*Runnable, error) {
	r := &Runnable{
		Command: config.Command,
		Status:  &types.CommandStatus{},
	}
	r.init()
	return r, nil
}

func (r *Runnable) setStatus(out map[string]types.CommandOutput) {
	var totalTimetaken float64
	for _, op := range out {
		totalTimetaken = totalTimetaken + op.ExecutionTime.ValueInSeconds
		if op.Error != "" {
			r.Status.Counter.ErrorCount++
		}
		if op.Warning != "" {
			r.Status.Counter.WarnCount++
		}
		if op.Timedout {
			r.Status.Counter.TimeoutCount++
		}
	}
	switch r.enabled {
	case types.EnabledOnce:
		// Command that is meant to run only once is initialised to
		// Completed phase
		r.Status.Phase = types.CommandPhaseCompleted
	default:
		// Command that is meant to be run periodically is initialised
		// to Running phase
		r.Status.Phase = types.CommandPhaseRunning
	}
	if r.Status.Counter.TimeoutCount > 0 {
		r.Status.Phase = types.CommandPhaseTimedOut
		r.Status.Timedout = true
		r.Status.Reason = fmt.Sprintf(
			"%d timeout(s) found",
			r.Status.Counter.TimeoutCount,
		)
	}
	if r.Status.Counter.ErrorCount > 0 {
		r.Status.Phase = types.CommandPhaseError
		r.Status.Reason = fmt.Sprintf(
			"%d error(s) found",
			r.Status.Counter.ErrorCount,
		)
	}
	totalTimeTakenSecs := time.Duration(totalTimetaken) * time.Second
	totalTimeTakenSecsFmt := totalTimeTakenSecs.Round(time.Millisecond).String()
	r.Status.ExecutionTime = types.ExecutionTime{
		ValueInSeconds: totalTimeTakenSecs.Seconds() + 0.0001,
		ReadableValue:  totalTimeTakenSecsFmt,
	}
	r.Status.Outputs = out
}

// Run executes the commands in a sequential order
func (r *Runnable) Run() (status *types.CommandStatus, err error) {
	if r.enabled == types.EnabledNever {
		return &types.CommandStatus{
			Phase:   types.CommandPhaseSkipped,
			Message: "Resource is not enabled",
		}, nil
	}
	runcmdlist, err := NewShellListRunner(r.Command.Spec)
	if err != nil {
		return nil, err
	}
	out := runcmdlist.Run()
	r.setStatus(out)
	return r.Status, nil
}
