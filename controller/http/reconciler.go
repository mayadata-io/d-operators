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

package http

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"openebs.io/metac/controller/generic"

	ctrlutil "mayadata.io/d-operators/common/controller"
	"mayadata.io/d-operators/common/unstruct"
	"mayadata.io/d-operators/pkg/http"
	types "mayadata.io/d-operators/types/http"
)

// Reconciler manages reconciliation of HTTP resources
type Reconciler struct {
	// common Reconciler is embedded to enable using its
	// properties & methods
	ctrlutil.Reconciler

	// http object that is received as part of reconciliation
	observedHTTP *types.HTTP

	// secret object that is received as part of reconciliation
	observedSecret *unstructured.Unstructured

	// name of the secret that is received as part of reconciliation
	observedSecretName string

	// username to invoke authenticated APIs
	observedUsername string
	// passwrod to invoke authenticated APIs
	observedPassword string

	// response received after invoking the http request
	response types.HTTPResponse
}

// evalObservedHTTP tranforms the HTTP custom resource observed in Kubernetes
// cluster to its corresponding typed definition. The observed HTTP
// resource is received as an unstructured instance. Unstructured instances are
// not friendly to parse. Hence, the need for this method.
//
// NOTE:
//	Executing this method does not imply catching validation errors found
// in HTTP custom resource
func (r *Reconciler) evalObservedHTTP() {
	var http types.HTTP
	err := unstruct.ToTyped(
		r.HookRequest.Watch,
		&http,
	)
	if err != nil {
		r.Err = err
		return
	}
	r.observedHTTP = &http

	// validate presence of secret
	r.observedSecretName = http.Spec.SecretName
	if r.observedSecretName == "" {
		r.Err = errors.Errorf("Missing spec.secretName")
		return
	}
}

// evalObservedSecret parses the relevant secret found in Kubernetes
// cluster. This secret is used to authenticate the request invocation.
func (r *Reconciler) evalObservedSecret() {
	r.observedSecret = r.HookRequest.Attachments.FindByGroupKindName(
		"v1",
		"Secret",
		r.observedSecretName,
	)
	if r.observedSecret == nil {
		r.Err = errors.Errorf(
			"Secret not found: Name %q",
			r.observedSecretName,
		)
		return
	}
	// add it back to response attachments without any
	// changes
	//
	// NOTE:
	//	Adding back the request attachments to response
	// attachments help in evaluating the completion state
	r.HookResponse.Attachments = append(
		r.HookResponse.Attachments,
		r.observedSecret,
	)
	// parse secret type
	stype, found, err := unstructured.NestedString(
		r.observedSecret.Object,
		"type",
	)
	if err != nil {
		r.Err = err
		return
	}
	if !found || stype == "" {
		r.Warns = append(
			r.Warns,
			"Secret type not found: Name %q: Will default to %q",
			r.observedSecretName,
			types.SecretTypeBasicAuth,
		)
	}
	// BasicAuth is the only supported secret
	if stype != types.SecretTypeBasicAuth {
		r.Err = errors.Errorf(
			"Unsupported secret: Name %q: Type %q",
			r.observedSecretName,
			stype,
		)
		return
	}
	data, found, err := unstructured.NestedMap(
		r.observedSecret.Object,
		"data",
	)
	if err != nil {
		r.Err = err
		return
	}
	if data == nil || !found {
		r.Err = errors.Errorf(
			"Secret data not found: Name %q",
			r.observedSecretName,
		)
	}
	// save the credentials to be used later
	r.observedUsername = fmt.Sprintf("%v", data[types.BasicAuthUsernameKey])
	r.observedPassword = fmt.Sprintf("%v", data[types.BasicAuthPasswordKey])
	// validate presence of credentials
	if r.observedUsername == "" || r.observedPassword == "" {
		r.Err = errors.Errorf(
			"Missing credentials in secret: Name %q",
			r.observedSecretName,
		)
		return
	}
}

func (r *Reconciler) skipReconcile() {
	// HTTP resource does not have any attachments to be
	// reconciled. Hence skip reconcile flag is set to true
	r.HookResponse.SkipReconcile = true
}

func (r *Reconciler) invokeHTTP() {
	i := http.Invoker(http.InvocableConfig{
		Username:    r.observedUsername,
		Password:    r.observedPassword,
		HTTPMethod:  r.observedHTTP.Spec.Method,
		URL:         r.observedHTTP.Spec.URL,
		Body:        r.observedHTTP.Spec.Body,
		Headers:     r.observedHTTP.Spec.Headers,
		PathParams:  r.observedHTTP.Spec.PathParams,
		QueryParams: r.observedHTTP.Spec.QueryParams,
	})
	r.response, r.Err = i.Invoke()
}

// handleRuntimeError handles runtime error if any
func (r *Reconciler) handleRuntimeError() {
	if r.Err == nil {
		// nothing to do if there was no error
		return
	}
	r.HookResponse.Status = map[string]interface{}{
		"phase":  types.HTTPStatusPhaseError,
		"reason": r.Err.Error(),
	}
}

// updateWatchStatus updates the watched HTTP resource's
// status field with the response received due to invocation
// of HTTP url.
//
// NOTE:
//	Status forms the core business logic of reconciling a HTTP
// custom resource.
func (r *Reconciler) updateWatchStatus() {
	// check for runtime errors
	if r.Err != nil {
		r.handleRuntimeError()
		// skip setting other status fields
		return
	}

	// initialise phase to Online
	var phase = types.HTTPStatusPhaseOnline
	var warn, reason string

	// check for warnings
	if len(r.Warns) != 0 {
		warn = fmt.Sprintf(
			"%d warnings: %s",
			len(r.Warns),
			strings.Join(r.Warns, ": "),
		)
	}

	// check for error from response
	if r.response.IsError {
		phase = types.HTTPStatusPhaseError
	}

	// set the desired status
	r.HookResponse.Status = map[string]interface{}{
		"phase":  phase,
		"reason": reason,
		"warn":   warn,
		"response": map[string]interface{}{
			"body":           r.response.Body,
			"httpStatusCode": r.response.HTTPStatusCode,
			"httpStatus":     r.response.HTTPStatus,
			"httpError":      r.response.HTTPError,
			"isError":        r.response.IsError,
		},
	}
}

// Sync implements the idempotent logic to reconcile HTTP
// custom resource.
//
// NOTE:
// 	SyncHookRequest is the payload received as part of reconcile
// request. Similarly, SyncHookResponse is the payload sent as a
// response.
//
// NOTE:
//	This controller watches HTTP custom resource. In other words,
// HTTP custom resource is the watch.
func Sync(request *generic.SyncHookRequest, response *generic.SyncHookResponse) error {
	r := &Reconciler{
		Reconciler: ctrlutil.Reconciler{
			HookRequest:  request,
			HookResponse: response,
		},
	}

	// Add functions to achieve desired state by parsing
	// the watch & attachments observed in the Kubernetes cluster
	r.ReconcileFns = []func(){
		r.evalObservedHTTP,
		r.evalObservedSecret,
		r.skipReconcile,
		r.invokeHTTP,
	}

	// Add functions to update the watch
	//
	// NOTE:
	//	One can add functions to set the watch's labels,
	// annotations, & status.
	//
	// NOTE:
	//	HTTP custom resource is the watch here
	r.DesiredWatchFns = []func(){
		r.updateWatchStatus,
	}

	// run reconcile
	return r.Reconcile()
}
