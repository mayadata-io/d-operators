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
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/json"
	types "mayadata.io/d-operators/types/director"
	"mayadata.io/d-operators/types/gvk"
	http "mayadata.io/d-operators/types/http"
)

func (r *Reconciler) addAPIForActiveNodesOrNone() {
	if !r.observedIncludes.ContainsExact(types.IncludeAllAPIs) &&
		!r.observedIncludes.ContainsExact(types.GetActiveNodes) {
		// nothing to be done since this is not included
		// as part of DIrectorHTTP
		return
	}
	var body []byte
	values := r.observedHTTPData.Spec.Values
	if len(values) != 0 && values[types.GetActiveNodes] != nil {
		body, r.Err = json.Marshal(values[types.GetActiveNodes])
		if r.Err != nil {
			r.Err = errors.Wrapf(
				r.Err,
				"Invalid http request body for %s",
				types.GetActiveNodes,
			)
			return
		}
	}
	r.desiredStates = append(
		r.desiredStates,
		&unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       gvk.KindHTTP,
				"apiVersion": gvk.APIVersionDAOV1Alpha1,
				"metadata": map[string]interface{}{
					"name":      types.GetActiveNodes,
					"namespace": r.observedDirectorNamespace,
					"labels": map[string]interface{}{
						http.LabelKeyHTTPDataName: r.observedDirectorName,
					},
				},
				"spec": map[string]interface{}{
					"url":    types.URLActiveNodes,
					"method": http.GET,
					"body":   body,
				},
			},
		},
	)
}
