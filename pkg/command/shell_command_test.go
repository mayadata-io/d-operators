/*
Copyright 2020 The MayaData Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package command

import (
	"testing"

	"mayadata.io/d-operators/common/pointer"
	types "mayadata.io/d-operators/types/command"
)

func TestNewShellRunner(t *testing.T) {
	var tests = map[string]struct {
		CommandInfo       types.CommandInfo
		ExpectedCMD       string
		ExpectedArgsCount int
		IsError           bool
	}{
		"no command info": {
			IsError: true,
		},
		"missing command name": {
			CommandInfo: types.CommandInfo{},
			IsError:     true,
		},
		"missing cmd & script": {
			CommandInfo: types.CommandInfo{
				Name: "testit",
			},
			IsError: true,
		},
		"cmd as well as script": {
			CommandInfo: types.CommandInfo{
				Name: "testing",
				CMD: []string{
					"ls",
					"-ltr",
				},
				Script: "ls -ltr",
			},
			IsError: true,
		},
		"with cmd": {
			CommandInfo: types.CommandInfo{
				Name: "testing",
				CMD: []string{
					"ls",
					"-ltr",
				},
			},
			ExpectedArgsCount: 1,
			ExpectedCMD:       "ls",
		},
		"with script": {
			CommandInfo: types.CommandInfo{
				Name:   "testing",
				Script: "ls -ltr",
			},
			ExpectedArgsCount: 2,
			ExpectedCMD:       "/bin/sh",
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			got, err := NewShellRunner(mock.CommandInfo)
			if mock.IsError && err == nil {
				t.Fatalf("Expected error got none")
			}
			if !mock.IsError && err != nil {
				t.Fatalf("Expected no error got %s", err.Error())
			}
			if mock.IsError {
				return
			}
			if mock.ExpectedCMD != got.CMD {
				t.Fatalf(
					"Expected CMD %s got %s",
					mock.ExpectedCMD,
					got.CMD,
				)
			}
			if mock.ExpectedArgsCount != len(got.Args) {
				t.Fatalf(
					"Expected args count %d got %d",
					mock.ExpectedArgsCount,
					len(got.Args),
				)
			}
		})
	}
}

func TestNewShellListRunner(t *testing.T) {
	var tests = map[string]struct {
		Spec                     types.CommandSpec
		ExpectedENVCount         int
		ExpectedItemCount        int
		ExpectedTimeoutInSeconds int64
		IsContinueOnError        bool
		IsError                  bool
	}{
		"no spec": {
			ExpectedTimeoutInSeconds: 300,
		},
		"with timeout": {
			Spec: types.CommandSpec{
				TimeoutInSeconds: pointer.Int64(20),
			},
			ExpectedTimeoutInSeconds: 20,
		},
		"with continue on error": {
			Spec: types.CommandSpec{
				MustRunAllCommands: pointer.Bool(true),
			},
			IsContinueOnError:        true,
			ExpectedTimeoutInSeconds: 300,
		},
		"with env": {
			Spec: types.CommandSpec{
				Env: map[string]string{
					"IP": "12.12.1.1",
				},
			},
			ExpectedENVCount:         1,
			ExpectedTimeoutInSeconds: 300,
		},
		"with one invalid command": {
			Spec: types.CommandSpec{
				Commands: []types.CommandInfo{
					{
						Name: "cmd-01",
					},
				},
			},
			IsError: true,
		},
		"with one valid command": {
			Spec: types.CommandSpec{
				Commands: []types.CommandInfo{
					{
						Name: "cmd-01",
						CMD: []string{
							"ls",
							"-ltr",
						},
					},
				},
			},
			ExpectedItemCount:        1,
			ExpectedTimeoutInSeconds: 300,
		},
		"with one valid & one invalid command": {
			Spec: types.CommandSpec{
				Commands: []types.CommandInfo{
					{
						Name: "cmd-01",
						CMD: []string{
							"ls",
							"-ltr",
						},
					},
					{
						Name: "missing-cmd-02",
					},
				},
			},
			IsError: true,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			got, err := NewShellListRunner(mock.Spec)
			if mock.IsError && err == nil {
				t.Fatalf("Expected error got none")
			}
			if !mock.IsError && err != nil {
				t.Fatalf("Expected no error got %s", err.Error())
			}
			if mock.IsError {
				return
			}
			if mock.ExpectedENVCount != len(got.Env) {
				t.Fatalf(
					"Expected env count %d got %d",
					mock.ExpectedENVCount,
					len(got.Env),
				)
			}
			if mock.ExpectedItemCount != len(got.Items) {
				t.Fatalf(
					"Expected item count %d got %d",
					mock.ExpectedItemCount,
					len(got.Items),
				)
			}
			if mock.ExpectedTimeoutInSeconds != got.TimeoutInSeconds {
				t.Fatalf(
					"Expected timeout %d got %d",
					mock.Spec.TimeoutInSeconds,
					got.TimeoutInSeconds,
				)
			}
			if mock.IsContinueOnError != got.ContinueOnError {
				t.Fatalf(
					"Expected continue on error %t got %t",
					mock.IsContinueOnError,
					got.ContinueOnError,
				)
			}
		})
	}
}
