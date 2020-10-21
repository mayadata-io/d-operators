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

package pkg

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-cmd/cmd"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

// RunnableShell represents the shell command that gets executed
//
// This has borrowed concepts from cluster api project
// ref - https://github.com/kubernetes-sigs/cluster-api/blob/master/test/infrastructure/docker/cloudinit/runcmd.go
type RunnableShell struct {
	Name             string
	CMD              string
	Args             []string
	Stdin            string
	TimeoutInSeconds *int64
}

// NewShellRunner returns a new instance of RunnableShell
func NewShellRunner(cmdInfo CommandInfo) (RunnableShell, error) {
	if cmdInfo.Name == "" {
		return RunnableShell{},
			errors.New("Invalid command: Missing name")
	}
	if len(cmdInfo.CMD) != 0 && cmdInfo.Script != "" {
		return RunnableShell{},
			errors.Errorf(
				"Invalid command %q: Both cmd & script are not allowed",
				cmdInfo.Name,
			)
	}
	if len(cmdInfo.CMD) == 0 && cmdInfo.Script == "" {
		return RunnableShell{},
			errors.Errorf(
				"Invalid command %q: Either cmd or script must be set",
				cmdInfo.Name,
			)
	}
	rs := &RunnableShell{
		Name: cmdInfo.Name,
	}
	if len(cmdInfo.CMD) > 0 {
		rs.CMD = cmdInfo.CMD[0]
		if len(cmdInfo.CMD) > 1 {
			rs.Args = cmdInfo.CMD[1:]
		}
	} else {
		rs.CMD = "/bin/sh"
		rs.Args = []string{"-c", cmdInfo.Script}
	}
	return *rs, nil
}

// RunnableShellList defines the list of shell commands that
// gets executed in the given order
type RunnableShellList struct {
	// list of commands executed in order
	Items []RunnableShell

	// set of env variables that can be used by all commands
	Env map[string]string

	// default timeout for all
	TimeoutInSeconds int64

	ContinueOnError bool
}

// NewShellListRunner returns a new RunnableShellList from the provided
// command specifications
func NewShellListRunner(spec CommandSpec) (RunnableShellList, error) {
	var timeoutInSecs int64 = 300 // defaults to 5 minutes
	var continueOnErr bool        // defaults to false

	if spec.TimeoutInSeconds != nil && *spec.TimeoutInSeconds > 0 {
		timeoutInSecs = *spec.TimeoutInSeconds
	}
	if spec.MustRunAllCommands != nil {
		continueOnErr = *spec.MustRunAllCommands
	}
	cmdlist := &RunnableShellList{
		Env:              spec.Env,
		TimeoutInSeconds: timeoutInSecs,
		ContinueOnError:  continueOnErr,
	}

	for _, cmd := range spec.Commands {
		rc, err := NewShellRunner(cmd)
		if err != nil {
			return RunnableShellList{}, err
		}
		cmdlist.Items = append(cmdlist.Items, rc)
	}
	return *cmdlist, nil
}

func (l RunnableShellList) getCMDEnv() (out []string) {
	for k, v := range l.Env {
		out = append(out, fmt.Sprintf("%s=%s", k, v))
	}
	return
}

// Run executes all the shell commands in order
func (l RunnableShellList) Run() map[string]CommandOutput {
	var timeoutInSecs int64
	// initialise the run output
	output := make(map[string]CommandOutput, len(l.Items))
	cmdEnv := l.getCMDEnv()

	// commands are executed serially one by one
	for _, rc := range l.Items {
		var stdout, stderr strings.Builder
		var stoperr error
		var warn string
		var isTimeout bool

		// default to global setting
		timeoutInSecs = l.TimeoutInSeconds
		if rc.TimeoutInSeconds != nil && *rc.TimeoutInSeconds > 0 {
			timeoutInSecs = *rc.TimeoutInSeconds
		}

		// Disable output buffering, enable streaming
		options := cmd.Options{
			Buffered:  false,
			Streaming: true,
		}
		execCmd := cmd.NewCmdOptions(options, rc.CMD, rc.Args...)
		// set environment variables
		execCmd.Env = cmdEnv

		doneChan := make(chan struct{})

		// Capture STDOUT and STDERR lines streaming due to execution
		// of command
		go func() {
			defer close(doneChan)

			timeout := time.Duration(timeoutInSecs) * time.Second
			timeoutChan := time.After(timeout) // does not block

			for execCmd.Stdout != nil || execCmd.Stderr != nil {
				select {
				case line, open := <-execCmd.Stdout:
					if !open {
						execCmd.Stdout = nil
						continue
					}
					// build the output
					stdout.WriteString(line)
				case line, open := <-execCmd.Stderr:
					if !open {
						execCmd.Stderr = nil
						continue
					}
					// build the error
					stderr.WriteString(line)
				case <-timeoutChan:
					klog.V(1).Infof(
						"Command timed out: Name %q: Timeout %s",
						rc.Name,
						timeout.Round(time.Millisecond).String(),
					)
					isTimeout = true
					execCmd.Stdout = nil
					execCmd.Stderr = nil
					stoperr = execCmd.Stop()
				}
			}
		}()

		// Run and wait for command to return
		statusChan := <-execCmd.Start()
		// Wait for goroutine to capture stdout &/or stderr
		<-doneChan

		if stoperr != nil {
			// Any error while stopping the command is considered
			// as a warning
			warn = stoperr.Error()
		}

		timeTaken := time.Duration(statusChan.Runtime) * time.Second
		timeTakenFmt := timeTaken.Round(time.Millisecond).String()
		var statusChanErr = ""
		if statusChan.Error != nil {
			statusChanErr = statusChan.Error.Error()
		}
		output[rc.Name] = CommandOutput{
			CMD:       statusChan.Cmd,
			Completed: statusChan.Complete,
			Timedout:  isTimeout,
			Error:     statusChanErr,
			Exit:      statusChan.Exit,
			PID:       statusChan.PID,
			Stderr:    stderr.String(),
			Stdout:    stdout.String(),
			ExecutionTime: ExecutionTime{
				ValueInSeconds: timeTaken.Seconds() + 0.0001,
				ReadableValue:  timeTakenFmt,
			},
			Warning: warn,
		}
		// check for errors
		if isTimeout || stderr.Len() != 0 || statusChanErr != "" {
			// verify if logic should continue or break out
			if !l.ContinueOnError {
				break
			}
		}
	}
	return output
}
