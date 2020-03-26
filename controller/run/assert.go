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
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	types "mayadata.io/d-operators/types/run"

	"openebs.io/metac/controller/common/selector"
)

// ResourceListCondition enables filtering, matching against
// a list of resources by runnings these resources against
// one condition
type ResourceListCondition struct {
	Items     []*unstructured.Unstructured
	Condition *types.Condition

	successfulMatches map[*unstructured.Unstructured]bool
	isSuccess         bool
	err               error
}

// NewResourceListCondition returns a new instance of ResourceCondition
// from the provided condition _(read resource selectors)_ & resources
func NewResourceListCondition(
	cond types.Condition,
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
		Condition: &types.Condition{
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

// verify if the provided resource matches the condition
func (c *ResourceListCondition) tryMatchAndRegisterFor(
	resource *unstructured.Unstructured,
) {
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
		c.tryMatchAndRegisterFor(resource)
		if c.err != nil {
			return false, c.err
		}
		if isOperatorExists && len(c.successfulMatches) > 0 {
			// any resource match is a success
			return true, nil
		}
	}
	successfulCount := len(c.successfulMatches)
	if isOperatorNotExist && successfulCount == 0 {
		// success if there are no matches
		c.isSuccess = true
	} else if isOperatorEqualsCount && successfulCount == *c.Condition.Count {
		// success if count matches the selected resources
		c.isSuccess = true
	} else if isOperatorLTE && successfulCount <= *c.Condition.Count {
		// success if count is less than or equal to selected resource count
		c.isSuccess = true
	} else if isOperatorGTE && successfulCount >= *c.Condition.Count {
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

	successfulConditions int
	isSuccess            bool
	err                  error
}

func (a *Assertion) runAllConditions() {
	isOperatorOR :=
		a.Request.Assert.AssertOperator == types.AssertOperatorOR
	isOperatorAND :=
		a.Request.Assert.AssertOperator == types.AssertOperatorAND
	var atleastOneSuccess bool
	// run each condition against all the resources
	for _, cond := range a.Request.Assert.Conditions {
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

// Assert verifies this assert condition
func (a *Assertion) Assert() (bool, error) {
	if a.Request.Assert == nil || len(a.Request.Assert.Conditions) == 0 {
		// nothing needs to be done
		// return successful assert
		return true, nil
	}
	if len(a.Request.Resources) == 0 {
		// raise error if there were conditions without
		// any resources since these conditions need to
		// be executed against resources
		return false, errors.Errorf(
			"Can't assert: No resources provided",
		)
	}
	// run all the conditions specified in this assertion
	a.runAllConditions()
	return a.isSuccess, a.err
}

// ExecuteAssert executes the assert based on the provided request
func ExecuteAssert(req AssertRequest) (bool, error) {
	if req.Assert == nil {
		return false, errors.Errorf(
			"Can't assert: Nil assert",
		)
	}
	var op = req.Assert.AssertOperator
	if op == "" {
		// OR is the default AssertOperator
		op = types.AssertOperatorOR
	}
	var newreq = AssertRequest{
		Assert: &types.Assert{
			AssertOperator: op,
			Conditions:     req.Assert.Conditions,
		},
		Resources: req.Resources,
	}
	a := &Assertion{
		Request: newreq,
	}
	return a.Assert()
}
