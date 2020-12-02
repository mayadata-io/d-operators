// +build !integration

/*
Copyright 2020 The MayaData Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package schema

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	stringutil "mayadata.io/d-operators/common/string"
)

func TestValidationError(t *testing.T) {
	var tests = map[string]struct {
		Fails             []ErrorMessage
		Verbose           []string
		ExpectNonEmptyMsg bool
	}{
		"assert no panic with nil fails": {
			Fails: nil,
		},
		"assert no panic with empty fails": {
			Fails: []ErrorMessage{},
		},
		"assert no panic with one fail item": {
			Fails: []ErrorMessage{
				{
					Error:  "err",
					Remedy: "do something",
				},
			},
			ExpectNonEmptyMsg: true,
		},
		"assert no panic with many fail items": {
			Fails: []ErrorMessage{
				{
					Error:  "err",
					Remedy: "do something",
				},
				{
					Error:  "err again",
					Remedy: "do something again",
				},
			},
			ExpectNonEmptyMsg: true,
		},
		"assert no panic with many fail items & verbose messages": {
			Fails: []ErrorMessage{
				{
					Error:  "got some err",
					Remedy: "do something",
				},
				{
					Error:  "got some err again",
					Remedy: "do something again",
				},
			},
			Verbose: []string{
				"Path is not supported",
			},
			ExpectNonEmptyMsg: true,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			v := &FieldPathValidationResult{
				Failures: mock.Fails,
				Verbose:  mock.Verbose,
			}
			got := v.Error()
			if mock.ExpectNonEmptyMsg && got == "" {
				t.Fatalf("Expected error message got none")
			}
			if !mock.ExpectNonEmptyMsg && got != "" {
				t.Fatalf("Expected no error message got: \n%s", got)
			}
		})
	}
}

func TestValidationIsSupportedPath(t *testing.T) {
	var tests = map[string]struct {
		FieldPath               string
		SupportedAbsolutePaths  []string
		UserAllowedPathPrefixes []string
		IsSupported             bool
	}{
		"No field path & No Supported path": {},
		"No field path": {
			SupportedAbsolutePaths: []string{
				"spec.metadata.labels",
			},
		},
		"starts with metadata as field path": {
			FieldPath: "metadata",
			SupportedAbsolutePaths: []string{
				"metadata.labels",
			},
			IsSupported: true,
		},
		"with metadata as field path": {
			FieldPath: "metadata",
			SupportedAbsolutePaths: []string{
				"metadata",
			},
			IsSupported: true,
		},
		"with spec.employees1.name as field path": {
			FieldPath: "spec.employees1.name",
			SupportedAbsolutePaths: []string{
				"spec.employees1.name",
			},
			IsSupported: true,
		},
		"starts with spec.employees12.emp as field path": {
			FieldPath: "spec.employees12.emp",
			SupportedAbsolutePaths: []string{
				"spec.employees12.emp.name",
			},
			IsSupported: true,
		},
		"with spec.employees.[1] as field path - negative": {
			FieldPath: "spec.employees.[1]",
			SupportedAbsolutePaths: []string{
				"spec.employees",
				"spec.employees.name",
			},
		},
		"with spec.employees.[1] as field path": {
			FieldPath: "spec.employees.[1]",
			SupportedAbsolutePaths: []string{
				"spec.employees.[*]",
				"spec.employees.[*].name",
			},
			IsSupported: true,
		},
		"with spec.employees.[1].name as field path - negative": {
			FieldPath: "spec.employees.[1].name",
			SupportedAbsolutePaths: []string{
				"spec.employees.[*]",
				"spec.employees.[*].id",
				"spec.employees.[*].age",
			},
		},
		"with spec.employees.[11].name as field path": {
			FieldPath: "spec.employees.[11].name",
			SupportedAbsolutePaths: []string{
				"spec.employees.[*].name",
			},
			IsSupported: true,
		},
		"starts with spec.employees.[11].emp as field path": {
			FieldPath: "spec.employees.[11].emp",
			SupportedAbsolutePaths: []string{
				"spec.employees.[*].emp.age",
			},
			IsSupported: true,
		},
		"with spec.employees.[111].emp as field path": {
			FieldPath: "spec.employees.[111].emp",
			SupportedAbsolutePaths: []string{
				"spec.employees.[*].emp",
			},
			IsSupported: true,
		},
		"starts with spec.employees.[111].emp as field path": {
			FieldPath: "spec.employees.[111].emp",
			SupportedAbsolutePaths: []string{
				"spec.employees.[*].emp.name",
			},
			IsSupported: true,
		},
		"starts with spec.employee as field path - partial match - negative": {
			FieldPath: "spec.emp",
			SupportedAbsolutePaths: []string{
				"spec.employee",
			},
		},
		"starts with spec.employees.[111].emp as field path - partial match - negative": {
			FieldPath: "spec.employees.[111].emp",
			SupportedAbsolutePaths: []string{
				"spec.employees.[*].employee.name",
			},
		},
		"with spec.lob.[111].employees.[1] as field path": {
			FieldPath: "spec.lob.[111].employees.[1]",
			SupportedAbsolutePaths: []string{
				"spec.lob.[*].employees.[*]",
			},
			IsSupported: true,
		},
		"starts with spec.lob.[111].employees.[1] as field path": {
			FieldPath: "spec.lob.[111].employees.[1]",
			SupportedAbsolutePaths: []string{
				"spec.lob.[*].employees.[*].id",
			},
			IsSupported: true,
		},
		"any path starting with status is supported": {
			FieldPath:               "status.schema.failures.[11].msg.[1]",
			UserAllowedPathPrefixes: []string{"status."},
			IsSupported:             true,
		},
		"any path starting with metadata is supported": {
			FieldPath:               "metadata.name",
			UserAllowedPathPrefixes: []string{"metadata."},
			IsSupported:             true,
		},
		"any finalizer is supported": {
			FieldPath:               "metadata.finalizers",
			UserAllowedPathPrefixes: []string{"metadata."},
			IsSupported:             true,
		},
		"any path starting with status.schema is supported": {
			FieldPath:               "status.schema.failures.[11].msg.[1]",
			UserAllowedPathPrefixes: []string{"status.schema."},
			IsSupported:             true,
		},
		"any path starting with status.schema is supported - negative": {
			FieldPath:               "status.tasks.failures.[11].msg.[1]",
			UserAllowedPathPrefixes: []string{"status.schema."},
		},
		"any path starting with status.schema.failures is supported": {
			FieldPath:               "status.schema.failures.[11].msg.[1]",
			UserAllowedPathPrefixes: []string{"status.schema.failures."},
			IsSupported:             true,
		},
		"any path starting with status.schema.failures is supported - negative": {
			FieldPath:               "status.schema.verbose.[11].msg.[1]",
			UserAllowedPathPrefixes: []string{"status.schema.failures."},
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			v := &FieldPathValidation{
				SupportedAbsolutePaths:  mock.SupportedAbsolutePaths,
				UserAllowedPathPrefixes: mock.UserAllowedPathPrefixes,
			}
			got := v.isSupportedPath(mock.FieldPath)
			if mock.IsSupported != got {
				t.Fatalf(
					"Expected %t got %t",
					mock.IsSupported,
					got,
				)
			}
		})
	}
}

func TestValidationGetSupportedPathsForField(t *testing.T) {
	var supportedAbsolutePaths = []string{
		"metadata.labels",
		"metadata.annotations",
		"spec.replicas",
		"spec.containers",
		"status.phase",
		"status.replicas.running",
	}
	var tests = map[string]struct {
		FieldName              string
		ExpectedSupportedPaths []string
	}{
		"labels": {
			FieldName: "labels",
			ExpectedSupportedPaths: []string{
				"metadata.labels",
			},
		},
		"replicas": {
			FieldName: "replicas",
			ExpectedSupportedPaths: []string{
				"spec.replicas",
				"status.replicas.running",
			},
		},
		"spec": {
			FieldName: "spec",
			ExpectedSupportedPaths: []string{
				"spec.replicas",
				"spec.containers",
			},
		},
		"running": {
			FieldName: "running",
			ExpectedSupportedPaths: []string{
				"status.replicas.running",
			},
		},
		"does not exist": {
			FieldName:              "does_not_exist",
			ExpectedSupportedPaths: nil,
		},
		"empty name": {
			FieldName:              "",
			ExpectedSupportedPaths: nil,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			v := &FieldPathValidation{
				SupportedAbsolutePaths: supportedAbsolutePaths,
			}
			got := v.getSupportedPathsForField(mock.FieldName)
			eq := stringutil.NewEquality(got, mock.ExpectedSupportedPaths)
			if eq.IsDiff() {
				t.Fatalf(
					"Expected no diff got \n%s",
					cmp.Diff(got, mock.ExpectedSupportedPaths),
				)
			}
		})
	}
}

func TestValidationGetRemedyMsgForField(t *testing.T) {
	var supportedAbsolutePaths = []string{
		"metadata.labels",
		"metadata.annotations",
		"spec.replicas",
		"spec.containers",
		"status.phase",
		"status.replicas.running",
	}
	var tests = map[string]struct {
		FieldName  string
		IsEmptyMsg bool
	}{
		"labels": {
			FieldName: "labels",
		},
		"replicas": {
			FieldName: "replicas",
		},
		"spec": {
			FieldName: "spec",
		},
		"running": {
			FieldName: "running",
		},
		"does not exist": {
			FieldName: "does_not_exist",
		},
		"empty name": {
			FieldName: "",
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			v := &FieldPathValidation{
				SupportedAbsolutePaths: supportedAbsolutePaths,
			}
			got := v.getRemedyMsgForField("", mock.FieldName)
			if mock.IsEmptyMsg && got != "" {
				t.Fatalf("Expected no message got %s", got)
			}
			if !mock.IsEmptyMsg && got == "" {
				t.Fatalf("Expected message got none")
			}
		})
	}
}

func TestValidationValidateFieldPath(t *testing.T) {
	var supportedAbsolutePaths = []string{
		"metadata.labels",
		"metadata.annotations",
		"spec.replicas",
		"spec.containers",
		"status.phase",
		"status.replicas.running",
	}
	var tests = map[string]struct {
		FieldPath      string
		ExpectFailures bool
	}{
		"metadata": {
			FieldPath: "metadata",
		},
		"labels": {
			FieldPath:      "labels",
			ExpectFailures: true,
		},
		"metadata.labels": {
			FieldPath: "metadata.labels",
		},
		"replicas": {
			FieldPath:      "replicas",
			ExpectFailures: true,
		},
		"spec": {
			FieldPath: "spec",
		},
		"spec.replicas": {
			FieldPath: "spec.replicas",
		},
		"running": {
			FieldPath:      "running",
			ExpectFailures: true,
		},
		"status.replicas.running": {
			FieldPath: "status.replicas.running",
		},
		"does not exist": {
			FieldPath:      "does_not_exist",
			ExpectFailures: true,
		},
		"empty name": {
			FieldPath:      "",
			ExpectFailures: true,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			v := &FieldPathValidation{
				SupportedAbsolutePaths: supportedAbsolutePaths,
			}
			v.validateFieldPaths(mock.FieldPath)
			if mock.ExpectFailures && len(v.failures) == 0 {
				t.Fatalf("Expected failures got none")
			}
			if !mock.ExpectFailures && len(v.failures) != 0 {
				t.Fatalf("Expected no failures got %v", v.failures)
			}
		})
	}
}

func TestValidationIsListAMap(t *testing.T) {
	var tests = map[string]struct {
		List  []interface{}
		IsMap bool
	}{
		"nil list": {
			List: nil,
		},
		"empty list": {
			List: []interface{}{},
		},
		"list of map": {
			List: []interface{}{
				map[string]interface{}{},
			},
			IsMap: true,
		},
		"list of nils": {
			List: []interface{}{
				nil,
				nil,
			},
		},
		"list of maps": {
			List: []interface{}{
				map[string]interface{}{
					"key": "value",
				},
				map[string]interface{}{
					"key1": "value1",
				},
			},
			IsMap: true,
		},
		"list of scalars - string": {
			List: []interface{}{
				"hi",
				"there",
			},
		},
		"list of scalars - int": {
			List: []interface{}{
				10,
				200,
			},
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			v := &FieldPathValidation{}
			got := v.isListAMap(mock.List)
			if got != mock.IsMap {
				t.Fatalf(
					"Expected ismap as %t got %t",
					mock.IsMap,
					got,
				)
			}
		})
	}
}

func TestValidationMakeMapFromList(t *testing.T) {
	var tests = map[string]struct {
		Given    []interface{}
		Expected map[string]interface{}
	}{
		"nil list": {
			Given:    nil,
			Expected: nil,
		},
		"empty list": {
			Given:    []interface{}{},
			Expected: nil,
		},
		"list of nils": {
			Given: []interface{}{
				nil,
				nil,
			},
			Expected: map[string]interface{}{
				"[0]": nil,
				"[1]": nil,
			},
		},
		"list of string": {
			Given: []interface{}{
				"hi",
				"there",
			},
			Expected: map[string]interface{}{
				"[0]": "hi",
				"[1]": "there",
			},
		},
		"list of int": {
			Given: []interface{}{
				10,
				11,
			},
			Expected: map[string]interface{}{
				"[0]": 10,
				"[1]": 11,
			},
		},
		"list of map": {
			Given: []interface{}{
				map[string]interface{}{
					"hi": 10,
				},
				map[string]interface{}{
					"hello": 20,
				},
			},
			Expected: map[string]interface{}{
				"[0]": map[string]interface{}{
					"hi": 10,
				},
				"[1]": map[string]interface{}{
					"hello": 20,
				},
			},
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			v := &FieldPathValidation{}
			got := v.makeMapFromList(mock.Given)
			diff := cmp.Diff(mock.Expected, got)
			if diff != "" {
				t.Fatalf("Expected no diff got: \n%s", diff)
			}
		})
	}
}

func TestValidationValidateFieldPathsOfMap(t *testing.T) {
	var supportedAbsolutePaths = []string{
		"metadata.labels",
		"metadata.annotations",
		"spec.replicas",
		"spec.containers",
		"status.phase",
		"status.replicas.running",
	}
	var tests = map[string]struct {
		BasePath string
		Given    map[string]interface{}
		IsValid  bool
	}{
		// ---------------------------------------
		// spec.replicas
		// ---------------------------------------
		"missing base path for replicas": {
			BasePath: "",
			Given: map[string]interface{}{
				"replicas": 1,
			},
		},
		"invalid base path for replicas": {
			BasePath: "metadata",
			Given: map[string]interface{}{
				"replicas": 2,
			},
		},
		"valid base path for 0 spec replicas": {
			BasePath: "spec",
			Given: map[string]interface{}{
				"replicas": 0,
			},
			IsValid: true,
		},
		"valid path spec.replicas": {
			BasePath: "spec",
			Given: map[string]interface{}{
				"replicas": 1,
			},
			IsValid: true,
		},
		// ---------------------------------------
		// status.replicas.running
		// ---------------------------------------
		"invalid base path for status.replicas.running": {
			BasePath: "status.replicas",
			Given: map[string]interface{}{
				"replicas": map[string]interface{}{
					"running": 1,
				},
			},
		},
		"valid path status.replicas.running": {
			BasePath: "status",
			Given: map[string]interface{}{
				"replicas": map[string]interface{}{
					"running": 1,
				},
			},
			IsValid: true,
		},
		// ---------------------------------------
		// metadata.labels
		// ---------------------------------------
		"missing base path for labels": {
			BasePath: "",
			Given: map[string]interface{}{
				"labels": map[string]interface{}{
					"app": "dope",
				},
			},
		},
		"invalid base path for labels": {
			BasePath: "meta",
			Given: map[string]interface{}{
				"labels": map[string]interface{}{
					"app": "dope",
				},
			},
		},
		"valid base path for nil labels": {
			BasePath: "metadata",
			Given: map[string]interface{}{
				"labels": nil,
			},
			IsValid: true,
		},
		"valid base path for labels": {
			BasePath: "metadata",
			Given: map[string]interface{}{
				"labels": map[string]interface{}{
					"app": "dope",
				},
			},
			IsValid: true,
		},
		// ---------------------------------------
		// metadata.annotations
		// ---------------------------------------
		"valid base path for nil annotations": {
			BasePath: "metadata",
			Given: map[string]interface{}{
				"annotations": nil,
			},
			IsValid: true,
		},
		"valid base path for annotations": {
			BasePath: "metadata",
			Given: map[string]interface{}{
				"annotations": map[string]interface{}{
					"app": "dope",
				},
			},
			IsValid: true,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			v := &FieldPathValidation{
				SupportedAbsolutePaths: supportedAbsolutePaths,
				UserAllowedPathPrefixes: []string{
					"metadata.labels.",
					"metadata.annotations.",
				},
			}
			v.validateFieldPathsOfMap(mock.BasePath, mock.Given)
			result := &FieldPathValidationResult{
				Failures: v.failures,
				Verbose:  v.verbose,
			}
			if mock.IsValid && len(v.failures) != 0 {
				t.Fatalf("Expected valid map got: \n%s", result)
			}
			if !mock.IsValid && len(v.failures) == 0 {
				t.Fatalf("Expected invalid map got none")
			}
		})
	}
}

func TestValidationValidateFieldPathsOfListViaMap(t *testing.T) {
	var supportedAbsolutePaths = []string{
		"spec.replicas",
		"spec.containers",
		"status.phase",
		"status.replicas.[*].running",
	}
	var tests = map[string]struct {
		BasePath string
		Given    []interface{}
		IsValid  bool
	}{
		// ---------------------------------------
		// spec.replicas
		// ---------------------------------------
		"invalid list based path spec.replicas": {
			BasePath: "spec",
			Given: []interface{}{
				map[string]interface{}{
					"replicas": 1,
				},
			},
		},
		// ---------------------------------------
		// status.replicas.running
		// ---------------------------------------
		"invalid list based path status.replicas.running": {
			BasePath: "status",
			Given: []interface{}{
				map[string]interface{}{
					"replicas": map[string]interface{}{
						"running": 1,
					},
				},
			},
		},
		"valid list based path status.replicas.[*].running": {
			BasePath: "status.replicas",
			Given: []interface{}{
				map[string]interface{}{
					"running": true,
				},
				map[string]interface{}{
					"running": false,
				},
			},
			IsValid: true,
		},
		// ---------------------------------------
		// metadata.labels
		// ---------------------------------------
		"labels cannot be used as list": {
			BasePath: "metadata",
			Given: []interface{}{
				map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "dope",
					},
				},
			},
		},
		// ---------------------------------------
		// metadata.annotations
		// ---------------------------------------
		"annotations can not be used as list": {
			BasePath: "metadata",
			Given: []interface{}{
				map[string]interface{}{
					"annotations": map[string]interface{}{
						"app": "dope",
					},
				},
			},
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			v := &FieldPathValidation{
				SupportedAbsolutePaths: supportedAbsolutePaths,
				UserAllowedPathPrefixes: []string{
					"metadata.labels.",
					"metadata.annotations.",
				},
			}
			v.validateFieldPathsOfListViaMap(mock.BasePath, mock.Given)
			result := &FieldPathValidationResult{
				Failures: v.failures,
				Verbose:  v.verbose,
			}
			if mock.IsValid && len(v.failures) != 0 {
				t.Fatalf("Expected valid list got: \n%s", result)
			}
			if !mock.IsValid && len(v.failures) == 0 {
				t.Fatalf("Expected invalid list got none")
			}
		})
	}
}

func TestValidationValidateFieldPathsOfArray(t *testing.T) {
	var supportedAbsolutePaths = []string{
		"spec.tags", // scalar list
		"spec.replicas",
		"spec.containers.[*].name",  // list of maps
		"spec.containers.[*].image", // list of maps
		"status.phase",
		"status.replicas.[*].running", // list of maps
	}
	var tests = map[string]struct {
		BasePath string
		Given    []interface{}
		IsValid  bool
	}{
		// ---------------------------------------
		// spec.tags
		// ---------------------------------------
		"valid scalar list spec.tags": {
			BasePath: "spec.tags",
			Given: []interface{}{
				"hi",
				"there",
				"how-do-you-do",
			},
			IsValid: true,
		},
		// ---------------------------------------
		// spec.replicas
		// ---------------------------------------
		"invalid list based path spec.replicas": {
			BasePath: "spec",
			Given: []interface{}{
				map[string]interface{}{
					"replicas": 1,
				},
			},
		},
		// ---------------------------------------
		// status.replicas.[*].running
		// ---------------------------------------
		"invalid list based path status.replicas.running": {
			BasePath: "status",
			Given: []interface{}{
				map[string]interface{}{
					"replicas": map[string]interface{}{
						"running": 1,
					},
				},
			},
		},
		"valid list based path status.replicas.[*].running": {
			BasePath: "status.replicas",
			Given: []interface{}{
				map[string]interface{}{
					"running": true,
				},
				map[string]interface{}{
					"running": false,
				},
			},
			IsValid: true,
		},
		// -------------------------------------
		// spec.containers.[*].image
		// spec.containers.[*].name
		// -------------------------------------
		"invalid list based path spec.containers.[*].desc": {
			BasePath: "spec.containers",
			Given: []interface{}{
				map[string]interface{}{
					"name":  "my-nginx-con",
					"image": "nginx",
					"desc":  "nginx",
				},
				map[string]interface{}{
					"image": "dope",
				},
			},
		},
		"valid list based path spec.containers.[*].image": {
			BasePath: "spec.containers",
			Given: []interface{}{
				map[string]interface{}{
					"name":  "my-nginx-con",
					"image": "nginx",
				},
				map[string]interface{}{
					"image": "dope",
				},
			},
			IsValid: true,
		},
		// ---------------------------------------
		// metadata.labels
		// ---------------------------------------
		"labels cannot be used as list": {
			BasePath: "metadata",
			Given: []interface{}{
				map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "dope",
					},
				},
			},
		},
		// ---------------------------------------
		// metadata.annotations
		// ---------------------------------------
		"annotations can not be used as list": {
			BasePath: "metadata",
			Given: []interface{}{
				map[string]interface{}{
					"annotations": map[string]interface{}{
						"app": "dope",
					},
				},
			},
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			v := &FieldPathValidation{
				SupportedAbsolutePaths: supportedAbsolutePaths,
				UserAllowedPathPrefixes: []string{
					"metadata.labels.",
					"metadata.annotations.",
				},
			}
			v.validateFieldPathsOfArray(mock.BasePath, mock.Given)
			result := &FieldPathValidationResult{
				Failures: v.failures,
				Verbose:  v.verbose,
			}
			if mock.IsValid && len(v.failures) != 0 {
				t.Fatalf("Expected valid array got: \n%s", result)
			}
			if !mock.IsValid && len(v.failures) == 0 {
				t.Fatalf("Expected invalid array got none")
			}
		})
	}
}

func TestValidationValidatePrivate(t *testing.T) {
	var supportedAbsolutePaths = []string{
		"spec.tags", // scalar list
		"spec.replicas",
		"spec.containers.[*].name",  // list of maps
		"spec.containers.[*].image", // list of maps
		"status.phase",
		"status.replicas.[*].running", // list of maps
	}
	var tests = map[string]struct {
		BasePath string
		Given    interface{}
		IsValid  bool
	}{
		// ---------------------------------------
		// spec.tags
		// ---------------------------------------
		"valid scalar list spec.tags": {
			BasePath: "spec.tags",
			Given: []interface{}{
				"hi",
				"there",
				"how-do-you-do",
			},
			IsValid: true,
		},
		// ---------------------------------------
		// spec.replicas
		// ---------------------------------------
		"invalid list based path spec.replicas": {
			BasePath: "spec",
			Given: []interface{}{
				map[string]interface{}{
					"replicas": 1,
				},
			},
		},
		"valid spec.replicas": {
			BasePath: "spec",
			Given: map[string]interface{}{
				"replicas": 1,
			},
			IsValid: true,
		},
		// ---------------------------------------
		// status.replicas.[*].running
		// ---------------------------------------
		"invalid list based path status.replicas.running": {
			BasePath: "status",
			Given: []interface{}{
				map[string]interface{}{
					"replicas": map[string]interface{}{
						"running": 1,
					},
				},
			},
		},
		"valid list based path status.replicas.[*].running": {
			BasePath: "status.replicas",
			Given: []interface{}{
				map[string]interface{}{
					"running": true,
				},
				map[string]interface{}{
					"running": false,
				},
			},
			IsValid: true,
		},
		// -------------------------------------
		// spec.containers.[*].image
		// spec.containers.[*].name
		// -------------------------------------
		"invalid list based path spec.containers.[*].desc": {
			BasePath: "spec.containers",
			Given: []interface{}{
				map[string]interface{}{
					"name":  "my-nginx-con",
					"image": "nginx",
					"desc":  "nginx",
				},
				map[string]interface{}{
					"image": "dope",
				},
			},
		},
		"valid list based path spec.containers.[*].image": {
			BasePath: "spec.containers",
			Given: []interface{}{
				map[string]interface{}{
					"name":  "my-nginx-con",
					"image": "nginx",
				},
				map[string]interface{}{
					"image": "dope",
				},
			},
			IsValid: true,
		},
		// ---------------------------------------
		// metadata.labels
		// ---------------------------------------
		"valid metadata.labels": {
			BasePath: "metadata",
			Given: map[string]interface{}{
				"labels": map[string]interface{}{
					"app": "dope",
				},
			},
			IsValid: true,
		},
		"invalid labels": {
			BasePath: "",
			Given: map[string]interface{}{
				"labels": map[string]interface{}{
					"app": "dope",
				},
			},
		},
		// ---------------------------------------
		// metadata.annotations
		// ---------------------------------------
		"valid metadata.annotations": {
			BasePath: "metadata",
			Given: map[string]interface{}{
				"annotations": map[string]interface{}{
					"app": "dope",
				},
			},
			IsValid: true,
		},
		"invalid annotations": {
			BasePath: "",
			Given: map[string]interface{}{
				"annotations": map[string]interface{}{
					"app": "dope",
				},
			},
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			v := &FieldPathValidation{
				SupportedAbsolutePaths: supportedAbsolutePaths,
				UserAllowedPathPrefixes: []string{
					"metadata.labels.",
					"metadata.annotations.",
				},
			}
			v.validate(mock.BasePath, mock.Given)
			result := &FieldPathValidationResult{
				Failures: v.failures,
				Verbose:  v.verbose,
			}
			if mock.IsValid && len(v.failures) != 0 {
				t.Fatalf("Expected valid schema got: \n%s", result)
			}
			if !mock.IsValid && len(v.failures) == 0 {
				t.Fatalf("Expected invalid schema got none")
			}
		})
	}
}

func TestValidationValidate(t *testing.T) {
	var supportedAbsolutePaths = []string{
		"spec.tags", // scalar list
		"spec.replicas",
		"spec.containers.[*].name",  // list of maps
		"spec.containers.[*].image", // list of maps
		"status.phase",
		"status.replicas.[*].running", // list of maps
	}
	var tests = map[string]struct {
		Given   map[string]interface{}
		IsValid bool
	}{
		// ---------------------------------------
		// spec.tags
		// ---------------------------------------
		"valid scalar list spec.tags": {
			Given: map[string]interface{}{
				"spec": map[string]interface{}{
					"tags": []interface{}{
						"hi",
						"there",
						"how-do-you-do",
					},
				},
			},
			IsValid: true,
		},
		// ---------------------------------------
		// spec.replicas
		// ---------------------------------------
		"valid spec.replicas": {
			Given: map[string]interface{}{
				"spec": map[string]interface{}{
					"replicas": 1,
				},
			},
			IsValid: true,
		},
		// ---------------------------------------
		// status.replicas.[*].running
		// ---------------------------------------
		"invalid list based path status.replicas.running": {
			Given: map[string]interface{}{
				"status": map[string]interface{}{
					"replicas": map[string]interface{}{
						"running": 1,
					},
				},
			},
		},
		"valid list based path status.replicas.[*].running": {
			Given: map[string]interface{}{
				"status": map[string]interface{}{
					"replicas": []interface{}{
						map[string]interface{}{
							"running": true,
						},
						map[string]interface{}{
							"running": false,
						},
					},
				},
			},
			IsValid: true,
		},
		// -------------------------------------
		// spec.containers.[*].image
		// spec.containers.[*].name
		// -------------------------------------
		"invalid list based path spec.containers.[*].desc": {
			Given: map[string]interface{}{
				"spec": map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"name":  "my-nginx-con",
							"image": "nginx",
							"desc":  "nginx",
						},
						map[string]interface{}{
							"image": "dope",
						},
					},
				},
			},
		},
		"valid list based path spec.containers.[*].image": {
			Given: map[string]interface{}{
				"spec": map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"name":  "my-nginx-con",
							"image": "nginx",
						},
						map[string]interface{}{
							"image": "dope",
						},
					},
				},
			},
			IsValid: true,
		},
		// ---------------------------------------
		// metadata.labels
		// ---------------------------------------
		"valid metadata.labels": {
			Given: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "dope",
					},
				},
			},
			IsValid: true,
		},
		"invalid labels": {
			Given: map[string]interface{}{
				"labels": map[string]interface{}{
					"app": "dope",
				},
			},
		},
		// ---------------------------------------
		// metadata.annotations
		// ---------------------------------------
		"valid metadata.annotations": {
			Given: map[string]interface{}{
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						"app": "dope",
					},
				},
			},
			IsValid: true,
		},
		"invalid annotations": {
			Given: map[string]interface{}{
				"annotations": map[string]interface{}{
					"app": "dope",
				},
			},
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			v := &FieldPathValidation{
				Target:                 mock.Given,
				SupportedAbsolutePaths: supportedAbsolutePaths,
				UserAllowedPathPrefixes: []string{
					"metadata.labels.",
					"metadata.annotations.",
				},
			}
			got := v.Validate()
			if mock.IsValid && got.Status != FieldPathValidationStatusValid {
				t.Fatalf("Expected valid schema got: \n%s", got)
			}
			if !mock.IsValid && got.Status == FieldPathValidationStatusValid {
				t.Fatalf("Expected invalid schema got \n%s", got)
			}
		})
	}
}

func TestValidationValidateUserAllowedPathPrefixes(t *testing.T) {
	var supportedAbsolutePaths = []string{
		"spec.tags", // scalar list
		"spec.replicas",
		"spec.containers.[*].name",  // list of maps
		"spec.containers.[*].image", // list of maps
		"status.phase",
		"status.replicas.[*].running", // list of maps
	}
	var tests = map[string]struct {
		Given                   map[string]interface{}
		UserAllowedPathPrefixes []string
		IsValid                 bool
	}{
		// ---------------------------------------
		// spec.tags
		// ---------------------------------------
		"valid scalar list spec.tags": {
			Given: map[string]interface{}{
				"spec": map[string]interface{}{
					"tags": []interface{}{
						"hi",
						"there",
						"how-do-you-do",
					},
				},
			},
			IsValid: true,
		},
		// ---------------------------------------
		// spec.replicas
		// ---------------------------------------
		"valid spec.replicas": {
			Given: map[string]interface{}{
				"spec": map[string]interface{}{
					"replicas": 1,
				},
			},
			IsValid: true,
		},
		// ---------------------------------------
		// status.replicas.[*].running
		// ---------------------------------------
		"invalid list based path status.replicas.running": {
			Given: map[string]interface{}{
				"status": map[string]interface{}{
					"replicas": map[string]interface{}{
						"running": 1,
					},
				},
			},
		},
		"valid list based path status.replicas.[*].running": {
			Given: map[string]interface{}{
				"status": map[string]interface{}{
					"replicas": []interface{}{
						map[string]interface{}{
							"running": true,
						},
						map[string]interface{}{
							"running": false,
						},
					},
				},
			},
			IsValid: true,
		},
		// -------------------------------------
		// spec.containers.[*].image
		// spec.containers.[*].name
		// -------------------------------------
		"invalid list based path spec.containers.[*].desc": {
			Given: map[string]interface{}{
				"spec": map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"name":  "my-nginx-con",
							"image": "nginx",
							"desc":  "nginx",
						},
						map[string]interface{}{
							"image": "dope",
						},
					},
				},
			},
		},
		"valid list based path spec.containers.[*].image": {
			Given: map[string]interface{}{
				"spec": map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"name":  "my-nginx-con",
							"image": "nginx",
						},
						map[string]interface{}{
							"image": "dope",
						},
					},
				},
			},
			IsValid: true,
		},
		// ---------------------------------------
		// metadata.labels
		// ---------------------------------------
		"valid metadata.labels": {
			Given: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "dope",
					},
				},
			},
			UserAllowedPathPrefixes: []string{
				"metadata.labels.",
			},
			IsValid: true,
		},
		"invalid metadata.labels user allowed path prefixx": {
			Given: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "dope",
					},
				},
			},
			UserAllowedPathPrefixes: []string{
				"metadata.labels", // missing dot as suffix
			},
		},
		"invalid labels": {
			Given: map[string]interface{}{
				"labels": map[string]interface{}{
					"app": "dope",
				},
			},
		},
		// ---------------------------------------
		// metadata.annotations
		// ---------------------------------------
		"valid metadata.annotations": {
			Given: map[string]interface{}{
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						"app": "dope",
					},
				},
			},
			UserAllowedPathPrefixes: []string{
				"metadata.annotations.",
			},
			IsValid: true,
		},
		"invalid metadata.annotations user allowed path prefixx": {
			Given: map[string]interface{}{
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						"app": "dope",
					},
				},
			},
			UserAllowedPathPrefixes: []string{
				"metadata.annotations", // missing dot as suffix
			},
		},
		"invalid annotations": {
			Given: map[string]interface{}{
				"annotations": map[string]interface{}{
					"app": "dope",
				},
			},
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			v := &FieldPathValidation{
				Target:                  mock.Given,
				SupportedAbsolutePaths:  supportedAbsolutePaths,
				UserAllowedPathPrefixes: mock.UserAllowedPathPrefixes,
			}
			got := v.Validate()
			if mock.IsValid && got.Status != FieldPathValidationStatusValid {
				t.Fatalf("Expected valid schema got: \n%s", got)
			}
			if !mock.IsValid && got.Status == FieldPathValidationStatusValid {
				t.Fatalf("Expected invalid schema got \n%s", got)
			}
		})
	}
}
