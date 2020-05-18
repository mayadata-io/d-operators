// +build mage

package main

import (
	"os"
	"time"

	"github.com/magefile/mage/sh"
)

const (
	operatorBinName = "dope"
	moduleName      = "mayadata.io/d-operators"
)

// allow user to override by running as GOEXE=xxx
var goexe = "go"

// allow user to override by running as OPERATOR_BIN_PATH=xxx
var operatorBinPath = "test/bin"

func initGo() {
	if exe := os.Getenv("GOEXE"); exe != "" {
		goexe = exe
	}
	// We want to use Go 1.11 modules even if the source lives
	// inside GOPATH. The default is "auto".
	os.Setenv("GO111MODULE", "on")
}

func initTargetBinPath() {
	if binpath := os.Getenv("OPERATOR_BIN_PATH"); binpath != "" {
		operatorBinPath = binpath
	}
}

func init() {
	initGo()
	initTargetBinPath()
}

// default environment settings
func defaultEnvs() map[string]string {
	hash, _ := sh.Output("git", "rev-parse", "--short", "HEAD")
	return map[string]string{
		"PACKAGE":     moduleName,
		"COMMIT_HASH": hash,
		"BUILD_DATE":  time.Now().Format("2006-01-02T15:04:05Z0700"),
		"CGO_ENABLED": "0",
		"GOOS":        "linux",
	}
}

// builds d-operators binary
func Dope() error {
	env := defaultEnvs()
	// this build defaults to amd64 architecture
	env["GOARCH"] = "amd64"
	return sh.RunWith(
		env,
		goexe,
		"build",
		"-o",
		operatorBinName,
		"cmd/main.go",
	)
}

// removes d-operators binary
func Clean() error {
	return sh.Run("rm", "-f", operatorBinName)
}

// moves d-operators binary to test location
func moveDopeToTestLoc() error {
	return sh.Run("mv", operatorBinName, operatorBinPath)
}

// manages dependencies before running tests
func TestPrep() error {
	var fns = []func() error{
		Clean,
		Dope,
		moveDopeToTestLoc,
	}
	for _, f := range fns {
		err := f()
		if err != nil {
			return err
		}
	}
	return nil
}
