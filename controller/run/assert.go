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

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"mayadata.io/d-operators/common/labels"
	types "mayadata.io/d-operators/types/run"

	"openebs.io/metac/controller/common/selector"
	"openebs.io/metac/dynamic/apply"
)

// ResourceListCondition enables filtering, matching against
// a list of resources by runnings these resources against
// one condition
type ResourceListCondition struct {
	Items     []*unstructured.Unstructured
	Condition *types.IfCondition

	matches           []string
	nomatches         []string
	successfulMatches map[*unstructured.Unstructured]bool
	successCount      int

	isSuccess bool
	err       error
}

// NewResourceListCondition returns a new instance of ResourceCondition
// from the provided condition _(read resource selectors)_ & resources
func NewResourceListCondition(
	cond types.IfCondition,
	resources []*unstructured.Unstructured,
) *ResourceListCondition {
	if len(cond.ResourceSelector.SelectorTerms) == 0 {
		return &ResourceListCondition{
			err: errors.Errorf(
				"Invalid resource condition: Empty selector terms",
			),
		}
	}
	if cond.Count == nil {
		if cond.ResourceOperator == types.ResourceOperatorEqualsCount ||
			cond.ResourceOperator == types.ResourceOperatorGTE ||
			cond.ResourceOperator == types.ResourceOperatorLTE {
			return &ResourceListCondition{
				err: errors.Errorf(
					"Invalid resource condition: Count must be set when operator is %q",
					cond.ResourceOperator,
				),
			}
		}
	}
	if len(resources) == 0 {
		return &ResourceListCondition{
			err: errors.Errorf(
				"Invalid resource condition: No resources provided",
			),
		}
	}
	rc := &ResourceListCondition{
		Condition: &types.IfCondition{
			ResourceSelector: cond.ResourceSelector,
			ResourceOperator: cond.ResourceOperator,
			Count:            cond.Count,
		},
		Items:             resources,
		successfulMatches: make(map[*unstructured.Unstructured]bool),
	}
	// set default(s)
	if rc.Condition.ResourceOperator == "" {
		// Exists is the default operator
		rc.Condition.ResourceOperator = types.ResourceOperatorExists
	}
	return rc
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
		c.matches = append(
			c.matches,
			fmt.Sprintf(
				"Assert matched for %q / %q: %s",
				resource.GetNamespace(),
				resource.GetName(),
				resource.GetObjectKind().GroupVersionKind().String(),
			),
		)
	} else {
		c.nomatches = append(
			c.nomatches,
			fmt.Sprintf(
				"Assert failed for %q / %q: %s",
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
	Assert    *types.Assert
	Resources []*unstructured.Unstructured
}

// Assertion asserts by running the conditions against the
// resources
type Assertion struct {
	Request AssertRequest

	matches   []string
	nomatches []string

	isSuccess bool
	err       error
}

func (a *Assertion) verifyAllConditions() {
	// flag OR operator if all conditions need to OR-ed
	isOperatorOR :=
		a.Request.Assert.IfOperator == types.IfOperatorOR
	// flag AND operator if all conditions need to be AND-ed
	isOperatorAND :=
		a.Request.Assert.IfOperator == types.IfOperatorAND
	var atleastOneSuccess bool
	// run each condition against all the available resources
	for _, cond := range a.Request.Assert.IfConditions {
		// check each condition
		success, err := NewResourceListCondition(
			cond,
			a.Request.Resources,
		).IsSuccess()
		if err != nil {
			a.err = err
			return
		}
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
	kind := state.GetKind()
	apiVersion := state.GetAPIVersion()
	name := state.GetName()
	namespace := state.GetNamespace()
	lbls := state.GetLabels()
	anns := state.GetAnnotations()
	// assert against all the available resources
	for _, resource := range a.Request.Resources {
		if resource == nil || resource.Object == nil {
			a.err = errors.Errorf(
				"Can't verify state: Nil resource found",
			)
			return
		}
		if resource.GetKind() != kind || resource.GetAPIVersion() != apiVersion {
			// this is not the resource we want to assert against
			continue
		}
		if name != "" && name != resource.GetName() {
			// this is not the resource we want to assert against
			continue
		}
		if namespace != "" && namespace != resource.GetNamespace() {
			// this is not the resource we want to assert against
			continue
		}
		if len(lbls) != 0 && !labels.New(resource.GetLabels()).Has(lbls) {
			// this is not the resource we want to assert against
			continue
		}
		if len(anns) != 0 && !labels.New(resource.GetAnnotations()).Has(anns) {
			// this is not the resource we want to assert against
			continue
		}
		// at this point we want to assert the given state with the
		// current resource by running a 3 way merge & finally matching
		// the resulting merge with the original resource
		final, err := apply.Merge(
			resource.UnstructuredContent(), // observed
			state.UnstructuredContent(),    // last applied
			state.UnstructuredContent(),    // desired
		)
		if err != nil {
			a.err = errors.Wrapf(
				err,
				"Failed to assert state",
			)
			return
		}
		if !reflect.DeepEqual(final, resource.UnstructuredContent()) {
			a.nomatches = append(
				a.nomatches,
				fmt.Sprintf(
					"Assert failed for %q / %q: %s",
					resource.GetNamespace(),
					resource.GetName(),
					resource.GetObjectKind().GroupVersionKind().String(),
				),
			)
		} else {
			a.matches = append(
				a.matches,
				fmt.Sprintf(
					"Assert matched for %q / %q: %s",
					resource.GetNamespace(),
					resource.GetName(),
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
				"No assert matches found: Tried %d resources: Assert.State may be invalid",
				len(a.Request.Resources),
			),
		)
	}
	if len(a.nomatches) == 0 {
		// assert is a success if there were no failed matches
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
func ExecuteAssertConditions(req AssertRequest) (bool, error) {
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
	}
	return a.AssertAllConditions()
}

// ExecuteAssertState asserts based on the provided state
func ExecuteAssertState(req AssertRequest) (bool, error) {
	a := &Assertion{
		Request: req,
	}
	return a.AssertState()
}

// ExecuteAssert executes the assert based on the provided request
func ExecuteAssert(req AssertRequest) (bool, error) {
	if req.Assert == nil {
		return false, errors.Errorf(
			"Can't assert: Missing assert specs",
		)
	}
	if len(req.Assert.State) != 0 && len(req.Assert.IfConditions) != 0 {
		return false, errors.Errorf(
			"Can't assert: Both assert state & conditions can't be together",
		)
	}
	if len(req.Assert.State) == 0 && len(req.Assert.IfConditions) == 0 {
		return false, errors.Errorf(
			"Can't assert: Either assert state or conditions need to be set",
		)
	}
	if len(req.Resources) == 0 {
		// raise error if there were conditions without
		// any resources since these conditions need to
		// be executed against resources
		return false, errors.Errorf(
			"Can't assert: No resources provided",
		)
	}
	// assertion can be either be done against the provided state
	// or with the provided conditions
	if len(req.Assert.State) != 0 {
		return ExecuteAssertState(req)
	}
	return ExecuteAssertConditions(req)
}
