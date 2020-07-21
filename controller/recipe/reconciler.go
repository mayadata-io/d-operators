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

package recipe

import (
	"k8s.io/utils/pointer"
	"openebs.io/metac/controller/generic"
	k8s "openebs.io/metac/third_party/kubernetes"

	commonctrl "mayadata.io/d-operators/common/controller"
	"mayadata.io/d-operators/common/unstruct"
	"mayadata.io/d-operators/pkg/recipe"
	types "mayadata.io/d-operators/types/recipe"
)

// Reconciler manages reconciliation of Recipe custom resource
type Reconciler struct {
	commonctrl.Reconciler

	ObservedRecipe *types.Recipe
	RecipeStatus   *types.RecipeStatus
}

func (r *Reconciler) eval() {
	var j types.Recipe
	// convert from unstructured instance to typed instance
	err := unstruct.ToTyped(r.HookRequest.Watch, &j)
	if err != nil {
		r.Err = err
		return
	}
	r.ObservedRecipe = &j
}

func (r *Reconciler) invoke() {
	runner := recipe.NewRunner(
		recipe.RunnerConfig{
			Recipe: *r.ObservedRecipe,
		},
	)
	r.RecipeStatus, r.Err = runner.Run()
}

func (r *Reconciler) setSyncResponse() {
	// we skip the reconcile always since there are no attachments
	// to reconcile
	r.HookResponse.SkipReconcile = true
	r.SkipReason = "No attachments to reconcile"
	// update the skip reason for locked recipes
	if r.RecipeStatus.Phase == types.RecipeStatusLocked {
		r.SkipReason = r.RecipeStatus.Reason
	}
	// set resync period for recipes with errors
	if r.Err != nil {
		// resync since this might be a temporary error
		//
		// TODO:
		// 	Might be better to expose this from recipe.spec
		r.HookResponse.ResyncAfterSeconds = 5.0
	}
}

func (r *Reconciler) setWatchStatusAsError() {
	r.HookResponse.Status = map[string]interface{}{
		"phase":  "Error",
		"reason": r.Err.Error(),
	}
	r.HookResponse.Labels = map[string]*string{
		"recipe.dope.metacontroller.io/phase": k8s.StringPtr("Error"),
	}
}

func (r *Reconciler) setWatchStatusFromRecipeStatus() {
	r.HookResponse.Status = map[string]interface{}{
		"phase":           string(r.RecipeStatus.Phase),
		"reason":          r.RecipeStatus.Reason,
		"message":         r.RecipeStatus.Message,
		"failedTaskCount": int64(r.RecipeStatus.FailedTaskCount),
		"taskCount":       int64(r.RecipeStatus.TaskCount),
		"taskListStatus":  r.RecipeStatus.TaskListStatus,
	}
	r.HookResponse.Labels = map[string]*string{
		"recipe.dope.metacontroller.io/phase": pointer.StringPtr(string(r.RecipeStatus.Phase)),
	}
	if r.ObservedRecipe != nil &&
		r.ObservedRecipe.Spec.Refresh.ResyncAfterSeconds != nil {
		r.HookResponse.ResyncAfterSeconds = *r.ObservedRecipe.Spec.Refresh.ResyncAfterSeconds
	}
}

func (r *Reconciler) setWatchStatus() {
	if r.Err != nil {
		if r.ObservedRecipe != nil &&
			r.ObservedRecipe.Spec.Refresh.OnErrorResyncAfterSeconds != nil {
			// resync based on configuration
			r.HookResponse.ResyncAfterSeconds =
				*r.ObservedRecipe.Spec.Refresh.OnErrorResyncAfterSeconds
		}
		r.setWatchStatusAsError()
		return
	}
	if r.RecipeStatus.Phase == types.RecipeStatusLocked {
		// nothing needs to be done
		// old status will persist
		return
	}
	r.setWatchStatusFromRecipeStatus()
}

// Sync implements the idempotent logic to sync Recipe resource
//
// NOTE:
// 	SyncHookRequest is the payload received as part of reconcile
// request. Similarly, SyncHookResponse is the payload sent as a
// response as part of reconcile response.
//
// NOTE:
//	This controller watches Recipe custom resource
func Sync(request *generic.SyncHookRequest, response *generic.SyncHookResponse) error {
	r := &Reconciler{
		Reconciler: commonctrl.Reconciler{
			Name:         "recipe-sync-reconciler",
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
