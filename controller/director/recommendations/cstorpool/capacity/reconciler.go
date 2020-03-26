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

package capacity

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/json"
	"openebs.io/metac/controller/generic"

	ctrlutil "mayadata.io/d-operators/common/controller"
	"mayadata.io/d-operators/common/unstruct"
	types "mayadata.io/d-operators/types/director"
	"mayadata.io/d-operators/types/gvk"
	http "mayadata.io/d-operators/types/http"
)

// Reconciler manages reconciliation of HTTP resources
type Reconciler struct {
	ctrlutil.Reconciler

	observedCSPCapRecommend    *types.CStorPoolCapacityRecommendation
	observedHTTPDataName       string
	observedHTTPData           *http.HTTPData
	observedSecretName         string
	observedClusterID          string
	observedRAIDDeviceCount    int
	observedRAIDType           string
	observedRecommendationList *http.HTTP
	observedPathParams         map[string]string

	desiredRecommendationListName                 string
	desiredCSPCapRecommendationName               string
	desiredRecommendationID                       string
	desiredCSPCapacityRecommendationBody          string
	desiredPathParamsForCSPCapacityRecommendation map[string]string
}

func (r *Reconciler) walkObservedCSPCapacityRecommendation() {
	var recommend types.CStorPoolCapacityRecommendation
	err := unstruct.ToTyped(
		r.HookRequest.Watch,
		&recommend,
	)
	if err != nil {
		r.Err = err
		return
	}
	r.observedCSPCapRecommend = &recommend
	// set httpdata name
	r.observedHTTPDataName = recommend.Spec.HTTPDataName
	if r.observedHTTPDataName == "" {
		r.Err = errors.Errorf("Missing spec.httpDataName")
	}
	// set secret name
	r.observedSecretName = recommend.Spec.SecretName
	if r.observedSecretName == "" {
		r.Err = errors.Errorf("Missing spec.secretName")
		return
	}
	// set device count
	r.observedRAIDDeviceCount = recommend.Spec.RAIDDeviceCount
	if r.observedRAIDDeviceCount == 0 {
		r.Err = errors.Errorf("Missing / Invalid spec.raidDeviceCount")
		return
	}
	// set raid type
	r.observedRAIDType = recommend.Spec.RAIDType
	if r.observedRAIDType == "" {
		r.Err = errors.Errorf("Missing spec.raidType")
		return
	}
	// set desired recommendation list name
	r.desiredRecommendationListName =
		recommend.GetName() + "-recommend-list"
	// set desired csp capacity recommendation name
	r.desiredCSPCapRecommendationName =
		recommend.GetName() + "-csp-capacity-recommend"
}

func (r *Reconciler) walkObservedHTTPData() {
	var httpdata http.HTTPData
	obj := r.HookRequest.Attachments.FindByGroupKindName(
		gvk.GroupDAOMayadataIO,
		gvk.KindHTTPData,
		r.observedHTTPDataName,
	)
	if obj == nil {
		r.Err = errors.Errorf(
			"HTTPData %q not found",
			r.observedHTTPDataName,
		)
		return
	}
	// add it back to response attachments
	//
	// NOTE:
	//	Adding back the request attachments to response
	// attachments help in evaluating the completion state
	r.HookResponse.Attachments = append(
		r.HookResponse.Attachments,
		obj,
	)
	// extract path params from HTTPData
	err := unstruct.ToTyped(
		obj,
		&httpdata,
	)
	if err != nil {
		r.Err = err
		return
	}
	r.observedHTTPData = &httpdata
	// extract cluster id from HTTPData
	r.observedPathParams = r.observedHTTPData.Spec.PathParams
	if len(r.observedPathParams) == 0 {
		r.Err = errors.Errorf(
			"Missing spec.pathParams in %q / %q: %s",
			r.observedHTTPData.GetNamespace(),
			r.observedHTTPData.GetName(),
			r.observedHTTPData.GetObjectKind().GroupVersionKind().String(),
		)
		return
	}
	r.observedClusterID = r.observedPathParams["clusterId"]
	// this is used in the body of some requests
	if r.observedClusterID == "" {
		r.Err = errors.Errorf(
			"Missing spec.pathParams.clusterId in %q / %q: %s",
			r.observedHTTPData.GetNamespace(),
			r.observedHTTPData.GetName(),
			r.observedHTTPData.GetObjectKind().GroupVersionKind().String(),
		)
		return
	}
}

func (r *Reconciler) walkObservedRecommendationList() {
	var observedRecommendationList http.HTTP
	obj := r.HookRequest.Attachments.FindByGroupKindName(
		gvk.GroupDAOMayadataIO,
		gvk.KindHTTP,
		r.desiredRecommendationListName,
	)
	if obj == nil {
		// reconciliation has not yet happened
		r.Warns = append(
			r.Warns,
			"No recommendation list observed",
		)
		return
	}
	err := unstruct.ToTyped(
		obj,
		&observedRecommendationList,
	)
	if err != nil {
		r.Err = errors.Wrapf(
			err,
			"Failed to convert recommendation list to typed",
		)
		return
	}
	r.observedRecommendationList = &observedRecommendationList
}

func (r *Reconciler) setDesiredRecommendationIDForCStorPool() {
	if r.observedRecommendationList == nil {
		// nothing to do
		return
	}
	rList := r.observedRecommendationList
	var isComplete = "false"
	if rList.Status.Completion != nil {
		isComplete = fmt.Sprintf(
			"%v",
			rList.Status.Completion["state"],
		)
	}
	// extract current phase
	phase := rList.Status.Phase
	if isComplete == "false" || phase == "" || phase == http.HTTPStatusPhaseError {
		r.Warns = append(
			r.Warns,
			"Yet to receive recommendation list response: IsComplete=%t: Phase=%q",
			isComplete,
			phase,
		)
		return
	}
	// extract the http response of recommendation list API
	recommendations, err := json.Marshal(rList.Status.Body)
	if err != nil {
		r.Err = errors.Wrapf(
			err,
			"Failed to marshal status.body of %q / %q: %s",
			rList.GetNamespace(),
			rList.GetName(),
			rList.GetObjectKind().GroupVersionKind().String(),
		)
		return
	}
	// extract recommendation id relevant to cStorPool
	r.desiredRecommendationID = gjson.Get(
		string(recommendations),
		`data.#(name=="cStorPool").id`,
	).String()
	// verify if it was found
	if r.desiredRecommendationID == "" {
		r.Err = errors.Wrapf(
			err,
			"Missing recommendation id for cStorPool in %q / %q: %s",
			rList.GetNamespace(),
			rList.GetName(),
			rList.GetObjectKind().GroupVersionKind().String(),
		)
		return
	}
}

func (r *Reconciler) setDesiredPathParamsForCSPCapRecommendation() {
	if r.desiredRecommendationID == "" {
		return
	}
	r.desiredPathParamsForCSPCapacityRecommendation = map[string]string{
		"recommendation_csp_id": r.desiredRecommendationID,
	}
	for k, v := range r.observedPathParams {
		r.desiredPathParamsForCSPCapacityRecommendation[k] = v
	}
}

func (r *Reconciler) setDesiredRecommendationList() {
	cspcCapRecommend := r.observedCSPCapRecommend
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": gvk.DAOMayadataIOV1Alpha1,
			"kind":       gvk.KindHTTP,
			"metadata": map[string]interface{}{
				"name":      r.desiredRecommendationListName,
				"namespace": cspcCapRecommend.GetNamespace(),
				"annotations": map[string]interface{}{
					// mention the resource that created this
					"cspc-capacity-recommend.dao.mayadata.io/uid": cspcCapRecommend.GetUID(),
				},
			},
			"spec": map[string]interface{}{
				"secretName": r.observedSecretName,
				"url":        types.URLRecommendationList,
				"pathParams": r.observedPathParams,
				"method":     http.GET,
			},
		},
	}
	r.HookResponse.Attachments = append(
		r.HookResponse.Attachments,
		obj,
	)
}

func (r *Reconciler) buildDesiredCSPCapacityRecommendationBody() {
	if r.desiredRecommendationID == "" {
		// yet to receive list of recommendations
		return
	}
	body := map[string]interface{}{
		"clusterId": r.observedClusterID,
		"raidGroupConfig": map[string]interface{}{
			"groupDeviceCount": r.observedRAIDDeviceCount,
			"type":             r.observedRAIDType,
		},
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		r.Err = errors.Wrapf(
			err,
			"Failed to build capacity based cstor pool recommendation",
		)
		return
	}
	r.desiredCSPCapacityRecommendationBody = string(bodyBytes)
}

func (r *Reconciler) setDesiredCSPCapacityRecommendation() {
	if r.desiredRecommendationID == "" {
		// yet to receive list of recommendations
		return
	}
	cspcCapRecommend := r.observedCSPCapRecommend
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": gvk.DAOMayadataIOV1Alpha1,
			"kind":       gvk.KindHTTP,
			"metadata": map[string]interface{}{
				"name":      r.desiredCSPCapRecommendationName,
				"namespace": cspcCapRecommend.GetNamespace(),
				"annotations": map[string]interface{}{
					// mention the resource that created this
					"cspc-capacity-recommend.dao.mayadata.io/uid": cspcCapRecommend.GetUID(),
				},
			},
			"spec": map[string]interface{}{
				"secretName": r.observedSecretName,
				"url":        types.URLGetCSPCapacityRecommendation,
				"pathParams": r.desiredPathParamsForCSPCapacityRecommendation,
				"method":     http.POST,
				"body":       r.desiredCSPCapacityRecommendationBody,
			},
		},
	}
	r.HookResponse.Attachments = append(
		r.HookResponse.Attachments,
		obj,
	)
}

func (r *Reconciler) updateWatchStatus() {
	var status = map[string]interface{}{}
	var completion = map[string]interface{}{
		"state": false,
	}
	var warn string
	// init with Online
	status["phase"] = types.HTTPStatusOnline
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
		status["phase"] = types.HTTPStatusError
		status["reason"] = r.Err.Error()
	}
	// hook request has the observed state of children
	observedAttachments := r.HookRequest.Attachments.Len()
	// hook response has the desired state of children
	desiredAttachments := len(r.HookResponse.Attachments)
	// we expect 3 attachments in response:
	//
	// - HTTPData
	// - HTTP for RecommendationList
	// - HTTP for CStorPool Capacity Recommendation
	if r.Err == nil && desiredAttachments == 3 {
		completion["state"] = true
	}
	completion["observedAttachmentCount"] = observedAttachments
	completion["desiredAttachmentCount"] = desiredAttachments
	// set completion status
	status["completion"] = completion
	// set the desired status against hook response
	r.HookResponse.Status = status
}

// Sync implements the idempotent logic to sync DirectorHTTP
//
// NOTE:
// 	SyncHookRequest is the payload received as part of reconcile
// request. Similarly, SyncHookResponse is the payload sent as a
// response as part of reconcile request.
//
// NOTE:
//	This controller watches CStorPoolCapacityRecommendation custom
// resource
func Sync(request *generic.SyncHookRequest, response *generic.SyncHookResponse) error {
	r := &Reconciler{
		Reconciler: ctrlutil.Reconciler{
			HookRequest:  request,
			HookResponse: response,
		},
	}

	// add logic to achieve desired state of attachments/children
	r.ReconcileFns = []func(){
		r.walkObservedCSPCapacityRecommendation,
		r.walkObservedHTTPData,
		r.walkObservedRecommendationList,
		r.setDesiredRecommendationList,
		r.setDesiredRecommendationIDForCStorPool,
		r.setDesiredPathParamsForCSPCapRecommendation,
		r.buildDesiredCSPCapacityRecommendationBody,
		r.setDesiredCSPCapacityRecommendation,
	}

	// add logic to achieve desired state of watch
	r.DesiredWatchFns = []func(){
		r.updateWatchStatus,
	}

	// run reconcile
	return r.Reconcile()
}
