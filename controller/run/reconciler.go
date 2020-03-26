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
	"openebs.io/metac/controller/generic"

	ctrlutil "mayadata.io/d-operators/common/controller"
	"mayadata.io/d-operators/common/unstruct"
	types "mayadata.io/d-operators/types/run"
)

// Reconciler manages reconciliation of HTTP resources
type Reconciler struct {
	ctrlutil.Reconciler

	observedRun *types.Run
}

func (r *Reconciler) walkObservedRun() {
	var run types.Run
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
	r := &Reconciler{
		Reconciler: ctrlutil.Reconciler{
			HookRequest:  request,
			HookResponse: response,
		},
	}
	// add functions to achieve desired state
	r.ReconcileFns = []func(){
		r.walkObservedRun,
	}
	// add functions to achieve desired watch
	r.DesiredWatchFns = []func(){}
	// run reconcile
	return r.Reconcile()
}
