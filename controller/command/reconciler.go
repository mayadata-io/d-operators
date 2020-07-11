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

package command

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"openebs.io/metac/controller/generic"

	commonctrl "mayadata.io/d-operators/common/controller"
	"mayadata.io/d-operators/common/unstruct"
	"mayadata.io/d-operators/pkg/command"
	types "mayadata.io/d-operators/types/command"
)

// Reconciler manages reconciliation of Command custom resource
type Reconciler struct {
	commonctrl.Reconciler

	ObservedCommand *types.Command
	Attachment      *unstructured.Unstructured
}

func (r *Reconciler) eval() {
	var c types.Command
	// convert from unstructured instance to typed instance
	err := unstruct.ToTyped(r.HookRequest.Watch, &c)
	if err != nil {
		r.Err = err
		return
	}
	r.ObservedCommand = &c
}

func (r *Reconciler) invoke() {
	if r.ObservedCommand.Status.Phase != "" && !r.HookRequest.Attachments.IsEmpty() {
		return
	}
	builder := command.NewJobBuilder(
		command.JobBuilderConfig{
			Command: *r.ObservedCommand,
		},
	)
	r.Attachment, r.Err = builder.Build()
}

func (r *Reconciler) setSyncResponse() {
	if len(r.HookResponse.Attachments) == 0 {
		r.HookResponse.SkipReconcile = true
	}

	if r.ObservedCommand != nil &&
		r.ObservedCommand.Spec.Refresh.ResyncAfterSeconds != nil {
		r.HookResponse.ResyncAfterSeconds = *r.ObservedCommand.Spec.Refresh.ResyncAfterSeconds
	}
}

// Sync implements the idempotent logic to sync Command resource
//
// NOTE:
// 	SyncHookRequest is the payload received as part of reconcile
// request. Similarly, SyncHookResponse is the payload sent as a
// response as part of reconcile response.
//
// NOTE:
//	This controller watches Command custom resource
func Sync(request *generic.SyncHookRequest, response *generic.SyncHookResponse) error {
	r := &Reconciler{
		Reconciler: commonctrl.Reconciler{
			Name:         "command-sync-reconciler",
			HookRequest:  request,
			HookResponse: response,
		},
	}
	// add functions to achieve desired state
	r.ReconcileFns = []func(){
		r.eval,
		r.invoke,
		r.setSyncResponse,
	}
	// run reconcile
	return r.Reconcile()
}
