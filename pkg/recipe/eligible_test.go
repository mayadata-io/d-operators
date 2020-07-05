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
	"testing"

	"mayadata.io/d-operators/common/pointer"
	types "mayadata.io/d-operators/types/recipe"
)

func TestEligibilityInitAndValidate(t *testing.T) {
	var tests = map[string]struct {
		eligible *types.Eligible
		isErr    bool
	}{
		"no error": {
			isErr: false,
		},
		"nil eligible checks": {
			eligible: &types.Eligible{
				Checks: nil,
			},
			isErr: false,
		},
		"empty eligible checks": {
			eligible: &types.Eligible{
				Checks: []types.EligibleItem{},
			},
			isErr: false,
		},
		"one nil eligible check item": {
			eligible: &types.Eligible{
				Checks: []types.EligibleItem{
					{},
				},
			},
			isErr: false,
		},
		"invalid eligible check item": {
			eligible: &types.Eligible{
				Checks: []types.EligibleItem{
					{
						Count: nil,
						When:  types.EligibleItemRuleListCountGTE,
					},
				},
			},
			isErr: true,
		},
		"valid eligible check item": {
			eligible: &types.Eligible{
				Checks: []types.EligibleItem{
					{
						Count: pointer.Int(2),
						When:  types.EligibleItemRuleListCountGTE,
					},
				},
			},
			isErr: false,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			e := &Eligibility{
				Eligible:  mock.eligible,
				eligibles: make(map[string]types.EligibleItem),
			}
			err := e.initAndValidate()
			if mock.isErr && err == nil {
				t.Fatal("Expected error got none")
			}
			if !mock.isErr && err != nil {
				t.Fatalf("Expected no error got %s", err.Error())
			}
		})
	}
}

func TestEligibilityIsEligibleCheck(t *testing.T) {
	var tests = map[string]struct {
		grants     map[string]bool
		when       types.EligibleRule
		isEligible bool
	}{
		"nil grants": {
			isEligible: false,
		},
		"empty grants": {
			grants:     make(map[string]bool),
			isEligible: false,
		},
		"successful grant with all checks pass": {
			grants: map[string]bool{
				"v1-Pod":     true,
				"v2-Service": true,
			},
			when:       types.EligibleRuleAllChecksPass,
			isEligible: true,
		},
		"successful grant with all checks pass - 2": {
			grants: map[string]bool{
				"v1-Pod":     true,
				"v2-Service": true,
			},
			when:       types.EligibleRuleAnyCheckPass,
			isEligible: true,
		},
		"un-successful grant with some checks pass": {
			grants: map[string]bool{
				"v1-Pod":     true,
				"v2-Service": false,
			},
			when:       types.EligibleRuleAllChecksPass,
			isEligible: false,
		},
		"successful grant with some checks pass": {
			grants: map[string]bool{
				"v1-Pod":     true,
				"v2-Service": false,
			},
			when:       types.EligibleRuleAnyCheckPass,
			isEligible: true,
		},
		"un-successful grant with all checks fail": {
			grants: map[string]bool{
				"v1-Pod":     false,
				"v2-Service": false,
			},
			when:       types.EligibleRuleAllChecksPass,
			isEligible: false,
		},
		"un-successful grant with all checks fail - 2": {
			grants: map[string]bool{
				"v1-Pod":     false,
				"v2-Service": false,
			},
			when:       types.EligibleRuleAnyCheckPass,
			isEligible: false,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			e := &Eligibility{
				grants: mock.grants,
				Eligible: &types.Eligible{
					When: mock.when,
				},
			}
			got := e.isEligibleCheck()
			if got != mock.isEligible {
				t.Fatalf("Expected %t got %t", mock.isEligible, got)
			}
		})
	}
}

func TestEligibilitySetGrants(t *testing.T) {
	var tests = map[string]struct {
		selected       map[string][]NamespaceName
		eligibles      map[string]types.EligibleItem
		expectedGrants map[string]bool
		isErr          bool
	}{
		"nil selected": {},
		"empty selected": {
			selected: make(map[string][]NamespaceName),
		},
		"1 selected resource with 1 instance: criteria not provided": {
			selected: map[string][]NamespaceName{
				"v1-Pod": {
					NamespaceName{
						Name:      "my-pod",
						Namespace: "default",
					},
				},
			},
			eligibles: map[string]types.EligibleItem{
				"v1-Pod": {},
			},
			expectedGrants: map[string]bool{
				"v1-Pod": true,
			},
			isErr: false,
		},
		"unsupported criteria": {
			selected: map[string][]NamespaceName{
				"v1-Pod": {
					NamespaceName{
						Name:      "my-pod",
						Namespace: "default",
					},
				},
			},
			eligibles: map[string]types.EligibleItem{
				"v1-Pod": {
					When: types.EligibleItemRule("JUNK"),
				},
			},
			isErr: true,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			e := &Eligibility{
				selected:  mock.selected,
				eligibles: mock.eligibles,
				grants:    make(map[string]bool),
			}
			err := e.setGrants()
			if mock.isErr && err == nil {
				t.Fatal("Expected error got none")
			}
			if !mock.isErr && err != nil {
				t.Fatalf("Expected no error got %s", err.Error())
			}
			if len(e.grants) != len(mock.expectedGrants) {
				t.Fatalf(
					"Expected grant count %d got %d",
					len(mock.expectedGrants),
					len(e.grants),
				)
			}
		})
	}
}
