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
	"strings"

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

	// getChildJob will hold function to fetch the child object
	// from k8s cluster
	// NOTE: This is helpful to mocking
	getChildJob func() (*unstructured.Unstructured, bool, error)

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
	var got *unstructured.Unstructured
	var found bool
	var err error

	if r.childJob == nil || r.childJob.Object == nil {
		return
	}

	if r.childJob.GetKind() != types.KindJob ||
		r.childJob.GetAPIVersion() != types.JobAPIVersion {
		r.err = errors.Errorf(
			"Invalid child: Expected %q %q: Got %q %q: Command %q / %q",
			types.JobAPIVersion,
			types.KindJob,
			r.childJob.GetAPIVersion(),
			r.childJob.GetKind(),
			r.command.GetNamespace(),
			r.command.GetName(),
		)
		return
	}

	if r.getChildJob != nil {
		got, found, err = r.getChildJob()
	} else {
		got, found, err = r.isChildJobAvailable()
	}
	if err != nil {
		r.err = err
		return
	}

	if !found {
		klog.V(3).Infof(
			"Job is not available: Command %q / %q",
			r.command.GetNamespace(),
			r.command.GetName(),
		)
		return
	}

	// At this point Job is present in Kubernetes cluster
	r.isChildJobFound = true

	// Extract status.failed of this Job
	failedCount, found, err := unstructured.NestedInt64(
		got.Object,
		"status",
		"failed",
	)
	if err != nil {
		r.err = errors.Wrapf(
			err,
			"Failed to get Job status.failed: Kind %q: Job %q / %q",
			r.childJob.GetKind(),
			r.childJob.GetNamespace(),
			r.childJob.GetName(),
		)
		return
	}
	if !found {
		klog.V(1).Infof(
			"Job status.failed is not set: Kind %q: Job %q / %q",
			r.childJob.GetKind(),
			r.childJob.GetNamespace(),
			r.childJob.GetName(),
		)
	}
	if failedCount > 0 {
		r.isChildJobCompleted = false
		return
	}

	// Extract status.conditions of this Job to know whether
	// job has completed its execution
	jobConditions, found, err := unstructured.NestedSlice(
		got.Object,
		"status",
		"conditions",
	)
	if err != nil {
		r.err = errors.Wrapf(
			err,
			"Failed to get Job status.conditions: Kind %q: Job %q / %q",
			r.childJob.GetKind(),
			r.childJob.GetNamespace(),
			r.childJob.GetName(),
		)
		return
	}
	if !found {
		klog.V(1).Infof(
			"Job status.conditions is not set: Kind %q: Job %q / %q",
			r.childJob.GetKind(),
			r.childJob.GetNamespace(),
			r.childJob.GetName(),
		)
		// Job's status.conditions is not set
		//
		// Nothing to do
		// Wait for next reconcile
		return
	}
	// Look for condition type complete
	// if found then mark isChildJobCompleted as true
	for _, value := range jobConditions {
		condition, ok := value.(map[string]interface{})
		if !ok {
			r.err = errors.Errorf(
				"Job status.condition is not map[string]interface{} got %T: "+
					"kind %q: Job %q / %q",
				value,
				r.childJob.GetKind(),
				r.childJob.GetNamespace(),
				r.childJob.GetName(),
			)
			return
		}
		condType := condition["type"].(string)
		if condType == types.JobPhaseCompleted {
			condStatus := condition["status"].(string)
			if strings.ToLower(condStatus) == "true" {
				r.isChildJobCompleted = true
			}
		}
	}

	// If there is no condtion with complete type then
	// nothing to do

	// wait for next reconciliation
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
			"Failed to init lock: Nil kube details: Command %q / %q",
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
		r.initKubernetes,
		r.initJobClient,
		r.initChildJobDetails,
		r.initCommandDetails,
		r.initLocking,
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
		return types.CommandStatus{}, errors.Wrapf(
			err,
			"Failed to create job: Name %q / %q",
			r.childJob.GetNamespace(),
			r.childJob.GetName(),
		)
	}
	return types.CommandStatus{
		Phase: types.CommandPhaseJobCreated,
		Message: fmt.Sprintf(
			"Command Job created: %q %q: %q",
			got.GetNamespace(),
			got.GetName(),
			got.GetUID(),
		),
	}, nil
}

func (r *Reconciliation) isChildJobAvailable() (*unstructured.Unstructured, bool, error) {
	got, err := r.jobClient.
		Namespace(r.childJob.GetNamespace()).
		Get(
			r.childJob.GetName(),
			v1.GetOptions{},
		)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, errors.Wrapf(
			err,
			"Failed to get job: Name %q / %q",
			r.childJob.GetNamespace(),
			r.childJob.GetName(),
		)
	}
	return got, got != nil, nil
}

func (r *Reconciliation) deleteChildJob() (types.CommandStatus, error) {
	// If propagationPolicy is set to background then the garbage collector will
	// delete dependents in the background
	propagationPolicy := v1.DeletePropagationBackground
	err := r.jobClient.
		Namespace(r.childJob.GetNamespace()).
		Delete(
			r.childJob.GetName(),
			&v1.DeleteOptions{
				// Delete immediately
				GracePeriodSeconds: pointer.Int64(0),
				PropagationPolicy:  &propagationPolicy,
			},
		)
	if err != nil && !apierrors.IsNotFound(err) {
		return types.CommandStatus{}, err
	}
	return types.CommandStatus{
		Phase: types.CommandPhaseJobDeleted,
		Message: fmt.Sprintf(
			"Command Job deleted: %q / %q: %q",
			r.childJob.GetNamespace(),
			r.childJob.GetName(),
			r.childJob.GetUID(),
		),
	}, nil
}

func (r *Reconciliation) reconcileRunOnceCommand() (types.CommandStatus, error) {
	klog.V(1).Infof(
		"Reconcile started: Run once: Command %q / %q",
		r.command.GetNamespace(),
		r.command.GetName(),
	)
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
		klog.V(1).Infof(
			"Will create command job: Command %q / %q",
			r.command.GetNamespace(),
			r.command.GetName(),
		)
		return r.createChildJob()
	}
	if isDeleteChildJob() {
		klog.V(1).Infof(
			"Will delete command job: Command %q / %q",
			r.command.GetNamespace(),
			r.command.GetName(),
		)
		return r.deleteChildJob()
	}
	klog.V(1).Infof(
		"Previous reconciliation is in-progress: Command %q / %q",
		r.command.GetNamespace(),
		r.command.GetName(),
	)
	return types.CommandStatus{
		Phase:   types.CommandPhaseInProgress,
		Message: "Previous reconciliation is in-progress",
	}, nil
}

func (r *Reconciliation) reconcileRunAlwaysCommand() (types.CommandStatus, error) {
	klog.V(1).Infof(
		"Reconcile started: Run always: Command %q / %q",
		r.command.GetNamespace(),
		r.command.GetName(),
	)
	if !r.isChildJobFound {
		klog.V(1).Infof(
			"Will create command job: Command %q / %q",
			r.command.GetNamespace(),
			r.command.GetName(),
		)
		return r.createChildJob()
	}
	if r.isStatusSet && r.isChildJobCompleted {
		// Since this is for run always we are performing below steps
		// 1. Delete Job and wait til it gets deleted from etcd
		// 2. Create Job in the same reconciliation
		klog.V(1).Infof(
			"Will delete command job: Command %q / %q",
			r.command.GetNamespace(),
			r.command.GetName(),
		)
		_, err := r.deleteChildJob()
		if err != nil {
			return types.CommandStatus{}, err
		}

		// Logic to wait for Job resource deletion from etcd
		var message = fmt.Sprintf(
			"Waiting for command job: %q / %q deletion",
			r.childJob.GetNamespace(),
			r.childJob.GetName(),
		)
		err = r.Retry.Waitf(
			func() (bool, error) {
				_, err := r.jobClient.
					Namespace(r.childJob.GetNamespace()).
					Get(r.childJob.GetName(), v1.GetOptions{})
				if err != nil {
					if apierrors.IsNotFound(err) {
						return true, nil
					}
					return false, err
				}
				return false, nil
			},
			message,
		)

		klog.V(1).Infof("Deleted command job: Command %q / %q successfully",
			r.command.GetNamespace(),
			r.command.GetName(),
		)

		klog.V(1).Infof(
			"Will create command job: Command %q / %q",
			r.command.GetNamespace(),
			r.command.GetName(),
		)
		return r.createChildJob()
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
	klog.V(1).Infof(
		"Will delete command job: Command is disabled: Command %q / %q",
		r.command.GetNamespace(),
		r.command.GetName(),
	)
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
		klog.Errorf(
			"Failed to check lock status for command: %q / %q: %s",
			r.command.GetNamespace(),
			r.command.GetName(),
			err.Error(),
		)
		return types.CommandStatus{}, err
	}
	if isLocked {
		klog.V(3).Infof(
			"Will skip command reconciliation: It is locked: Command %q / %q",
			r.command.GetNamespace(),
			r.command.GetName(),
		)
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
					"Forced unlock failed: Command %q / %q: Status %q %q: %s",
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
				"Forced unlock was successful: Command %q / %q: Status %q %q",
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
				"Graceful unlock failed: Command %q / %q: Status %q %q: %s",
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
			"Unlocked gracefully: Command %q / %q: Status %q %q: %s",
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
