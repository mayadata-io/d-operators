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
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"openebs.io/metac/controller/generic"

	ctrl "mayadata.io/d-operators/common/controller"
	jsonutil "mayadata.io/d-operators/common/json"
	"mayadata.io/d-operators/common/pointer"
	"mayadata.io/d-operators/common/unstruct"
	"mayadata.io/d-operators/pkg/command"
	types "mayadata.io/d-operators/types/command"
)

// Reconciler manages reconciliation of Command custom resource
type Reconciler struct {
	// Common reconciler fields & functions are exposed
	// from this embedded structure
	ctrl.Reconciler

	observedCommand types.Command
	status          types.CommandStatus
}

func (r *Reconciler) eval() {
	var c types.Command
	// convert unstructured instance to typed instance
	err := unstruct.ToTyped(r.HookRequest.Watch, &c)
	if err != nil {
		r.Err = errors.Wrapf(
			err,
			"Failed to convert unstruct command to typed instance: Name %q / %q",
			r.HookRequest.Watch.GetNamespace(),
			r.HookRequest.Watch.GetName(),
		)
		klog.Errorf(
			"Resource %s \nError %s",
			jsonutil.New(r.HookRequest.Watch).MustMarshal(),
			err.Error(),
		)
		return
	}
	r.observedCommand = c
}

func (r *Reconciler) sync() {
	// creating / deleting a Kubernetes Job is part of Command reconciliation
	jobBuilder := command.NewJobBuilder(
		command.JobBuildingConfig{
			Command: r.observedCommand,
		},
	)
	// This gets the desired job specifications
	job, err := jobBuilder.Build()
	if err != nil {
		r.Err = err
		return
	}
	cmdReconciler, err := command.NewReconciler(command.ReconciliationConfig{
		Command: r.observedCommand,
		Child:   job,
	})
	if err != nil {
		r.Err = err
		return
	}
	r.status, r.Err = cmdReconciler.Reconcile()
}

func (r *Reconciler) setResyncInterval() {
	// reset metac's resync interval time if set
	if r.observedCommand.Spec.Resync.IntervalInSeconds != nil {
		r.HookResponse.ResyncAfterSeconds =
			float64(*r.observedCommand.Spec.Resync.IntervalInSeconds)
	}
	// error based resync interval overrides normal resync interval
	if r.Err != nil &&
		r.observedCommand.Spec.Resync.OnErrorResyncInSeconds != nil {
		r.HookResponse.ResyncAfterSeconds =
			float64(*r.observedCommand.Spec.Resync.OnErrorResyncInSeconds)
	}
}

func (r *Reconciler) setWatchAttributes() {
	if r.status.Phase == types.CommandPhaseSkipped {
		r.HookResponse.Status = map[string]interface{}{
			"phase":  r.status.Phase,
			"reason": r.status.Reason,
		}
		r.HookResponse.Labels = map[string]*string{
			types.LblKeyCommandPhase: pointer.String(string(types.CommandPhaseSkipped)),
		}
	}
	if r.Err != nil {
		r.HookResponse.Status = map[string]interface{}{
			"phase":  types.CommandPhaseError,
			"reason": r.Err.Error(),
		}
		r.HookResponse.Labels = map[string]*string{
			types.LblKeyCommandPhase: pointer.String(string(types.CommandPhaseError)),
		}
	}
}

func (r *Reconciler) setResponse() {
	// Reconciling atachments are skipped since attachments
	// are not reconciled as part of reconciling Command resource
	r.HookResponse.SkipReconcile = true
	r.setResyncInterval()
	r.setWatchAttributes()
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
		Reconciler: ctrl.Reconciler{
			Name:         "command-sync",
			HookRequest:  request,
			HookResponse: response,
		},
	}
	// add functions to achieve desired state
	r.ReconcileFns = []func(){
		r.eval,
		r.sync,
	}

	r.DesiredWatchFns = []func(){
		r.setResponse,
	}

	// run reconcile
	return r.Reconcile()
}
