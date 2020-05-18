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

package job

import (
	apiextnclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	dynamicclientset "openebs.io/metac/dynamic/clientset"
	dynamicdiscovery "openebs.io/metac/dynamic/discovery"
)

// FixtureConfig is used to create a new instance of Fixture
type FixtureConfig struct {
	KubeConfig   *rest.Config
	APIDiscovery *dynamicdiscovery.APIResourceDiscovery
	IsTearDown   bool
}

// Fixture is the base structure that ties a job specification
// with one or more kubernetes api operations.
type Fixture struct {
	apiDiscovery *dynamicdiscovery.APIResourceDiscovery

	// dynamic clientset to invoke kubernetes operations
	// against kubernetes native as well as custom resources
	dynamicClientset *dynamicclientset.Clientset

	// typed clientset to invoke kubernetes operations
	kubeClientset kubernetes.Interface

	// crdClient is based on k8s.io/apiextensions-apiserver &
	// is all about invoking operations against custom
	// resource definitions aka CRDs
	crdClient apiextnclientset.ApiextensionsV1beta1Interface

	tearDown      bool
	teardownFuncs []func() error

	// error as value
	err error
}

func (f *Fixture) setTearDown(config FixtureConfig) {
	f.tearDown = config.IsTearDown
}

func (f *Fixture) setAPIDiscovery(config FixtureConfig) {
	f.apiDiscovery = config.APIDiscovery
}

func (f *Fixture) setCRDClient(config FixtureConfig) {
	f.crdClient, f.err = apiextnclientset.NewForConfig(
		config.KubeConfig,
	)
}

func (f *Fixture) setDynamicClientset(config FixtureConfig) {
	f.dynamicClientset, f.err = dynamicclientset.New(
		config.KubeConfig,
		config.APIDiscovery,
	)
}

func (f *Fixture) setKubeClientset(config FixtureConfig) {
	f.kubeClientset, f.err = kubernetes.NewForConfig(
		config.KubeConfig,
	)
}

// NewFixture returns a new instance of Fixture
func NewFixture(config FixtureConfig) (*Fixture, error) {
	f := &Fixture{}
	var setters = []func(FixtureConfig){
		f.setTearDown,
		f.setAPIDiscovery,
		f.setCRDClient,
		f.setDynamicClientset,
		f.setKubeClientset,
	}
	for _, setFn := range setters {
		setFn(config)
		if f.err != nil {
			return nil, f.err
		}
	}
	return f, nil
}

// GetTypedClientset returns the Kubernetes clientset that is
// used to invoke CRUD operations against Kubernetes native
// resources.
// func (f *Fixture) GetTypedClientset() kubernetes.Interface {
// 	return f.typedClientset
// }

// GetCRDClient returns the CRD client that can be used to invoke
// CRUD operations against CRDs itself
// func (f *Fixture) GetCRDClient() apiextensionsclientset.ApiextensionsV1beta1Interface {
// 	return f.crdClient
// }

// TearDown cleans up resources created through this instance
// of the test fixture.
func (f *Fixture) TearDown() {
	if !f.tearDown {
		return
	}
	// cleanup in descending order
	for i := len(f.teardownFuncs) - 1; i >= 0; i-- {
		teardown := f.teardownFuncs[i]
		err := teardown()
		if err != nil {
			if apierrors.IsNotFound(err) {
				klog.V(2).Infof(
					"Teardown ignored: Resource not found: %+v",
					err,
				)
				continue
			}
			if apierrors.IsConflict(err) {
				klog.V(2).Infof(
					"Teardown ignored: Conflict: %+v",
					err,
				)
				continue
			}
			// we treat the teardown error as level 1 Info
			klog.V(1).Infof(
				"Teardown failed: %s: %+v",
				apierrors.ReasonForError(err),
				err,
			)
		}
	}
}

// AddToTeardown adds the given teardown func
func (f *Fixture) AddToTeardown(teardown func() error) {
	if !f.tearDown {
		return
	}
	f.teardownFuncs = append(f.teardownFuncs, teardown)
}
