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

package doperator

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"openebs.io/metac/controller/generic"

	ctrlutil "mayadata.io/d-operators/common/controller"
	"mayadata.io/d-operators/common/unstruct"
	types "mayadata.io/d-operators/types/doperator"
)

// Reconciler manages CStorPoolAuto operational needs
// by reconciling CStorPoolAuto custom resource
type Reconciler struct {
	ctrlutil.Reconciler

	observedDOperator *types.DOperator
}

func (r *Reconciler) walkObservedDOperator() {
	var dope types.DOperator
	err := unstruct.ToTyped(
		r.HookRequest.Watch,
		&dope,
	)
	if err != nil {
		r.Err = err
		return
	}
	// set observed watch
	r.observedDOperator = &dope
}

func (r *Reconciler) setDesiredCRDBlockDeviceSet() {
	crd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1beta1",
			"kind":       "CustomResourceDefinition",
			"metadata": map[string]interface{}{
				"name": "blockdevicesets.dao.mayadata.io",
				"annotations": map[string]interface{}{
					// refer to the object that triggered this creation
					"doperator.dao.mayadata.io/uid": r.observedDOperator.UID,
				},
			},
			"spec": map[string]interface{}{
				"group":   "dao.mayadata.io",
				"version": "v1alpha1",
				"scope":   "Namespaced",
				"names": map[string]interface{}{
					"plural":   "blockdevicesets",
					"singular": "blockdeviceset",
					"kind":     "BlockDeviceSet",
					"shortNames": []interface{}{
						"bdset",
					},
				},
			},
		},
	}
	r.HookResponse.Attachments = append(
		r.HookResponse.Attachments,
		crd,
	)
}

func (r *Reconciler) setDesiredCRDHTTP() {
	crd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1beta1",
			"kind":       "CustomResourceDefinition",
			"metadata": map[string]interface{}{
				"name": "https.dao.mayadata.io",
				"annotations": map[string]interface{}{
					// refer to the object that triggered this creation
					"doperator.dao.mayadata.io/uid": r.observedDOperator.UID,
				},
			},
			"spec": map[string]interface{}{
				"group":   "dao.mayadata.io",
				"version": "v1alpha1",
				"scope":   "Namespaced",
				"names": map[string]interface{}{
					"plural":   "https",
					"singular": "http",
					"kind":     "HTTP",
					"shortNames": []interface{}{
						"http",
					},
				},
			},
		},
	}
	r.HookResponse.Attachments = append(
		r.HookResponse.Attachments,
		crd,
	)
}

func (r *Reconciler) setDesiredCRDHTTPData() {
	crd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1beta1",
			"kind":       "CustomResourceDefinition",
			"metadata": map[string]interface{}{
				"name": "httpdatas.dao.mayadata.io",
				"annotations": map[string]interface{}{
					// refer to the object that triggered this creation
					"doperator.dao.mayadata.io/uid": r.observedDOperator.UID,
				},
			},
			"spec": map[string]interface{}{
				"group":   "dao.mayadata.io",
				"version": "v1alpha1",
				"scope":   "Namespaced",
				"names": map[string]interface{}{
					"plural":   "httpdatas",
					"singular": "httpdata",
					"kind":     "HTTPData",
					"shortNames": []interface{}{
						"httpdata",
					},
				},
			},
		},
	}
	r.HookResponse.Attachments = append(
		r.HookResponse.Attachments,
		crd,
	)
}

func (r *Reconciler) setDesiredCRDCStorPoolCapacityRecommendation() {
	crd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1beta1",
			"kind":       "CustomResourceDefinition",
			"metadata": map[string]interface{}{
				"name": "cstorpoolcapacityrecommendations.dao.mayadata.io",
				"annotations": map[string]interface{}{
					// refer to the object that triggered this creation
					"doperator.dao.mayadata.io/uid": r.observedDOperator.UID,
				},
			},
			"spec": map[string]interface{}{
				"group":   "dao.mayadata.io",
				"version": "v1alpha1",
				"scope":   "Namespaced",
				"names": map[string]interface{}{
					"plural":   "cstorpoolcapacityrecommendations",
					"singular": "cstorpoolcapacityrecommendation",
					"kind":     "CStorPoolCapacityRecommendation",
					"shortNames": []interface{}{
						"cspcapr",
					},
				},
			},
		},
	}
	r.HookResponse.Attachments = append(
		r.HookResponse.Attachments,
		crd,
	)
}

func (r *Reconciler) setDesiredCRDDirectorHTTP() {
	crd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1beta1",
			"kind":       "CustomResourceDefinition",
			"metadata": map[string]interface{}{
				"name": "directorhttps.dao.mayadata.io",
				"annotations": map[string]interface{}{
					// refer to the object that triggered this creation
					"doperator.dao.mayadata.io/uid": r.observedDOperator.UID,
				},
			},
			"spec": map[string]interface{}{
				"group":   "dao.mayadata.io",
				"version": "v1alpha1",
				"scope":   "Namespaced",
				"names": map[string]interface{}{
					"plural":   "directorhttps",
					"singular": "directorhttp",
					"kind":     "DirectorHTTP",
					"shortNames": []interface{}{
						"drhttp",
					},
				},
			},
		},
	}
	r.HookResponse.Attachments = append(
		r.HookResponse.Attachments,
		crd,
	)
}

func (r *Reconciler) setDesiredCRDCStorPoolAuto() {
	crd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1beta1",
			"kind":       "CustomResourceDefinition",
			"metadata": map[string]interface{}{
				"name": "cstorpoolautos.dao.mayadata.io",
				"annotations": map[string]interface{}{
					// refer to the object that triggered this creation
					"doperator.dao.mayadata.io/uid": r.observedDOperator.UID,
				},
			},
			"spec": map[string]interface{}{
				"group":   "dao.mayadata.io",
				"version": "v1alpha1",
				"scope":   "Namespaced",
				"names": map[string]interface{}{
					"plural":   "cstorpoolautos",
					"singular": "cstorpoolauto",
					"kind":     "CStorPoolAuto",
					"shortNames": []interface{}{
						"cspauto",
					},
				},
			},
		},
	}
	r.HookResponse.Attachments = append(
		r.HookResponse.Attachments,
		crd,
	)
}

func (r *Reconciler) setDesiredCRDs() {
	r.setDesiredCRDBlockDeviceSet()
	r.setDesiredCRDHTTP()
	r.setDesiredCRDHTTPData()
	r.setDesiredCRDCStorPoolCapacityRecommendation()
	r.setDesiredCRDDirectorHTTP()
	r.setDesiredCRDCStorPoolAuto()
}

// Sync implements the idempotent logic to sync HTTP
//
// NOTE:
// 	SyncHookRequest is the payload received as part of reconcile
// request. Similarly, SyncHookResponse is the payload sent as a
// response as part of reconcile request.
//
// NOTE:
//	This controller watches DOperator custom resource
func Sync(request *generic.SyncHookRequest, response *generic.SyncHookResponse) error {
	r := &Reconciler{}
	r.HookRequest = request
	r.HookResponse = response
	// add functions to achieve desired state
	r.ReconcileFns = []func(){
		r.walkObservedDOperator,
		r.setDesiredCRDs,
	}
	// add functions to achieve desired watch
	r.DesiredWatchFns = []func(){}
	// run reconcile
	return r.Reconcile()
}
