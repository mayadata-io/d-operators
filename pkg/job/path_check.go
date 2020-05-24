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
	"fmt"
	"strings"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
	types "mayadata.io/d-operators/types/job"
)

// PathChecking helps in verifying expected field value
// against the field value of observed resource found in
// the cluster
type PathChecking struct {
	*Fixture
	Retry *Retryable

	TaskName  string
	State     *unstructured.Unstructured
	PathCheck types.PathCheck

	operator       types.PathCheckOperator
	dataType       types.PathValueDataType
	pathOnlyCheck  bool
	valueOnlyCheck bool

	retryIfValueNotLTE    bool
	retryIfValueNotGTE    bool
	retryIfValueNotEquals bool
	retryIfValueEquals    bool
	retryIfPathNotExists  bool
	retryIfPathExists     bool

	result *types.PathCheckResult
	err    error
}

// PathCheckingConfig is used to create an instance of PathChecking
type PathCheckingConfig struct {
	Fixture   *Fixture
	Retry     *Retryable
	TaskName  string
	State     *unstructured.Unstructured
	PathCheck types.PathCheck
}

// NewPathChecker returns a new instance of PathChecking
func NewPathChecker(config PathCheckingConfig) *PathChecking {
	return &PathChecking{
		TaskName:  config.TaskName,
		Fixture:   config.Fixture,
		State:     config.State,
		Retry:     config.Retry,
		PathCheck: config.PathCheck,
		result:    &types.PathCheckResult{},
	}
}

func (pc *PathChecking) init() {
	if pc.PathCheck.Operator == "" {
		// defaults to Exists operator
		klog.V(3).Infof(
			"Will default PathCheck operator to PathCheckOperatorExists",
		)
		pc.operator = types.PathCheckOperatorExists
	} else {
		pc.operator = pc.PathCheck.Operator
	}
	if pc.PathCheck.DataType == "" {
		// defaults to int64 data type
		klog.V(3).Infof(
			"Will default PathCheck datatype to PathValueDataTypeInt64",
		)
		pc.dataType = types.PathValueDataTypeInt64
	} else {
		pc.dataType = pc.PathCheck.DataType
	}
	switch pc.operator {
	case types.PathCheckOperatorExists,
		types.PathCheckOperatorNotExists:
		pc.pathOnlyCheck = true
	default:
		pc.valueOnlyCheck = true
	}
}

func (pc *PathChecking) validate() {
	switch pc.operator {
	case types.PathCheckOperatorExists,
		types.PathCheckOperatorNotExists:
		if pc.valueOnlyCheck {
			pc.err = errors.Errorf(
				"Invalid PathCheck: Operator %q can't be used with value %v: TaskName %s",
				pc.operator,
				pc.PathCheck.Value,
				pc.TaskName,
			)
		}
	}
}

func (pc *PathChecking) assertValueInt64(obj *unstructured.Unstructured) (bool, error) {
	got, found, err := unstructured.NestedInt64(
		obj.UnstructuredContent(),
		strings.Split(pc.PathCheck.Path, ".")...,
	)
	if err != nil {
		return false, err
	}
	if !found {
		return false, errors.Errorf(
			"PathCheck failed: Path %q not found: TaskName %s",
			pc.PathCheck.Path,
			pc.TaskName,
		)
	}
	val := pc.PathCheck.Value
	expected, ok := val.(int64)
	if !ok {
		return false, errors.Errorf(
			"PathCheck failed: %v is of type %T, expected int64: TaskName %s",
			val,
			val,
			pc.TaskName,
		)
	}
	pc.result.Verbose = fmt.Sprintf(
		"Expected value %d got %d",
		expected,
		got,
	)
	if pc.retryIfValueEquals && got == expected {
		return false, nil
	}
	if pc.retryIfValueNotEquals && got != expected {
		return false, nil
	}
	if pc.retryIfValueNotGTE && got < expected {
		return false, nil
	}
	if pc.retryIfValueNotLTE && got > expected {
		return false, nil
	}
	// returning true will no longer retry
	return true, nil
}

func (pc *PathChecking) assertValueFloat64(obj *unstructured.Unstructured) (bool, error) {
	got, found, err := unstructured.NestedFloat64(
		obj.UnstructuredContent(),
		strings.Split(pc.PathCheck.Path, ".")...,
	)
	if err != nil {
		return false, err
	}
	if !found {
		return false, errors.Errorf(
			"PathCheck failed: Path %q not found: TaskName %s",
			pc.PathCheck.Path,
			pc.TaskName,
		)
	}
	val := pc.PathCheck.Value
	expected, ok := val.(float64)
	if !ok {
		return false, errors.Errorf(
			"PathCheck failed: Value %v is of type %T, expected float64: TaskName %s",
			val,
			val,
			pc.TaskName,
		)
	}
	pc.result.Verbose = fmt.Sprintf(
		"Expected value %f got %f",
		expected,
		got,
	)
	if pc.retryIfValueEquals && got == expected {
		return false, nil
	}
	if pc.retryIfValueNotEquals && got != expected {
		return false, nil
	}
	if pc.retryIfValueNotGTE && got < expected {
		return false, nil
	}
	if pc.retryIfValueNotLTE && got > expected {
		return false, nil
	}
	// returning true will no longer retry
	return true, nil
}

func (pc *PathChecking) assertValue(obj *unstructured.Unstructured) (bool, error) {
	if pc.PathCheck.DataType == types.PathValueDataTypeInt64 {
		return pc.assertValueInt64(obj)
	}
	// currently float & int64 are supported data types
	return pc.assertValueFloat64(obj)
}

func (pc *PathChecking) assertPath(obj *unstructured.Unstructured) (bool, error) {
	_, found, err := unstructured.NestedFieldNoCopy(
		obj.UnstructuredContent(),
		strings.Split(pc.PathCheck.Path, ".")...,
	)
	if err != nil {
		return false, err
	}
	if pc.retryIfPathNotExists && !found {
		return false, nil
	}
	if pc.retryIfPathExists && found {
		return false, nil
	}
	// returning true will no longer retry
	return true, nil
}

func (pc *PathChecking) assertPathAndValue(context string) (bool, error) {
	err := pc.Retry.Waitf(
		func() (bool, error) {
			client, err := pc.dynamicClientset.
				GetClientForAPIVersionAndKind(
					pc.State.GetAPIVersion(),
					pc.State.GetKind(),
				)
			if err != nil {
				return false, err
			}
			observed, err := client.
				Namespace(pc.State.GetNamespace()).
				Get(
					pc.State.GetName(),
					metav1.GetOptions{},
				)
			if err != nil {
				return false, err
			}
			if pc.pathOnlyCheck {
				return pc.assertPath(observed)
			}
			return pc.assertValue(observed)
		},
		context,
	)
	return err == nil, err
}

func (pc *PathChecking) assertPathExists() (success bool, err error) {
	var message = fmt.Sprintf(
		"PathCheckExists: Resource %s %s: GVK %s: TaskName %s",
		pc.State.GetNamespace(),
		pc.State.GetName(),
		pc.State.GroupVersionKind(),
		pc.TaskName,
	)
	pc.result.Message = message
	// We want to retry if path does not exist in observed state.
	// This is done with the expectation of eventually having an
	// observed state with expected path
	pc.retryIfPathNotExists = true
	return pc.assertPathAndValue(message)
}

func (pc *PathChecking) assertPathNotExists() (success bool, err error) {
	var message = fmt.Sprintf(
		"PathCheckNotExists: Resource %s %s: GVK %s: TaskName %s",
		pc.State.GetNamespace(),
		pc.State.GetName(),
		pc.State.GroupVersionKind(),
		pc.TaskName,
	)
	pc.result.Message = message
	// We want to retry if path does not exist in observed state.
	// This is done with the expectation of eventually having an
	// observed state with expected path
	pc.retryIfPathExists = true
	return pc.assertPathAndValue(message)
}

func (pc *PathChecking) assertPathValueNotEquals() (success bool, err error) {
	var message = fmt.Sprintf(
		"PathCheckValueNotEquals: Resource %s %s: GVK %s: TaskName %s",
		pc.State.GetNamespace(),
		pc.State.GetName(),
		pc.State.GroupVersionKind(),
		pc.TaskName,
	)
	pc.result.Message = message
	// We want to retry if path values of expected & observed
	// match. This is done with the expectation of having
	// observed value not equal to the expected value eventually.
	pc.retryIfValueEquals = true
	return pc.assertPathAndValue(message)
}

func (pc *PathChecking) assertPathValueEquals() (success bool, err error) {
	var message = fmt.Sprintf(
		"PathCheckValueEquals: Resource %s %s: GVK %s: TaskName %s",
		pc.State.GetNamespace(),
		pc.State.GetName(),
		pc.State.GroupVersionKind(),
		pc.TaskName,
	)
	pc.result.Message = message
	// We want to retry if path values of expected & observed
	// does not match. This is done with the expectation of having
	// observed value equal to the expected value eventually.
	pc.retryIfValueNotEquals = true
	return pc.assertPathAndValue(message)
}

func (pc *PathChecking) assertPathValueGTE() (success bool, err error) {
	var message = fmt.Sprintf(
		"PathCheckValueGTE: Resource %s %s: GVK %s: TaskName %s",
		pc.State.GetNamespace(),
		pc.State.GetName(),
		pc.State.GroupVersionKind(),
		pc.TaskName,
	)
	pc.result.Message = message
	// We want to retry if path values of expected & observed
	// does not match. This is done with the expectation of having
	// observed value equal to the expected value eventually.
	pc.retryIfValueNotGTE = true
	return pc.assertPathAndValue(message)
}

func (pc *PathChecking) assertPathValueLTE() (success bool, err error) {
	var message = fmt.Sprintf(
		"PathCheckValueLTE: Resource %s %s: GVK %s: TaskName %s",
		pc.State.GetNamespace(),
		pc.State.GetName(),
		pc.State.GroupVersionKind(),
		pc.TaskName,
	)
	pc.result.Message = message
	// We want to retry if path values of expected & observed
	// does not match. This is done with the expectation of having
	// observed value equal to the expected value eventually.
	pc.retryIfValueNotLTE = true
	return pc.assertPathAndValue(message)
}

func (pc *PathChecking) postAssert(success bool, err error) {
	if err != nil {
		if _, ok := err.(*RetryTimeout); !ok {
			pc.err = err
			return
		}
		pc.result.Timeout = err.Error()
	}
	// initialise phase to failed
	pc.result.Phase = types.PathCheckResultFailed
	if success {
		pc.result.Phase = types.PathCheckResultPassed
	}
}

func (pc *PathChecking) assert() {
	switch pc.operator {
	case types.PathCheckOperatorExists:
		pc.postAssert(pc.assertPathExists())

	case types.PathCheckOperatorNotExists:
		pc.postAssert(pc.assertPathNotExists())

	case types.PathCheckOperatorEquals:
		pc.postAssert(pc.assertPathValueEquals())

	case types.PathCheckOperatorNotEquals:
		pc.postAssert(pc.assertPathValueNotEquals())

	case types.PathCheckOperatorGTE:
		pc.postAssert(pc.assertPathValueGTE())

	case types.PathCheckOperatorLTE:
		pc.postAssert(pc.assertPathValueLTE())

	default:
		pc.err = errors.Errorf(
			"PathCheck %q failed: Invalid operator %q",
			pc.TaskName,
			pc.operator,
		)
	}
}

// Run executes the assertion
func (pc *PathChecking) Run() (types.PathCheckResult, error) {
	var fns = []func(){
		pc.init,
		pc.validate,
		pc.assert,
	}
	for _, fn := range fns {
		fn()
		if pc.err != nil {
			return types.PathCheckResult{}, errors.Wrapf(
				pc.err,
				"Info %s: Warn %s: Verbose %s",
				pc.result.Message,
				pc.result.Warning,
				pc.result.Verbose,
			)
		}
	}
	return *pc.result, nil
}
