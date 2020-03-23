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

package types

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CStorPoolAuto is a kubernetes custom resource that defines
// the specifications to manage CStorPoolAuto needs
type CStorPoolAuto struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec CStorPoolAutoSpec `json:"spec"`
}

// CStorPoolAutoSpec defines the configuration required
// to manage CStorPoolAuto
//
// rough draft
// ```yaml
// kind: CStorPoolAuto
// spec:
//   rbac:
//     disable: 			// if true disables installing any rbac
//     items:
//     - cstorpoolauto-local
//     - "*"  				// default internally
//   crd:
//     disable: 			// if true disables installing any crds
//     items:
//     - group: "*" 		// default internally
//       version: "*"
//       resource: "*"
//     - group: dao.mayadata.io
//       version: v1alpha1
//       resource: cstorclusterconfigs
//   deploy:
//     uses:
//       namespace:
//       serviceAccountName:
//     cstorpoolauto:
//       image:
//       loglevel:
//       config:
//     storageprovisioner:
//       image:
//       loglevel:
//       config:
// ```
type CStorPoolAutoSpec struct {
	InstallRBAC             *bool  `json:"installRBAC,omitempty"`
	InstallCRD              *bool  `json:"installCRD,omitempty"`
	TargetNamespace         string `json:"targetNamespace"`
	ServiceAccountName      string `json:"serviceAccountName"`
	CStorPoolAutoImage      string `json:"cstorPoolAutoImage"`
	StorageProvisionerImage string `json:"storageProvisionerImage"`
	LogLevel                *int   `json:"logLevel,omitempty"`
}
