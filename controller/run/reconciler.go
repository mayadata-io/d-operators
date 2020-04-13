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

package run

import (
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"openebs.io/metac/controller/generic"

	ctrlutil "mayadata.io/d-operators/common/controller"
	"mayadata.io/d-operators/common/unstruct"
	types "mayadata.io/d-operators/types/run"
)

// Reconciler manages reconciliation of HTTP resources
type Reconciler struct {
	ctrlutil.Reconciler

	isWatchAndRunSame bool
	observedRun       *types.Run

	runResponse *Response
}

func (r *Reconciler) evalRun() {
	var run types.Run
	// convert from unstructured instance to typed Run instance
	err := unstruct.ToTyped(
		r.HookRequest.Watch,
		&run,
	)
	if err != nil {
		r.Err = err
		return
	}
	r.observedRun = &run
}

func (r *Reconciler) invokeRun() {
	// if this run is for a resource that is being reconciled
	// via the Run reconciler
	var runForWatch *unstructured.Unstructured
	runAnns := r.observedRun.GetAnnotations()
	if len(runAnns) == 0 {
		// watch & run as same
		r.isWatchAndRunSame = true
		runForWatch = r.HookRequest.Watch
	} else if runAnns[string(types.RunForWatchEnabled)] == "true" {
		runForWatch := r.HookRequest.Attachments.FindByGroupKindName(
			runAnns[string(types.RunForWatchAPIGroup)],
			runAnns[string(types.RunForWatchKind)],
			runAnns[string(types.RunForWatchName)],
		)
		if runForWatch == nil {
			r.Err = errors.Errorf(
				"Can't reconcile run: Watch not found: %s/%s %s: %s/%s %s",
				r.observedRun.GetNamespace(),
				r.observedRun.GetName(),
				r.observedRun.GroupVersionKind().String(),
				runAnns[string(types.RunForWatchAPIGroup)],
				runAnns[string(types.RunForWatchKind)],
				runAnns[string(types.RunForWatchName)],
			)
			return
		}
	}
	r.runResponse, r.Err = ExecRun(Request{
		ObservedResources: r.HookRequest.Attachments.List(),
		Run:               r.HookRequest.Watch,
		Watch:             runForWatch,
		RunCond:           r.observedRun.Spec.RunIf,
		Tasks:             r.observedRun.Spec.Tasks,
		IncludeInfo:       r.observedRun.Spec.IncludeInfo,
	})
}

func (r *Reconciler) fillSyncResponse() {
	r.HookResponse.Attachments = append(
		r.HookResponse.Attachments,
		r.runResponse.DesiredResources...,
	)
	// TODO (@amitkumardas) @ metac -then-> here:
	// add explicit deletes
	// add explicit updates
}

func (r *Reconciler) trySetWatchStatus() {
	if r.isWatchAndRunSame {
		r.HookResponse.Status = map[string]interface{}{
			"status": r.runResponse.RunStatus,
		}
	} else {
		// TODO (@amitkumardas):
		//
		// add one event against the watch resource into
		// response attachments
		//
		// event may be error or normal or warning based
		// on status
	}
}

// Sync implements the idempotent logic to sync Run resource
//
// NOTE:
// 	SyncHookRequest is the payload received as part of reconcile
// request. Similarly, SyncHookResponse is the payload sent as a
// response as part of reconcile response.
//
// NOTE:
//	This controller watches Run custom resource
func Sync(request *generic.SyncHookRequest, response *generic.SyncHookResponse) error {
	r := &Reconciler{
		Reconciler: ctrlutil.Reconciler{
			HookRequest:  request,
			HookResponse: response,
		},
	}
	// add functions to achieve desired state
	r.ReconcileFns = []func(){
		r.evalRun,
		r.invokeRun,
		r.fillSyncResponse,
	}
	// add functions to achieve desired watch
	r.DesiredWatchFns = []func(){
		r.trySetWatchStatus,
	}
	// run reconcile
	return r.Reconcile()
}
