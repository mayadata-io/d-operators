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

package controller

import (
	"fmt"
	"strings"

	"k8s.io/klog/v2"
	"openebs.io/metac/controller/generic"

	metacutil "mayadata.io/d-operators/common/metac"
)

// Reconciler is the base structure to faciliate reconciliation
type Reconciler struct {
	Name         string
	HookRequest  *generic.SyncHookRequest
	HookResponse *generic.SyncHookResponse

	// desired state functions
	PreReconcileFns  []func()
	ReconcileFns     []func()
	PostReconcileFns []func()

	// desired state for watch
	DesiredWatchFns []func()

	SkipReason string
	Fatal      error
	Err        error
	Warns      []string
}

// validateHook validates the hook request & response
func (r *Reconciler) validateHook() {
	// validation failure of request &/ response is a fatal error
	r.Fatal = metacutil.ValidateGenericControllerArgs(
		r.HookRequest,
		r.HookResponse,
	)
}

// logSyncStart logs the start of sync
func (r *Reconciler) logSyncStart() {
	klog.V(3).Infof(
		"Starting sync: Controller %q: %s",
		r.Name,
		metacutil.GetDetailsFromRequest(r.HookRequest),
	)
}

// logSyncFinish logs the completion of sync
func (r *Reconciler) logSyncFinish() {
	klog.V(3).Infof(
		"Completed sync: Controller %q: Request %s: Response %s",
		r.Name,
		metacutil.GetDetailsFromRequest(r.HookRequest),
		metacutil.GetDetailsFromResponse(r.HookResponse),
	)
}

// updateWatchStatus updates the watch's status fields
func (r *Reconciler) updateWatchStatus() {
	var status = map[string]interface{}{}
	var completion = map[string]interface{}{
		"state": false,
	}
	var warn string
	if r.Err != nil {
		status["phase"] = "Error"
		status["reason"] = r.Err.Error()
	} else {
		status["phase"] = "Online"
	}
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
	observedAttachments := r.HookRequest.Attachments.Len()
	desiredAttachments := len(r.HookResponse.Attachments)
	if r.Err == nil && observedAttachments == desiredAttachments {
		completion["state"] = true
	}
	// set completion against the status
	status["completion"] = completion
	// set the desired status
	r.HookResponse.Status = status
}

// handleError logs the error if any
func (r *Reconciler) handleError() {
	if r.Err == nil {
		// nothing to do if there was no error
		return
	}
	// log this error with context
	klog.Errorf(
		"Reconcile failed: Controller %q: Name %q %q: Error %s",
		r.Name,
		r.HookRequest.Watch.GetNamespace(),
		r.HookRequest.Watch.GetName(),
		r.Err.Error(),
	)
	// stop further reconciliation at metac since there was an error
	r.HookResponse.SkipReconcile = true
	r.SkipReason = r.Err.Error()
}

// Reconcile runs the reconciliation to achieve to desired state
func (r *Reconciler) Reconcile() error {
	if len(r.PreReconcileFns) == 0 {
		r.PreReconcileFns = []func(){
			r.logSyncStart,
			r.validateHook,
		}
	}
	if len(r.PostReconcileFns) == 0 {
		r.PostReconcileFns = []func(){
			r.logSyncFinish,
		}
	}
	if len(r.DesiredWatchFns) == 0 {
		r.DesiredWatchFns = []func(){
			r.updateWatchStatus,
		}
	}
	var reconFns = []func(){}
	reconFns = append(reconFns, r.PreReconcileFns...)
	reconFns = append(reconFns, r.ReconcileFns...)
	reconFns = append(reconFns, r.PostReconcileFns...)
	for _, fn := range reconFns {
		fn()
		// post operation checks
		if r.Fatal != nil {
			// this will panic
			return r.Fatal
		}
		if r.Err != nil {
			// this logs the error thus avoiding panic in the
			// controller
			r.handleError()
			// break out of the reconcile functions
			break
		}
	}
	// desired watch functions are run even if reconcile functions
	// result in any error
	for _, dWatchFn := range r.DesiredWatchFns {
		dWatchFn()
	}
	// check if attachments / children need not be reconciled
	if r.HookResponse.SkipReconcile {
		klog.V(3).Infof(
			"Will skip reconcile: Controller %q: Name %q %q: %s",
			r.Name,
			r.HookRequest.Watch.GetNamespace(),
			r.HookRequest.Watch.GetName(),
			r.SkipReason,
		)
	}
	return nil
}
