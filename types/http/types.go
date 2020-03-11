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
	// LabelKeyHTTPDataName is the key used in labels to
	// provides the name of HTTPData CR
	LabelKeyHTTPDataName string = "httpdata-dao-mayadata-io/name"
)

// HTTP is a kubernetes custom resource that defines
// the specifications to invoke http request & store its
// response
type HTTP struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec HTTPRequestSpec `json:"spec"`
}

const (
	// POST based http request
	POST string = "post"

	// GET based http request
	GET string = "get"
)

// HTTPRequestSpec defines the configuration required
// to invoke http request
type HTTPRequestSpec struct {
	URL     string            `json:"url"`
	Method  string            `json:"method,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Params  map[string]string `json:"params,omitempty"`
	Body    string            `json:"body,omitempty"`
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
	PathParams map[string]string      `json:"params,omitempty"`
	Values     map[string]interface{} `json:"values,omitempty"`
}
