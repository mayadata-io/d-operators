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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
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
			TaskListStatus: map[string]types.TaskStatus{},
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

func (r *Runner) isRunEnabled() (bool, error) {
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
							"recipe.dope.metacontroller.io/lock": "true",
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
		// no of tasks + elapsed time task
		ProtectedTaskCount: len(r.Recipe.Spec.Tasks) + 1,
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

func (r *Runner) addRecipeElapsedTimeInSeconds(elapsedtime float64) {
	r.RecipeStatus.TaskListStatus["recipe-elapsed-time"] = types.TaskStatus{
		Step:                 len(r.Recipe.Spec.Tasks) + 1,
		Internal:             pointer.BoolPtr(true),
		Phase:                types.TaskStatusPassed,
		ElapsedTimeInSeconds: pointer.Float64Ptr(elapsedtime),
	}
}

// runAll runs all the tasks
func (r *Runner) runAllTasks() (status *types.RecipeStatus, err error) {
	defer func() {
		r.fixture.TearDown()
	}()
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
			return nil, errors.Wrapf(
				err,
				"Failed to run task [%d] %q",
				idx+1,
				task.Name,
			)
		}
		r.RecipeStatus.TaskListStatus[task.Name] = got
		if got.Phase == types.TaskStatusFailed {
			// We run subsequent tasks even if current task
			// failed
			failedTasks++
		}
	}
	// time taken for this recipe
	elapsedSeconds := time.Since(start).Seconds()
	r.addRecipeElapsedTimeInSeconds(elapsedSeconds)
	// build the result
	if failedTasks > 0 {
		r.RecipeStatus.Phase = types.RecipeStatusFailed
		r.RecipeStatus.FailedTaskCount = failedTasks
	} else {
		r.RecipeStatus.Phase = r.mayBePassedOrCompletedStatus()
	}
	r.RecipeStatus.TaskCount = len(r.Recipe.Spec.Tasks)
	return r.RecipeStatus, nil
}

// Run executes the tasks in a sequential order
func (r *Runner) Run() (status *types.RecipeStatus, err error) {
	err = r.init()
	if err != nil {
		return nil, err
	}

	lockrunner := r.buildLockRunner()
	locked, err := lockrunner.IsLocked()
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Failed to check for existence of lock",
		)
	}
	if locked {
		klog.V(2).Infof(
			"Will skip executing recipe %s %s: Previous lock exists",
			r.Recipe.GetNamespace(),
			r.Recipe.GetName(),
		)
		// if this recipe is locked then skip its execution
		r.RecipeStatus.Phase = types.RecipeStatusLocked
		r.RecipeStatus.Reason = "Recipe was skipped: Previous lock exists"
		return r.RecipeStatus, nil
	}

	lockstatus, unlock, err := lockrunner.Lock()
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Failed to take lock",
		)
	}
	// we make use of defer to execute unlock
	defer func() {
		if err != nil {
			// force unlock in case of recipe execution error
			_, unlockerr := r.buildLockRunner().MustUnlock()
			if unlockerr != nil {
				klog.Errorf("Failed to force unlock: %s", unlockerr.Error())
			}
			return
		}
		// graceful unlock
		unlockstatus, unlockerr := unlock()
		if unlockerr != nil {
			klog.Errorf("Failed to unlock: %s", unlockerr.Error())
			return
		}
		r.RecipeStatus.TaskListStatus[r.Recipe.GetName()+"-unlock"] = unlockstatus
	}()
	r.RecipeStatus.TaskListStatus[r.Recipe.GetName()+"-lock"] = lockstatus

	err = r.evalAllTasks()
	if err != nil {
		return nil, err
	}

	isEnabled, err := r.isRunEnabled()
	if err != nil {
		return nil, err
	}
	if !isEnabled {
		klog.V(2).Infof(
			"Will skip executing recipe %s %s: It is disabled",
			r.Recipe.GetNamespace(),
			r.Recipe.GetName(),
		)
		return &types.RecipeStatus{
			Phase: types.RecipeStatusDisabled,
		}, nil
	}

	return r.runAllTasks()
}
