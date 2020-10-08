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
	"fmt"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
	metac "openebs.io/metac/start"

	"mayadata.io/d-operators/common/pointer"
	"mayadata.io/d-operators/pkg/kubernetes"
	"mayadata.io/d-operators/pkg/lock"
	types "mayadata.io/d-operators/types/command"
	"openebs.io/metac/dynamic/clientset"
)

// ReconciliationConfig helps constructing a new
// instance of Reconciliation
type ReconciliationConfig struct {
	Command types.Command
	Child   *unstructured.Unstructured
}

// Reconciliation helps reconciling Command resource
type Reconciliation struct {
	*kubernetes.Utility
	locker  *lock.Locking
	command types.Command

	// childJob represents either the k8s job or daemonjob
	//
	// NOTE:
	//	Parent child concept refers to Command custom resource
	// as the parent and k8s Job as the child related to this
	// Command resource
	childJob *unstructured.Unstructured

	// client to invoke CRUD operations against k8s Job
	jobClient *clientset.ResourceClient

	// is Command resource supposed to run Once
	isRunOnce bool

	// is Command resource supposed to run Always
	isRunAlways bool

	// is Command resource never supposed to run
	isRunNever bool

	// if child i.e. k8s Job needs to be applied
	isApplyAction bool

	// if child i.e. k8s Job needs to be deleted
	isDeleteAction bool

	// if child i.e. k8s Job is found in the cluster
	isChildJobFound bool

	// if k8s Job has run to completion
	isChildJobCompleted bool

	// if command resource should be retried when it
	// results into error
	isRetryOnError bool

	// if command resource should be retried when it
	// results in timeout
	isRetryOnTimeout bool

	// various status flags of the Command resource
	isStatusSet            bool
	isStatusSetAsError     bool
	isStatusSetAsCompleted bool

	commandStatus *types.CommandStatus

	// error as value
	err error
}

func (r *Reconciliation) initChildJobDetails() {
	if r.childJob == nil || r.childJob.Object == nil {
		return
	}

	if r.childJob.GetKind() != types.KindJob ||
		r.childJob.GetAPIVersion() != types.JobAPIVersion {
		r.err = errors.Errorf(
			"Invalid child: Expected %s/%s: Got %s/%s",
			types.JobAPIVersion,
			types.KindJob,
			r.childJob.GetAPIVersion(),
			r.childJob.GetKind(),
		)
		return
	}
	// At this point Job is present in Kubernetes cluster
	r.isChildJobFound = true
	// Extract status.phase of this Job
	phase, found, err := unstructured.NestedString(
		r.childJob.Object,
		"status",
		"phase",
	)
	if err != nil {
		r.err = err
		return
	}
	if !found {
		// Job's status.phase is not set
		return
	}
	r.isChildJobCompleted = phase == types.JobPhaseCompleted
}

func (r *Reconciliation) initCommandDetails() {
	r.isStatusSet = r.command.Status.Phase != ""
	r.isStatusSetAsCompleted = r.command.Status.Phase == types.CommandPhaseCompleted
	r.isStatusSetAsError = r.command.Status.Phase == types.CommandPhaseError

	// enabled defaults to Once
	// In other words Command can reconcile only once by default
	var enabled = types.EnabledOnce
	// override with user specified value if set
	if r.command.Spec.Enabled.When != "" {
		enabled = r.command.Spec.Enabled.When
	}
	r.isRunOnce = enabled == types.EnabledOnce
	r.isRunAlways = enabled == types.EnabledAlways
	r.isRunNever = enabled == types.EnabledNever

	if r.command.Spec.Retry.When != nil && *r.command.Spec.Retry.When != "" {
		r.isRetryOnError = *r.command.Spec.Retry.When == types.RetryOnError
		r.isRetryOnTimeout = *r.command.Spec.Retry.When == types.RetryOnTimeout
	}
}

func (r *Reconciliation) initLocking() {
	if metac.KubeDetails == nil {
		r.err = errors.Errorf(
			"Failed to init lock: Nil metac kube details: Command %s %s",
			r.command.GetNamespace(),
			r.command.GetName(),
		)
		return
	}
	r.locker, r.err = lock.NewLocker(lock.LockingConfig{
		// D-Operators uses metac as a library
		// Metac on its part populates the kube config & api discovery
		KubeConfig:   metac.KubeDetails.Config,
		APIDiscovery: metac.KubeDetails.GetMetacAPIDiscovery(),
		// LockingObj is the config map that will be used as a
		// lock to stop multiple goroutines trying to run the
		// Command resource simultaneously
		LockingObj: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       "ConfigMap",
				"apiVersion": "v1",
				"metadata": map[string]interface{}{
					"name":      r.command.GetName() + "-lock",
					"namespace": r.command.GetNamespace(),
					"labels": map[string]interface{}{
						types.LblKeyIsCommandLock: "true",
						types.LblKeyCommandName:   r.command.GetName(),
						types.LblKeyCommandUID:    string(r.command.GetUID()),
					},
				},
			},
		},
		// This lock will not be removed if this Command is meant
		// to be run only once
		IsLockForever: r.isRunOnce,
	})
}

func (r *Reconciliation) initKubernetes() {
	r.Utility, r.err = kubernetes.NewUtility(kubernetes.UtilityConfig{
		KubeConfig:   metac.KubeDetails.Config,
		APIDiscovery: metac.KubeDetails.GetMetacAPIDiscovery(),
	})
}

func (r *Reconciliation) initJobClient() {
	r.jobClient, r.err = r.GetClientForAPIVersionAndKind(
		types.JobAPIVersion,
		types.KindJob,
	)
}

func (r *Reconciliation) init() error {
	// ---------------------------------
	// this ORDER of functions should be maintained
	// ---------------------------------
	var fns = []func(){
		r.initChildJobDetails,
		r.initCommandDetails,
		r.initKubernetes,
		r.initLocking,
		r.initJobClient,
	}
	for _, fn := range fns {
		fn()
		if r.err != nil {
			return r.err
		}
	}
	return nil
}

// NewReconciler returns a new instance of Reconciliation
func NewReconciler(config ReconciliationConfig) (*Reconciliation, error) {
	// build a new instance of Reconciliation
	r := &Reconciliation{
		command:       config.Command,
		childJob:      config.Child,
		commandStatus: &types.CommandStatus{},
	}
	// initialize fields
	err := r.init()
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Reconciliation) createChildJob() (types.CommandStatus, error) {
	got, err := r.jobClient.
		Namespace(r.childJob.GetNamespace()).
		Create(
			r.childJob,
			v1.CreateOptions{},
		)
	if err != nil {
		return types.CommandStatus{}, err
	}
	return types.CommandStatus{
		Phase: types.CommandPhaseJobCreated,
		Message: fmt.Sprintf(
			"Command Job created: %s %s: %s",
			got.GetNamespace(),
			got.GetName(),
			got.GetUID(),
		),
	}, nil
}

func (r *Reconciliation) deleteChildJob() (types.CommandStatus, error) {
	err := r.jobClient.
		Namespace(r.childJob.GetNamespace()).
		Delete(
			r.childJob.GetName(),
			&v1.DeleteOptions{
				// Delete immediately
				GracePeriodSeconds: pointer.Int64(0),
			},
		)
	if err != nil && !apierrors.IsNotFound(err) {
		return types.CommandStatus{}, err
	}
	return types.CommandStatus{
		Phase: types.CommandPhaseJobDeleted,
		Message: fmt.Sprintf(
			"Command Job deleted: %s %s: %s",
			r.childJob.GetNamespace(),
			r.childJob.GetName(),
			r.childJob.GetUID(),
		),
	}, nil
}

func (r *Reconciliation) reconcileRunOnceCommand() (types.CommandStatus, error) {
	var isDeleteChildJob = func() bool {
		if !r.isChildJobFound || !r.isChildJobCompleted {
			return false
		}
		if r.isStatusSetAsCompleted || r.isStatusSetAsError {
			return true
		}
		return false
	}
	var isCreateChildJob = func() bool {
		if !r.isStatusSet && !r.isChildJobFound {
			return true
		}
		if r.isStatusSetAsError && r.isRetryOnError && !r.isChildJobFound {
			return true
		}
		return false
	}
	if isCreateChildJob() {
		return r.createChildJob()
	}
	if isDeleteChildJob() {
		return r.deleteChildJob()
	}
	return types.CommandStatus{
		Phase:   types.CommandPhaseInProgress,
		Message: "Previous reconciliation is in-progress",
	}, nil
}

func (r *Reconciliation) reconcileRunAlwaysCommand() (types.CommandStatus, error) {
	if !r.isChildJobFound {
		return r.createChildJob()
	}
	if r.isStatusSet && r.isChildJobCompleted {
		return r.deleteChildJob()
	}
	return types.CommandStatus{
		Phase:   types.CommandPhaseInProgress,
		Message: "Previous reconciliation is in-progress",
	}, nil
}

func (r *Reconciliation) deleteChildJobOnDisabledCommand() (types.CommandStatus, error) {
	var output = types.CommandStatus{
		Phase:  types.CommandPhaseSkipped,
		Reason: "Resource is not enabled",
	}
	if !r.isChildJobFound {
		// nothing to do
		return output, nil
	}
	// Delete without any checks
	_, err := r.deleteChildJob()
	if err != nil {
		return types.CommandStatus{}, err
	}
	return output, nil
}

// Reconcile either creates or deletes a Kubernetes job or does nothing
// as part of reconciling a Command resource.
func (r *Reconciliation) Reconcile() (status types.CommandStatus, err error) {
	if r.isRunNever {
		return r.deleteChildJobOnDisabledCommand()
	}
	isLocked, err := r.locker.IsLocked()
	if err != nil {
		klog.Errorf("Failed to check lock status: %s", err.Error())
		return types.CommandStatus{}, err
	}
	if isLocked {
		return types.CommandStatus{
			Phase: types.CommandPhaseLocked,
		}, nil
	}
	_, unlock, err := r.locker.Lock()
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return types.CommandStatus{
				Phase: types.CommandPhaseLocked,
			}, nil
		}
		return types.CommandStatus{}, err
	}
	// make use of defer to UNLOCK
	defer func() {
		// FORCE UNLOCK in case of the following:
		// - Executing command resulted in error
		//
		// NOTE:
		//	Lock is removed to enable subsequent reconcile attempts
		if err != nil {
			_, unlockerr := r.locker.MustUnlock()
			if unlockerr != nil {
				// swallow unlock error by logging
				klog.Errorf(
					"Forced unlock failed: Command %q %q: Status %q %q: %s",
					r.command.Namespace,
					r.command.Name,
					r.commandStatus.Phase,
					r.commandStatus.Message,
					unlockerr.Error(),
				)
				// bubble up the original error
				return
			}
			klog.V(3).Infof(
				"Forced unlock was successful: Command %q %q: Status %q %q",
				r.command.Namespace,
				r.command.Name,
				r.commandStatus.Phase,
				r.commandStatus.Message,
			)
			// bubble up the original error if any
			return
		}
		// GRACEFUL UNLOCK
		//
		// NOTE:
		//	Unlocking lets this Command to be executed in its next
		// reconcile attempts if they are meant to be run ALWAYS
		//
		// NOTE:
		// 	Command that is set to be run ALWAYS follow below steps:
		// 1/ Lock,
		// 2/ Execute, &
		// 3/ Unlock
		unlockstatus, unlockerr := unlock()
		if unlockerr != nil {
			// swallow the unlock error by logging
			klog.Errorf(
				"Graceful unlock failed: Command %q %q: Status %q %q: %s",
				r.command.Namespace,
				r.command.Name,
				r.commandStatus.Phase,
				r.commandStatus.Message,
				unlockerr.Error(),
			)
			// return the executed state
			return
		}
		klog.V(3).Infof(
			"Unlocked gracefully: Command %q %q: Status %q %q: %s",
			r.command.Namespace,
			r.command.Name,
			r.commandStatus.Phase,
			r.commandStatus.Message,
			unlockstatus,
		)
	}()
	if r.isRunOnce {
		return r.reconcileRunOnceCommand()
	}
	return r.reconcileRunAlwaysCommand()
}
