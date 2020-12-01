// +build !integration

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

package unstruct

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"mayadata.io/d-operators/common/pointer"
)

type MockSchemaFailure struct {
	Error  string `json:"error"`
	Remedy string `json:"remedy,omitempty"`
}

type MockSchemaResult struct {
	Phase    string              `json:"phase"`
	Failures []MockSchemaFailure `json:"failures,omitempty"`
	Verbose  []string            `json:"verbose,omitempty"`
}

type MockTaskCount struct {
	Failed  int `json:"failed"`  // Number of failed tasks
	Skipped int `json:"skipped"` // Number of skipped tasks
	Warning int `json:"warning"` // Number of tasks with warnings
	Total   int `json:"total"`   // Total number of tasks in the Recipe
}

type MockExecutionTime struct {
	ValueInSeconds float64 `json:"valueInSeconds"`
	ReadableValue  string  `json:"readableValue"`
}

type MockTaskResult struct {
	Step          int               `json:"step"`
	Phase         string            `json:"phase"`
	ExecutionTime MockExecutionTime `json:"executionTime,omitempty"`
	Internal      *bool             `json:"internal,omitempty"`
	Message       string            `json:"message,omitempty"`
	Verbose       string            `json:"verbose,omitempty"`
	Warning       string            `json:"warning,omitempty"`
	Timeout       string            `json:"timeout,omitempty"`
}

type MockRecipeStatus struct {
	Phase          string                    `json:"phase"`
	Reason         string                    `json:"reason,omitempty"`
	Message        string                    `json:"message,omitempty"`
	Schema         *MockSchemaResult         `json:"schema,omitempty"`
	ExecutionTime  *MockExecutionTime        `json:"executionTime,omitempty"`
	TaskCount      *MockTaskCount            `json:"taskCount,omitempty"`
	TaskResultList map[string]MockTaskResult `json:"taskResultList"`
}

func TestMarshalThenUnmarshalThenSetUnstructStatus(t *testing.T) {
	var tests = map[string]struct {
		Given interface{}
		IsOk  bool
	}{
		"all is ok": {
			Given: &MockRecipeStatus{
				Phase:   "Ok",
				Message: "All is okay",
				ExecutionTime: &MockExecutionTime{
					ValueInSeconds: 123.23,
					ReadableValue:  "2m23s",
				},
				Schema: &MockSchemaResult{
					Phase:    "ValidSchema",
					Failures: nil,
					Verbose: []string{
						"No errors",
					},
				},
				TaskCount: &MockTaskCount{
					Failed:  0,
					Skipped: 0,
					Total:   3,
					Warning: 0,
				},
				TaskResultList: map[string]MockTaskResult{
					"task-0": {
						Step:    0,
						Phase:   "Completed",
						Message: "Task ran successfully",
						ExecutionTime: MockExecutionTime{
							ValueInSeconds: 10.000,
							ReadableValue:  "0m10s",
						},
						Internal: pointer.Bool(false),
					},
					"task-1": {
						Step:    1,
						Phase:   "Completed",
						Message: "Task ran successfully",
						ExecutionTime: MockExecutionTime{
							ValueInSeconds: 20.000,
							ReadableValue:  "0m20s",
						},
						Internal: pointer.Bool(false),
					},
				},
			},
			IsOk: true,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			// part 1
			var dest interface{}
			err := MarshalThenUnmarshal(mock.Given, &dest)
			if mock.IsOk && err != nil {
				t.Fatalf(
					"Expected no error during marshal/unmarshal got %s",
					err.Error(),
				)
			}
			if !mock.IsOk && err == nil {
				t.Fatalf("Expected error during marshal/unmarshal got none")
			}
			// part 2
			result := &unstructured.Unstructured{
				Object: make(map[string]interface{}),
			}
			err = unstructured.SetNestedField(
				result.Object,
				dest,
				"status",
			)
			if mock.IsOk && err != nil {
				t.Fatalf("Expected no error during set got %s", err.Error())
			}
			if !mock.IsOk && err == nil {
				t.Fatalf("Expected error during set got none")
			}
		})
	}
}
