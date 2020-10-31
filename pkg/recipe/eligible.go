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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
	metac "openebs.io/metac/apis/metacontroller/v1alpha1"
	"openebs.io/metac/controller/common/selector"

	"mayadata.io/d-operators/pkg/kubernetes"
	typesgvk "mayadata.io/d-operators/types/gvk"
	types "mayadata.io/d-operators/types/recipe"
)

// EligibleItemKey defines a key based on APIVersion & Kind
type EligibleItemKey struct {
	ID         string
	APIVersion string
	Kind       string
}

// NewEligibleItemKey returns a new APIVersionKindKey
func NewEligibleItemKey(item types.EligibleItem) EligibleItemKey {
	return EligibleItemKey{
		ID:         item.ID,
		APIVersion: item.APIVersion,
		Kind:       item.Kind,
	}
}

// NewDefaultingEligibleItemKey returns a new APIVersionKindKey that
// defaults to Recipe's types if kind or apiversion is not specified
func NewDefaultingEligibleItemKey(item types.EligibleItem) EligibleItemKey {
	var apiversion = item.APIVersion
	var kind = item.Kind
	if apiversion == "" {
		apiversion = typesgvk.APIVersionRecipe
	}
	if kind == "" {
		kind = typesgvk.KindRecipe
	}
	return EligibleItemKey{
		ID:         item.ID,
		APIVersion: apiversion,
		Kind:       kind,
	}
}

// String implements Stringer interface
func (k EligibleItemKey) String() string {
	var contents = []string{k.APIVersion, k.Kind}
	if k.ID != "" {
		contents = append(contents, k.ID)
	}
	return strings.Join(contents, "-")
}

// EligibilityConfig helps in constructing new instance of
// Eligibility
type EligibilityConfig struct {
	RecipeName string
	Fixture    *Fixture
	Eligible   *types.Eligible
	// Retry      *Retryable
	Retry *kubernetes.Retryable
}

// Eligibility flags if a Recipe should get executed or not
type Eligibility struct {
	*Fixture
	Eligible *types.Eligible

	// Retry      *Retryable
	Retry *kubernetes.Retryable

	RecipeName string

	eligibles   map[string]types.EligibleItem
	observed    map[string][]unstructured.Unstructured
	selected    map[string][]NamespaceName
	grants      map[string]bool
	evalMessage string

	// error as value
	err error
}

func (e *Eligibility) initAndValidate() error {
	var duplicates = map[string]bool{}
	var errs []string

	if e.Eligible == nil {
		// nothing to do
		return nil
	}
	if len(e.Eligible.Checks) == 0 && e.Eligible.When != "" {
		return errors.Errorf(
			"Invalid Eligible check: When can not be set with nil checks: %s",
			e.RecipeName,
		)
	}

	for _, eligibleitem := range e.Eligible.Checks {
		when := eligibleitem.When
		if when == "" {
			// defaults to exists
			when = types.EligibleItemRuleExists
		}
		switch when {
		case types.EligibleItemRuleExists,
			types.EligibleItemRuleNotFound:
			// supported
		case types.EligibleItemRuleListCountEquals,
			types.EligibleItemRuleListCountNotEquals,
			types.EligibleItemRuleListCountGTE,
			types.EligibleItemRuleListCountLTE:
			// supported
			if eligibleitem.Count == nil {
				errs = append(
					errs,
					fmt.Sprintf(
						"Invalid Eligible check %q: Missing count: RecipeName %q",
						when,
						e.RecipeName,
					),
				)
				continue
			}
		default:
			errs = append(
				errs,
				fmt.Sprintf(
					"Unsupported Eligible check %q: RecipeName %q",
					when,
					e.RecipeName,
				),
			)
			continue
		}
		key := NewDefaultingEligibleItemKey(eligibleitem)
		if duplicates[key.String()] {
			errs = append(
				errs,
				fmt.Sprintf(
					"Duplicate Eligible check %q: RecipeName %q",
					when,
					e.RecipeName,
				),
			)
			continue
		}
		duplicates[key.String()] = true
		e.eligibles[key.String()] = eligibleitem
	}
	if len(errs) != 0 {
		return errors.Errorf(
			"%d error(s) found: %s",
			len(errs),
			strings.Join(errs, ": "),
		)
	}
	return nil
}

// NewEligibility returns a new instance of Eligibility
func NewEligibility(config EligibilityConfig) (*Eligibility, error) {
	e := &Eligibility{
		RecipeName: config.RecipeName,
		Fixture:    config.Fixture,
		Eligible:   config.Eligible,
		Retry:      config.Retry,
		eligibles:  make(map[string]types.EligibleItem),
		observed:   make(map[string][]unstructured.Unstructured),
		selected:   make(map[string][]NamespaceName),
		grants:     make(map[string]bool),
	}
	err := e.initAndValidate()
	if err != nil {
		return nil, err
	}
	return e, nil
}

// EligibilityLog helps in debugging
type EligibilityLog struct {
	RecipeName        string                        `json:"recipeName"`
	IsEligible        bool                          `json:"isEligible"`
	Attempts          int                           `json:"attempts"`
	IsTimeout         bool                          `json:"isTimeout"`
	EligibileCriteria map[string]types.EligibleItem `json:"eligibileCriteria"`
	Observed          map[string][]NamespaceName    `json:"observed"`
	Selected          map[string][]NamespaceName    `json:"selected"`
	Grants            map[string]bool               `json:"grants"`
	EvalMessage       string                        `json:"evalMessage"`
	Error             string                        `json:"error"`
}

func (e *Eligibility) isEligibleCheck() (result bool) {
	defer func() {
		klog.V(3).Infof("IsEligible=%t: %s", result, e.RecipeName)
	}()
	if len(e.grants) == 0 {
		klog.V(2).Infof(
			"Failed to evaluate: No grants found: %s",
			e.RecipeName,
		)
		return
	}

	when := e.Eligible.When
	switch when {
	case types.EligibleRuleAnyCheckPass:
		for _, ok := range e.grants {
			if ok {
				// at-least one grant should pass to be eligible
				result = true
				return
			}
		}
		return
	default:
		for _, ok := range e.grants {
			if !ok {
				// all grants should pass to be eligible
				return
			}
		}
		result = true
		return
	}
}

func (e *Eligibility) setGrants() error {
	if len(e.selected) == 0 {
		klog.V(1).Infof(
			"Nothing to grant: No objects selected",
		)
		return nil
	}
	for key, selectedObjs := range e.selected {
		// pick the eligibility criteria for this key
		eligiblecriteria := e.eligibles[key]

		// default the condition if not set
		var when = eligiblecriteria.When
		var actualCount = len(selectedObjs)
		var expectedCount int
		var msgs = []string{
			fmt.Sprintf("Key %s", key),
		}
		if when == "" {
			when = types.EligibleItemRuleExists
		}
		msgs = append(
			msgs,
			fmt.Sprintf("When %s", when),
			fmt.Sprintf("Actual count %d", actualCount),
		)
		if eligiblecriteria.Count != nil {
			expectedCount = *eligiblecriteria.Count
			msgs = append(
				msgs,
				fmt.Sprintf("Expected count %d", expectedCount),
			)
		}
		// build the eligibility evaluation message
		e.evalMessage = strings.Join(msgs, ": ")

		switch when {
		case types.EligibleItemRuleExists:
			e.grants[key] = actualCount > 0
		case types.EligibleItemRuleNotFound:
			e.grants[key] = actualCount == 0
		case types.EligibleItemRuleListCountEquals:
			e.grants[key] = actualCount == expectedCount
		case types.EligibleItemRuleListCountNotEquals:
			e.grants[key] = actualCount != expectedCount
		case types.EligibleItemRuleListCountGTE:
			e.grants[key] = actualCount >= expectedCount
		case types.EligibleItemRuleListCountLTE:
			e.grants[key] = actualCount <= expectedCount
		default:
			return errors.Errorf(
				"Invalid eligible criteria: Unsupported when %q",
				when,
			)
		}
	}
	return nil
}

func (e *Eligibility) isResourceLabelMatch(
	target unstructured.Unstructured,
	lblselector v1.LabelSelector,
) (bool, error) {
	eval := selector.Evaluation{
		Target: &target,
		Terms: []*metac.SelectorTerm{
			{
				MatchLabels:           lblselector.MatchLabels,
				MatchLabelExpressions: lblselector.MatchExpressions,
			},
		},
	}
	return eval.RunMatch()
}

func (e *Eligibility) setSelectedFromObservedResources() error {
	for key, observedInstances := range e.observed {
		eligibleitem := e.eligibles[key]
		if len(eligibleitem.LabelSelector.MatchExpressions) == 0 &&
			len(eligibleitem.LabelSelector.MatchLabels) == 0 {
			// if there are no selections then all observed instance
			// are considered selected
			e.selected[key] = NewNamespaceNameList(observedInstances)
			continue
		}
		// initialise this key with empty list
		//
		// NOTE:
		// 	It is important to initialise the key to properly evaluate
		// the final eligibility result
		e.selected[key] = []NamespaceName{}
		for _, instance := range observedInstances {
			if instance.Object == nil {
				continue
			}
			ismatch, err := e.isResourceLabelMatch(
				instance,
				eligibleitem.LabelSelector,
			)
			if err != nil {
				return err
			}
			if ismatch {
				e.selected[key] = append(
					e.selected[key],
					NewNamespaceName(instance),
				)
			}
		}
	}
	return nil
}

func (e *Eligibility) setObservedResources() error {
	for _, eligibleitem := range e.Eligible.Checks {
		key := NewDefaultingEligibleItemKey(eligibleitem)
		client, err := e.dynamicClientset.GetClientForAPIVersionAndKind(
			key.APIVersion,
			key.Kind,
		)
		if err != nil {
			return &DiscoveryError{err.Error()}
		}
		list, err := client.List(v1.ListOptions{})
		if err != nil {
			return err
		}
		e.observed[key.String()] = list.Items // Items might be nil
	}
	return nil
}

// IsEligible returns true if Recipe is eligible to be executed
func (e *Eligibility) IsEligible() (ok bool, err error) {
	if e.Eligible == nil || len(e.Eligible.Checks) == 0 {
		// nothing to check
		return true, nil
	}

	var attempts int
	var timeout bool
	var errmsg string
	defer func() {
		var islog = false
		if err != nil {
			// we should log in-case of any error
			islog = true
			errmsg = err.Error()
		}
		if !klog.V(3).Enabled() || islog {
			return
		}
		logfn := func() string {
			log := EligibilityLog{
				RecipeName:        e.RecipeName,
				IsEligible:        ok,
				Attempts:          attempts,
				IsTimeout:         timeout,
				EligibileCriteria: e.eligibles,
				Observed:          ResourceMappedNamespaceNames(e.observed),
				Selected:          e.selected,
				Grants:            e.grants,
				EvalMessage:       e.evalMessage,
				Error:             errmsg,
			}
			raw, _ := json.MarshalIndent(log, "", ".")
			return string(raw)
		}
		klog.V(1).Infof("Eligibility evaluation:-\n%s", logfn())
	}()

	err = e.Retry.Waitf(
		func() (bool, error) {
			attempts++

			var fns = []func() error{
				e.setObservedResources,
				e.setSelectedFromObservedResources,
				e.setGrants,
			}
			for _, fn := range fns {
				err = fn()
				if err != nil {
					// Keep retrying
					return false, err
				}
			}
			ok = e.isEligibleCheck()
			return ok, nil
		},
		"IsEnable check: RecipeName %s",
		e.RecipeName,
	)
	if err != nil {
		if _, timeout = err.(*RetryTimeout); timeout {
			// in case of timeout we swallow the error &
			// return false i.e. mark the recipe as not eligible
			errmsg = err.Error()
			// eligibility check has failed after several retries
			return false, nil
		}
	}
	return ok, err
}
