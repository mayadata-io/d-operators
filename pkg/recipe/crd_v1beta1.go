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

// ExecutableCRDV1Beta1 helps to apply or create desired CRD state
// against the cluster
type ExecutableCRDV1Beta1 struct {
	ExecutableCRD
}

// ExecutableCRDV1Beta1Config helps in creating new instance of
// ExecutableCRDV1Beta1
type ExecutableCRDV1Beta1Config struct {
	BaseRunner      BaseRunner
	IgnoreDiscovery bool
	State           *unstructured.Unstructured
}

// NewCRDV1Beta1Executor returns a new instance of ExecutableCRDV1Beta1
func NewCRDV1Beta1Executor(config ExecutableCRDV1Beta1Config) (*ExecutableCRDV1Beta1, error) {
	e := &ExecutableCRDV1Beta1{
		ExecutableCRD: ExecutableCRD{
			BaseRunner:      config.BaseRunner,
			IgnoreDiscovery: config.IgnoreDiscovery,
			State:           config.State,
		},
	}
	err := e.setCRResourceAndAPIVersion()
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (e *ExecutableCRDV1Beta1) setCRResourceAndAPIVersion() error {
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
	ver, found, err := unstructured.NestedString(
		e.State.Object,
		"spec",
		"version",
	)
	if err != nil {
		return errors.Wrapf(err, "Failed to get spec.version")
	}
	if !found || ver == "" {
		return errors.Errorf("Missing spec.version")
	}
	apiver := fmt.Sprintf("%s/%s", group, ver)

	// memoize
	e.CRResource = plural
	e.CRAPIVersion = apiver

	// return resource name & apiVersion of the resource
	return nil
}
