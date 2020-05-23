package types

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ControlPlaneSetup is a typed representation of a
// ControlPlane setup
type ControlPlaneSetup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec ControlPlaneSetupSpec `json:"spec"`
}

// ControlPlaneSetupSpec that defines how ControlPlane should
// be brought up
type ControlPlaneSetupSpec struct {
	ETCD       Downloadable `json:"etcd"`
	Kubernetes Downloadable `json:"kubernetes"`
	Target     Target       `json:"target"`
	Test       Test         `json:"test"`
}

// Downloadable defines a downloadable resource
type Downloadable struct {
	DownloadURL string `json:"downloadURL"`
	Version     string `json:"version"`
}

// Target defines the kubernetes controller i.e. target under test
type Target struct {
	Binary Binary `json:"binary"`
	Deploy Deploy `json:"deploy"`
}

// Test defines the test cases that are verified against the target
type Test struct {
	Experiments Deploy    `json:"experiments"`
	Deploy      Deploy    `json:"deploy"`
	Inference   Inference `json:"inference"`
}

// Inference exposes the tunables that decide the outcome
// of running the experiments against the target on this setup
type Inference struct {
	ExperimentName         string          `json:"experimentName"`
	ExperimentNamespace    string          `json:"experimentNamespace"`
	MaxRetryAttempt        *int            `json:"maxRetryAttempt"`
	RetryIntervalInSeconds *int            `json:"retryIntervalInSeconds"`
	DisplaySelector        DisplaySelector `json:"displaySelector"`
}

// DisplaySelector defines the filters to display the test results
type DisplaySelector struct {
	MatchLabels map[string]string `json:"matchLabels"`
	MatchPhases []string          `json:"matchPhases"`
}

// Binary defines a binary & its properties
type Binary struct {
	Path                 string   `json:"path"`
	Name                 string   `json:"name"`
	Args                 []string `json:"args"`
	KubeAPIServerURLFlag string   `json:"kubeAPIServerURLFlag"`
}

// Deploy defines the path & files available in this path
type Deploy struct {
	Path  string   `json:"path"`
	Files []string `json:"files"`
}
