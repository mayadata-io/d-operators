package runner

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
	"mayadata.io/d-operators/test/pkg/controlplane"
)

// FailedRun defines a run that has resulted in error
type FailedRun struct {
	Err string
}

// Error implements error interface
func (fr *FailedRun) Error() string {
	return fr.Err
}

const (
	notFoundErrMsg    = "no such file or directory"
	inferenceTestName = "inference"
)

// name of the namespace that host all the declarative tests
var declarativeTestNamespace = "d-testing"

// default kubernetes download url
var k8sDownloadURL = "https://dl.k8s.io"

// default etcd download url
var etcdDownloadURL = "https://github.com/coreos/etcd/releases/download"

// default etcd version
var etcdVersion = "v3.4.3"

// default kubernetes version
var k8sVersion = "v1.16.4"

func loadAllTestSetups() (SetupConfigs, error) {
	l := &SetupConfigsLoader{Path: "setup/"}
	return l.Load()
}

type cmdargs struct {
	cmd  string
	args []string
}

// TestRunnerConfig is used to create a new instance of TestRunner
type TestRunnerConfig struct {
	SetupConfig *unstructured.Unstructured
}

// TestRunner helps creating a kubernetes setup & executing tests in it
type TestRunner struct {
	SetupConfig *unstructured.Unstructured

	operatorBinPath string
	operatorBinName string
	operatorBinArgs []string

	// flag to connect operator binary to k8s apiserver
	kubeAPIServerURLFlag string

	operatorManifestsPath string
	operatorManifestFiles []string

	// operatorDeployPath string
	// operatorDeployFile string

	operatorTestsPath string
	operatorTestFiles []string

	kubeBinPath string

	k8sDownloadURL string
	k8sVersion     string

	etcdVersion     string
	etcdDownloadURL string

	cp *controlplane.ControlPlane

	stopOperatorFn func() error

	setupTeardownFns []func() error

	// err as value
	err error
}

// String implements Stringer interface
func (r *TestRunner) String() string {
	if r.SetupConfig == nil {
		return "testrunner"
	}
	return r.SetupConfig.GetName()
}

func (r *TestRunner) addToSetupTeardown(fn func() error) {
	r.setupTeardownFns = append(r.setupTeardownFns, fn)
}

// TeardownSetup brings down the test setup
func (r *TestRunner) TeardownSetup() {
	// cleanup in descending order
	for i := len(r.setupTeardownFns) - 1; i >= 0; i-- {
		err := r.setupTeardownFns[i]()
		if err != nil {
			// log & continue
			klog.V(1).Infof(
				"Errors during setup teardown: %s: %+v",
				r,
				err,
			)
		}
	}
}

func (r *TestRunner) initOperatorBinaryFields() error {
	binfields, found, err := unstructured.NestedMap(
		r.SetupConfig.UnstructuredContent(),
		"spec",
		"operator",
		"binary",
	)
	if err != nil {
		return err
	}
	if !found {
		return errors.Errorf(
			"Missing spec.operator.binary: %s",
			r,
		)
	}
	var ok bool
	r.operatorBinPath, ok = binfields["path"].(string)
	if !ok {
		return errors.Errorf(
			"Invalid spec.operator.binary.path: %s",
			r,
		)
	}
	r.operatorBinName, ok = binfields["name"].(string)
	if !ok {
		return errors.Errorf(
			"Invalid spec.operator.binary.name: %s",
			r,
		)
	}
	binArgs, ok := binfields["args"].([]interface{})
	if !ok {
		return errors.Errorf(
			"Invalid spec.operator.binary.args: %s",
			r,
		)
	}
	for _, arg := range binArgs {
		r.operatorBinArgs = append(
			r.operatorBinArgs,
			arg.(string),
		)
	}
	r.kubeAPIServerURLFlag, ok = binfields["kubeAPIServerURLFlag"].(string)
	if !ok {
		return errors.Errorf(
			"Invalid spec.operator.binary.kubeAPIServerURLFlag: %s",
			r,
		)
	}
	return nil
}

func (r *TestRunner) initOperatorManifestFields() error {
	manifestfields, found, err := unstructured.NestedMap(
		r.SetupConfig.UnstructuredContent(),
		"spec",
		"operator",
		"manifest",
	)
	if err != nil {
		return err
	}
	if !found {
		return errors.Errorf(
			"Missing spec.operator.manifest: %s",
			r,
		)
	}
	var ok bool
	r.operatorManifestsPath, ok = manifestfields["path"].(string)
	if !ok {
		return errors.Errorf(
			"Invalid spec.operator.manifest.path: %s",
			r,
		)
	}
	files, ok := manifestfields["files"].([]interface{})
	if !ok {
		return errors.Errorf(
			"Invalid spec.operator.manifest.files: %s",
			r,
		)
	}
	for _, file := range files {
		r.operatorManifestFiles = append(
			r.operatorManifestFiles,
			file.(string),
		)
	}
	return nil
}

func (r *TestRunner) initOperatorTestFields() error {
	testfields, found, err := unstructured.NestedMap(
		r.SetupConfig.UnstructuredContent(),
		"spec",
		"operator",
		"tests",
	)
	if err != nil {
		return err
	}
	if !found {
		return errors.Errorf(
			"Missing spec.operator.tests: %s",
			r,
		)
	}
	var ok bool
	r.operatorTestsPath, ok = testfields["path"].(string)
	if !ok {
		return errors.Errorf(
			"Invalid spec.operator.tests.path: %s",
			r,
		)
	}
	if testfields["files"] != nil {
		files, ok := testfields["files"].([]interface{})
		if !ok {
			return errors.Errorf(
				"Invalid spec.operator.tests.files: %s",
				r,
			)
		}
		for _, file := range files {
			r.operatorTestFiles = append(
				r.operatorTestFiles,
				file.(string),
			)
		}
	}
	return nil
}

// func (r *testrunner) initOperatorDeployFields() error {
// 	deployfields, found, err := unstructured.NestedMap(
// 		r.config.UnstructuredContent(),
// 		"spec",
// 		"operator",
// 		"deploy",
// 	)
// 	if err != nil {
// 		return err
// 	}
// 	if !found {
// 		return errors.Errorf(
// 			"Missing spec.operator.deploy: %s",
// 			r,
// 		)
// 	}
// 	var ok bool
// 	r.operatorDeployPath, ok = deployfields["path"].(string)
// 	if !ok {
// 		return errors.Errorf(
// 			"Invalid spec.operator.deploy.path: %s",
// 			r,
// 		)
// 	}
// 	r.operatorDeployFile, ok = deployfields["file"].(string)
// 	if !ok {
// 		return errors.Errorf(
// 			"Invalid spec.operator.deploy.file: %s",
// 			r,
// 		)
// 	}
// 	return nil
// }

func (r *TestRunner) initKubeBinPath() error {
	r.kubeBinPath = "kubebin"
	return nil
}

func (r *TestRunner) initK8sVersion() error {
	ver, _, err := unstructured.NestedString(
		r.SetupConfig.UnstructuredContent(),
		"spec",
		"kubernetes",
		"version",
	)
	if err != nil {
		return err
	}
	if ver != "" {
		r.k8sVersion = ver
	} else {
		r.k8sVersion = k8sVersion
	}
	return nil
}

func (r *TestRunner) initEtcdVersion() error {
	ver, _, err := unstructured.NestedString(
		r.SetupConfig.UnstructuredContent(),
		"spec",
		"etcd",
		"version",
	)
	if err != nil {
		return err
	}
	if ver != "" {
		r.etcdVersion = ver
	} else {
		r.etcdVersion = etcdVersion
	}
	return nil
}

func (r *TestRunner) initK8sDownloadURL() error {
	url, _, err := unstructured.NestedString(
		r.SetupConfig.UnstructuredContent(),
		"spec",
		"kubernetes",
		"downloadURL",
	)
	if err != nil {
		return err
	}
	if url != "" {
		r.k8sDownloadURL = url
	} else {
		r.k8sDownloadURL = k8sDownloadURL
	}
	return nil
}

func (r *TestRunner) initEtcdDownloadURL() error {
	url, _, err := unstructured.NestedString(
		r.SetupConfig.UnstructuredContent(),
		"spec",
		"etcd",
		"downloadURL",
	)
	if err != nil {
		return err
	}
	if url != "" {
		r.etcdDownloadURL = url
	} else {
		r.etcdDownloadURL = etcdDownloadURL
	}
	return nil
}

func (r *TestRunner) init() error {
	var fns = []func() error{
		r.initKubeBinPath,
		r.initK8sVersion,
		r.initK8sDownloadURL,
		r.initEtcdVersion,
		r.initEtcdDownloadURL,
		r.initOperatorManifestFields,
		r.initOperatorBinaryFields,
		r.initOperatorTestFields,
		//r.initOperatorDeployFields,
	}
	for _, fn := range fns {
		err := fn()
		if err != nil {
			return err
		}
	}
	return nil
}

// NewTestRunner returns a new instance of TestRunner
func NewTestRunner(config TestRunnerConfig) (*TestRunner, error) {
	if config.SetupConfig == nil || len(config.SetupConfig.Object) == 0 {
		return nil, errors.Errorf("Nil test setup")
	}
	r := &TestRunner{
		SetupConfig: config.SetupConfig,
	}
	err := r.init()
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (r *TestRunner) wgetKubectlToTestLoc() error {
	var cmds = []cmdargs{
		{
			cmd: "wget",
			args: []string{
				"-q",
				fmt.Sprintf(
					"%s/%s/bin/linux/amd64/kubectl",
					r.k8sDownloadURL,
					r.k8sVersion,
				),
			},
		},
		{
			cmd: "chmod",
			args: []string{
				"+x",
				"kubectl",
			},
		},
		{
			cmd: "mv",
			args: []string{
				"kubectl",
				r.kubeBinPath,
			},
		},
	}
	for _, cmd := range cmds {
		err := sh.Run(cmd.cmd, cmd.args...)
		if err != nil {
			return errors.Wrapf(err, fmt.Sprintf("%s", r))
		}
	}
	return nil
}

func (r *TestRunner) wgetKubectlToTestLocIfNotFound() {
	cmd := fmt.Sprintf("%s/kubectl", r.kubeBinPath)
	err := sh.Run(cmd, "version", "--client")
	if err == nil {
		return
	}
	if !strings.Contains(err.Error(), notFoundErrMsg) {
		// this is different than not found error
		r.err = errors.Wrapf(err, fmt.Sprintf("%s", r))
		return
	}
	r.err = r.wgetKubectlToTestLoc()
}

func (r *TestRunner) wgetETCDToTestLoc() error {
	var etcdFileName = fmt.Sprintf("etcd-%s-linux-amd64", r.etcdVersion)
	var etcdFileNameAsTarGZ = fmt.Sprintf("%s.tar.gz", etcdFileName)
	var etcdDownloadURL = fmt.Sprintf(
		"%s/%s/%s",
		r.etcdDownloadURL,
		etcdVersion,
		etcdFileNameAsTarGZ,
	)
	var cmds = []cmdargs{
		{
			cmd: "wget",
			args: []string{
				"-q",
				etcdDownloadURL,
			},
		},
		{
			cmd: "tar",
			args: []string{
				"-zxf",
				etcdFileNameAsTarGZ,
			},
		},
		{
			cmd: "mv",
			args: []string{
				fmt.Sprintf("%s/etcd", etcdFileName),
				r.kubeBinPath,
			},
		},
		{
			cmd: "rm",
			args: []string{
				"-rf",
				etcdFileName,
				etcdFileNameAsTarGZ,
			},
		},
	}
	for _, cmd := range cmds {
		err := sh.Run(cmd.cmd, cmd.args...)
		if err != nil {
			return errors.Wrapf(err, fmt.Sprintf("%s", r))
		}
	}
	return nil
}

func (r *TestRunner) wgetETCDToTestLocIfNotFound() {
	cmd := fmt.Sprintf("%s/etcd", r.kubeBinPath)
	err := sh.Run(cmd, "--version")
	if err == nil {
		return
	}
	if !strings.Contains(err.Error(), notFoundErrMsg) {
		// this is different than not found error
		r.err = errors.Wrapf(err, fmt.Sprintf("%s", r))
		return
	}
	// download etcd
	r.err = r.wgetETCDToTestLoc()
}

func (r *TestRunner) wgetKubeAPIServerToTestLoc() error {
	var cmds = []cmdargs{
		{
			cmd: "wget",
			args: []string{
				"-q",
				fmt.Sprintf(
					"%s/%s/bin/linux/amd64/kube-apiserver",
					r.k8sDownloadURL,
					r.k8sVersion,
				),
			},
		},
		{
			cmd: "chmod",
			args: []string{
				"+x",
				"kube-apiserver",
			},
		},
		{
			cmd: "mv",
			args: []string{
				"kube-apiserver",
				r.kubeBinPath,
			},
		},
	}
	for _, cmd := range cmds {
		err := sh.Run(cmd.cmd, cmd.args...)
		if err != nil {
			return errors.Wrapf(err, fmt.Sprintf("%s", r))
		}
	}
	return nil
}

func (r *TestRunner) wgetKubeAPIServerToTestLocIfNotFound() {
	cmd := fmt.Sprintf("%s/kube-apiserver", r.kubeBinPath)
	err := sh.Run(cmd, "--version")
	if err == nil {
		return
	}
	if !strings.Contains(err.Error(), notFoundErrMsg) {
		// this is different than not found error
		r.err = errors.Wrapf(err, fmt.Sprintf("%s", r))
		return
	}
	// download kube api server
	r.err = r.wgetKubeAPIServerToTestLoc()
}

// startKube starts kubernetes control plane
func (r *TestRunner) startKube() {
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
func (r *TestRunner) stopKube() (err error) {
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

// startOperator starts the operator
func (r *TestRunner) startOperator() {
	klog.V(2).Infof("Starting operator: %s", r)
	cmd := controlplane.NewCommand(
		controlplane.CommandConfig{
			Err: os.Stderr,
			Out: os.Stdout,
		},
	)
	var allArgs []string
	allArgs = append(
		allArgs,
		r.operatorBinArgs...,
	)
	allArgs = append(
		allArgs,
		fmt.Sprintf(
			"%s=%s",
			r.kubeAPIServerURLFlag,
			r.cp.APIURL().String(),
		),
	)
	operator := fmt.Sprintf("%s/%s", r.operatorBinPath, r.operatorBinName)
	stop, err := cmd.Start(
		operator,
		allArgs...,
	)
	if err != nil {
		r.err = errors.Wrapf(err, fmt.Sprintf("%s", r))
		return
	}
	r.stopOperatorFn = stop
	r.addToSetupTeardown(r.stopOperator)
	klog.V(2).Infof("Operator started successfully: %s", r)
}

// StopOperator stops operator
func (r *TestRunner) stopOperator() (err error) {
	if r.stopOperatorFn == nil {
		return
	}
	klog.V(2).Infof("Stopping operator: %s", r)
	defer func() {
		if err == nil {
			klog.V(2).Infof("Operator stopped successfully: %s", r)
		}
	}()
	return r.stopOperatorFn()
}

// applyOperatorManifests applies operator's manifest files
func (r *TestRunner) applyOperatorManifests() {
	klog.V(2).Infof(
		"Applying manifests: Path %s: %s",
		r.operatorManifestsPath,
		r,
	)
	defer func() {
		if r.err == nil {
			klog.V(2).Infof(
				"Applied manifests successfully: %s",
				r,
			)
		}
	}()
	r.err = r.cp.KubeCtl().Apply(
		controlplane.ApplyConfig{
			Path:      r.operatorManifestsPath,
			YAMLFiles: r.operatorManifestFiles,
		},
	)
}

func (r *TestRunner) createDeclarativeTestNamespace() {
	klog.V(2).Infof(
		"Creating declarative test namespace %s: %s",
		declarativeTestNamespace,
		r,
	)
	defer func() {
		if r.err == nil {
			klog.V(2).Infof(
				"Created declarative test namespace %s successfully: %s",
				declarativeTestNamespace,
				r,
			)
		}

	}()
	r.err = r.cp.KubeCtl().Run(
		"create",
		"namespace",
		declarativeTestNamespace,
	)
}

// applyDeclarativeTests applies files to test an operator
func (r *TestRunner) applyDeclarativeTests() {
	klog.V(2).Infof(
		"Applying declarative tests: Path %s: %s",
		r.operatorTestsPath,
		r,
	)
	defer func() {
		if r.err == nil {
			klog.V(2).Infof(
				"Applied declarative tests successfully: %s",
				r,
			)
		}

	}()
	r.err = r.cp.KubeCtl().Apply(
		controlplane.ApplyConfig{
			Path:      r.operatorTestsPath,
			YAMLFiles: r.operatorTestFiles,
		},
	)
}

// waitTillAllJobsAreCompleted waits till all jobs are completed
func (r *TestRunner) waitTillAllJobsAreCompleted() {
	var counter int
	for {
		counter++
		// TODO (@amitkumardas):
		// Need to expose counter as a tunable in setup.yaml
		if counter >= 25 {
			break
		}
		op, err := r.cp.KubeCtl().RunOp(
			"get",
			"jobs.metacontroller.app",
			"-n",
			declarativeTestNamespace,
			inferenceTestName,
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
			// test has failed
			break
		}
		klog.V(2).Infof("Job 'inference' has status.phase=%q", op)
		// TODO (@amitkumardas):
		// Need to expose sleep seconds as a tunable in setup.yaml
		time.Sleep(3 * time.Second)
	}
	// At this point inference job has **failed**
	// It is wise to display the output & then set error
	op, _ := r.cp.KubeCtl().RunOp(
		"get",
		"jobs.metacontroller.app",
		"-n",
		declarativeTestNamespace,
		inferenceTestName,
		"-oyaml",
	)
	klog.Infof("\n%s", op)
	r.err = &FailedRun{Err: "Tests might have failed or timedout"}
}

// printDeclarativeTestResultsV prints verbose test results on the terminal
func (r *TestRunner) printDeclarativeTestResultsV() {
	if r.err != nil {
		if _, ok := r.err.(*FailedRun); !ok {
			// will not print for runtime errors
			return
		}
	}
	op, err := r.cp.KubeCtl().RunOp(
		"get",
		"jobs.metacontroller.app",
		"-n",
		declarativeTestNamespace,
		"-l",
		"d-testing.metacontroller.app/enabled=true",
		"-oyaml",
	)
	if err != nil {
		// log this error & continue
		// we are not interested on this err
		klog.Infof("%s", err)
		return
	}
	klog.Infof("\n%s", op)
}

func (r *TestRunner) run() error {
	defer func() {
		r.printDeclarativeTestResultsV()
	}()
	var fns = []func(){
		r.wgetETCDToTestLocIfNotFound,
		r.wgetKubeAPIServerToTestLocIfNotFound,
		r.wgetKubectlToTestLocIfNotFound,
		r.startKube,
		r.applyOperatorManifests,
		r.startOperator,
		r.createDeclarativeTestNamespace,
		r.applyDeclarativeTests,
		r.waitTillAllJobsAreCompleted,
	}
	for _, f := range fns {
		f()
		if r.err != nil {
			return r.err
		}
	}
	return nil
}

// Run tests on kubernetes
func Run() error {
	testsetups, err := loadAllTestSetups()
	if err != nil {
		return err
	}
	for _, testsetup := range testsetups {
		r, err := NewTestRunner(TestRunnerConfig{
			SetupConfig: &testsetup,
		})
		if err != nil {
			return err
		}
		err = r.run()
		r.TeardownSetup()
		if err != nil {
			return err
		}
	}
	return nil
}
