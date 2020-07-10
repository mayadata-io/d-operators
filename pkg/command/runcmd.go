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
	"fmt"
	"strings"
	"time"

	"github.com/go-cmd/cmd"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	types "mayadata.io/d-operators/types/command"
)

// RunCmd represents the shell command that gets executed
//
// This has borrowed concepts from cluster api project
// ref - https://github.com/kubernetes-sigs/cluster-api/blob/master/test/infrastructure/docker/cloudinit/runcmd.go
type RunCmd struct {
	Name             string
	Cmd              string
	Args             []string
	Stdin            string
	TimeoutInSeconds *float64
}

// NewRunCmdFromCommand returns a new RunCmd from the
// given command info
func NewRunCmdFromCommand(cmdInfo types.CommandInfo) (RunCmd, error) {
	if cmdInfo.Name == "" {
		return RunCmd{}, errors.New("Invalid command: Missing name")
	}
	if len(cmdInfo.Cmd) != 0 && cmdInfo.Sh != "" {
		return RunCmd{},
			errors.Errorf(
				"Invalid command %q: Both Cmd & Sh are not allowed",
				cmdInfo.Name,
			)
	}
	if len(cmdInfo.Cmd) == 0 && cmdInfo.Sh == "" {
		return RunCmd{},
			errors.Errorf(
				"Invalid command %q: Either Cmd or Sh must be set",
				cmdInfo.Name,
			)
	}
	rc := &RunCmd{
		Name: string(cmdInfo.Name),
	}
	if len(cmdInfo.Cmd) > 0 {
		rc.Cmd = cmdInfo.Cmd[0]
		if len(cmdInfo.Cmd) > 1 {
			rc.Args = cmdInfo.Cmd[1:]
		}
	} else {
		rc.Cmd = "/bin/sh"
		rc.Args = []string{"-c", cmdInfo.Sh}
	}
	return *rc, nil
}

// RunCmdList defines the list of shell commands that gets executed
// in the given order
type RunCmdList struct {
	Items            []RunCmd          // commands executed in order
	Env              map[string]string // env for all
	TimeoutInSeconds float64           // default timeout for all
	ContinueOnError  bool
}

// NewRunCmdListFromSpec returns a new RunCmdList from the provided
// command specifications
func NewRunCmdListFromSpec(spec types.CommandSpec) (RunCmdList, error) {
	var timeoutInSecs float64 = 300 // defaults to 5 minutes
	var continueOnErr bool          // defaults to false

	if spec.TimeoutInSeconds != nil && *spec.TimeoutInSeconds > 0 {
		timeoutInSecs = *spec.TimeoutInSeconds
	}
	if spec.ContinueOnError != nil {
		continueOnErr = *spec.ContinueOnError
	}
	cmdlist := &RunCmdList{
		Env:              spec.Env,
		TimeoutInSeconds: timeoutInSecs,
		ContinueOnError:  continueOnErr,
	}

	for _, cmd := range spec.Commands {
		rc, err := NewRunCmdFromCommand(cmd)
		if err != nil {
			return RunCmdList{}, err
		}
		cmdlist.Items = append(cmdlist.Items, rc)
	}
	return *cmdlist, nil
}

func (l RunCmdList) toCmdEnv() (out []string) {
	for k, v := range l.Env {
		out = append(out, fmt.Sprintf("%s=%s", k, v))
	}
	return
}

// Run executes all the commands in order
func (l RunCmdList) Run() map[types.CommandName]types.CommandOutput {
	output := make(map[types.CommandName]types.CommandOutput, len(l.Items))
	cmdEnv := l.toCmdEnv()
	var timeoutInSecs float64

	// commands are executed serially one by one
	for _, rc := range l.Items {
		var stdout, stderr strings.Builder
		var stoperr error
		var warn string
		var isTimeout bool

		timeoutInSecs = l.TimeoutInSeconds // defaults to global setting
		if rc.TimeoutInSeconds != nil && *rc.TimeoutInSeconds > 0 {
			timeoutInSecs = *rc.TimeoutInSeconds
		}

		// Disable output buffering, enable streaming
		options := cmd.Options{
			Buffered:  false,
			Streaming: true,
		}
		execCmd := cmd.NewCmdOptions(options, rc.Cmd, rc.Args...)
		execCmd.Env = cmdEnv

		doneChan := make(chan struct{})

		// Capture STDOUT and STDERR lines streaming from Cmd
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
					stdout.WriteString(line)
				case line, open := <-execCmd.Stderr:
					if !open {
						execCmd.Stderr = nil
						continue
					}
					stderr.WriteString(line)
				case <-timeoutChan:
					klog.V(1).Infof(
						"Command timed out: Name %q: Timeout %d",
						rc.Name,
						timeout,
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
			// stop err is treated as a warning
			warn = stoperr.Error()
		}

		output[types.CommandName(rc.Name)] = types.CommandOutput{
			Cmd:                statusChan.Cmd,
			Completed:          statusChan.Complete,
			Timedout:           isTimeout,
			Error:              statusChan.Error,
			Exit:               statusChan.Exit,
			PID:                statusChan.PID,
			Stderr:             stderr.String(),
			Stdout:             stdout.String(),
			TimetakenInSeconds: statusChan.Runtime,
			Warning:            warn,
		}
		if isTimeout || stderr.Len() != 0 {
			if !l.ContinueOnError {
				break
			}
		}
	}
	return output
}
