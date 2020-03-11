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

	resty "github.com/go-resty/resty/v2"
	"openebs.io/metac/controller/generic"

	ctrlutil "mayadata.io/d-operators/common/controller"
	"mayadata.io/d-operators/common/unstruct"
	"mayadata.io/d-operators/types/gvk"
	types "mayadata.io/d-operators/types/http"
)

// Reconciler manages reconciliation of HTTP resources
type Reconciler struct {
	ctrlutil.Reconciler

	observedHTTP         *types.HTTP
	observedHTTPData     *types.HTTPData
	observedHTTPDataName string

	response *resty.Response
}

func (r *Reconciler) setObservedHTTP() {
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
}

func (r *Reconciler) setObservedHTTPDataName() {
	lbls := r.observedHTTP.GetLabels()
	if len(lbls) == 0 {
		return
	}
	r.observedHTTPDataName = lbls[types.LabelKeyHTTPDataName]
}

func (r *Reconciler) setObservedHTTPData() {
	if r.observedHTTPDataName == "" {
		r.Warns = append(
			r.Warns,
			"HTTPData resource not found",
		)
		return
	}
	var httpdata types.HTTPData
	obj := r.HookRequest.Attachments.FindByGroupKindName(
		gvk.APIVersionDAOV1Alpha1,
		gvk.KindHTTPData,
		r.observedHTTPDataName,
	)
	if obj == nil {
		r.Warns = append(
			r.Warns,
			"HTTPData resource %s not found",
			r.observedHTTPDataName,
		)
		return
	}
	err := unstruct.ToTyped(
		obj,
		&httpdata,
	)
	if err != nil {
		r.Err = err
		return
	}
	r.observedHTTPData = &httpdata
}

func (r *Reconciler) invokeHTTP() {
	req := resty.New().R()
	req.SetBody(r.observedHTTP.Spec.Body)
	req.SetHeaders(r.observedHTTP.Spec.Headers)
	req.SetQueryParams(r.observedHTTP.Spec.Params)
	req.SetPathParams(r.observedHTTPData.Spec.PathParams)

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
	if r.response != nil {
		var completion = map[string]interface{}{
			"state": true,
		}
		// set completion status
		status["completion"] = completion
	}
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
//	SyncHookRequest uses HTTP as the watched resource.
// The same watched resource forms the desired state by updating
// the its status.
func Sync(request *generic.SyncHookRequest, response *generic.SyncHookResponse) error {
	r := &Reconciler{}
	r.HookRequest = request
	r.HookResponse = response

	// add functions to achieve desired state
	r.ReconcileFns = []func(){
		r.setObservedHTTP,
		r.setObservedHTTPDataName,
		r.setObservedHTTPData,
		r.invokeHTTP,
	}

	// add functions to achieve desired watch
	r.DesiredWatchFns = []func(){
		r.updateWatchStatus,
	}
	// run reconcile
	return r.Reconcile()
}
