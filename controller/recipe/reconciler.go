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
	"github.com/pkg/errors"
	"openebs.io/metac/controller/generic"
	k8s "openebs.io/metac/third_party/kubernetes"

	ctrl "mayadata.io/d-operators/common/controller"
	"mayadata.io/d-operators/common/unstruct"
	"mayadata.io/d-operators/pkg/recipe"
	"mayadata.io/d-operators/pkg/schema"
	types "mayadata.io/d-operators/types/recipe"
)

// Reconciler manages reconciliation of Recipe custom resource
type Reconciler struct {
	ctrl.Reconciler

	ObservedRecipe            *types.Recipe
	FieldPathValidationResult *schema.FieldPathValidationResult

	// resulting status after executing the observed Recipe
	recipeRunStatus types.RecipeStatus
}

func (r *Reconciler) eval() {
	// validate the received structured instance
	v := &schema.FieldPathValidation{
		Target:                  r.HookRequest.Watch.Object,
		SupportedAbsolutePaths:  types.SupportedAbsolutePaths,
		UserAllowedPathPrefixes: types.UserAllowedPathPrefixes,
	}
	valResult := v.Validate()

	var j types.Recipe
	// convert from unstructured instance to typed instance
	err := unstruct.ToTyped(r.HookRequest.Watch, &j)
	if err != nil {
		r.Err = errors.Wrapf(
			err,
			"Failed to convert from unstructured type to recipe type",
		)
		return
	}
	r.ObservedRecipe = &j
	r.FieldPathValidationResult = valResult
}

func (r *Reconciler) invoke() {
	runner := recipe.NewRunner(
		recipe.RunnerConfig{
			Recipe:                    *r.ObservedRecipe,
			FieldPathValidationResult: *r.FieldPathValidationResult,
		},
	)
	r.recipeRunStatus, r.Err = runner.Run()
}

func (r *Reconciler) setSyncResponse() {
	// we skip the reconcile always since there are no attachments
	// to reconcile
	r.HookResponse.SkipReconcile = true
	// default skip reason
	r.SkipReason = "No attachments to reconcile"
}

func (r *Reconciler) setRecipeStatusFromResult() {
	if r.ObservedRecipe == nil {
		// nothing needs to be done
		return
	}
	if r.recipeRunStatus.Phase == types.RecipeStatusNotEligible &&
		r.ObservedRecipe.Spec.Resync.OnNotEligibleResyncInSeconds != nil {
		// set configured resync interval when Recipe's phase
		// is set to NotEligible
		r.HookResponse.ResyncAfterSeconds =
			float64(*r.ObservedRecipe.Spec.Resync.OnNotEligibleResyncInSeconds)

		// Further logic should not be executed
		//
		// NOTE:
		//	Setting resync interval when phase is set to NotEligible
		// holds more priority
		return
	}
	if r.ObservedRecipe.Spec.Resync.IntervalInSeconds != nil {
		// set configured resync interval during normal conditions
		//
		// NOTE:
		// 	This is set when Recipe's phase is NOT IN following phases:
		// - NotEligible
		// - Error
		r.HookResponse.ResyncAfterSeconds =
			float64(*r.ObservedRecipe.Spec.Resync.IntervalInSeconds)
	}
}

func (r *Reconciler) setRecipeStatusFromError() {
	if r.ObservedRecipe != nil &&
		r.ObservedRecipe.Spec.Resync.OnErrorResyncInSeconds != nil {
		// set configured resync interval when Recipe's phase
		// is set to Error
		r.HookResponse.ResyncAfterSeconds =
			float64(*r.ObservedRecipe.Spec.Resync.OnErrorResyncInSeconds)
	}
	r.HookResponse.Status = map[string]interface{}{
		"phase":  "Error",
		"reason": r.Err.Error(),
	}
	r.HookResponse.Labels = map[string]*string{
		types.LblKeyRecipePhase: k8s.StringPtr("Error"),
	}
}

func (r *Reconciler) setRecipeStatus() {
	if r.Err != nil {
		// reconciler is only concerned about the Recipes that
		// result in error
		r.setRecipeStatusFromError()
		return
	}
	// normal conditions i.e. when non error conditions
	r.setRecipeStatusFromResult()
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
		Reconciler: ctrl.Reconciler{
			Name:         "sync-recipe",
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
		r.setRecipeStatus,
	}
	// run reconcile
	return r.Reconcile()
}
