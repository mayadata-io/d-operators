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
	"reflect"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"
	types "mayadata.io/d-operators/types/recipe"
	dynamicapply "openebs.io/metac/dynamic/apply"
)

// StateChecking helps in verifying expected state against the
// state observed in the cluster
type StateChecking struct {
	BaseRunner
	State      *unstructured.Unstructured
	StateCheck types.StateCheck

	actualListCount int
	operator        types.StateCheckOperator
	retryOnDiff     bool
	retryOnEqual    bool
	result          *types.StateCheckResult
	err             error
}

// StateCheckingConfig is used to create an instance of StateChecking
type StateCheckingConfig struct {
	BaseRunner
	State      *unstructured.Unstructured
	StateCheck types.StateCheck
}

// NewStateChecker returns a new instance of StateChecking
func NewStateChecker(config StateCheckingConfig) *StateChecking {
	return &StateChecking{
		BaseRunner: config.BaseRunner,
		State:      config.State,
		StateCheck: config.StateCheck,
		result:     &types.StateCheckResult{},
	}
}

func (sc *StateChecking) init() {
	if sc.StateCheck.Operator == "" {
		klog.V(3).Infof(
			"Will default StateCheck operator to StateCheckOperatorEquals",
		)
		sc.operator = types.StateCheckOperatorEquals
	} else {
		sc.operator = sc.StateCheck.Operator
	}
}

func (sc *StateChecking) validate() {
	var isCountBasedAssert bool
	// evaluate if this is count based assertion
	// based on value
	if sc.StateCheck.Count != nil {
		isCountBasedAssert = true
	}
	// evaluate if its count based assertion
	// based on operator
	if !isCountBasedAssert {
		switch sc.operator {
		case types.StateCheckOperatorListCountEquals,
			types.StateCheckOperatorListCountNotEquals:
			isCountBasedAssert = true
		}
	}
	if isCountBasedAssert && sc.StateCheck.Count == nil {
		sc.err = errors.Errorf(
			"Invalid StateCheck %q: Operator %q can't be used with nil count",
			sc.TaskName,
			sc.operator,
		)
	} else if isCountBasedAssert {
		switch sc.operator {
		case types.StateCheckOperatorEquals,
			types.StateCheckOperatorNotEquals,
			types.StateCheckOperatorNotFound:
			sc.err = errors.Errorf(
				"Invalid StateCheck %q: Operator %q can't be used with count %d",
				sc.TaskName,
				sc.operator,
				*sc.StateCheck.Count,
			)
		}
	}
}

func (sc *StateChecking) isMergeEqualsObserved(context string) (bool, error) {
	var observed, merged *unstructured.Unstructured
	err := sc.Retry.Waitf(
		// returning true will stop retrying this func
		func() (bool, error) {
			client, err := sc.GetClientForAPIVersionAndKind(
				sc.State.GetAPIVersion(),
				sc.State.GetKind(),
			)
			if err != nil {
				// retry based on condition
				return sc.IsFailFastOnDiscoveryError(), err
			}
			observed, err = client.
				Namespace(sc.State.GetNamespace()).
				Get(
					sc.State.GetName(),
					metav1.GetOptions{},
				)
			if err != nil {
				// Keep retrying
				return false, err
			}
			merged = &unstructured.Unstructured{}
			merged.Object, err = dynamicapply.Merge(
				observed.UnstructuredContent(), // observed
				sc.State.UnstructuredContent(), // last applied
				sc.State.UnstructuredContent(), // desired
			)
			if err != nil {
				// Exit in case of merge error
				// Stop retrying
				return true, err
			}
			if sc.retryOnDiff && !reflect.DeepEqual(merged, observed) {
				// Keep retrying
				return false, nil
			}
			if sc.retryOnEqual && reflect.DeepEqual(merged, observed) {
				// Keep retrying
				return false, nil
			}
			// Stop retrying
			return true, nil
		},
		context,
	)
	klog.V(2).Infof(
		"Is state equal? %t: Diff\n%s",
		reflect.DeepEqual(merged, observed),
		cmp.Diff(merged, observed),
	)
	if err != nil {
		return false, err
	}
	return reflect.DeepEqual(merged, observed), nil
}

func (sc *StateChecking) assertEquals() {
	var message = fmt.Sprintf(
		"StateCheckEquals: Resource %s %s: GVK %s: TaskName %s",
		sc.State.GetNamespace(),
		sc.State.GetName(),
		sc.State.GroupVersionKind(),
		sc.TaskName,
	)
	// We want to retry in case of any difference between
	// expected and observed states. This is done with the
	// expectation of having observed state equal to the
	// expected state eventually.
	sc.retryOnDiff = true
	success, err := sc.isMergeEqualsObserved(message)
	if err != nil {
		sc.err = err
		return
	}
	// init phase as failed
	sc.result.Phase = types.StateCheckResultFailed
	if success {
		sc.result.Phase = types.StateCheckResultPassed
	}
	sc.result.Message = message
}

func (sc *StateChecking) assertNotEquals() {
	var message = fmt.Sprintf(
		"StateCheckNotEquals: Resource %s %s: GVK %s: TaskName %s",
		sc.State.GetNamespace(),
		sc.State.GetName(),
		sc.State.GroupVersionKind(),
		sc.TaskName,
	)
	// We want to retry if expected and observed states are found
	// to be equal. This is done with the expectation of having
	// observed state not equal to the expected state eventually.
	sc.retryOnEqual = true
	success, err := sc.isMergeEqualsObserved(message)
	if err != nil {
		sc.err = err
		return
	}
	// init phase as failed
	sc.result.Phase = types.StateCheckResultFailed
	if !success {
		sc.result.Phase = types.StateCheckResultPassed
	}
	sc.result.Message = message
}

func (sc *StateChecking) assertNotFound() {
	var message = fmt.Sprintf(
		"StateCheckNotFound: Resource %s %s: GVK %s: TaskName %s",
		sc.State.GetNamespace(),
		sc.State.GetName(),
		sc.State.GroupVersionKind(),
		sc.TaskName,
	)
	var warning string
	// init result to Failed
	var phase = types.StateCheckResultFailed
	err := sc.Retry.Waitf(
		func() (bool, error) {
			client, err := sc.GetClientForAPIVersionAndKind(
				sc.State.GetAPIVersion(),
				sc.State.GetKind(),
			)
			if err != nil {
				// retry based on condition
				return sc.IsFailFastOnDiscoveryError(), err
			}
			got, err := client.
				Namespace(sc.State.GetNamespace()).
				Get(
					sc.State.GetName(),
					metav1.GetOptions{},
				)
			if err != nil {
				if apierrors.IsNotFound(err) {
					// phase is set to Passed here
					phase = types.StateCheckResultPassed
					// Stop retrying
					return true, nil
				}
				// Keep retrying
				return false, err
			}
			if len(got.GetFinalizers()) == 0 && got.GetDeletionTimestamp() != nil {
				phase = types.StateCheckResultWarning
				warning = fmt.Sprintf(
					"Marking StateCheck %q to passed: Finalizer count %d: Deletion timestamp %s",
					sc.TaskName,
					len(got.GetFinalizers()),
					got.GetDeletionTimestamp(),
				)
				// Stop retrying
				return true, nil
			}
			// Keep retrying
			return false, nil
		},
		message,
	)
	if err != nil {
		if _, ok := err.(*RetryTimeout); !ok {
			sc.err = err
			return
		}
		sc.result.Timeout = err.Error()
	}
	sc.result.Phase = phase
	sc.result.Message = message
	sc.result.Warning = warning
}

func (sc *StateChecking) isListCountMatch() (bool, error) {
	client, err := sc.GetClientForAPIVersionAndKind(
		sc.State.GetAPIVersion(),
		sc.State.GetKind(),
	)
	if err != nil {
		return false, &DiscoveryError{err.Error()}
	}
	list, err := client.
		Namespace(sc.State.GetNamespace()).
		List(metav1.ListOptions{
			LabelSelector: labels.Set(
				sc.State.GetLabels(),
			).String(),
		})
	if err != nil {
		return false, err
	}
	// store this to be used later
	sc.actualListCount = len(list.Items)
	return sc.actualListCount == *sc.StateCheck.Count, nil
}

func (sc *StateChecking) assertListCountEquals() {
	var message = fmt.Sprintf(
		"AssertListCountEquals: Resource %s: GVK %s: TaskName %s",
		sc.State.GetNamespace(),
		sc.State.GroupVersionKind(),
		sc.TaskName,
	)
	// init result to Failed
	var phase = types.StateCheckResultFailed
	err := sc.Retry.Waitf(
		func() (bool, error) {
			match, err := sc.isListCountMatch()
			if err != nil {
				// Retry on condition
				return sc.IsFailFastOnError(err), err
			}
			if match {
				phase = types.StateCheckResultPassed
				// Stop retrying
				return true, nil
			}
			// Keep retrying
			return false, nil
		},
		message,
	)
	if err != nil {
		if _, ok := err.(*RetryTimeout); !ok {
			sc.err = err
			return
		}
		sc.result.Timeout = err.Error()
	}
	sc.result.Phase = phase
	sc.result.Message = message
	sc.result.Warning = fmt.Sprintf(
		"Expected count %d got %d",
		*sc.StateCheck.Count,
		sc.actualListCount,
	)
}

func (sc *StateChecking) assertListCountNotEquals() {
	var message = fmt.Sprintf(
		"AssertListCountNotEquals: Resource %s: GVK %s: TaskName %s",
		sc.State.GetNamespace(),
		sc.State.GroupVersionKind(),
		sc.TaskName,
	)
	// init result to Failed
	var phase = types.StateCheckResultFailed
	err := sc.Retry.Waitf(
		func() (bool, error) {
			match, err := sc.isListCountMatch()
			if err != nil {
				// Retry on condition
				return sc.IsFailFastOnError(err), err
			}
			if !match {
				phase = types.StateCheckResultPassed
				// Stop retrying
				return true, nil
			}
			// Keep retrying
			return false, nil
		},
		message,
	)
	if err != nil {
		if _, ok := err.(*RetryTimeout); !ok {
			sc.err = err
			return
		}
		sc.result.Timeout = err.Error()
	}
	sc.result.Phase = phase
	sc.result.Message = message
	sc.result.Warning = fmt.Sprintf(
		"Expected count %d got %d",
		*sc.StateCheck.Count,
		sc.actualListCount,
	)
}

func (sc *StateChecking) assert() {
	switch sc.operator {
	case types.StateCheckOperatorEquals:
		sc.assertEquals()

	case types.StateCheckOperatorNotEquals:
		sc.assertNotEquals()

	case types.StateCheckOperatorNotFound:
		sc.assertNotFound()

	case types.StateCheckOperatorListCountEquals:
		sc.assertListCountEquals()

	case types.StateCheckOperatorListCountNotEquals:
		sc.assertListCountNotEquals()

	default:
		sc.err = errors.Errorf(
			"StateCheck %q failed: Invalid operator %q",
			sc.TaskName,
			sc.operator,
		)
	}
}

// Run executes the assertion
func (sc *StateChecking) Run() (types.StateCheckResult, error) {
	var fns = []func(){
		sc.init,
		sc.validate,
		sc.assert,
	}
	for _, fn := range fns {
		fn()
		if sc.err != nil {
			return types.StateCheckResult{}, errors.Wrapf(
				sc.err,
				"Info %s: Warn %s: Verbose %s",
				sc.result.Message,
				sc.result.Warning,
				sc.result.Verbose,
			)
		}
	}
	return *sc.result, nil
}
