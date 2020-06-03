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
	"fmt"
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
	types "mayadata.io/d-operators/types/job"
	metac "openebs.io/metac/start"
)

// RunnerConfig helps constructing new Runner instances
type RunnerConfig struct {
	Job   types.Job
	Retry *Retryable
}

// Runner helps executing a Job
type Runner struct {
	Job       types.Job
	JobStatus *types.JobStatus
	Retry     *Retryable

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
	if config.Job.Spec.Teardown != nil {
		isTearDown = *config.Job.Spec.Teardown
	}
	// check retry
	var retry = NewRetry(RetryConfig{})
	if config.Retry != nil {
		retry = config.Retry
	}
	return &Runner{
		isTearDown: isTearDown,
		Job:        config.Job,
		JobStatus: &types.JobStatus{
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
	if r.Job.Spec.Enabled == nil {
		r.Job.Spec.Enabled = &types.Enabled{
			When: types.EnabledRuleOnce,
		}
	}
}

func (r *Runner) waitTillThinkTimeExpires() {
	if r.Job.Spec.ThinkTimeInSeconds == nil {
		return
	}
	wait := *r.Job.Spec.ThinkTimeInSeconds
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
		JobName:  fmt.Sprintf("%s %s", r.Job.GetNamespace(), r.Job.GetName()),
		Fixture:  r.fixture,
		Eligible: r.Job.Spec.Eligible,
		Retry:    r.Retry,
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
	if r.Job.Spec.Enabled.When == types.EnabledRuleOnce {
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
						"name":      r.Job.GetName() + "-lock",
						"namespace": r.Job.GetNamespace(),
						"labels": map[string]interface{}{
							"job.dope.metacontroller.io/lock": "true",
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
		ProtectedTaskCount: len(r.Job.Spec.Tasks) + 1,
	}
}

// evalAll evaluates all tasks
func (r *Runner) evalAll() error {
	for _, task := range r.Job.Spec.Tasks {
		err := r.eval(task)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) mayBePassedOrCompletedStatus() types.JobStatusPhase {
	if r.Job.Spec.Enabled.When == types.EnabledRuleOnce {
		return types.JobStatusCompleted
	}
	return types.JobStatusPassed
}

// func (r *Runner) getAPIDiscovery() *metacdiscovery.APIResourceDiscovery {
// 	// if !r.hasCRDTask {
// 	klog.V(3).Infof("Using metac api discovery instance")
// 	return metac.KubeDetails.GetMetacAPIDiscovery()
// 	// }

// 	// TODO
// 	//	If we need a api discovery with more frequent refreshes
// 	// then we might use below. We need to stop the discovery
// 	// once this Job instance is done.
// 	//
// 	// klog.V(3).Infof("Using new instance of api discovery")
// 	// // return a discovery that refreshes more frequently
// 	// apiDiscovery := metac.KubeDetails.NewAPIDiscovery()
// 	// apiDiscovery.Start(5 * time.Second)
// 	// return apiDiscovery
// }

func (r *Runner) addJobElapsedTimeInSeconds(elapsedtime float64) {
	r.JobStatus.TaskListStatus["job-elapsed-time"] = types.TaskStatus{
		Step:                 len(r.Job.Spec.Tasks) + 1,
		Internal:             pointer.BoolPtr(true),
		Phase:                types.TaskStatusPassed,
		ElapsedTimeInSeconds: pointer.Float64Ptr(elapsedtime),
	}
}

// runAll runs all the tasks
func (r *Runner) runAll() (status *types.JobStatus, err error) {
	defer func() {
		r.fixture.TearDown()
	}()
	var failedTasks int
	var start = time.Now()
	for idx, task := range r.Job.Spec.Tasks {
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
			return nil, errors.Wrapf(
				err,
				"Failed to run task [%d] %q",
				idx+1,
				task.Name,
			)
		}
		r.JobStatus.TaskListStatus[task.Name] = got
		if got.Phase == types.TaskStatusFailed {
			failedTasks++
		}
	}
	// time taken for this job
	elapsedSeconds := time.Since(start).Seconds()
	r.addJobElapsedTimeInSeconds(elapsedSeconds)
	// build the result
	if failedTasks > 0 {
		r.JobStatus.Phase = types.JobStatusFailed
		r.JobStatus.FailedTaskCount = failedTasks
	} else {
		r.JobStatus.Phase = r.mayBePassedOrCompletedStatus()
	}
	r.JobStatus.TaskCount = len(r.Job.Spec.Tasks)
	return r.JobStatus, nil
}

// Run executes the tasks in a sequential order
func (r *Runner) Run() (status *types.JobStatus, err error) {
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
			"Will skip executing job %s %s: Previous lock exists",
			r.Job.GetNamespace(),
			r.Job.GetName(),
		)
		// if this job is locked then skip its execution
		r.JobStatus.Phase = types.JobStatusLocked
		r.JobStatus.Reason = "Job was skipped: Previous lock exists"
		return r.JobStatus, nil
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
			// force unlock in case of job execution error
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
		r.JobStatus.TaskListStatus[r.Job.GetName()+"-unlock"] = unlockstatus
	}()
	r.JobStatus.TaskListStatus[r.Job.GetName()+"-lock"] = lockstatus

	err = r.evalAll()
	if err != nil {
		return nil, err
	}

	isEnabled, err := r.isRunEnabled()
	if err != nil {
		return nil, err
	}
	if !isEnabled {
		klog.V(2).Infof(
			"Will skip executing job %s %s: It is disabled",
			r.Job.GetNamespace(),
			r.Job.GetName(),
		)
		return &types.JobStatus{
			Phase: types.JobStatusDisabled,
		}, nil
	}

	return r.runAll()
}
