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
	"sync"
	"time"

	"github.com/pkg/errors"
	apiextnv1beta1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/client-go/discovery"
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

// Utility provides options to invoke kubernetes APIs
type Utility struct {
	*UtilityFuncs
	Retry *Retryable

	kubeConfig *rest.Config

	apiResourceDiscovery *dynamicdiscovery.APIResourceDiscovery

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

var doOnce sync.Once

// singleton instance of Utility to be used across the
// project to invoke Kubernetes operations
var singleton *Utility

// Singleton returns a new or existing instance of Utility
func Singleton(config UtilityConfig) (*Utility, error) {
	if singleton != nil {
		return singleton, nil
	}
	var err error
	doOnce.Do(func() {
		singleton, err = NewUtility(config)
	})
	return singleton, err
}

// NewUtility returns a new instance of Kubernetes utility
func NewUtility(config UtilityConfig) (*Utility, error) {
	// initialize retry instance to default behaviour
	var retry = NewRetry(RetryConfig{})
	if config.Retry != nil {
		// override from config if available
		retry = config.Retry
	}

	// initialize utility instance
	u := &Utility{
		UtilityFuncs: &UtilityFuncs{},
		Retry:        retry,
		kubeConfig:   config.KubeConfig,
	}

	if u.kubeConfig == nil {
		// Set kube config to in-cluster config
		u.kubeConfig, u.err = rest.InClusterConfig()
		if u.err != nil {
			return nil, errors.Wrapf(
				u.err,
				"Failed to create k8s utility",
			)
		}
	}

	// setup options to mutate utility instance
	// based on the provided config
	//
	// NOTE:
	// 	Following order needs to be maintained
	var setters = []func(UtilityConfig){
		// pre settings
		u.setTeardownFlag,
		u.setAPIResourceDiscoveryOrDefault,

		// post settings
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

func (u *Utility) setTeardownFlag(config UtilityConfig) {
	u.isTeardown = config.IsTeardown
}

func (u *Utility) setAPIResourceDiscoveryOrDefault(config UtilityConfig) {
	u.apiResourceDiscovery = config.APIDiscovery
	if u.apiResourceDiscovery != nil {
		return
	}
	discoveryClient := discovery.NewDiscoveryClientForConfigOrDie(
		u.kubeConfig,
	)
	d := dynamicdiscovery.NewAPIResourceDiscoverer(discoveryClient)

	// This needs to be started with appropriate refresh interval
	// before being used. We set to 30 seconds discovery interval.
	// In other words, if new CRDs are applied then it will take
	// this interval for these new CRDs to be discovered.
	d.Start(time.Duration(30) * time.Second)

	u.apiResourceDiscovery = d
}

func (u *Utility) setCRDClient(config UtilityConfig) {
	u.crdClient, u.err = apiextnv1beta1.NewForConfig(
		u.kubeConfig,
	)
}

func (u *Utility) setDynamicClientset(config UtilityConfig) {
	u.dynamicClientset, u.err = dynamicclientset.New(
		u.kubeConfig,
		u.apiResourceDiscovery, // this must be set previously
	)
}

func (u *Utility) setKubeClientset(config UtilityConfig) {
	u.kubeClientset, u.err = kubernetes.NewForConfig(
		u.kubeConfig,
	)
}

// MustTeardown deletes resources created through this instance
func (u *Utility) MustTeardown() {
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

// Teardown optionally deletes resources created through this
// instance
func (u *Utility) Teardown() {
	if !u.isTeardown {
		return
	}
	u.MustTeardown()
}

// MustAddToTeardown adds the given teardown function to
// the list of teardown functions
func (u *Utility) MustAddToTeardown(teardown func() error) {
	if teardown == nil {
		return
	}
	u.teardownFuncs = append(u.teardownFuncs, teardown)
}

// AddToTeardown optionally adds the given teardown function to
// the list of teardown functions
func (u *Utility) AddToTeardown(teardown func() error) {
	if !u.isTeardown {
		return
	}
	u.MustAddToTeardown(teardown)
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
// given api version & resource name
func (u *Utility) GetClientForAPIVersionAndResource(
	apiversion string,
	resourceName string,
) (*clientset.ResourceClient, error) {
	if u.getClientForAPIVersionAndResourceFn != nil {
		return u.getClientForAPIVersionAndResourceFn(apiversion, resourceName)
	}
	return u.dynamicClientset.GetClientForAPIVersionAndResource(
		apiversion,
		resourceName,
	)
}

// GetAPIResourceForAPIVersionAndResourceName returns the
// discovered API resource based on the provided api version &
// resource name
func (u *Utility) GetAPIResourceForAPIVersionAndResourceName(
	apiversion string,
	resourceName string,
) *dynamicdiscovery.APIResource {
	if u.getAPIForAPIVersionAndResourceFn != nil {
		return u.getAPIForAPIVersionAndResourceFn(apiversion, resourceName)
	}
	return u.apiResourceDiscovery.GetAPIForAPIVersionAndResource(
		apiversion,
		resourceName,
	)
}

// GetAPIResourceDiscovery returns the api resource discovery instance
func (u *Utility) GetAPIResourceDiscovery() *dynamicdiscovery.APIResourceDiscovery {
	return u.apiResourceDiscovery
}

// GetKubeConfig returns the Kubernetes config instance
func (u *Utility) GetKubeConfig() *rest.Config {
	return u.kubeConfig
}
