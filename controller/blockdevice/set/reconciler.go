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

package set

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"openebs.io/metac/controller/generic"

	ctrlutil "mayadata.io/d-operators/common/controller"
	"mayadata.io/d-operators/common/unstruct"
	types "mayadata.io/d-operators/types/blockdevice"
)

// Reconciler manages reconciliation of HTTP resources
type Reconciler struct {
	ctrlutil.Reconciler

	observedBlockDeviceSet      *types.BlockDeviceSet
	observedBlockDeviceReplicas int

	desiredBlockDeviceTemplate *unstructured.Unstructured
	desiredBlockDeviceName     string
}

func (r *Reconciler) walkObservedBlockDeviceSet() {
	var bdset types.BlockDeviceSet
	err := unstruct.ToTyped(
		r.HookRequest.Watch,
		&bdset,
	)
	if err != nil {
		r.Err = err
		return
	}
	r.observedBlockDeviceSet = &bdset
	if bdset.Spec.Replicas == nil {
		// by default only one block device gets created
		r.observedBlockDeviceReplicas = 1
	} else {
		r.observedBlockDeviceReplicas = *bdset.Spec.Replicas
	}
}

func (r *Reconciler) buildDesiredBlockDeviceTemplate() {
	bd := r.observedBlockDeviceSet.Spec.Device
	if len(bd) == 0 {
		r.Err = errors.Errorf(
			"Invalid BlockDeviceSet: Missing spec.device: %+v",
			r.observedBlockDeviceSet,
		)
		return
	}
	obj := &unstructured.Unstructured{}
	obj.SetUnstructuredContent(bd)
	// verify if its a proper unstructured instance
	_, r.Err = obj.MarshalJSON()
	if r.Err != nil {
		r.Err = errors.Wrapf(
			r.Err,
			"Invalid BlockDeviceSet %q / %q",
			r.observedBlockDeviceSet.GetNamespace(),
			r.observedBlockDeviceSet.GetName(),
		)
		return
	}
	r.desiredBlockDeviceTemplate = obj
}

func (r *Reconciler) updateDesiredBlockDeviceTemplateAnnotations() {
	anns := r.desiredBlockDeviceTemplate.GetAnnotations()
	if anns == nil {
		anns = make(map[string]string)
	}
	// add BlockDeviceSet UID to the annotations
	anns["blockdeviceset.dao.mayadata.io/uid"] =
		string(r.observedBlockDeviceSet.GetUID())
	// set the updated labels
	r.desiredBlockDeviceTemplate.SetAnnotations(anns)
}

func (r *Reconciler) resetAndBuildBlockDeviceName() {
	// start with generate name if any
	name := r.desiredBlockDeviceTemplate.GetGenerateName()
	// reset desired's generate name to empty
	r.desiredBlockDeviceTemplate.SetGenerateName("")
	if name == "" {
		// use desired's name if any
		name = r.desiredBlockDeviceTemplate.GetName()
	}
	if name == "" {
		// use observed BlockDeviceSet name otherwise
		name = r.observedBlockDeviceSet.GetName()
	}
	// final desired name for block device(s)
	r.desiredBlockDeviceName = name
}

func (r *Reconciler) setAllDesiredBlockDevices() {
	for i := 0; i < r.observedBlockDeviceReplicas; i++ {
		obj := r.desiredBlockDeviceTemplate
		name := fmt.Sprintf("%s-%d", r.desiredBlockDeviceName, i)
		obj.SetName(name)
		r.HookResponse.Attachments = append(
			r.HookResponse.Attachments,
			obj,
		)
	}
}

// updateWatchStatus updates the watched HTTP resource's
// status field with the response received due to invocation
// of HTTP url.
//
// NOTE:
//	This forms the core business logic of reconciling a HTTP
// custom resource.
func (r *Reconciler) updateWatchStatus() {
	var status = map[string]interface{}{}
	var completion = map[string]interface{}{
		"state": false,
	}
	var warn string
	// init with Online
	status["phase"] = types.BlockDeviceSetStatusOnline
	// check for warnings
	if len(r.Warns) != 0 {
		warn = fmt.Sprintf(
			"%d warnings: %s",
			len(r.Warns),
			strings.Join(r.Warns, ": "),
		)
	}
	if warn != "" {
		status["warn"] = warn
	}
	// is runtime error
	if r.Err != nil {
		status["phase"] = types.BlockDeviceSetStatusError
		status["reason"] = r.Err.Error()
	}
	// hook request has the observed state of children
	observedReplicas := r.HookRequest.Attachments.Len()
	// hook response has the desired state of children
	desiredReplicas := len(r.HookResponse.Attachments)
	if r.Err == nil && observedReplicas == desiredReplicas {
		completion["state"] = true
	}
	completion["observedReplicas"] = observedReplicas
	completion["desiredReplicas"] = desiredReplicas
	// set completion status
	status["completion"] = completion
	// set the desired status
	r.HookResponse.Status = status
}

// Sync implements the idempotent logic to sync HTTP
//
// NOTE:
// 	SyncHookRequest is the payload received as part of reconcile
// request. Similarly, SyncHookResponse is the payload sent as a
// response as part of reconcile request.
//
// NOTE:
//	This controller watches BlockDeviceSet custom resource
func Sync(request *generic.SyncHookRequest, response *generic.SyncHookResponse) error {
	r := &Reconciler{
		Reconciler: ctrlutil.Reconciler{
			HookRequest:  request,
			HookResponse: response,
		},
	}

	// add functions to achieve desired state
	r.ReconcileFns = []func(){
		r.walkObservedBlockDeviceSet,
		r.buildDesiredBlockDeviceTemplate,
		r.updateDesiredBlockDeviceTemplateAnnotations,
		r.resetAndBuildBlockDeviceName,
		r.setAllDesiredBlockDevices,
	}

	// add functions to achieve desired watch
	r.DesiredWatchFns = []func(){
		r.updateWatchStatus,
	}
	// run reconcile
	return r.Reconcile()
}
