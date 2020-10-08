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

package kubernetes

import (
	"github.com/pkg/errors"
	apiextnv1beta1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"openebs.io/metac/dynamic/clientset"
	dynamicclientset "openebs.io/metac/dynamic/clientset"
	dynamicdiscovery "openebs.io/metac/dynamic/discovery"
)

// UtilityConfig helps creating new instances of Utility
type UtilityConfig struct {
	KubeConfig   *rest.Config
	APIDiscovery *dynamicdiscovery.APIResourceDiscovery
	Retry        *Retryable
	IsTeardown   bool
}

// UtilityFuncs exposes Utility fields as functional options
//
// NOTE:
// 	These functions can be mocked during unit tests
type UtilityFuncs struct {
	getClientForAPIVersionAndKindFn     func(string, string) (*clientset.ResourceClient, error)
	getClientForAPIVersionAndResourceFn func(string, string) (*clientset.ResourceClient, error)
	getAPIForAPIVersionAndResourceFn    func(string, string) *dynamicdiscovery.APIResource
}

// Utility exposes instances needed to invoke kubernetes
// api operations
type Utility struct {
	*UtilityFuncs
	Retry *Retryable

	apiDiscovery *dynamicdiscovery.APIResourceDiscovery

	// dynamic client to invoke kubernetes operations
	// against kubernetes native as well as custom resources
	dynamicClientset *dynamicclientset.Clientset

	// clientset to invoke kubernetes operations
	kubeClientset kubernetes.Interface

	// client to invoke operations against k8s.io/apiextensions-apiserver
	// In other words invoke operations against custom resource
	// definitions aka CRDs
	crdClient apiextnv1beta1.ApiextensionsV1beta1Interface

	// If resources created using this utility should be deleted
	// when this instance's Teardown method is invoked
	isTeardown bool

	// list of teardown functions invoked when this instance's
	// Teardown method is invoked
	teardownFuncs []func() error

	// error as value
	err error
}

func (u *Utility) setTeardownFlag(config UtilityConfig) {
	u.isTeardown = config.IsTeardown
}

func (u *Utility) setAPIDiscovery(config UtilityConfig) {
	u.apiDiscovery = config.APIDiscovery
}

func (u *Utility) setCRDClient(config UtilityConfig) {
	if config.KubeConfig == nil {
		u.err = errors.Errorf(
			"Failed to set crd client: Nil kube config provided",
		)
		return
	}
	u.crdClient, u.err = apiextnv1beta1.NewForConfig(
		config.KubeConfig,
	)
}

func (u *Utility) setDynamicClientset(config UtilityConfig) {
	u.dynamicClientset, u.err = dynamicclientset.New(
		config.KubeConfig,
		config.APIDiscovery,
	)
}

func (u *Utility) setKubeClientset(config UtilityConfig) {
	u.kubeClientset, u.err = kubernetes.NewForConfig(
		config.KubeConfig,
	)
}

// NewUtility returns a new instance of Fixture
func NewUtility(config UtilityConfig) (*Utility, error) {
	// check retry
	var retry = NewRetry(RetryConfig{})
	if config.Retry != nil {
		retry = config.Retry
	}
	u := &Utility{
		UtilityFuncs: &UtilityFuncs{},
		Retry:        retry,
	}
	var setters = []func(UtilityConfig){
		u.setTeardownFlag,
		u.setAPIDiscovery,
		u.setCRDClient,
		u.setDynamicClientset,
		u.setKubeClientset,
	}
	for _, set := range setters {
		set(config)
		if u.err != nil {
			return nil, u.err
		}
	}
	return u, nil
}

// Teardown deletes resources created through this instance
func (u *Utility) Teardown() {
	if !u.isTeardown {
		return
	}
	// cleanup in descending order
	for i := len(u.teardownFuncs) - 1; i >= 0; i-- {
		teardown := u.teardownFuncs[i]
		err := teardown()
		if err != nil {
			if apierrors.IsNotFound(err) {
				klog.V(3).Infof(
					"Teardown ignored: Resource not found: %+v",
					err,
				)
				continue
			}
			if apierrors.IsConflict(err) {
				klog.V(3).Infof(
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

// AddToTeardown adds the given teardown function to
// the list of teardown functions
func (u *Utility) AddToTeardown(teardown func() error) {
	if !u.isTeardown {
		return
	}
	u.teardownFuncs = append(u.teardownFuncs, teardown)
}

// GetClientForAPIVersionAndKind returns the dynamic client for the
// given api version & kind
func (u *Utility) GetClientForAPIVersionAndKind(
	apiversion string,
	kind string,
) (*clientset.ResourceClient, error) {
	if u.getClientForAPIVersionAndKindFn != nil {
		return u.getClientForAPIVersionAndKindFn(apiversion, kind)
	}
	return u.dynamicClientset.GetClientForAPIVersionAndKind(
		apiversion,
		kind,
	)
}

// GetClientForAPIVersionAndResource returns the dynamic client for the
// given api version & resource
func (u *Utility) GetClientForAPIVersionAndResource(
	apiversion string,
	resource string,
) (*clientset.ResourceClient, error) {
	if u.getClientForAPIVersionAndResourceFn != nil {
		return u.getClientForAPIVersionAndResourceFn(apiversion, resource)
	}
	return u.dynamicClientset.GetClientForAPIVersionAndResource(
		apiversion,
		resource,
	)
}

// GetAPIForAPIVersionAndResource returns the discovered api based
// on the provided api version & resource
func (u *Utility) GetAPIForAPIVersionAndResource(
	apiversion string,
	resource string,
) *dynamicdiscovery.APIResource {
	if u.getAPIForAPIVersionAndResourceFn != nil {
		return u.getAPIForAPIVersionAndResourceFn(apiversion, resource)
	}
	return u.apiDiscovery.
		GetAPIForAPIVersionAndResource(
			apiversion,
			resource,
		)
}
