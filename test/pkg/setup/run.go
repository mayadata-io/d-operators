package setup

import (
	"k8s.io/klog/v2"
	"mayadata.io/d-operators/test/pkg/setup/controlplane"
)

// BuildAndRunTestsOnAllControlPlanes builds up the configured
// kubernetes control plane(s) one at a time. After setting up
// a control plane, experiments are run. This control plane is
// then tore down before repeating the process with a new
// control plane setup.
func BuildAndRunTestsOnAllControlPlanes() error {
	controlPlaneSetups, err := controlplane.LoadControlPlaneSetups()
	if err != nil {
		return err
	}
	if len(controlPlaneSetups) == 0 {
		klog.V(1).Infof("No control plane setups found")
	}
	for _, cpSetup := range controlPlaneSetups {
		bRunner, err := controlplane.NewBuildRunner(
			controlplane.RunnerConfig{
				Setup: cpSetup,
			},
		)
		if err != nil {
			return err
		}
		err = bRunner.Run()
		bRunner.TeardownSetup()
		if err != nil {
			return err
		}
	}
	return nil
}

// Run tests on kubernetes
func Run() error {
	return BuildAndRunTestsOnAllControlPlanes()
}
