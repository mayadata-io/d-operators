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

const (
	// LabelKeyBlockDeviceCloneName is the label key that holds
	// the name of BlockDeviceClone
	LabelKeyBlockDeviceCloneName string = "blockdeviceset-dao-mayadata-io/name"
)

const (
	// BlockDeviceSetStatusOnline represents no errors at BlockDeviceSet
	BlockDeviceSetStatusOnline string = "Online"

	// BlockDeviceSetStatusError represent error at BlockDeviceSet
	BlockDeviceSetStatusError string = "Error"
)

// BlockDeviceSet is a kubernetes custom resource that defines
// the specifications to create one or more BlockDevices
type BlockDeviceSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec BlockDeviceSetSpec `json:"spec"`
}

// BlockDeviceSetSpec defines the configuration required
// to create one or more BlockDevices
type BlockDeviceSetSpec struct {
	Device   map[string]interface{} `json:"device,omitempty"`
	Replicas *int                   `json:"replicas,omitempty"`
}
