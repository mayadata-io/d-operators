package controlplane

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"mayadata.io/d-operators/test/pkg/common"
	"mayadata.io/d-operators/test/pkg/controlplane"
	"mayadata.io/d-operators/test/pkg/types"
	k8s "openebs.io/metac/third_party/kubernetes"
)

// FailedControlPlaneRun defines a control plane run that
// has resulted in error
type FailedControlPlaneRun struct {
	Err string
}

// Error implements error interface
func (e *FailedControlPlaneRun) Error() string {
	return e.Err
}

// RunnerConfig is used to create a new instance of
// ControlPlaneRunner
type RunnerConfig struct {
	Setup types.ControlPlaneSetup
}

// BuildRunner helps creating a kubernetes setup & executes
// tests in it
type BuildRunner struct {
	Setup *types.ControlPlaneSetup

	kubeBinPath string

	cp               *controlplane.ControlPlane
	stopTargetFn     func() error
	setupTeardownFns []func() error

	// err as value
	err error
}

// String implements Stringer interface
func (r *BuildRunner) String() string {
	if r.Setup == nil {
		return "ControlPlaneBuildRunner"
	}
	return r.Setup.GetName()
}

func (r *BuildRunner) addToSetupTeardown(fn func() error) {
	r.setupTeardownFns = append(r.setupTeardownFns, fn)
}

// TeardownSetup brings down the test setup
func (r *BuildRunner) TeardownSetup() {
	// cleanup in descending order
	for i := len(r.setupTeardownFns) - 1; i >= 0; i-- {
		err := r.setupTeardownFns[i]()
		if err != nil {
			// log & continue
			klog.V(1).Infof(
				"Errors during ControlPlane teardown: %s: %+v",
				r,
				err,
			)
		}
	}
}

// evalTargetBinary validates &/ initialises target
// binary field
func (r *BuildRunner) evalTargetBinary() error {
	binary := r.Setup.Spec.Target.Binary
	if binary.Path == "" {
		return errors.Errorf(
			"Missing spec.target.binary.path: %s",
			r,
		)
	}
	if binary.Name == "" {
		return errors.Errorf(
			"Missing spec.target.binary.name: %s",
			r,
		)
	}
	if binary.KubeAPIServerURLFlag == "" {
		return errors.Errorf(
			"Missing spec.target.binary.kubeAPIServerURLFlag: %s",
			r,
		)
	}
	return nil
}

func (r *BuildRunner) evalTargetDeploy() error {
	deploy := r.Setup.Spec.Target.Deploy
	if deploy.Path == "" {
		return errors.Errorf(
			"Missing spec.target.deploy.path: %s",
			r,
		)
	}
	return nil
}

func (r *BuildRunner) evalTestExperiments() error {
	experiments := r.Setup.Spec.Test.Experiments
	if experiments.Path == "" {
		return errors.Errorf(
			"Missing spec.test.experiments.path: %s",
			r,
		)
	}
	return nil
}

func (r *BuildRunner) evalTestInference() error {
	if r.Setup.Spec.Test.Inference.ExperimentName == "" {
		return errors.Errorf(
			"Missing spec.inference.experimentName: %s",
			r,
		)
	}
	if r.Setup.Spec.Test.Inference.ExperimentNamespace == "" {
		return errors.Errorf(
			"Missing spec.inference.experimentNamespace: %s",
			r,
		)
	}
	if r.Setup.Spec.Test.Inference.MaxRetryAttempt == nil {
		r.Setup.Spec.Test.Inference.MaxRetryAttempt =
			k8s.IntPtr(DefaultMaxRetryAttempt)
	}
	if r.Setup.Spec.Test.Inference.RetryIntervalInSeconds == nil {
		r.Setup.Spec.Test.Inference.RetryIntervalInSeconds =
			k8s.IntPtr(DefaultRetryIntervalInSeconds)
	}
	return nil
}

func (r *BuildRunner) initKubeBinPath() error {
	r.kubeBinPath = DefaultKubeControlPlaneBinariesPath
	return nil
}

func (r *BuildRunner) evalKubeSetup() error {
	if r.Setup.Spec.Kubernetes.DownloadURL == "" {
		// set to default if empty
		r.Setup.Spec.Kubernetes.DownloadURL = DefaultK8sDownloadURL
	}
	if r.Setup.Spec.Kubernetes.Version == "" {
		// set to default if empty
		r.Setup.Spec.Kubernetes.Version = DefaultK8sVersion
	}
	return nil
}

func (r *BuildRunner) evalETCDSetup() error {
	if r.Setup.Spec.ETCD.DownloadURL == "" {
		// set to default if empty
		r.Setup.Spec.ETCD.DownloadURL = DefaultETCDDownloadURL
	}
	if r.Setup.Spec.ETCD.Version == "" {
		// set to default if empty
		r.Setup.Spec.ETCD.Version = DefaultETCDVersion
	}
	return nil
}

func (r *BuildRunner) init() error {
	if r.Setup == nil {
		return errors.Errorf("Failed to init: Nil ControlPlaneSetup")
	}
	var fns = []func() error{
		r.initKubeBinPath,
		r.evalKubeSetup,
		r.evalETCDSetup,
		r.evalTargetBinary,
		r.evalTargetDeploy,
		r.evalTestInference,
		r.evalTestExperiments,
	}
	for _, fn := range fns {
		err := fn()
		if err != nil {
			return err
		}
	}
	return nil
}

// NewBuildRunner returns a new instance of BuildRunner
func NewBuildRunner(config RunnerConfig) (*BuildRunner, error) {
	r := &BuildRunner{
		Setup: &types.ControlPlaneSetup{
			TypeMeta:   config.Setup.TypeMeta,
			ObjectMeta: config.Setup.ObjectMeta,
			Spec:       config.Setup.Spec,
		},
	}
	err := r.init()
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (r *BuildRunner) wgetKubectlToDownloadPath() error {
	var cmds = []common.CMDArgs{
		{
			CMD: "wget",
			Args: []string{
				"-q",
				fmt.Sprintf(
					"%s/%s/bin/linux/amd64/kubectl",
					r.Setup.Spec.Kubernetes.DownloadURL,
					r.Setup.Spec.Kubernetes.Version,
				),
			},
		},
		{
			CMD: "chmod",
			Args: []string{
				"+x",
				"kubectl",
			},
		},
		{
			CMD: "mv",
			Args: []string{
				"kubectl",
				r.kubeBinPath,
			},
		},
	}
	for _, cmd := range cmds {
		err := sh.Run(cmd.CMD, cmd.Args...)
		if err != nil {
			return errors.Wrapf(err, fmt.Sprintf("%s", r))
		}
	}
	return nil
}

func (r *BuildRunner) wgetKubectlToDownloadPathIfNotFound() {
	cmd := fmt.Sprintf("%s/kubectl", r.kubeBinPath)
	op, err := sh.Output(cmd, "version", "--client")
	if err == nil {
		klog.V(1).Infof("%s", op)
		return
	}
	if !strings.Contains(err.Error(), NotFoundErrMsg) {
		// this is different than not found error
		r.err = errors.Wrapf(err, fmt.Sprintf("%s", r))
		return
	}
	r.err = r.wgetKubectlToDownloadPath()
}

func (r *BuildRunner) wgetETCDToDownloadPath() error {
	var etcdFileName = fmt.Sprintf("etcd-%s-linux-amd64", r.Setup.Spec.ETCD.Version)
	var etcdFileNameAsTarGZ = fmt.Sprintf("%s.tar.gz", etcdFileName)
	var etcdDownloadURL = fmt.Sprintf(
		"%s/%s/%s",
		r.Setup.Spec.ETCD.DownloadURL,
		DefaultETCDVersion,
		etcdFileNameAsTarGZ,
	)
	var cmds = []common.CMDArgs{
		{
			CMD: "wget",
			Args: []string{
				"-q",
				etcdDownloadURL,
			},
		},
		{
			CMD: "tar",
			Args: []string{
				"-zxf",
				etcdFileNameAsTarGZ,
			},
		},
		{
			CMD: "mv",
			Args: []string{
				fmt.Sprintf("%s/etcd", etcdFileName),
				r.kubeBinPath,
			},
		},
		{
			CMD: "rm",
			Args: []string{
				"-rf",
				etcdFileName,
				etcdFileNameAsTarGZ,
			},
		},
	}
	for _, cmd := range cmds {
		err := sh.Run(cmd.CMD, cmd.Args...)
		if err != nil {
			return errors.Wrapf(err, fmt.Sprintf("%s", r))
		}
	}
	return nil
}

func (r *BuildRunner) wgetETCDToDownloadPathIfNotFound() {
	cmd := fmt.Sprintf("%s/etcd", r.kubeBinPath)
	op, err := sh.Output(cmd, "--version")
	if err == nil {
		klog.V(1).Infof("%s", op)
		return
	}
	if !strings.Contains(err.Error(), NotFoundErrMsg) {
		// this is different than not found error
		r.err = errors.Wrapf(err, fmt.Sprintf("%s", r))
		return
	}
	// download etcd
	r.err = r.wgetETCDToDownloadPath()
}

func (r *BuildRunner) wgetKubeAPIServerToDownloadPath() error {
	var cmds = []common.CMDArgs{
		{
			CMD: "wget",
			Args: []string{
				"-q",
				fmt.Sprintf(
					"%s/%s/bin/linux/amd64/kube-apiserver",
					r.Setup.Spec.Kubernetes.DownloadURL,
					r.Setup.Spec.Kubernetes.Version,
				),
			},
		},
		{
			CMD: "chmod",
			Args: []string{
				"+x",
				"kube-apiserver",
			},
		},
		{
			CMD: "mv",
			Args: []string{
				"kube-apiserver",
				r.kubeBinPath,
			},
		},
	}
	for _, cmd := range cmds {
		err := sh.Run(cmd.CMD, cmd.Args...)
		if err != nil {
			return errors.Wrapf(err, fmt.Sprintf("%s", r))
		}
	}
	return nil
}

func (r *BuildRunner) wgetKubeAPIServerToDownloadPathIfNotFound() {
	cmd := fmt.Sprintf("%s/kube-apiserver", r.kubeBinPath)
	err := sh.Run(cmd, "--version")
	if err == nil {
		return
	}
	if !strings.Contains(err.Error(), NotFoundErrMsg) {
		// this is different than not found error
		r.err = errors.Wrapf(err, fmt.Sprintf("%s", r))
		return
	}
	// download kube api server
	r.err = r.wgetKubeAPIServerToDownloadPath()
}

// startKube starts kubernetes control plane
func (r *BuildRunner) startKube() {
	klog.V(2).Infof("Starting k8s: %s", r)
	r.cp = &controlplane.ControlPlane{
		StartTimeout: 60 * time.Second,
		StopTimeout:  60 * time.Second,
	}
	err := r.cp.Start()
	if err != nil {
		r.err = errors.Wrapf(err, fmt.Sprintf("%s", r))
		return
	}
	r.addToSetupTeardown(r.stopKube)
	klog.V(2).Infof("K8s started successfully: %s", r)
	return
}

// stopKube stops kubernetes control plane
func (r *BuildRunner) stopKube() (err error) {
	if r.cp == nil {
		return
	}
	klog.V(2).Infof("Stopping k8s: %s", r)
	defer func() {
		if err == nil {
			klog.V(2).Infof("K8s stopped successfully: %s", r)
		}
	}()
	err = r.cp.Stop()
	return errors.Wrapf(err, fmt.Sprintf("%s", r))
}

// startTargetBinary starts the target binary
func (r *BuildRunner) startTargetBinary() {
	klog.V(2).Infof("Starting target binary: %s", r)
	cmd := controlplane.NewCommand(
		controlplane.CommandConfig{
			Err: os.Stderr,
			Out: os.Stdout,
		},
	)
	var args []string
	args = append(
		args,
		r.Setup.Spec.Target.Binary.Args...,
	)
	args = append(
		args,
		fmt.Sprintf(
			"%s=%s",
			r.Setup.Spec.Target.Binary.KubeAPIServerURLFlag,
			r.cp.APIURL().String(),
		),
	)
	bin := fmt.Sprintf(
		"%s/%s",
		r.Setup.Spec.Target.Binary.Path,
		r.Setup.Spec.Target.Binary.Name,
	)
	stop, err := cmd.Start(bin, args...)
	if err != nil {
		r.err = errors.Wrapf(err, fmt.Sprintf("%s", r))
		return
	}
	r.stopTargetFn = stop
	r.addToSetupTeardown(r.stopTargetBinary)
	klog.V(2).Infof("Target binary started successfully: %s", r)
}

// stopTargetBinary stops the target binary
func (r *BuildRunner) stopTargetBinary() (err error) {
	if r.stopTargetFn == nil {
		return
	}
	klog.V(2).Infof("Stopping target binary: %s", r)
	defer func() {
		if err == nil {
			klog.V(2).Infof("Target binary stopped successfully: %s", r)
		}
	}()
	return r.stopTargetFn()
}

// applyTargetManifests applies target's manifests
func (r *BuildRunner) applyTargetManifests() {
	klog.V(2).Infof(
		"Applying target manifests: Path %s: %s",
		r.Setup.Spec.Target.Deploy.Path,
		r,
	)
	defer func() {
		if r.err == nil {
			klog.V(2).Infof(
				"Applied target manifests successfully: %s",
				r,
			)
		}
	}()
	r.err = r.cp.KubeCtl().Apply(
		controlplane.ApplyConfig{
			Path:      r.Setup.Spec.Target.Deploy.Path,
			YAMLFiles: r.Setup.Spec.Target.Deploy.Files,
		},
	)
}

// applyTestManifests applies test's manifests
func (r *BuildRunner) applyTestManifests() {
	if r.Setup.Spec.Test.Deploy.Path == "" {
		// Test manifests are optional
		klog.V(3).Infof("Skipping test manifests: Nothing to apply")
		return
	}
	klog.V(2).Infof(
		"Applying test manifests: Path %s: %s",
		r.Setup.Spec.Test.Deploy.Path,
		r,
	)
	defer func() {
		if r.err == nil {
			klog.V(2).Infof(
				"Applied test manifests successfully: %s",
				r,
			)
		}
	}()
	r.err = r.cp.KubeCtl().Apply(
		controlplane.ApplyConfig{
			Path:      r.Setup.Spec.Test.Deploy.Path,
			YAMLFiles: r.Setup.Spec.Test.Deploy.Files,
		},
	)
}

// applyExperiments applies all experiments to test the target
func (r *BuildRunner) applyExperiments() {
	klog.V(2).Infof(
		"Applying experiments: Path %s: %s",
		r.Setup.Spec.Test.Experiments.Path,
		r,
	)
	defer func() {
		if r.err == nil {
			klog.V(2).Infof(
				"Applied experiments successfully: %s",
				r,
			)
		}

	}()
	r.err = r.cp.KubeCtl().Apply(
		controlplane.ApplyConfig{
			Path:      r.Setup.Spec.Test.Experiments.Path,
			YAMLFiles: r.Setup.Spec.Test.Experiments.Files,
		},
	)
}

// waitTillAllExperimentsAreDone waits till targeted experiments
// are completed
func (r *BuildRunner) waitTillAllExperimentsAreDone() {
	var counter int
	maxRetries := *r.Setup.Spec.Test.Inference.MaxRetryAttempt
	if maxRetries <= 0 {
		klog.V(3).Infof(
			"Invalid max retries %d: Setting retries to 1",
			maxRetries,
		)
		maxRetries = 1
	}
	interval := *r.Setup.Spec.Test.Inference.RetryIntervalInSeconds
	intervalInSeconds := time.Duration(interval) * time.Second
	for {
		if counter >= maxRetries {
			break
		}
		counter++
		op, err := r.cp.KubeCtl().RunOp(
			"get",
			ExperimentAPIGroup,
			"-n",
			r.Setup.Spec.Test.Inference.ExperimentNamespace,
			r.Setup.Spec.Test.Inference.ExperimentName,
			"-o=jsonpath='{.status.phase}'",
		)
		if err != nil {
			r.err = err
			return
		}
		if strings.Contains(strings.TrimSpace(op), "Completed") {
			// testing is complete
			return
		}
		if strings.Contains(strings.TrimSpace(op), "Failed") {
			// test has failed; no need to retry
			break
		}
		klog.V(2).Infof(
			"Waiting to infer experiment results: %s / %s has status.phase=%q",
			r.Setup.Spec.Test.Inference.ExperimentNamespace,
			r.Setup.Spec.Test.Inference.ExperimentName,
			op,
		)
		time.Sleep(intervalInSeconds)
	}
	// Experiment used to infer has **failed** or **timed-out**
	// It is wise to display the output & then set error
	op, _ := r.cp.KubeCtl().RunOp(
		"get",
		ExperimentAPIGroup,
		"-n",
		r.Setup.Spec.Test.Inference.ExperimentNamespace,
		r.Setup.Spec.Test.Inference.ExperimentName,
		"-oyaml",
	)
	// Note that there is no verbose level assigned to this log
	// In other words we want this log to be displayed irrespective
	// of log level
	klog.Infof("\n%s", op)
	r.err = &FailedControlPlaneRun{
		Err: "Experiments might have failed or timed out",
	}
}

// printExperimentResultsV prints verbose experiment results on the terminal
func (r *BuildRunner) printExperimentResultsV() {
	if r.err != nil {
		if _, ok := r.err.(*FailedControlPlaneRun); !ok {
			// will not print for runtime errors
			return
		}
	}
	var args = []string{
		"get",
		ExperimentAPIGroup,
		"--all-namespaces",
	}
	selector := r.Setup.Spec.Test.Inference.DisplaySelector
	if len(selector.MatchLabels) != 0 {
		args = append(args, "-l")
		for k, v := range selector.MatchLabels {
			args = append(args, fmt.Sprintf("%s=%s", k, v))
		}
	}
	// ------------------------------------------------------
	// TODO (@amitkumardas):
	// Make use of Spec.Test.Inference.DisplaySelector.Phases
	// as label selector.
	//
	// This needs an enhancement in declarative job. It should
	// set the status.phase value as a label against the job.
	// ------------------------------------------------------

	args = append(args, "-oyaml")
	op, err := r.cp.KubeCtl().RunOp(args...)
	if err != nil {
		// log this error & continue
		//
		// NOTE:
		//  We just log this error as we are not really interested
		// in it
		klog.Infof("%s", err)
		return
	}
	klog.Infof("\n%s", op)
}

// Run builds control plane setup and executes specified
// experiments on this setup
func (r *BuildRunner) Run() error {
	defer func() {
		if rec := recover(); rec != nil {
			klog.Errorf("Recovering from panic: %v", rec)
			return
		}
		r.printExperimentResultsV()
	}()
	var fns = []func(){
		r.wgetETCDToDownloadPathIfNotFound,
		r.wgetKubeAPIServerToDownloadPathIfNotFound,
		r.wgetKubectlToDownloadPathIfNotFound,
		r.startKube,
		r.applyTargetManifests,
		r.startTargetBinary,
		r.applyTestManifests,
		r.applyExperiments,
		r.waitTillAllExperimentsAreDone,
	}
	for _, f := range fns {
		f()
		if r.err != nil {
			return r.err
		}
	}
	return nil
}
