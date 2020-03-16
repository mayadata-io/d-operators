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

	resty "github.com/go-resty/resty/v2"
	"openebs.io/metac/controller/generic"

	ctrlutil "mayadata.io/d-operators/common/controller"
	"mayadata.io/d-operators/common/unstruct"
	types "mayadata.io/d-operators/types/http"
)

const (
	// SecretTypeBasicAuth refers to basic authentication based secret
	SecretTypeBasicAuth string = "kubernetes.io/basic-auth"

	// BasicAuthUsernameKey is the key of the username for
	// SecretTypeBasicAuth secrets
	BasicAuthUsernameKey = "username"

	// BasicAuthPasswordKey is the key of the password or token for
	// SecretTypeBasicAuth secrets
	BasicAuthPasswordKey = "password"
)

// Reconciler manages reconciliation of HTTP resources
type Reconciler struct {
	ctrlutil.Reconciler

	observedHTTP       *types.HTTP
	observedSecretName string
	observedSecret     *unstructured.Unstructured
	observedUsername   string
	observedPassword   string

	response *resty.Response
}

func (r *Reconciler) walkObservedHTTP() {
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
	r.observedSecretName = http.Spec.SecretName
	if r.observedSecretName == "" {
		r.Err = errors.Errorf("Missing spec.secretName")
		return
	}
}

func (r *Reconciler) walkObservedSecret() {
	r.observedSecret = r.HookRequest.Attachments.FindByGroupKindName(
		"v1",
		"Secret",
		r.observedSecretName,
	)
	if r.observedSecret == nil {
		r.Err = errors.Errorf(
			"Secret %q not found",
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
			"Secret type not found in %q: Will use %q",
			r.observedSecretName,
			SecretTypeBasicAuth,
		)
	}
	if stype != SecretTypeBasicAuth {
		r.Err = errors.Errorf(
			"Unsupported secret type %s: %q",
			stype,
			r.observedSecretName,
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
			"Secret data not found: %q",
			r.observedSecretName,
		)
	}
	r.observedUsername = fmt.Sprintf("%v", data[BasicAuthUsernameKey])
	r.observedPassword = fmt.Sprintf("%v", data[BasicAuthPasswordKey])
	if r.observedUsername == "" || r.observedPassword == "" {
		r.Err = errors.Errorf(
			"Missing credentials in secret %q",
			r.observedSecretName,
		)
		return
	}
}

func (r *Reconciler) invokeHTTP() {
	req := resty.New().R().
		SetBasicAuth(r.observedUsername, r.observedPassword).
		SetBody(r.observedHTTP.Spec.Body).
		SetHeaders(r.observedHTTP.Spec.Headers).
		SetQueryParams(r.observedHTTP.Spec.QueryParams).
		SetPathParams(r.observedHTTP.Spec.PathParams)

	switch strings.ToLower(r.observedHTTP.Spec.Method) {
	case types.POST:
		r.response, r.Err = req.Post(r.observedHTTP.Spec.URL)
	case types.GET:
		r.response, r.Err = req.Get(r.observedHTTP.Spec.URL)
	default:
		r.Err = errors.Errorf(
			"Unsupported http method %s",
			r.observedHTTP.Spec.Method,
		)
	}
}

// updateWatchStatus updates the watched HTTP resource's
// status field with the response received due to invocation
// of HTTP url.
//
// NOTE:
//	This forms the core business logic of reconciling a HTTP
// custom resource.
func (r *Reconciler) updateWatchStatus() {
	var status = map[string]interface{}{}
	var completion = map[string]interface{}{
		"state": false,
	}
	var warn string
	// init with Online
	status["phase"] = types.HTTPStatusPhaseOnline
	// check for warnings
	if len(r.Warns) != 0 {
		warn = fmt.Sprintf(
			"%d warnings: %s",
			len(r.Warns),
			strings.Join(r.Warns, ": "),
		)
	}
	if warn != "" {
		status["warn"] = warn
	}
	// is runtime error
	if r.Err != nil {
		status["phase"] = types.HTTPStatusPhaseError
		status["reason"] = r.Err.Error()
	}
	// set http response details
	if r.response != nil {
		status["body"] = r.response.Result()
		status["httpStatusCode"] = r.response.StatusCode()
		status["httpStatus"] = r.response.Status()
		status["httpError"] = r.response.Error()
	}
	// check for error again & set/override the phase
	if r.response != nil && r.response.IsError() {
		status["phase"] = types.HTTPStatusPhaseError
	}
	// check for completion state
	if r.Err == nil && r.response != nil && !r.response.IsError() {
		completion["state"] = true
	}
	// set completion status
	status["completion"] = completion
	// set the desired status
	r.HookResponse.Status = status
}

// Sync implements the idempotent logic to sync HTTP
//
// NOTE:
// 	SyncHookRequest is the payload received as part of reconcile
// request. Similarly, SyncHookResponse is the payload sent as a
// response as part of reconcile request.
//
// NOTE:
//	This controller watches HTTP custom resource
func Sync(request *generic.SyncHookRequest, response *generic.SyncHookResponse) error {
	r := &Reconciler{}
	r.HookRequest = request
	r.HookResponse = response

	// add functions to achieve desired state
	r.ReconcileFns = []func(){
		r.walkObservedHTTP,
		r.walkObservedSecret,
		r.invokeHTTP,
	}

	// add functions to achieve desired watch
	r.DesiredWatchFns = []func(){
		r.updateWatchStatus,
	}
	// run reconcile
	return r.Reconcile()
}
