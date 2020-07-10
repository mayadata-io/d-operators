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
	"path/filepath"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"

	"mayadata.io/d-operators/common/unstruct"
	"mayadata.io/d-operators/pkg/command"
	types "mayadata.io/d-operators/types/command"
)

var (
	commandKind = flag.String(
		"command-kind",
		"Command",
		"Kind of Command custom resource",
	)
	commandResource = flag.String(
		"command-resource",
		"commands",
		"Resource name of Command custom resource",
	)
	commandGroup = flag.String(
		"command-group",
		"dope.metacontroller.io",
		"Group of Command custom resource",
	)
	commandVersion = flag.String(
		"command-version",
		"v1",
		"Version of Command custom resource",
	)
	commandName = flag.String(
		"command-name",
		"",
		"Name of Command custom resource",
	)
	commandNamespace = flag.String(
		"command-ns",
		"",
		"Namespace of Command custom resource",
	)

	kubeAPIServerURL = flag.String(
		"kube-apiserver-url",
		"",
		`Kubernetes api server url (same format as used by kubectl).
		If not specified, uses in-cluster config`,
	)
	kubeconfig *string

	clientGoQPS = flag.Float64(
		"client-go-qps",
		5,
		"Number of queries per second client-go is allowed to make (default 5)",
	)
	clientGoBurst = flag.Int(
		"client-go-burst",
		10,
		"Allowed burst queries for client-go (default 10)",
	)
)

// main function is the entry point of this binary.
//
// This binary is meant to be run to completion. In other
// words this does not expose any long running service.
//
// NOTE:
//	A kubernetes **Job** can make use of this binary
func main() {
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String(
			"kubeconfig",
			filepath.Join(home, ".kube", "config"),
			"(optional) absolute path to the kubeconfig file",
		)
	} else {
		kubeconfig = flag.String(
			"kubeconfig",
			"",
			"absolute path to the kubeconfig file",
		)
	}

	flag.Set("alsologtostderr", "true")
	flag.Parse()

	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(klogFlags)

	// Sync the glog and klog flags.
	flag.CommandLine.VisitAll(func(f1 *flag.Flag) {
		f2 := klogFlags.Lookup(f1.Name)
		if f2 != nil {
			value := f1.Value.String()
			f2.Value.Set(value)
		}
	})
	defer klog.Flush()

	if *commandName == "" {
		klog.Fatal("Invalid arguments: Flag 'command-name' must be set")
	}

	klog.V(1).Infof("Command custom resource: kind %s", *commandKind)
	klog.V(1).Infof("Command custom resource: resource %s", *commandResource)
	klog.V(1).Infof("Command custom resource: group %s", *commandGroup)
	klog.V(1).Infof("Command custom resource: version %s", *commandVersion)
	klog.V(1).Infof("Command custom resource: name %s", *commandName)
	klog.V(1).Infof("Command custom resource: namespace %s", *commandNamespace)

	runCommand(getRestConfig())
	os.Exit(0)
}

func getRestConfig() *rest.Config {
	var config *rest.Config
	var err error
	if *kubeconfig != "" {
		klog.V(1).Infof("Using kubeconfig %s", *kubeconfig)
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	} else if *kubeAPIServerURL != "" {
		klog.V(1).Infof("Using kubernetes api server url %s", *kubeAPIServerURL)
		config, err = clientcmd.BuildConfigFromFlags(*kubeAPIServerURL, "")
	} else {
		klog.V(1).Info("Using in-cluster kubeconfig")
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		klog.Fatal(err)
	}
	config.QPS = float32(*clientGoQPS)
	config.Burst = *clientGoBurst
	return config
}

func runCommand(config *rest.Config) {
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		klog.Fatal(err)
	}

	gvr := schema.GroupVersionResource{
		Group:    *commandGroup,
		Version:  *commandVersion,
		Resource: *commandResource,
	}

	got, err := client.Resource(gvr).
		Namespace(*commandNamespace).
		Get(
			*commandName,
			v1.GetOptions{},
		)
	if err != nil {
		klog.Fatal(err)
	}

	var c types.Command
	// convert from unstructured instance to typed instance
	err = unstruct.ToTyped(got, &c)
	if err != nil {
		klog.Fatal(err)
	}

	cmder, err := command.NewCommander(
		command.CommandableConfig{
			Command: &c,
		},
	)
	if err != nil {
		klog.Fatal(err)
	}
	status, err := cmder.Run()
	if err != nil {
		klog.Fatal(err)
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Command before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		got, err = client.Resource(gvr).
			Namespace(*commandNamespace).
			Get(
				*commandName,
				v1.GetOptions{},
			)
		if err != nil {
			klog.Fatal(err)
		}

		// update labels
		lbls := got.GetLabels()
		if len(lbls) == 0 {
			lbls = make(map[string]string)
		}
		lbls["command.dope.metacontroller.io/phase"] = string(status.Phase)
		got.SetLabels(lbls)

		// update status
		err = unstructured.SetNestedField(
			got.Object,
			status,
			"status",
		)
		if err != nil {
			klog.Fatal(err)
		}

		// update command resource
		_, updateErr := client.Resource(gvr).
			Namespace(*commandNamespace).
			Update(
				got,
				v1.UpdateOptions{},
			)
		return updateErr
	})
	if retryErr != nil {
		klog.Fatal(retryErr)
	}
}
