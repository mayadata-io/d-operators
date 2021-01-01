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

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ExecutableCRDV1 helps to apply or create desired CRD state
// against the cluster
type ExecutableCRDV1 struct {
	ExecutableCRD
}

// ExecutableCRDV1Config helps in creating new instance of
// ExecutableCRDV1
type ExecutableCRDV1Config struct {
	BaseRunner       BaseRunner
	IgnoreDiscovery  bool
	State            *unstructured.Unstructured
	DesiredCRVersion string
}

// NewCRDV1Executor returns a new instance of ExecutableCRDV1Beta1
func NewCRDV1Executor(config ExecutableCRDV1Config) (*ExecutableCRDV1, error) {
	e := &ExecutableCRDV1{
		ExecutableCRD: ExecutableCRD{
			BaseRunner:       config.BaseRunner,
			IgnoreDiscovery:  config.IgnoreDiscovery,
			State:            config.State,
			DesiredCRVersion: config.DesiredCRVersion,
		},
	}
	err := e.setCRResourceAndAPIVersion()
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (e *ExecutableCRDV1) setCRResourceAndAPIVersion() error {
	plural, found, err := unstructured.NestedString(
		e.State.Object,
		"spec",
		"names",
		"plural",
	)
	if err != nil {
		return errors.Wrapf(err, "Failed to get spec.names.plural")
	}
	if !found {
		return errors.Errorf("Missing spec.names.plural")
	}
	group, found, err := unstructured.NestedString(
		e.State.Object,
		"spec",
		"group",
	)
	if err != nil {
		return errors.Wrapf(err, "Failed to get spec.group")
	}
	if !found {
		return errors.Errorf("Missing spec.group")
	}
	// Get the version that is found first in the list
	// if nothing has been set
	if e.DesiredCRVersion == "" {
		vers, found, err := unstructured.NestedSlice(
			e.State.Object,
			"spec",
			"versions",
		)
		if err != nil {
			return errors.Wrapf(err, "Failed to get spec.versions")
		}
		if !found {
			return errors.Errorf("Missing spec.versions")
		}
		for _, item := range vers {
			itemObj, ok := item.(map[string]interface{})
			if !ok {
				return errors.Errorf(
					"Expected spec.versions type as map[string]interface{} got %T",
					item,
				)
			}
			e.DesiredCRVersion = itemObj["name"].(string)
			// --
			// First version in the list is only considered
			// This value is only used for CRD discovery which
			// again is an optional feature
			// --
			break
		}
	}

	apiver := fmt.Sprintf("%s/%s", group, e.DesiredCRVersion)

	// memoize
	e.CRResource = plural
	e.CRAPIVersion = apiver

	// return resource name & apiVersion of the resource
	return nil
}
