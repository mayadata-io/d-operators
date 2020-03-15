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

package http

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	types "mayadata.io/d-operators/types/director"
	"mayadata.io/d-operators/types/gvk"
	http "mayadata.io/d-operators/types/http"
)

func (r *Reconciler) addGetActiveNodes() {
	if !r.observedIncludes.ContainsExact(types.IncludeAllAPIs) &&
		!r.observedIncludes.ContainsExact(types.GetActiveNodes) {
		// nothing to be done since this is not included
		// as part of DIrectorHTTP
		return
	}
	var headers, pathParams map[string]string
	if r.observedHTTPData != nil {
		headers = r.observedHTTPData.Spec.Headers
		pathParams = r.observedHTTPData.Spec.PathParams
	}
	r.desiredStates = append(
		r.desiredStates,
		&unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       gvk.KindHTTP,
				"apiVersion": gvk.VersionV1Alpha1,
				"metadata": map[string]interface{}{
					"name":      types.GetActiveNodes,
					"namespace": r.observedDirectorNamespace,
					"annotations": map[string]interface{}{
						"directorhttp.dao.mayadata.io/uid": r.observedDirector.GetUID(),
					},
				},
				"spec": map[string]interface{}{
					"secretName": r.observedSecretName,
					"headers":    headers,
					"pathParams": pathParams,
					"url":        types.URLActiveNodes,
					"method":     http.GET,
				},
			},
		},
	)
}
