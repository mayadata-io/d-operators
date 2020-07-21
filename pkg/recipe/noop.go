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

package recipe

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"openebs.io/metac/dynamic/clientset"
	dynamicdiscovery "openebs.io/metac/dynamic/discovery"
)

// NoopFixture is a BaseFixture instance useful for unit testing
var NoopFixture = &BaseFixture{

	getClientForAPIVersionAndKindFn: func(
		apiversion string,
		kind string,
	) (*clientset.ResourceClient, error) {
		di := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())
		nri := di.Resource(schema.GroupVersionResource{})
		ri := nri.Namespace("")
		return &clientset.ResourceClient{
			ResourceInterface: ri,
			APIResource:       &dynamicdiscovery.APIResource{},
		}, nil
	},

	getClientForAPIVersionAndResourceFn: func(
		apiversion string,
		resource string,
	) (*clientset.ResourceClient, error) {
		di := &dynamicfake.FakeDynamicClient{}
		nri := di.Resource(schema.GroupVersionResource{})
		ri := nri.Namespace("")
		return &clientset.ResourceClient{
			ResourceInterface: ri,
			APIResource:       &dynamicdiscovery.APIResource{},
		}, nil
	},

	getAPIForAPIVersionAndResourceFn: func(
		apiversion string,
		resource string,
	) *dynamicdiscovery.APIResource {
		return &dynamicdiscovery.APIResource{}
	},
}

// NoopConfigMapFixture is a BaseFixture instance useful for unit testing
var NoopConfigMapFixture = &BaseFixture{

	getClientForAPIVersionAndKindFn: func(
		apiversion string,
		kind string,
	) (*clientset.ResourceClient, error) {
		di := dynamicfake.NewSimpleDynamicClient(
			runtime.NewScheme(),
			&unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":       "ConfigMap",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name": "cm-1",
					},
					"spec": nil,
				},
			},
		)
		nri := di.Resource(schema.GroupVersionResource{
			Version:  "v1",
			Resource: "configmaps",
		})
		ri := nri.Namespace("")
		return &clientset.ResourceClient{
			ResourceInterface: ri,
			APIResource:       &dynamicdiscovery.APIResource{},
		}, nil
	},

	getClientForAPIVersionAndResourceFn: func(
		apiversion string,
		resource string,
	) (*clientset.ResourceClient, error) {
		di := dynamicfake.NewSimpleDynamicClient(
			runtime.NewScheme(),
			&unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":       "ConfigMap",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name": "cm-1",
					},
					"spec": nil,
				},
			},
		)
		nri := di.Resource(schema.GroupVersionResource{
			Version:  "v1",
			Resource: "configmaps",
		})
		ri := nri.Namespace("")
		return &clientset.ResourceClient{
			ResourceInterface: ri,
			APIResource:       &dynamicdiscovery.APIResource{},
		}, nil
	},

	getAPIForAPIVersionAndResourceFn: func(
		apiversion string,
		resource string,
	) *dynamicdiscovery.APIResource {
		return &dynamicdiscovery.APIResource{}
	},
}
