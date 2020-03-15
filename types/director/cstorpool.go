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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	// HTTPStatusOnline represents no errors in CStorPoolCapacityRecommendation
	HTTPStatusOnline string = "Online"

	// HTTPStatusError represents errors in CStorPoolCapacityRecommendation
	HTTPStatusError string = "Error"
)

const (
	// URLRecommendationList is the URL to list the recommendations
	URLRecommendationList string = URLDirector +
		"/v3/groups/{group_id}/recommendations/"

	// URLGetCSPCapacityRecommendation is the URL to get capacity
	// based recommendation for cstor pool
	URLGetCSPCapacityRecommendation string = URLDirector +
		"/v3/groups/{group_id}/recommendations/{recommendation_csp_id}/?action=getcapacityrecommendation"
)

// CStorPoolCapacityRecommendation is a kubernetes custom resource that
// defines the specifications to apply recommendation for cstorpool
// based on capacity
type CStorPoolCapacityRecommendation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec CStorPoolCapacityRecommendationSpec `json:"spec"`
}

// CStorPoolCapacityRecommendationSpec defines the configuration required
// to apply recommendation for cstorpool based on capacity
type CStorPoolCapacityRecommendationSpec struct {
	HTTPDataName    string `json:"httpDataName"`
	SecretName      string `json:"secretName"`
	RAIDDeviceCount int    `json:"raidDeviceCount"`
	RAIDType        string `json:"raidType"`
}
