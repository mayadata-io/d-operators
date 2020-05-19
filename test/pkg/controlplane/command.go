package controlplane

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

// Command is a wrapper around the exec.Command invocation.
type Command struct {
	// Out, Err specify where Command should write its StdOut &
	// StdErr to.
	//
	// If not specified, the output will be discarded in case of
	// no errors or added to error details in case of errors.
	Out io.Writer
	Err io.Writer

	// out & err buffers are used to provide additional details
	// in case of errors. These buffers are used only if Out & Err
	// are not set by the callers of Command.
	outBuf *bytes.Buffer
	errBuf *bytes.Buffer
}

// CommandConfig is used to create a new instance of Command
type CommandConfig struct {
	Out io.Writer
	Err io.Writer
}

// NewCommand returns a new instance of Command
func NewCommand(config CommandConfig) *Command {
	cmd := &Command{
		Out: config.Out,
		Err: config.Err,
	}
	if cmd.Out == nil {
		cmd.outBuf = new(bytes.Buffer)
	}
	if cmd.Err == nil {
		cmd.errBuf = new(bytes.Buffer)
	}
	return cmd
}

// wrapError wraps the given error with additional information
func (c *Command) wrapErrorf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	var details []string
	details = append(
		details,
		fmt.Sprintf(format, args...),
	)
	if c.errBuf != nil {
		errstr := c.errBuf.String()
		if errstr != "" {
			details = append(details, errstr)
		}
	}
	if c.outBuf != nil {
		outstr := c.outBuf.String()
		if outstr != "" {
			details = append(details, outstr)
		}
	}
	return errors.Wrapf(
		err,
		strings.Join(details, ": "),
	)
}

// build returns a new instance of exec.Cmd
func (c *Command) build(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	if c.Out != nil {
		cmd.Stdout = c.Out
	} else if c.outBuf != nil {
		cmd.Stdout = c.outBuf
	}
	if c.Err != nil {
		cmd.Stderr = c.Err
	} else if c.errBuf != nil {
		cmd.Stderr = c.errBuf
	}
	return cmd
}

// Run executes the given command along with its arguments
func (c *Command) Run(name string, args ...string) error {
	cmd := c.build(name, args...)
	return c.wrapErrorf(
		cmd.Run(),
		"Failed to run %s",
		cmd.Args,
	)
}

// RunOp executes the given command along with its arguments
// and returns the output
func (c *Command) RunOp(name string, args ...string) (output string, err error) {
	cmd := c.build(name, args...)
	err = c.wrapErrorf(
		cmd.Run(),
		"Failed to run %s",
		cmd.Args,
	)
	if c.outBuf != nil {
		output = c.outBuf.String()
	}
	return
}

// Start starts the given command and returns corresponding
// stop function. Start does not wait for the command to
// complete.
func (c *Command) Start(name string, args ...string) (func() error, error) {
	cmd := c.build(name, args...)
	err := cmd.Start()
	if err != nil {
		return nil, c.wrapErrorf(
			err,
			"Failed to start %s",
			cmd.Args,
		)
	}
	klog.V(2).Infof(
		"Command started successfully: %s",
		cmd.Args,
	)
	return func() error {
		pid := cmd.Process.Pid
		klog.V(2).Infof(
			"Stopping command: PID %d: %s",
			pid,
			cmd.Args,
		)
		return c.wrapErrorf(
			cmd.Process.Kill(),
			"Failed to kill command: PID %d: %s",
			pid,
			cmd.Args,
		)
	}, nil
}
