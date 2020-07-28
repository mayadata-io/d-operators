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
	// SecretTypeBasicAuth refers to basic authentication based secret
	SecretTypeBasicAuth string = "kubernetes.io/basic-auth"

	// BasicAuthUsernameKey is the key used to refer the username for
	// SecretTypeBasicAuth secrets
	BasicAuthUsernameKey = "username"

	// BasicAuthPasswordKey is the key used to refer the password or
	// token for SecretTypeBasicAuth secrets
	BasicAuthPasswordKey = "password"
)

const (
	// POST based http request
	POST string = "POST"

	// GET based http request
	GET string = "GET"
)

// When is a typed definition to determine if HTTP custom resource
// is enabled to be reconciled
type When string

const (
	// Always flags the HTTP custom resource to be reconciled always
	Always When = "Always"

	// Never disables the HTTP custom resource from being reconciled
	Never When = "Never"

	// Once flags the HTTP custom resource to be reconciled only once
	// This is useful to invoke http requests to create or delete entity
	Once When = "Once"
)

// Enabled determines if HTTP custom resource is enabled to be
// reconciled
type Enabled struct {
	When When `json:"when,omitempty"`
}

// HTTP is a kubernetes custom resource that defines
// the specifications to invoke http request & store its
// response
type HTTP struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   HTTPRequestSpec   `json:"spec"`
	Status HTTPRequestStatus `json:"status,omitempty"`
}

// HTTPRequestSpec defines the configuration required
// to invoke http request
type HTTPRequestSpec struct {
	// Enabled flags this custom resource as enabled or disabled
	// for reconciliation
	Enabled *Enabled `json:"enabled,omitempty"`

	// Kubernetes secret to authorise the HTTP request
	SecretName string `json:"secretName"`

	// URL to be invoked
	URL string `json:"url"`

	// Post or Get call
	Method string `json:"method,omitempty"`

	// Headers used during API invocation
	Headers map[string]string `json:"headers,omitempty"`

	// QueryParams set against the URL query parameters
	QueryParams map[string]string `json:"queryParams,omitempty"`

	// PathParams set against the URL path
	PathParams map[string]string `json:"pathParams,omitempty"`

	// HTTP body used during API invocation
	Body string `json:"body,omitempty"`
}

// HTTPRequestStatus has the status & response of an
// invoked http URL
type HTTPRequestStatus struct {
	// Phase represents a single word status of http invocation
	Phase string `json:"phase"`

	// Reason reflects a decription of what happened after http invocation
	Reason string `json:"reason,omitempty"`

	// Warning message(s) if any
	Warn string `json:"warn,omitempty"`

	// Response received after invoking http request
	Response HTTPResponse `json:"response,omitempty"`
}

// HTTPResponse represents the response received after invoking http request
type HTTPResponse struct {
	Body           interface{} `json:"body,omitempty"`
	HTTPStatusCode int         `json:"httpStatusCode"`
	HTTPStatus     string      `json:"httpStatus"`
	HTTPError      interface{} `json:"httpError,omitempty"`
	IsError        bool        `json:"isError"`
}

const (
	// HTTPStatusPhaseError indicates error in HTTP invocation
	// if any
	HTTPStatusPhaseError string = "Error"

	// HTTPStatusPhaseOnline indicates last HTTP invocation
	// was successful
	HTTPStatusPhaseOnline string = "Online"
)

// HTTPData is a kubernetes custom resource that contains
// values requried to invoke HTTP APIs
type HTTPData struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec HTTPDataSpec `json:"spec"`
}

// HTTPDataSpec defines the values required to invoke
// http request
type HTTPDataSpec struct {
	Headers    map[string]string `json:"headers,omitempty"`
	PathParams map[string]string `json:"pathParams,omitempty"`
}
