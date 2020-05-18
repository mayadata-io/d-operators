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
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
	types "mayadata.io/d-operators/types/job"
	metacdiscovery "openebs.io/metac/dynamic/discovery"
	metac "openebs.io/metac/start"
)

// RunnerConfig helps constructing new Runner instances
type RunnerConfig struct {
	Job types.Job
}

// Runner helps executing a Job
type Runner struct {
	Job       types.Job
	JobStatus *types.JobStatus

	fixture    *Fixture
	isTearDown bool
	hasCRDTask bool
	when       types.When
}

// NewRunner returns a new instance of Runner
func NewRunner(config RunnerConfig) *Runner {
	var isTearDown bool
	if config.Job.JobSpec.Teardown != nil {
		isTearDown = *config.Job.JobSpec.Teardown
	}
	return &Runner{
		isTearDown: isTearDown,
		Job:        config.Job,
		JobStatus: &types.JobStatus{
			TaskListStatus: map[string]types.TaskStatus{},
		},
	}
}

func (r *Runner) init() {
	if r.Job.JobSpec.Enabled == nil {
		r.when = types.Once
	} else {
		r.when = r.Job.JobSpec.Enabled.When
	}
}

func (r *Runner) isRunEnabled() bool {
	return r.when != types.Never
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

func (r *Runner) setFixture() error {
	f, err := NewFixture(FixtureConfig{
		KubeConfig:   metac.KubeDetails.Config,
		APIDiscovery: r.getAPIDiscovery(),
		IsTearDown:   r.isTearDown,
	})
	if err != nil {
		return err
	}
	r.fixture = f
	return nil
}

func (r *Runner) buildLockRunner() *LockRunner {
	var isLockForever = false
	if r.when == types.Once {
		isLockForever = true
	}
	lock := types.Task{
		Apply: &types.Apply{
			State: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":       "ConfigMap",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name":      r.Job.GetName() + "-lock",
						"namespace": r.Job.GetNamespace(),
						"labels": map[string]interface{}{
							"job.doperators.metacontroller.io/lock": "true",
						},
					},
				},
			},
		},
	}
	return &LockRunner{
		Fixture:            r.fixture,
		Task:               lock,
		LockForever:        isLockForever,
		Retry:              NewRetry(RetryConfig{}),
		ProtectedTaskCount: len(r.Job.JobSpec.Tasks),
	}
}

// evalAll evaluates all tasks
func (r *Runner) evalAll() error {
	for _, task := range r.Job.JobSpec.Tasks {
		err := r.eval(task)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) mayBePassedOrCompletedStatus() types.JobStatusPhase {
	if r.when == types.Once {
		return types.JobStatusCompleted
	}
	return types.JobStatusPassed
}

func (r *Runner) getAPIDiscovery() *metacdiscovery.APIResourceDiscovery {
	// if !r.hasCRDTask {
	klog.V(3).Infof("Using metac api discovery instance")
	return metac.KubeDetails.GetMetacAPIDiscovery()
	// }

	// TODO
	//	If we need a api discovery with more frequent refreshes
	// then we might use below. We need to stop the discovery
	// once this Job instance is done.
	//
	// klog.V(3).Infof("Using new instance of api discovery")
	// // return a discovery that refreshes more frequently
	// apiDiscovery := metac.KubeDetails.NewAPIDiscovery()
	// apiDiscovery.Start(5 * time.Second)
	// return apiDiscovery
}

// runAll runs all the tasks
func (r *Runner) runAll() (status *types.JobStatus, err error) {
	defer func() {
		r.fixture.TearDown()
	}()
	var failedTasks int
	for idx, task := range r.Job.JobSpec.Tasks {
		tr := &TaskRunner{
			Fixture:   r.fixture,
			TaskIndex: idx + 1,
			Task:      task,
			Retry:     NewRetry(RetryConfig{}),
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
	// build the result
	if failedTasks > 0 {
		r.JobStatus.Phase = types.JobStatusFailed
		r.JobStatus.FailedTaskCount = failedTasks
	} else {
		r.JobStatus.Phase = r.mayBePassedOrCompletedStatus()
	}
	r.JobStatus.TaskCount = len(r.Job.JobSpec.Tasks)
	return r.JobStatus, nil
}

// Run executes the tasks in a sequential order
func (r *Runner) Run() (status *types.JobStatus, err error) {
	r.init()

	if !r.isRunEnabled() {
		return &types.JobStatus{
			Phase: types.JobStatusDisabled,
		}, nil
	}

	err = r.setFixture()
	if err != nil {
		return nil, err
	}

	err = r.evalAll()
	if err != nil {
		return nil, err
	}

	lockstatus, unlock, err := r.buildLockRunner().Lock()
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			klog.V(3).Infof(
				"Will skip job %s %s: Previous lock exists: %+v",
				r.Job.GetNamespace(),
				r.Job.GetName(),
				err,
			)
			// if this job is locked then skip its execution
			r.JobStatus.Phase = types.JobStatusLocked
			r.JobStatus.Reason = "Job was skipped: Previous lock exists"
			return r.JobStatus, nil
		}
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

	return r.runAll()
}
