// +build mage

package test

import (
	"time"

	"github.com/magefile/mage/mg"
	"k8s.io/klog"
	"openebs.io/metac/test/integration/framework"
)

var cp *framework.ControlPlane

// Start kubernetes control plane
func StartKube() error {
	klog.V(2).Infof("Will start k8s")
	cp = framework.ControlPlane{
		StartTimeout: 60 * time.Second,
		StopTimeout:  60 * time.Second,
	}
	err := cp.Start()
	if err != nil {
		return err
	}
	klog.V(2).Infof("k8s started successfully")
	return nil
}

// Run integration tests
func Run() error {
	mg.Deps(StartKube)
	defer cp.Stop()

	// kubectl the namespace, crds, rbac
	// build d-operator binary
	// start d-operator binary
	// kubectl test files from artifacts
	// list passed test files
	// list failed test files
	// print -yaml of failed test files based on count
}
