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

package job

import (
	"openebs.io/metac/controller/generic"

	commonctrl "mayadata.io/d-operators/common/controller"
	"mayadata.io/d-operators/common/unstruct"
	"mayadata.io/d-operators/pkg/job"
	types "mayadata.io/d-operators/types/job"
)

// Reconciler manages reconciliation of Job custom resource
type Reconciler struct {
	commonctrl.Reconciler

	ObservedJob *types.Job
	JobStatus   *types.JobStatus
}

func (r *Reconciler) eval() {
	var j types.Job
	// convert from unstructured instance to typed instance
	err := unstruct.ToTyped(r.HookRequest.Watch, &j)
	if err != nil {
		r.Err = err
		return
	}
	r.ObservedJob = &j
}

func (r *Reconciler) invoke() {
	runner := job.NewRunner(
		job.RunnerConfig{
			Job: *r.ObservedJob,
		},
	)
	r.JobStatus, r.Err = runner.Run()
}

func (r *Reconciler) setSyncResponse() {
	// we skip the reconcile always since there are no attachments
	// to reconcile
	r.HookResponse.SkipReconcile = true
	r.SkipReason = "No attachments to reconcile"
	// update the skip reason for locked jobs
	if r.JobStatus.Phase == types.JobStatusLocked {
		r.SkipReason = r.JobStatus.Reason
	}
	// set resync period for jobs with errors
	if r.Err != nil {
		// resync since this might be a temporary error
		//
		// TODO:
		// 	Might be better to expose this from job.spec
		r.HookResponse.ResyncAfterSeconds = 5.0
	}
}

func (r *Reconciler) setWatchStatusAsError() {
	r.HookResponse.Status = map[string]interface{}{
		"phase":  "Error",
		"reason": r.Err.Error(),
	}
}

func (r *Reconciler) setWatchStatusFromJobStatus() {
	r.HookResponse.Status = map[string]interface{}{
		"phase":           string(r.JobStatus.Phase),
		"reason":          r.JobStatus.Reason,
		"message":         r.JobStatus.Message,
		"failedTaskCount": int64(r.JobStatus.FailedTaskCount),
		"taskCount":       int64(r.JobStatus.TaskCount),
		"taskListStatus":  r.JobStatus.TaskListStatus,
	}
}

func (r *Reconciler) setWatchStatus() {
	if r.Err != nil {
		// resync since this might be a temporary error
		//
		// TODO:
		// 	Might be better to expose this from job.spec
		r.HookResponse.ResyncAfterSeconds = 5.0
		r.setWatchStatusAsError()
		return
	}
	if r.JobStatus.Phase == types.JobStatusLocked {
		// nothing needs to be done
		// old status will persist
		return
	}
	r.setWatchStatusFromJobStatus()
}

// Sync implements the idempotent logic to sync Job resource
//
// NOTE:
// 	SyncHookRequest is the payload received as part of reconcile
// request. Similarly, SyncHookResponse is the payload sent as a
// response as part of reconcile response.
//
// NOTE:
//	This controller watches Job custom resource
func Sync(request *generic.SyncHookRequest, response *generic.SyncHookResponse) error {
	r := &Reconciler{
		Reconciler: commonctrl.Reconciler{
			Name:         "job-sync-reconciler",
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
	// add functions to achieve desired watch
	r.DesiredWatchFns = []func(){
		r.setWatchStatus,
	}
	// run reconcile
	return r.Reconcile()
}
