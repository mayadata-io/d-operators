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
	// URLDirector has the base URL path to invoke Director APIs
	URLDirector string = "http://{director_ip}:30380"

	// URLActiveNodes is the URL to fetch active nodes
	URLActiveNodes string = URLDirector +
		"/v3/groups/{group_id}/nodes?state=active&clusterId={cluster_id}"

	// URLProjectDetails is the URL to fetch project details
	URLProjectDetails string = URLDirector +
		"/v3/groups/{group_id}/project"
)

const (
	// DirectorHTTPStatusOnline represents no error in DirectorHTTP
	DirectorHTTPStatusOnline string = "Online"

	// DirectorHTTPStatusError represents error in DirectorHTTP
	DirectorHTTPStatusError string = "Error"
)

// DirectorHTTP is a kubernetes custom resource that defines
// the specifications to invoke Director APIs
type DirectorHTTP struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec DirectorHTTPSpec `json:"spec"`
}

// DirectorHTTPSpec defines the configuration required
// to invoke one or more Director APIs
type DirectorHTTPSpec struct {
	HTTPDataName string `json:"httpDataName"`
	SecretName   string `json:"secretName"`

	// Include the API names that should be invoked
	//
	// NOTE:
	// 	Use wildcard * to include all
	Include []string `json:"include,omitempty"`
}

const (
	// IncludeAllAPIs indicates DirectorHTTP to include all
	// its registered APIs
	IncludeAllAPIs string = "*"

	// GetActiveNodes represent the Director API to fetch
	// active node details
	GetActiveNodes string = "get-active-nodes"

	// GetProjectDetails represent the Director API to fetch
	// project details
	GetProjectDetails string = "get-project-details"
)
