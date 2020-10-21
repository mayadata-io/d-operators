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

package pkg

import (
	"testing"
)

func TestRunnableInit(t *testing.T) {
	var tests = map[string]struct {
		Command         Command
		ExpectedEnabled EnabledWhen
		IsError         bool
	}{
		"no command": {
			ExpectedEnabled: EnabledOnce,
		},
		"with command": {
			Command:         Command{},
			ExpectedEnabled: EnabledOnce,
		},
		"with command + enabled once": {
			Command: Command{
				Spec: CommandSpec{
					Enabled: Enabled{
						When: EnabledOnce,
					},
				},
			},
			ExpectedEnabled: EnabledOnce,
		},
		"with command + enabled never": {
			Command: Command{
				Spec: CommandSpec{
					Enabled: Enabled{
						When: EnabledNever,
					},
				},
			},
			ExpectedEnabled: EnabledNever,
		},
		"with command + enabled always": {
			Command: Command{
				Spec: CommandSpec{
					Enabled: Enabled{
						When: EnabledAlways,
					},
				},
			},
			ExpectedEnabled: EnabledAlways,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			r := &Runnable{
				Command: mock.Command,
			}
			r.init()
			if mock.IsError && r.err == nil {
				t.Fatalf("Expected error got none")
			}
			if !mock.IsError && r.err != nil {
				t.Fatalf("Expected no error got %s", r.err.Error())
			}
			if mock.IsError {
				return
			}
			if mock.ExpectedEnabled != r.enabled {
				t.Fatalf(
					"Expected enable %s got %s",
					mock.ExpectedEnabled,
					r.enabled,
				)
			}
		})
	}
}

func TestRunnableSetStatus(t *testing.T) {
	var tests = map[string]struct {
		Output               map[string]CommandOutput
		Enabled              EnabledWhen
		ExpectedWarnCount    int
		ExpectedErrorCount   int
		ExpectedTimeoutCount int
		ExpectedOutputCount  int
		ExpectedPhase        CommandPhase
	}{
		"no output": {
			ExpectedPhase: CommandPhaseRunning,
		},
		"run once command output": {
			Enabled:       EnabledOnce,
			ExpectedPhase: CommandPhaseCompleted,
		},
		"one error output": {
			Output: map[string]CommandOutput{
				"cmd-1": {
					CMD:   "cmd-1",
					Error: "err",
				},
			},
			ExpectedErrorCount:  1,
			ExpectedPhase:       CommandPhaseError,
			ExpectedOutputCount: 1,
		},
		"one error & one warning output": {
			Output: map[string]CommandOutput{
				"cmd-1": {
					CMD:   "cmd-1",
					Error: "err",
				},
				"cmd-2": {
					CMD:     "cmd-2",
					Warning: "blah blah",
				},
			},
			ExpectedErrorCount:  1,
			ExpectedWarnCount:   1,
			ExpectedPhase:       CommandPhaseError,
			ExpectedOutputCount: 2,
		},
		"one error & one warning & one timeout output": {
			Output: map[string]CommandOutput{
				"cmd-1": {
					CMD:   "cmd-1",
					Error: "err",
				},
				"cmd-2": {
					CMD:     "cmd-2",
					Warning: "blah blah",
				},
				"cmd-3": {
					CMD:      "cmd-3",
					Timedout: true,
				},
			},
			ExpectedErrorCount:   1,
			ExpectedWarnCount:    1,
			ExpectedTimeoutCount: 1,
			ExpectedPhase:        CommandPhaseError,
			ExpectedOutputCount:  3,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			r := Runnable{
				enabled: mock.Enabled,
				Status:  &CommandStatus{},
			}
			r.setStatus(mock.Output)
			if mock.ExpectedPhase != r.Status.Phase {
				t.Fatalf(
					"Expected phase %s got %s",
					mock.ExpectedPhase,
					r.Status.Phase,
				)
			}
			if mock.ExpectedErrorCount != r.Status.Counter.ErrorCount {
				t.Fatalf(
					"Expected error count %d got %d",
					mock.ExpectedErrorCount,
					r.Status.Counter.ErrorCount,
				)
			}
			if mock.ExpectedWarnCount != r.Status.Counter.WarnCount {
				t.Fatalf(
					"Expected warn count %d got %d",
					mock.ExpectedWarnCount,
					r.Status.Counter.WarnCount,
				)
			}
			if mock.ExpectedTimeoutCount != r.Status.Counter.TimeoutCount {
				t.Fatalf(
					"Expected timeout count %d got %d",
					mock.ExpectedTimeoutCount,
					r.Status.Counter.TimeoutCount,
				)
			}
			if mock.ExpectedErrorCount != r.Status.Counter.ErrorCount {
				t.Fatalf(
					"Expected error count %d got %d",
					mock.ExpectedErrorCount,
					r.Status.Counter.ErrorCount,
				)
			}
			if mock.ExpectedOutputCount != len(r.Status.Outputs) {
				t.Fatalf(
					"Expected command output count %d got %d",
					mock.ExpectedOutputCount,
					len(r.Status.Outputs),
				)
			}
		})
	}
}
