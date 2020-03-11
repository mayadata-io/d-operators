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

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"openebs.io/metac/controller/generic"

	ctrlutil "mayadata.io/d-operators/common/controller"
	stringutil "mayadata.io/d-operators/common/string"
	"mayadata.io/d-operators/common/unstruct"
	types "mayadata.io/d-operators/types/director"
	"mayadata.io/d-operators/types/gvk"
	http "mayadata.io/d-operators/types/http"
)

// Reconciler manages reconciliation of HTTP resources
type Reconciler struct {
	ctrlutil.Reconciler

	observedDirector          *types.DirectorHTTP
	observedHTTPData          *http.HTTPData
	observedIncludes          stringutil.List
	observedDirectorName      string
	observedDirectorNamespace string

	registeredAPIs []func()
	desiredStates  []*unstructured.Unstructured
}

func (r *Reconciler) setObservedDirector() {
	var director types.DirectorHTTP
	err := unstruct.ToTyped(
		r.HookRequest.Watch,
		&director,
	)
	if err != nil {
		r.Err = err
		return
	}
	r.observedDirector = &director
	r.observedIncludes = director.Spec.Include
	r.observedDirectorName = director.GetName()
	r.observedDirectorNamespace = director.GetNamespace()
}

func (r *Reconciler) setObservedHTTPData() {
	var httpdata http.HTTPData
	obj := r.HookRequest.Attachments.FindByGroupKindName(
		gvk.APIVersionDAOV1Alpha1,
		gvk.KindHTTPData,
		r.observedDirectorName,
	)
	if obj == nil {
		r.Warns = append(
			r.Warns,
			"HTTPData resource %s not found",
			r.observedDirectorName,
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

func (r *Reconciler) buildDesiredState() {
	// run the registered APIs
	for _, fn := range r.registeredAPIs {
		fn()
		if r.Err != nil {
			return
		}
	}
	r.HookResponse.Attachments = r.desiredStates
}

func (r *Reconciler) updateWatchStatus() {
	var status = map[string]interface{}{}
	var warn string
	// init with Online
	status["phase"] = types.DirectorHTTPStatusOnline
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
		status["phase"] = types.DirectorHTTPStatusError
		status["reason"] = r.Err.Error()
	}
	// check for completion state
	var completion = map[string]interface{}{}
	// hook request has the observed state of children
	observedAPIs := r.HookRequest.Attachments.Len()
	// hook response has the desired state of children
	desiredAPIs := len(r.HookResponse.Attachments)
	if observedAPIs != desiredAPIs {
		completion["state"] = false
	}
	completion["observedAPICount"] = observedAPIs
	completion["desiredAPICount"] = desiredAPIs
	// set completion status
	status["completion"] = completion
	// set the desired status against hook response
	r.HookResponse.Status = status
}

// Sync implements the idempotent logic to sync DirectorHTTP
//
// NOTE:
// 	SyncHookRequest is the payload received as part of reconcile
// request. Similarly, SyncHookResponse is the payload sent as a
// response as part of reconcile request.
//
// NOTE:
//	SyncHookRequest uses DirectorHTTP as the watched resource.
func Sync(request *generic.SyncHookRequest, response *generic.SyncHookResponse) error {
	r := &Reconciler{}
	r.HookRequest = request
	r.HookResponse = response

	// add logic to achieve desired state of attachments/children
	r.ReconcileFns = []func(){
		r.setObservedDirector,
		r.registerAPIs,
		r.buildDesiredState,
	}

	// add logic to achieve desired state of watch
	r.DesiredWatchFns = []func(){
		r.updateWatchStatus,
	}

	// run reconcile
	return r.Reconcile()
}
