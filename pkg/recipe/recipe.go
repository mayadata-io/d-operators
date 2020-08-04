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
	"fmt"
	"time"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
	"mayadata.io/d-operators/common/unstruct"
	types "mayadata.io/d-operators/types/recipe"
	metac "openebs.io/metac/start"
)

// RunnerConfig helps constructing new Runner instances
type RunnerConfig struct {
	Recipe types.Recipe
	Retry  *Retryable
}

// Runner helps executing a Recipe
type Runner struct {
	Recipe       types.Recipe
	RecipeStatus *types.RecipeStatus
	Retry        *Retryable

	fixture    *Fixture
	isTearDown bool
	hasCRDTask bool

	// err as value
	err error

	// dependency injection functions useful during unit testing
	updateRecipeWithRetriesFn func() error
}

// NewRunner returns a new instance of Runner
func NewRunner(config RunnerConfig) *Runner {
	// check teardown
	var isTearDown bool
	if config.Recipe.Spec.Teardown != nil {
		isTearDown = *config.Recipe.Spec.Teardown
	}
	// check retry
	var retry = NewRetry(RetryConfig{})
	if config.Retry != nil {
		retry = config.Retry
	}
	return &Runner{
		isTearDown: isTearDown,
		Recipe:     config.Recipe,
		RecipeStatus: &types.RecipeStatus{
			TaskResultList: map[string]types.TaskResult{},
		},
		Retry: retry,
	}
}

func (r *Runner) initFixture() {
	f, err := NewFixture(FixtureConfig{
		KubeConfig:   metac.KubeDetails.Config,
		APIDiscovery: metac.KubeDetails.GetMetacAPIDiscovery(),
		IsTearDown:   r.isTearDown,
	})
	if err != nil {
		r.err = err
	}
	r.fixture = f
}

func (r *Runner) initEnabled() {
	if r.Recipe.Spec.Enabled == nil {
		r.Recipe.Spec.Enabled = &types.Enabled{
			When: types.EnabledRuleOnce,
		}
	}
}

func (r *Runner) waitTillThinkTimeExpires() {
	if r.Recipe.Spec.ThinkTimeInSeconds == nil {
		return
	}
	wait := *r.Recipe.Spec.ThinkTimeInSeconds
	if wait < 0 {
		wait = 0
	}
	time.Sleep(time.Duration(wait) * time.Second)
}

func (r *Runner) init() error {
	var fns = []func(){
		r.initEnabled,
		r.initFixture,
	}
	for _, fn := range fns {
		fn()
		if r.err != nil {
			return r.err
		}
	}
	return nil
}

func (r *Runner) isRunEligible() (bool, error) {
	// we shall wait for think time to expire if it was set
	r.waitTillThinkTimeExpires()

	e, err := NewEligibility(EligibilityConfig{
		RecipeName: fmt.Sprintf("%s %s", r.Recipe.GetNamespace(), r.Recipe.GetName()),
		Fixture:    r.fixture,
		Eligible:   r.Recipe.Spec.Eligible,
		Retry:      r.Retry,
	})
	if err != nil {
		return false, err
	}
	return e.IsEligible()
}

func (r *Runner) eval(task types.Task) error {
	if task.Name == "" {
		return errors.Errorf(
			"Invalid task: Missing name",
		)
	}
	var action int
	var state *unstructured.Unstructured
	if task.Assert != nil {
		action++
		state = task.Assert.State
	}
	if task.Apply != nil {
		action++
		state = task.Apply.State
	}
	if task.Delete != nil {
		action++
		state = task.Delete.State
	}
	if task.Create != nil {
		action++
		state = task.Create.State
	}
	if action == 0 {
		return errors.Errorf(
			"Invalid task %q: Task needs one action",
			task.Name,
		)
	}
	if action > 1 {
		return errors.Errorf(
			"Invalid task %q: Task supports only one action",
			task.Name,
		)
	}
	if state.GetKind() == "CustomResourceDefinition" {
		r.hasCRDTask = true
	}
	return nil
}

func (r *Runner) buildLockRunner() *LockRunner {
	var isLockForever = false
	if r.Recipe.Spec.Enabled.When == types.EnabledRuleOnce {
		isLockForever = true
	}
	lock := types.Task{
		FailFast: &types.FailFast{
			When: types.FailFastOnDiscoveryError,
		},
		Apply: &types.Apply{
			State: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":       "ConfigMap",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name":      r.Recipe.GetName() + "-lock",
						"namespace": r.Recipe.GetNamespace(),
						"labels": map[string]interface{}{
							types.LblKeyIsRecipeLock: "true",
							types.LblKeyRecipeName:   r.Recipe.GetName(),
						},
					},
				},
			},
		},
	}
	return &LockRunner{
		BaseRunner: BaseRunner{
			Fixture:      r.fixture,
			Retry:        r.Retry,
			FailFastRule: lock.FailFast.When,
		},
		Task:        lock,
		LockForever: isLockForever,

		// no of tasks that are considered (read protected)
		// by this lock
		ProtectedTaskCount: len(r.Recipe.Spec.Tasks),
	}
}

// evalAll evaluates all tasks
func (r *Runner) evalAllTasks() error {
	for _, task := range r.Recipe.Spec.Tasks {
		err := r.eval(task)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) mayBePassedOrCompletedStatus() types.RecipeStatusPhase {
	if r.Recipe.Spec.Enabled.When == types.EnabledRuleOnce {
		return types.RecipeStatusCompleted
	}
	return types.RecipeStatusPassed
}

// func (r *Runner) getAPIDiscovery() *metacdiscovery.APIResourceDiscovery {
// 	// if !r.hasCRDTask {
// 	klog.V(3).Infof("Using metac api discovery instance")
// 	return metac.KubeDetails.GetMetacAPIDiscovery()
// 	// }

// 	// TODO
// 	//	If we need a api discovery with more frequent refreshes
// 	// then we might use below. We need to stop the discovery
// 	// once this Recipe instance is done.
// 	//
// 	// klog.V(3).Infof("Using new instance of api discovery")
// 	// // return a discovery that refreshes more frequently
// 	// apiDiscovery := metac.KubeDetails.NewAPIDiscovery()
// 	// apiDiscovery.Start(5 * time.Second)
// 	// return apiDiscovery
// }

// updateRecipeWithRetries updates the kubernetes cluster with
// desired recipe
func (r *Runner) updateRecipeWithRetries() error {
	if r.updateRecipeWithRetriesFn != nil {
		return r.updateRecipeWithRetriesFn()
	}
	// get the dynamic client for Recipe
	client, err := r.fixture.GetClientForAPIVersionAndKind(
		r.Recipe.APIVersion,
		r.Recipe.Kind,
	)
	if err != nil {
		return errors.Wrapf(
			err,
			"Get client failed: Recipe %q %q",
			r.Recipe.Namespace,
			r.Recipe.Name,
		)
	}

	status := map[string]interface{}{
		"phase":     string(r.RecipeStatus.Phase),
		"reason":    r.RecipeStatus.Reason,
		"message":   r.RecipeStatus.Message,
		"taskCount": r.RecipeStatus.TaskCount,
		// "failedTaskCount":        int64(r.RecipeStatus.TaskCount.Failed),
		// "totalTaskCount":         int64(r.RecipeStatus.TaskCount.Total),
		"executionTimeInSeconds": r.RecipeStatus.ExecutionTimeInSeconds,
		"taskListStatus":         r.RecipeStatus.TaskResultList,
	}
	var statusNew interface{}
	err = unstruct.MarshalThenUnmarshal(status, &statusNew)
	if err != nil {
		return errors.Wrapf(
			err,
			"Marshal unmarshal failed: Recipe %q %q",
			r.Recipe.Namespace,
			r.Recipe.Name,
		)
	}

	labels := map[string]string{
		types.LblKeyRecipePhase: string(r.RecipeStatus.Phase),
	}

	// runtimeErr is not retried
	var runtimeErr error

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Recipe before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		result, err := client.
			Namespace(r.Recipe.Namespace).
			Get(r.Recipe.Name, metav1.GetOptions{})
		if err != nil {
			runtimeErr = errors.Wrapf(
				err,
				"Get instance failed: Recipe %q %q",
				r.Recipe.Namespace,
				r.Recipe.Name,
			)
			// return nil to avoid retry
			return nil
		}

		// update recipe's status
		err = unstructured.SetNestedField(
			result.Object,
			statusNew,
			"status",
		)
		if err != nil {
			runtimeErr = errors.Wrapf(
				err,
				"Set unstruct failed: Recipe %q %q",
				r.Recipe.Namespace,
				r.Recipe.Name,
			)
			// return nil to avoid retry
			return nil
		}

		// merge recipe labels with new pairs
		unstruct.SetLabels(result, labels)

		updated, err := client.
			Namespace(r.Recipe.Namespace).
			Update(result, metav1.UpdateOptions{})

		if err != nil {
			// since this is an update error, this logic will be retried
			return errors.Wrapf(
				err,
				"Update failed: Recipe %q %q",
				r.Recipe.Namespace,
				r.Recipe.Name,
			)
		}

		if !client.HasSubresource("status") {
			return nil
		}

		// update the status as a sub resource by extracting the
		// latest resource version
		result.SetResourceVersion(updated.GetResourceVersion())
		_, err = client.
			Namespace(r.Recipe.Namespace).
			UpdateStatus(result, metav1.UpdateOptions{})

		return errors.Wrapf(
			err,
			"Update failed: Recipe %q %q",
			r.Recipe.Namespace,
			r.Recipe.Name,
		)
	})
	if runtimeErr != nil {
		return errors.Wrapf(
			runtimeErr,
			"Update failed: Runtime error: Recipe: %q %q",
			r.Recipe.Namespace,
			r.Recipe.Name,
		)
	}

	if retryErr == nil {
		klog.V(3).Infof(
			"Update succeeded: Recipe %q %q",
			r.Recipe.Namespace,
			r.Recipe.Name,
		)
	}

	return retryErr
}

// runAllTasks runs all the tasks
func (r *Runner) runAllTasks() (err error) {
	defer func() {
		r.fixture.TearDown()
		if err == nil {
			// update recipe's status if there was no error
			err = r.updateRecipeWithRetries()
		}
	}()

	// initialise the fields that can be set irrespective of Recipe
	// being run or not
	r.RecipeStatus.TaskCount.Total = len(r.Recipe.Spec.Tasks)

	// first thing to do even before running the Recipe is to
	// verify if this Recipe is eligible to run
	eligible, err := r.isRunEligible()
	if err != nil {
		return errors.Wrapf(
			err,
			"Eligibility check failed: Recipe %q %q: Status %q %q",
			r.Recipe.Namespace,
			r.Recipe.Name,
			r.Recipe.Status.Phase,
			r.Recipe.Status.Reason,
		)
	}
	if !eligible {
		klog.V(2).Infof(
			"Will skip execution: Not eligibile: Recipe %q %q",
			r.Recipe.GetNamespace(),
			r.Recipe.GetName(),
		)
		r.RecipeStatus.Phase = types.RecipeStatusNotEligible
		// short description
		r.RecipeStatus.Reason = "Did not meet eligibility criteria"
		// long description along with remedy
		r.RecipeStatus.Message =
			"Remedy: Wait for next reconciliation(s) or modify spec.eligible details"
		// all tasks are skipped since Recipe is not eligible for execution
		r.RecipeStatus.TaskCount.Skipped = len(r.Recipe.Spec.Tasks)
		return nil
	}

	var failedTasks int
	var start = time.Now()
	for idx, task := range r.Recipe.Spec.Tasks {
		var failFastRule types.FailFastRule
		if task.FailFast != nil {
			failFastRule = task.FailFast.When
		}
		tr := &TaskRunner{
			BaseRunner: BaseRunner{
				Fixture:      r.fixture,
				TaskIndex:    idx + 1,
				TaskName:     task.Name,
				Retry:        r.Retry,
				FailFastRule: failFastRule,
			},
			Task: task,
		}
		got, err := tr.Run()
		if err != nil {
			// We discontinue executing next tasks
			// if current task execution resulted in
			// error
			return errors.Wrapf(
				err,
				"Task failed: Index %d: Name %q: Recipe %q %q",
				idx+1,
				task.Name,
				r.Recipe.Namespace,
				r.Recipe.Name,
			)
		}
		r.RecipeStatus.TaskResultList[task.Name] = got
		if got.Phase == types.TaskStatusFailed {
			// We run subsequent tasks even if current task failed
			failedTasks++
		}
	}

	// time taken for this recipe to run all its tasks
	elapsedSeconds := time.Since(start).Seconds()
	r.RecipeStatus.ExecutionTimeInSeconds = pointer.Float64Ptr(elapsedSeconds)

	// set other fields of the status
	if failedTasks > 0 {
		// recipe is set to failed if any of its tasks resulted in failure
		r.RecipeStatus.Phase = types.RecipeStatusFailed
		r.RecipeStatus.TaskCount.Failed = failedTasks
	} else {
		r.RecipeStatus.Phase = r.mayBePassedOrCompletedStatus()
	}

	return nil
}

// Run executes the tasks in a sequential order
func (r *Runner) Run() (err error) {
	err = r.init()
	if err != nil {
		return err
	}
	// proceed further by verifying the presence of LOCK
	//
	// NOTE:
	//	Presence of LOCK indicates either of the following:
	// - Recipe is under execution by some other controller goroutine, OR
	// - Recipe was executed in its earlier reconcile attempt & is meant
	// to be run only once
	lockrunner := r.buildLockRunner()
	locked, err := lockrunner.IsLocked()
	if err != nil {
		return errors.Wrapf(
			err,
			"Verify lock failed: Recipe %q %q: Status %q %q",
			r.Recipe.GetNamespace(),
			r.Recipe.GetName(),
			r.Recipe.Status.Phase,
			r.Recipe.Status.Reason,
		)
	}
	if locked {
		klog.V(3).Infof(
			"Will skip execution: Previous lock exists: Recipe %q %q: Status %q %q",
			r.Recipe.GetNamespace(),
			r.Recipe.GetName(),
			r.Recipe.Status.Phase,
			r.Recipe.Status.Reason,
		)
		return nil
	}

	klog.V(2).Infof(
		"Will execute: Recipe %q %q: Status %q %q",
		r.Recipe.Namespace,
		r.Recipe.Name,
		r.Recipe.Status.Phase,
		r.Recipe.Status.Reason,
	)

	// Start executing by taking a LOCK
	_, unlock, err := lockrunner.Lock()
	if err != nil {
		return errors.Wrapf(
			err,
			"Create lock failed: Recipe %q %q: Status %q %q",
			r.Recipe.Namespace,
			r.Recipe.Name,
			r.Recipe.Status.Phase,
			r.Recipe.Status.Reason,
		)
	}
	// make use of defer to UNLOCK
	defer func() {
		if err != nil ||
			r.RecipeStatus.Phase == types.RecipeStatusNotEligible {
			// FORCE UNLOCK in case of following:
			// - Executing recipe resulted in error _OR_
			// - Recipe is not eligible to be executed _(in this attempt)_
			//
			// NOTE:
			//	Lock is removed to enable subsequent reconcile attempts
			_, unlockerr := lockrunner.MustUnlock()
			if unlockerr != nil {
				// swallow unlock error by logging
				klog.Errorf(
					"Forced unlock failed: Recipe %q %q: Status %q %q: %s",
					r.Recipe.Namespace,
					r.Recipe.Name,
					r.Recipe.Status.Phase,
					r.Recipe.Status.Reason,
					unlockerr.Error(),
				)
				// bubble up the original error
				return
			}
			klog.V(3).Infof(
				"Forced unlock was successful: Recipe %q %q: Status %q %q",
				r.Recipe.Namespace,
				r.Recipe.Name,
				r.Recipe.Status.Phase,
				r.Recipe.Status.Reason,
			)
			// bubble up the original error
			return
		}
		// GRACEFUL UNLOCK
		//
		// NOTE:
		//	Unlocking lets Recipe(s) to be executed in their next
		// reconcile attempts if these Recipe(s) are meant to be
		// run ALWAYS
		//
		// NOTE:
		// 	Recipes that are set to be run ALWAYS follow below steps:
		// 1/ Lock,
		// 2/ Execute, &
		// 3/ Unlock
		unlockstatus, unlockerr := unlock()
		if unlockerr != nil {
			// swallow the unlock error by logging
			klog.Errorf(
				"Graceful unlock failed: Recipe %q %q: Status %q %q: %s",
				r.Recipe.Namespace,
				r.Recipe.Name,
				r.Recipe.Status.Phase,
				r.Recipe.Status.Reason,
				unlockerr.Error(),
			)
			return
		}
		klog.V(3).Infof(
			"Unlocked gracefully: Recipe %q %q: Status %q %q: %s",
			r.Recipe.Namespace,
			r.Recipe.Name,
			r.Recipe.Status.Phase,
			r.Recipe.Status.Reason,
			unlockstatus,
		)
	}()

	err = r.evalAllTasks()
	if err != nil {
		return err
	}

	return r.runAllTasks()
}
