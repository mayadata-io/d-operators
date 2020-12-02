// +build !integration

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
	"testing"

	types "mayadata.io/d-operators/types/command"
)

func TestRunnableInit(t *testing.T) {
	var tests = map[string]struct {
		Command         types.Command
		ExpectedEnabled types.EnabledWhen
		IsError         bool
	}{
		"no command": {
			ExpectedEnabled: types.EnabledOnce,
		},
		"with command": {
			Command:         types.Command{},
			ExpectedEnabled: types.EnabledOnce,
		},
		"with command + enabled once": {
			Command: types.Command{
				Spec: types.CommandSpec{
					Enabled: types.Enabled{
						When: types.EnabledOnce,
					},
				},
			},
			ExpectedEnabled: types.EnabledOnce,
		},
		"with command + enabled never": {
			Command: types.Command{
				Spec: types.CommandSpec{
					Enabled: types.Enabled{
						When: types.EnabledNever,
					},
				},
			},
			ExpectedEnabled: types.EnabledNever,
		},
		"with command + enabled always": {
			Command: types.Command{
				Spec: types.CommandSpec{
					Enabled: types.Enabled{
						When: types.EnabledAlways,
					},
				},
			},
			ExpectedEnabled: types.EnabledAlways,
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
		Output               map[string]types.CommandOutput
		Enabled              types.EnabledWhen
		ExpectedWarnCount    int
		ExpectedErrorCount   int
		ExpectedTimeoutCount int
		ExpectedOutputCount  int
		ExpectedPhase        types.CommandPhase
	}{
		"no output": {
			ExpectedPhase: types.CommandPhaseRunning,
		},
		"run once command output": {
			Enabled:       types.EnabledOnce,
			ExpectedPhase: types.CommandPhaseCompleted,
		},
		"one error output": {
			Output: map[string]types.CommandOutput{
				"cmd-1": {
					CMD:   "cmd-1",
					Error: "err",
				},
			},
			ExpectedErrorCount:  1,
			ExpectedPhase:       types.CommandPhaseError,
			ExpectedOutputCount: 1,
		},
		"one error & one warning output": {
			Output: map[string]types.CommandOutput{
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
			ExpectedPhase:       types.CommandPhaseError,
			ExpectedOutputCount: 2,
		},
		"one error & one warning & one timeout output": {
			Output: map[string]types.CommandOutput{
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
			ExpectedPhase:        types.CommandPhaseError,
			ExpectedOutputCount:  3,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			r := Runnable{
				enabled: mock.Enabled,
				Status:  &types.CommandStatus{},
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
