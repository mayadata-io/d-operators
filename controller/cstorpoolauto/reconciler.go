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

package cstorpoolauto

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"openebs.io/metac/controller/generic"

	ctrlutil "mayadata.io/d-operators/common/controller"
	"mayadata.io/d-operators/common/unstruct"
	types "mayadata.io/d-operators/types/cstorpoolauto"
)

// Reconciler manages CStorPoolAuto operational needs
// by reconciling CStorPoolAuto custom resource
type Reconciler struct {
	ctrlutil.Reconciler

	observedCStorPoolAuto *types.CStorPoolAuto

	desiredLogLevel string
}

func (r *Reconciler) walkAndSetObservedCStorPoolAuto() {
	var cspauto types.CStorPoolAuto
	err := unstruct.ToTyped(
		r.HookRequest.Watch,
		&cspauto,
	)
	if err != nil {
		r.Err = err
		return
	}
	// set observed watch
	r.observedCStorPoolAuto = &cspauto
	// set the desired log level
	if cspauto.Spec.LogLevel == nil {
		r.desiredLogLevel = fmt.Sprintf("%d", 1)
	} else {
		r.desiredLogLevel = fmt.Sprintf("%d", *cspauto.Spec.LogLevel)
	}
}

func (r *Reconciler) setDesiredCRDCStorClusterConfig() {
	crd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1beta1",
			"kind":       "CustomResourceDefinition",
			"metadata": map[string]interface{}{
				"name": "cstorclusterconfigs.dao.mayadata.io",
				"annotations": map[string]interface{}{
					// refer to the watch that triggered this
					"cstorpoolauto.dao.mayadata.io/uid": r.observedCStorPoolAuto.UID,
				},
			},
			"spec": map[string]interface{}{
				"group":   "dao.mayadata.io",
				"version": "v1alpha1",
				"scope":   "Namespaced",
				"names": map[string]interface{}{
					"plural":   "cstorclusterconfigs",
					"singular": "cstorclusterconfig",
					"kind":     "CStorClusterConfig",
					"shortNames": []interface{}{
						"cscconfig",
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

func (r *Reconciler) setDesiredCRDCStorClusterPlan() {
	crd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1beta1",
			"kind":       "CustomResourceDefinition",
			"metadata": map[string]interface{}{
				"name": "cstorclusterplans.dao.mayadata.io",
				"annotations": map[string]interface{}{
					// refer to the watch that triggered this
					"cstorpoolauto.dao.mayadata.io/uid": r.observedCStorPoolAuto.UID,
				},
			},
			"spec": map[string]interface{}{
				"group":   "dao.mayadata.io",
				"version": "v1alpha1",
				"scope":   "Namespaced",
				"names": map[string]interface{}{
					"plural":   "cstorclusterplans",
					"singular": "cstorclusterplan",
					"kind":     "CStorClusterPlan",
					"shortNames": []interface{}{
						"cscplan",
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

func (r *Reconciler) setDesiredCRDCStorClusterStorageSet() {
	crd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1beta1",
			"kind":       "CustomResourceDefinition",
			"metadata": map[string]interface{}{
				"name": "cstorclusterstoragesets.dao.mayadata.io",
				"annotations": map[string]interface{}{
					// refer to the watch that triggered this
					"cstorpoolauto.dao.mayadata.io/uid": r.observedCStorPoolAuto.UID,
				},
			},
			"spec": map[string]interface{}{
				"group":   "dao.mayadata.io",
				"version": "v1alpha1",
				"scope":   "Namespaced",
				"names": map[string]interface{}{
					"plural":   "cstorclusterstoragesets",
					"singular": "cstorclusterstorageset",
					"kind":     "CStorClusterStorageSet",
					"shortNames": []interface{}{
						"cscstorageset",
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

func (r *Reconciler) setDesiredCRDStorage() {
	crd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1beta1",
			"kind":       "CustomResourceDefinition",
			"metadata": map[string]interface{}{
				"name": "storages.dao.mayadata.io",
				"annotations": map[string]interface{}{
					// refer to the watch that triggered this
					"cstorpoolauto.dao.mayadata.io/uid": r.observedCStorPoolAuto.UID,
				},
			},
			"spec": map[string]interface{}{
				"group":   "dao.mayadata.io",
				"version": "v1alpha1",
				"scope":   "Namespaced",
				"names": map[string]interface{}{
					"plural":   "storages",
					"singular": "storage",
					"kind":     "Storage",
					"shortNames": []interface{}{
						"stor",
					},
				},
				"additionalPrinterColumns": []interface{}{
					map[string]interface{}{
						"JSONPath":    ".spec.capacity",
						"name":        "Capacity",
						"description": "Capacity of storage",
						"type":        "string",
					},
					map[string]interface{}{
						"JSONPath":    ".spec.nodeName",
						"name":        "NodeName",
						"description": "Node where storage gets attached",
						"type":        "string",
					},
					map[string]interface{}{
						"JSONPath":    ".status.phase",
						"name":        "Status",
						"description": "Identifies the current status of storage",
						"type":        "string",
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
	if r.observedCStorPoolAuto.Spec.InstallCRD != nil &&
		!*r.observedCStorPoolAuto.Spec.InstallCRD {
		return
	}
	r.setDesiredCRDStorage()
	r.setDesiredCRDCStorClusterConfig()
	r.setDesiredCRDCStorClusterPlan()
	r.setDesiredCRDCStorClusterStorageSet()
}

func (r *Reconciler) setDesiredServiceAccount() {
	sa := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ServiceAccount",
			"metadata": map[string]interface{}{
				"name":      r.observedCStorPoolAuto.Spec.ServiceAccountName,
				"namespace": r.observedCStorPoolAuto.Spec.TargetNamespace,
				"annotations": map[string]interface{}{
					// refer to the watch that triggered this
					"cstorpoolauto.dao.mayadata.io/uid": r.observedCStorPoolAuto.UID,
				},
			},
		},
	}
	r.HookResponse.Attachments = append(
		r.HookResponse.Attachments,
		sa,
	)
}

func (r *Reconciler) setDesiredClusterRole() {
	crole := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "rbac.authorization.k8s.io/v1",
			"kind":       "ClusterRole",
			"metadata": map[string]interface{}{
				"name": r.observedCStorPoolAuto.Spec.ServiceAccountName,
				"annotations": map[string]interface{}{
					// refer to the watch that triggered this
					"cstorpoolauto.dao.mayadata.io/uid": r.observedCStorPoolAuto.UID,
				},
			},
			"rules": []interface{}{
				map[string]interface{}{
					"apiGroups": []interface{}{
						"openebs.io",
					},
					"resources": []interface{}{
						"blockdevices",
					},
					"verbs": []interface{}{
						"get",
						"list",
						"watch",
						"update",
					},
				},
				map[string]interface{}{
					"apiGroups": []interface{}{
						"",
					},
					"resources": []interface{}{
						"persistentvolumeclaims",
					},
					"verbs": []interface{}{
						"get",
						"list",
						"watch",
						"create",
						"update",
					},
				},
				map[string]interface{}{
					"apiGroups": []interface{}{
						"",
					},
					"resources": []interface{}{
						"customresourcedefinitions",
					},
					"verbs": []interface{}{
						"get",
						"list",
						"create",
						"update",
					},
				},
				map[string]interface{}{
					"apiGroups": []interface{}{
						"",
					},
					"resources": []interface{}{
						"nodes",
					},
					"verbs": []interface{}{
						"get",
						"list",
					},
				},
				map[string]interface{}{
					"apiGroups": []interface{}{
						"storage.k8s.io",
					},
					"resources": []interface{}{
						"volumeattachments",
					},
					"verbs": []interface{}{
						"get",
						"list",
						"watch",
						"create",
						"update",
					},
				},
				map[string]interface{}{
					"apiGroups": []interface{}{
						"dao.mayadata.io",
					},
					"resources": []interface{}{
						"storages",
						"cstorclusterconfigs",
						"cstorclusterplans",
						"cstorclusterstoragesets",
						"cstorpoolclusters",
					},
					"verbs": []interface{}{
						"get",
						"list",
						"watch",
						"create",
						"update",
					},
				},
			},
		},
	}
	r.HookResponse.Attachments = append(
		r.HookResponse.Attachments,
		crole,
	)
}

func (r *Reconciler) setDesiredClusterRoleBinding() {
	crb := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "rbac.authorization.k8s.io/v1",
			"kind":       "ClusterRoleBinding",
			"metadata": map[string]interface{}{
				"name": r.observedCStorPoolAuto.Spec.ServiceAccountName,
				"annotations": map[string]interface{}{
					// refer to the watch that triggered this
					"cstorpoolauto.dao.mayadata.io/uid": r.observedCStorPoolAuto.UID,
				},
			},
			"subjects": []interface{}{
				map[string]interface{}{
					"kind":      "ServiceAccount",
					"name":      r.observedCStorPoolAuto.Spec.ServiceAccountName,
					"namespace": r.observedCStorPoolAuto.Spec.TargetNamespace,
				},
			},
			"roleRef": map[string]interface{}{
				"kind":     "ClusterRole",
				"name":     r.observedCStorPoolAuto.Spec.ServiceAccountName,
				"apiGroup": "rbac.authorization.k8s.io",
			},
		},
	}
	r.HookResponse.Attachments = append(
		r.HookResponse.Attachments,
		crb,
	)
}

func (r *Reconciler) setDesiredRBAC() {
	if r.observedCStorPoolAuto.Spec.InstallRBAC != nil &&
		!*r.observedCStorPoolAuto.Spec.InstallRBAC {
		return
	}
	r.setDesiredServiceAccount()
	r.setDesiredClusterRole()
	r.setDesiredClusterRoleBinding()
}

func (r *Reconciler) setDesiredStorageProvisionerSTS() {
	cspauto := r.observedCStorPoolAuto
	sts := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "StatefulSet",
			"metadata": map[string]interface{}{
				"name":      "storage-provisioner",
				"namespace": cspauto.Spec.TargetNamespace,
				"labels": map[string]interface{}{
					"app.mayadata.io/name": "storage-provisioner",
				},
				"annotations": map[string]interface{}{
					// refer to the watch that triggered this
					"cstorpoolauto.dao.mayadata.io/uid": cspauto.UID,
				},
			},
			"spec": map[string]interface{}{
				"replicas": 1,
				"selector": map[string]interface{}{
					"matchLabels": map[string]interface{}{
						"app.mayadata.io/name": "storage-provisioner",
					},
				},
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"app.mayadata.io/name": "storage-provisioner",
						},
					},
					"spec": map[string]interface{}{
						"serviceAccountName": cspauto.Spec.ServiceAccountName,
						"containers": []interface{}{
							map[string]interface{}{
								"name":  "storage-provisioner",
								"image": cspauto.Spec.StorageProvisionerImage,
								"args": []interface{}{
									"--v=" + r.desiredLogLevel,
								},
								"env": []interface{}{
									map[string]interface{}{
										"name": "MY_NAME",
										"valueFrom": map[string]interface{}{
											"fieldRef": map[string]interface{}{
												"fieldPath": "metadata.name",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	r.HookResponse.Attachments = append(
		r.HookResponse.Attachments,
		sts,
	)
}

func (r *Reconciler) setDesiredCStorPoolAutoSTS() {
	cspauto := r.observedCStorPoolAuto
	sts := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "StatefulSet",
			"metadata": map[string]interface{}{
				"name":      "cstorpoolauto",
				"namespace": cspauto.Spec.TargetNamespace,
				"labels": map[string]interface{}{
					"app.mayadata.io/name": "cstorpoolauto",
				},
				"annotations": map[string]interface{}{
					// refer to the watch that triggered this
					"cstorpoolauto.dao.mayadata.io/uid": cspauto.UID,
				},
			},
			"spec": map[string]interface{}{
				"replicas": 1,
				"selector": map[string]interface{}{
					"matchLabels": map[string]interface{}{
						"app.mayadata.io/name": "cstorpoolauto",
					},
				},
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"app.mayadata.io/name": "cstorpoolauto",
						},
					},
					"spec": map[string]interface{}{
						"serviceAccountName": cspauto.Spec.ServiceAccountName,
						"containers": []interface{}{
							map[string]interface{}{
								"name":  "cstorpoolauto",
								"image": cspauto.Spec.CStorPoolAutoImage,
								"command": []interface{}{
									"/usr/bin/cstorpoolauto",
								},
								"args": []interface{}{
									"--logtostderr",
									"--run-as-local",
									"-v=" + r.desiredLogLevel,
									"--discovery-interval=40s",
									"--cache-flush-interval=240s",
								},
							},
						},
					},
				},
			},
		},
	}
	r.HookResponse.Attachments = append(
		r.HookResponse.Attachments,
		sts,
	)
}

// Sync implements the idempotent logic to sync HTTP
//
// NOTE:
// 	SyncHookRequest is the payload received as part of reconcile
// request. Similarly, SyncHookResponse is the payload sent as a
// response as part of reconcile request.
//
// NOTE:
//	This controller watches CStorPoolAuto custom resource
func Sync(request *generic.SyncHookRequest, response *generic.SyncHookResponse) error {
	r := &Reconciler{
		Reconciler: ctrlutil.Reconciler{
			HookRequest:  request,
			HookResponse: response,
		},
	}

	// add functions to achieve desired state
	r.ReconcileFns = []func(){
		r.setDesiredRBAC,
		r.setDesiredCRDs,
		r.setDesiredStorageProvisionerSTS,
		r.setDesiredCStorPoolAutoSTS,
	}

	// add functions to achieve desired watch
	r.DesiredWatchFns = []func(){}
	// run reconcile
	return r.Reconcile()
}
