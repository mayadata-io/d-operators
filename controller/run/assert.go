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

package run

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"mayadata.io/d-operators/common/labels"
	"mayadata.io/d-operators/common/pointer"
	types "mayadata.io/d-operators/types/run"

	"openebs.io/metac/controller/common/selector"
	"openebs.io/metac/dynamic/apply"
)

// ResourceListConditionConfig holds the info required to
// create a ResourceListCondition instance
type ResourceListConditionConfig struct {
	IncludeInfo map[types.IncludeInfoKey]bool
	Condition   types.IfCondition
	Resources   []*unstructured.Unstructured
}

// ResourceListCondition enables filtering, matching against
// a list of resources by runnings these resources against
// one condition
type ResourceListCondition struct {
	Items []*unstructured.Unstructured

	IncludeInfo map[types.IncludeInfoKey]bool
	Condition   *types.IfCondition

	// matches           []string
	// nomatches         []string
	successfulMatches map[*unstructured.Unstructured]bool
	successCount      int
	Result            *types.Result

	isSuccess bool
	err       error
}

// NewResourceListCondition returns a new instance of ResourceCondition
// from the provided condition _(read resource selectors)_ & resources
func NewResourceListCondition(config ResourceListConditionConfig) *ResourceListCondition {
	if len(config.Condition.ResourceSelector.SelectorTerms) == 0 {
		return &ResourceListCondition{
			err: errors.Errorf(
				"Invalid resource condition: Empty selector terms",
			),
		}
	}
	if config.Condition.Count == nil {
		if config.Condition.ResourceOperator == types.ResourceOperatorEqualsCount ||
			config.Condition.ResourceOperator == types.ResourceOperatorGTE ||
			config.Condition.ResourceOperator == types.ResourceOperatorLTE {
			return &ResourceListCondition{
				err: errors.Errorf(
					"Invalid resource condition: Count must be set when operator is %q",
					config.Condition.ResourceOperator,
				),
			}
		}
	}
	if len(config.Resources) == 0 {
		return &ResourceListCondition{
			err: errors.Errorf(
				"Invalid resource condition: No resources provided",
			),
		}
	}
	rc := &ResourceListCondition{
		IncludeInfo: config.IncludeInfo,
		Condition: &types.IfCondition{
			ResourceSelector: config.Condition.ResourceSelector,
			ResourceOperator: config.Condition.ResourceOperator,
			Count:            config.Condition.Count,
		},
		Items:             config.Resources,
		successfulMatches: make(map[*unstructured.Unstructured]bool),
		Result:            &types.Result{},
	}
	// set default(s)
	if rc.Condition.ResourceOperator == "" {
		// Exists is the default operator
		rc.Condition.ResourceOperator = types.ResourceOperatorExists
	}
	return rc
}

func (c *ResourceListCondition) includeMatchInfoIfEnabled(message ...string) {
	if c.IncludeInfo == nil {
		return
	}
	if !c.IncludeInfo[types.IncludeDesiredInfo] &&
		!c.IncludeInfo[types.IncludeAllInfo] {
		return
	}
	c.Result.DesiredResourcesInfo = append(
		c.Result.DesiredResourcesInfo,
		message...,
	)
}

func (c *ResourceListCondition) includeNoMatchInfoIfEnabled(message ...string) {
	if c.IncludeInfo == nil {
		return
	}
	if !c.IncludeInfo[types.IncludeSkippedInfo] &&
		!c.IncludeInfo[types.IncludeAllInfo] {
		return
	}
	c.Result.SkippedResourcesInfo = append(
		c.Result.SkippedResourcesInfo,
		message...,
	)
}

// verify if condition matches the provided resource matches the condition
func (c *ResourceListCondition) runMatchFor(resource *unstructured.Unstructured) {
	e := selector.Evaluation{
		Terms:  c.Condition.ResourceSelector.SelectorTerms,
		Target: resource,
	}
	isSuccess, err := e.RunMatch()
	if err != nil {
		c.err = err
		return
	}
	if isSuccess {
		c.successfulMatches[resource] = true
		c.successCount++
		c.includeMatchInfoIfEnabled(
			fmt.Sprintf(
				"Assert conditions matched for %q / %q: %s",
				resource.GetNamespace(),
				resource.GetName(),
				resource.GetObjectKind().GroupVersionKind().String(),
			),
		)
	} else {
		c.includeNoMatchInfoIfEnabled(
			fmt.Sprintf(
				"Assert conditions failed for %q / %q: %s",
				resource.GetNamespace(),
				resource.GetName(),
				resource.GetObjectKind().GroupVersionKind().String(),
			),
		)
	}
}

// IsSuccess returns true if condition matches its resources
func (c *ResourceListCondition) IsSuccess() (bool, error) {
	if c.err != nil {
		return false, c.err
	}
	isOperatorExists :=
		c.Condition.ResourceOperator == types.ResourceOperatorExists
	isOperatorNotExist :=
		c.Condition.ResourceOperator == types.ResourceOperatorNotExist
	isOperatorGTE :=
		c.Condition.ResourceOperator == types.ResourceOperatorGTE
	isOperatorLTE :=
		c.Condition.ResourceOperator == types.ResourceOperatorLTE
	isOperatorEqualsCount :=
		c.Condition.ResourceOperator == types.ResourceOperatorEqualsCount
	for _, resource := range c.Items {
		if resource == nil || resource.Object == nil {
			return false, errors.Errorf(
				"Can't match resource condition: Nil resource found",
			)
		}
		c.runMatchFor(resource)
		if c.err != nil {
			return false, c.err
		}
		if isOperatorExists && c.successCount > 0 {
			// any resource match is a success
			return true, nil
		}
	}
	if isOperatorNotExist && c.successCount == 0 {
		// success if there are no matches
		c.isSuccess = true
	} else if isOperatorEqualsCount && c.successCount == *c.Condition.Count {
		// success if count matches the selected resources
		c.isSuccess = true
	} else if isOperatorLTE && c.successCount <= *c.Condition.Count {
		// success if count is less than or equal to selected resource count
		c.isSuccess = true
	} else if isOperatorGTE && c.successCount >= *c.Condition.Count {
		// success if count is greater than or equal to selected resource count
		c.isSuccess = true
	}
	return c.isSuccess, nil
}

// AssertRequest forms the input required to execute an
// assertion
type AssertRequest struct {
	IncludeInfo map[types.IncludeInfoKey]bool
	TaskKey     string
	Assert      *types.Assert
	Resources   []*unstructured.Unstructured
}

// AssertResponse holds the response after executing an
// assert
type AssertResponse struct {
	AssertResult *types.Result
}

// Assertion asserts by running the conditions against the
// resources
type Assertion struct {
	Request AssertRequest
	Result  *types.Result

	matches   []string
	nomatches []string

	isSuccess bool
	err       error
}

func (a *Assertion) includeMatchInfoIfEnabled(message ...string) {
	if a.Request.IncludeInfo == nil {
		return
	}
	if !a.Request.IncludeInfo[types.IncludeDesiredInfo] &&
		!a.Request.IncludeInfo[types.IncludeAllInfo] {
		return
	}
	a.Result.DesiredResourcesInfo = append(
		a.Result.DesiredResourcesInfo,
		message...,
	)
}

func (a *Assertion) includeNoMatchInfoIfEnabled(message ...string) {
	if a.Request.IncludeInfo == nil {
		return
	}
	if !a.Request.IncludeInfo[types.IncludeSkippedInfo] &&
		!a.Request.IncludeInfo[types.IncludeAllInfo] {
		return
	}
	a.Result.SkippedResourcesInfo = append(
		a.Result.SkippedResourcesInfo,
		message...,
	)
}

func (a *Assertion) includeWarningIfEnabled(message ...string) {
	if a.Request.IncludeInfo == nil {
		return
	}
	if !a.Request.IncludeInfo[types.IncludeSkippedInfo] &&
		!a.Request.IncludeInfo[types.IncludeAllInfo] {
		return
	}
	a.Result.Warns = append(
		a.Result.Warns,
		message...,
	)
}

func (a *Assertion) verifyAllConditions() {
	// flag OR operator if all conditions need to OR-ed
	isOperatorOR :=
		a.Request.Assert.IfOperator == types.IfOperatorOR
	// flag AND operator if all conditions need to be AND-ed
	isOperatorAND :=
		a.Request.Assert.IfOperator == types.IfOperatorAND
	var atleastOneSuccess bool
	// run all conditions against all the available resources
	for _, cond := range a.Request.Assert.IfConditions {
		// check current condition against all the resources
		listCond := NewResourceListCondition(
			ResourceListConditionConfig{
				IncludeInfo: a.Request.IncludeInfo,
				Condition:   cond,
				Resources:   a.Request.Resources,
			},
		)
		success, err := listCond.IsSuccess()
		if err != nil {
			a.err = err
			return
		}
		// add matching conditions to info
		a.includeMatchInfoIfEnabled(listCond.Result.DesiredResourcesInfo...)
		// add non-matching conditions to info
		a.includeNoMatchInfoIfEnabled(listCond.Result.SkippedResourcesInfo...)
		if success && !atleastOneSuccess {
			atleastOneSuccess = true
		}
		if isOperatorOR && success {
			// at-least one success is a complete success
			a.isSuccess = true
			return
		}
		if isOperatorAND && !success {
			// any failure is a complete failure
			a.isSuccess = false
			return
		}
	}
	a.isSuccess = atleastOneSuccess
}

func (a *Assertion) verifyState() {
	// transform state into an unstructured instance
	state := &unstructured.Unstructured{
		Object: a.Request.Assert.State,
	}
	// extract essentials to match provided resources
	// with provided state
	stKind := state.GetKind()
	stAPIVersion := state.GetAPIVersion()
	stName := state.GetName()
	stNamespace := state.GetNamespace()
	stLbls := state.GetLabels()
	stAnns := state.GetAnnotations()
	// assert against all the available resources
	for _, resource := range a.Request.Resources {
		if resource == nil || resource.Object == nil {
			a.err = errors.Errorf(
				"Can't verify state: Nil resource found",
			)
			return
		}
		if resource.GetKind() != stKind || resource.GetAPIVersion() != stAPIVersion {
			// this is not the resource we want to assert against
			continue
		}
		// we do not do a exact name match instead try a prefix match
		if stName != "" && !strings.HasPrefix(resource.GetName(), stName) {
			// this is not the resource we want to assert against
			continue
		}
		if stNamespace != "" && stNamespace != resource.GetNamespace() {
			// this is not the resource we want to assert against
			continue
		}
		if len(stLbls) != 0 && !labels.New(resource.GetLabels()).Has(stLbls) {
			// this is not the resource we want to assert against
			continue
		}
		if len(stAnns) != 0 && !labels.New(resource.GetAnnotations()).Has(stAnns) {
			// this is not the resource we want to assert against
			continue
		}
		// get a deep copy of the resource
		resourceCopy := resource.DeepCopy()
		if stName != "" {
			// override the resource name with the name set in
			// assert state
			//
			// This is done to do support prefix based name match
			resourceCopy.SetName(stName)
		}
		// at this point we want to assert the given state with the
		// current resource by running a 3 way merge & finally matching
		// the resulting merge with the original resource
		final, err := apply.Merge(
			resourceCopy.UnstructuredContent(), // observed = current resource
			state.UnstructuredContent(),        // last applied = given state
			state.UnstructuredContent(),        // desired = given state
		)
		if err != nil {
			a.err = errors.Wrapf(
				err,
				"Failed to assert state",
			)
			return
		}
		if !reflect.DeepEqual(final, resourceCopy.UnstructuredContent()) {
			a.nomatches = append(
				a.nomatches,
				fmt.Sprintf(
					"Assert state didn't match for %q / %q: %s",
					resource.GetNamespace(),
					resource.GetName(), // use original resource name
					resource.GetObjectKind().GroupVersionKind().String(),
				),
			)
		} else {
			a.matches = append(
				a.matches,
				fmt.Sprintf(
					"Assert state matched for %q / %q: %s",
					resource.GetNamespace(),
					resource.GetName(), // use original resource name
					resource.GetObjectKind().GroupVersionKind().String(),
				),
			)
		}
	}
	if len(a.matches) == 0 && len(a.nomatches) == 0 {
		// its a failure if there are no successful matches
		a.nomatches = append(
			a.nomatches,
			fmt.Sprintf(
				"No matches for assert state: Tried against %d resources",
				len(a.Request.Resources),
			),
		)
		a.includeWarningIfEnabled(
			"No matches for given assert: Recheck assert state",
		)
	}
	// add matching asserts to desired info if any
	a.includeMatchInfoIfEnabled(a.matches...)
	// add non-matching asserts to skipped info if any
	a.includeNoMatchInfoIfEnabled(a.nomatches...)
	if len(a.nomatches) == 0 {
		// assert is a success since there were no failed matches
		a.isSuccess = true
	}
}

// AssertAllConditions asserts the provided conditions
func (a *Assertion) AssertAllConditions() (bool, error) {
	// assert all the conditions specified in this assertion
	a.verifyAllConditions()
	return a.isSuccess, a.err
}

// AssertState asserts the provided state
func (a *Assertion) AssertState() (bool, error) {
	// assert the provided state with resources
	a.verifyState()
	return a.isSuccess, a.err
}

// ExecuteAssertConditions asserts based on the provided
// conditions and resources
func ExecuteAssertConditions(req AssertRequest) (*AssertResponse, error) {
	var op = req.Assert.IfOperator
	if op == "" {
		// OR is the default AssertOperator
		op = types.IfOperatorOR
	}
	// a new & updated copy of AssertRequest
	var newreq = AssertRequest{
		Assert: &types.Assert{
			If: types.If{
				IfOperator:   op,
				IfConditions: req.Assert.IfConditions,
			},
		},
		Resources: req.Resources,
	}
	a := &Assertion{
		Request: newreq,
		Result:  &types.Result{},
	}
	ok, err := a.AssertAllConditions()
	if err != nil {
		return nil, err
	}
	if ok {
		return &AssertResponse{
			AssertResult: &types.Result{
				Phase:                types.ResultPhaseAssertPassed, // passed
				DesiredResourcesInfo: a.Result.DesiredResourcesInfo,
				SkippedResourcesInfo: a.Result.SkippedResourcesInfo,
				HasRunOnce:           pointer.Bool(true),
				Warns:                a.Result.Warns,
			},
		}, nil
	}
	return &AssertResponse{
		AssertResult: &types.Result{
			Phase:                types.ResultPhaseAssertFailed, // failed
			DesiredResourcesInfo: a.Result.DesiredResourcesInfo,
			SkippedResourcesInfo: a.Result.SkippedResourcesInfo,
			HasRunOnce:           pointer.Bool(true),
			Warns:                a.Result.Warns,
		},
	}, nil
}

// ExecuteAssertState asserts based on the provided state
func ExecuteAssertState(req AssertRequest) (*AssertResponse, error) {
	a := &Assertion{
		Request: req,
		Result:  &types.Result{},
	}
	ok, err := a.AssertState()
	if err != nil {
		return nil, err
	}
	if ok {
		return &AssertResponse{
			AssertResult: &types.Result{
				Phase:                types.ResultPhaseAssertPassed, // passed
				DesiredResourcesInfo: a.Result.DesiredResourcesInfo,
				SkippedResourcesInfo: a.Result.SkippedResourcesInfo,
				Warns:                a.Result.Warns,
				HasRunOnce:           pointer.Bool(true),
			},
		}, nil
	}
	return &AssertResponse{
		AssertResult: &types.Result{
			Phase:                types.ResultPhaseAssertFailed, // failed
			DesiredResourcesInfo: a.Result.DesiredResourcesInfo,
			SkippedResourcesInfo: a.Result.SkippedResourcesInfo,
			Warns:                a.Result.Warns,
			HasRunOnce:           pointer.Bool(true),
		},
	}, nil
}

// ExecuteCondition executes the assert based on the provided request
func ExecuteCondition(req AssertRequest) (*AssertResponse, error) {
	if req.TaskKey == "" {
		return nil, errors.Errorf(
			"Can't assert: Missing task key",
		)
	}
	if req.Assert == nil {
		return nil, errors.Errorf(
			"Can't assert: Missing assert specs: %s",
			req.TaskKey,
		)
	}
	if len(req.Assert.State) != 0 && len(req.Assert.IfConditions) != 0 {
		return nil, errors.Errorf(
			"Can't assert: Both assert state & conditions can't be used together: %s",
			req.TaskKey,
		)
	}
	if len(req.Assert.State) == 0 && len(req.Assert.IfConditions) == 0 {
		return nil, errors.Errorf(
			"Can't assert: Either assert state or conditions need to be set: %s",
			req.TaskKey,
		)
	}
	if len(req.Resources) == 0 {
		// raise error if there were conditions without
		// any resources since these conditions need to
		// be executed against resources
		return nil, errors.Errorf(
			"Can't assert: No resources provided: %s",
			req.TaskKey,
		)
	}
	// assertion can either be executed against the provided:
	// 1/ state, or
	// 2/ conditions
	if len(req.Assert.State) != 0 {
		return ExecuteAssertState(req)
	}
	return ExecuteAssertConditions(req)
}
