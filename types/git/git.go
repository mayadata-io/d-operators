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
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metac "openebs.io/metac/apis/metacontroller/v1alpha1"
)

// GitUpload is a kubernetes custom resource that defines
// the specifications to upload kubernetes resources and
// pod logs
type GitUpload struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   GitUploadSpec   `json:"spec"`
	Status GitUploadStatus `json:"status"`
}

// GitUploadSpec defines the specifications to upload
// kubernetes resources and pod logs
type GitUploadSpec struct {
	ResourceSelector []metac.GenericControllerResource `json:"resourceSelector,omitempty"`
}

// GitUploadStatus holds the status of executing a GitUpload
type GitUploadStatus struct {
	Phase   string `json:"phase"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
}

// String implements the Stringer interface
func (jr GitUploadStatus) String() string {
	raw, err := json.MarshalIndent(
		jr,
		" ",
		".",
	)
	if err != nil {
		panic(err)
	}
	return string(raw)
}
