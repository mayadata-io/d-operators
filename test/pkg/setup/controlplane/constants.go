package controlplane

const (
	// NotFoundErrMsg defines the message received when a binary
	// is not found
	NotFoundErrMsg = "no such file or directory"

	// ExperimentAPIGroup represent the api group of the experiment
	// considered for testcase implementation
	ExperimentAPIGroup = "recipes.dope.metacontroller.io"

	// DefaultK8sDownloadURL is the default kubernetes download url
	DefaultK8sDownloadURL = "https://dl.k8s.io"

	// DefaultETCDDownloadURL is the default etcd download url
	DefaultETCDDownloadURL = "https://github.com/coreos/etcd/releases/download"

	// DefaultETCDVersion is the default etcd version used to setup Kubernetes
	// ControlPlane
	DefaultETCDVersion = "v3.4.3"

	// DefaultK8sVersion is the default Kubernetes API Server & Kubectl version
	// used to setup Kubernetes ControlPlane
	DefaultK8sVersion = "v1.16.4"

	// DefaultKubeControlPlaneBinariesPath is the path to all binaries used to
	// setup Kubernetes ControlPlane
	DefaultKubeControlPlaneBinariesPath = "kubebin"

	// DefaultMaxRetryAttempt is the default retry attempt
	DefaultMaxRetryAttempt int = 25

	// DefaultRetryIntervalInSeconds is the default retry interval
	DefaultRetryIntervalInSeconds int = 1

	// DefaultControlPlaneSetupFilePath is the default path where control
	// plane setup files are expected to be available
	DefaultControlPlaneSetupFilePath string = "setup/controlplane"
)
