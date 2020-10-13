/*
Copyright 2020 The MayaData Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"os"

	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"mayadata.io/d-commander/pkg"
)

var (
	commandKind = flag.String(
		"command-kind",
		"Command", // default
		"Kubernetes custom resource kind",
	)

	commandResource = flag.String(
		"command-resource",
		"commands", // default
		"Kubernetes custom resource name",
	)

	commandGroup = flag.String(
		"command-group",
		"dope.mayadata.io", // default
		"Kubernetes custom resource group",
	)

	commandVersion = flag.String(
		"command-api-version",
		"v1", // default
		"Kubernetes custom resource api version",
	)

	commandName = flag.String(
		"command-name",
		"",
		"Name of the command",
	)

	commandNamespace = flag.String(
		"command-ns",
		"",
		"Namespace of the command",
	)

	kubeAPIServerURL = flag.String(
		"kube-apiserver-url",
		"",
		`Kubernetes api server url (same format as used by kubectl).
		If not specified, uses in-cluster config`,
	)

	kubeconfig = flag.String(
		"kubeconfig",
		"",
		"absolute path to the kubeconfig file",
	)

	clientGoQPS = flag.Float64(
		"client-go-qps",
		5, // default
		"Number of queries per second client-go is allowed to make (default 5)",
	)

	clientGoBurst = flag.Int(
		"client-go-burst",
		10, // default
		"Allowed burst queries for client-go (default 10)",
	)
)

// This is the entry point of this binary
//
// This binary is meant to be run to completion. In other
// words this does not expose any long running service.
// This executes the commands or scripts specified in the
// custom resource and updates this resource post execution.
//
//
// NOTE:
//	A kubernetes **Job** can make use of this binary
func main() {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Set("alsologtostderr", "true")
	flag.Parse()
	defer klog.Flush()

	if *commandName == "" {
		klog.Exit("Invalid arguments: Flag 'command-name' must be set")
	}
	if *commandNamespace == "" {
		klog.Exit("Invalid arguments: Flag 'command-ns' must be set")
	}

	klog.V(1).Infof("Command custom resource: group %q", *commandGroup)
	klog.V(1).Infof("Command custom resource: version %q", *commandVersion)
	klog.V(1).Infof("Command custom resource: kind %q", *commandKind)
	klog.V(1).Infof("Command custom resource: resource %q", *commandResource)
	klog.V(1).Infof("Command custom resource: namespace %q", *commandNamespace)
	klog.V(1).Infof("Command custom resource: name %q", *commandName)

	r, err := NewRunner()
	if err != nil {
		// This should lead to crashloopback if this
		// is running from within a Kubernetes pod
		klog.Exit(err)
	}
	err = r.Run()
	if err != nil {
		// This should lead to crashloopback if this
		// is running from within a Kubernetes pod
		klog.Exit(err)
	}
	os.Exit(0)
}

// Runnable helps in executing the Kubernetes command
// resource. It does so by executing the commands or scripts
// specified in the resource and updating this resource post
// execution.
type Runnable struct {
	Client dynamic.Interface
	GVR    schema.GroupVersionResource

	commandStatus *pkg.CommandStatus
}

// NewRunner returns a new instance of Runnable
func NewRunner() (*Runnable, error) {
	var config *rest.Config
	var err error

	if *kubeconfig != "" {
		klog.V(2).Infof("Using kubeconfig %q", *kubeconfig)
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	} else if *kubeAPIServerURL != "" {
		klog.V(2).Infof("Using kubernetes api server url %q", *kubeAPIServerURL)
		config, err = clientcmd.BuildConfigFromFlags(*kubeAPIServerURL, "")
	} else {
		klog.V(2).Info("Using in-cluster kubeconfig")
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		return nil, err
	}

	// configure kubernetes client config with additional settings
	// to manage deluge of requests to kubernetes API server
	config.QPS = float32(*clientGoQPS)
	config.Burst = *clientGoBurst

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	gvr := schema.GroupVersionResource{
		Group:    *commandGroup,
		Version:  *commandVersion,
		Resource: *commandResource,
	}

	return &Runnable{
		Client: client,
		GVR:    gvr,
	}, nil
}

func (a *Runnable) updateWithRetries() error {
	var statusNew interface{}
	err := pkg.MarshalThenUnmarshal(a.commandStatus, &statusNew)
	if err != nil {
		return errors.Wrapf(
			err,
			"Marshal unmarshal failed: Command %q / %q",
			*commandNamespace,
			*commandName,
		)
	}
	klog.V(1).Infof(
		"Command %q / %q: Status %s",
		*commandNamespace,
		*commandName,
		pkg.NewJSON(statusNew).MustMarshal(),
	)

	// Command is updated with latest labels
	labels := map[string]string{
		// this label key is set with same value as that of status.phase
		pkg.LblKeyCommandPhase: string(a.commandStatus.Phase),
	}

	var runtimeErr error

	// This uses exponential backoff to avoid exhausting
	// the apiserver
	retryErr := retry.RetryOnConflict(
		retry.DefaultRetry,
		func() error {
			// Retrieve the latest version of Command
			cmd, err := a.Client.
				Resource(a.GVR).
				Namespace(*commandNamespace).
				Get(*commandName, v1.GetOptions{})
			if err != nil {
				// Retry this error since this might be a temporary
				return errors.Wrapf(
					err,
					"Failed to get command: %q / %q",
					*commandNamespace,
					*commandName,
				)
			}

			// Mutate command resource's status field
			err = unstructured.SetNestedField(
				cmd.Object,
				statusNew,
				"status",
			)
			if err != nil {
				runtimeErr = errors.Wrapf(
					err,
					"Set unstruct failed: Path 'status': Command %q / %q",
					*commandNamespace,
					*commandName,
				)
				// Return nil to avoid retry
				//
				// NOTE:
				//	Setting unstructured instance should not be
				// retried since every retry will result in the
				// same error
				return nil
			}

			// Merge existing labels with desired pair(s)
			pkg.SetLabels(cmd, labels)

			updated, err := a.Client.
				Resource(a.GVR).
				Namespace(*commandNamespace).
				Update(cmd, v1.UpdateOptions{})

			if err != nil {
				// Update error is returned to be retried since this
				// might be temporary
				return errors.Wrapf(
					err,
					"Update failed: Command %q %q",
					*commandNamespace,
					*commandName,
				)
			}

			// Mutate command instance with latest resource version
			// before trying update status. This is done since previous
			// update would have modified resource version.
			cmd.SetResourceVersion(updated.GetResourceVersion())
			// Update command status as a **sub resource** update
			cmdUpdatedStatus, err := a.Client.
				Resource(a.GVR).
				Namespace(*commandNamespace).
				UpdateStatus(cmd, v1.UpdateOptions{})

			if err == nil {
				// This is an extra check to detect type conversion issues
				// if any during later stages
				var c pkg.Command
				tErr := pkg.ToTyped(cmdUpdatedStatus, &c)
				klog.V(1).Infof(
					"UnstructToTyped: IsError=%t: %v", tErr != nil, tErr,
				)
			}
			// If update status resulted in an error it will be
			// returned so that update can be retried
			return errors.Wrapf(
				err,
				"Update status failed: Command %q %q",
				*commandNamespace,
				*commandName,
			)
		})

	if runtimeErr != nil {
		return errors.Wrapf(
			runtimeErr,
			"Update failed: Runtime error: Command: %q %q",
			*commandNamespace,
			*commandName,
		)
	}
	return retryErr
}

// Run executes the command resource
func (a *Runnable) Run() error {
	got, err := a.Client.
		Resource(a.GVR).
		Namespace(*commandNamespace).
		Get(
			*commandName,
			v1.GetOptions{},
		)
	if err != nil {
		return errors.Wrapf(
			err,
			"Failed to get command: %q %q",
			*commandNamespace,
			*commandName,
		)
	}

	var c pkg.Command
	// convert from unstructured instance to typed instance
	err = pkg.ToTyped(got, &c)
	if err != nil {
		return errors.Wrapf(
			err,
			"Failed to convert unstructured command to typed instance: %q / %q",
			*commandNamespace,
			*commandName,
		)
	}

	cmdRunner, err := pkg.NewRunner(
		pkg.RunnableConfig{
			Command: c,
		},
	)
	if err != nil {
		return err
	}
	a.commandStatus, err = cmdRunner.Run()
	if err != nil {
		return err
	}

	return a.updateWithRetries()
}
