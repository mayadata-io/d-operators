package recipe

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	types "mayadata.io/d-operators/types/recipe"
)

func TestPathCheckingAssertValueString(t *testing.T) {
	var tests = map[string]struct {
		State                 *unstructured.Unstructured
		TaskName              string
		PathCheck             string
		Path                  string
		Value                 string
		retryIfValueEquals    bool
		retryIfValueNotEquals bool
		IsError               bool
		ExpectedAssert        bool
	}{
		"assert ipaddress None == None": {
			State: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Service",
					"metadata": map[string]interface{}{
						"name":      "test",
						"namespace": "default",
					},
					"spec": map[string]interface{}{
						"ipAddress": "None",
					},
				},
			},
			Path:                  "spec.ipAddress",
			Value:                 "None",
			retryIfValueNotEquals: true,
			ExpectedAssert:        true,
		},
		"assert ipaddress ip != 12.123.12.11": {
			State: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Service",
					"metadata": map[string]interface{}{
						"name":      "test",
						"namespace": "default",
					},
					"spec": map[string]interface{}{
						"ipAddress": "12.123.12.11",
					},
				},
			},
			Path:               "spec.ipAddress",
			Value:              "None",
			retryIfValueEquals: true,
			ExpectedAssert:     true,
		},
		"assert ipaddress ip != ": {
			State: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Service",
					"metadata": map[string]interface{}{
						"name":      "test",
						"namespace": "default",
					},
					"spec": map[string]interface{}{
						"ipAddress": "12.123.12.11",
					},
				},
			},
			Path:               "spec.ipAddress",
			Value:              "",
			retryIfValueEquals: true,
			ExpectedAssert:     true,
		},
		"assert ipaddress Nil != None": {
			State: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Service",
					"metadata": map[string]interface{}{
						"name":      "test",
						"namespace": "default",
					},
					"spec": map[string]interface{}{
						"ipAddress": "None",
					},
				},
			},
			Path:                  "spec.ipAddress",
			Value:                 "Nil",
			retryIfValueNotEquals: true,
			ExpectedAssert:        false,
		},
	}
	for scenario, tObj := range tests {
		scenario := scenario
		tObj := tObj
		t.Run(scenario, func(t *testing.T) {
			pc := &PathChecking{
				PathCheck: types.PathCheck{
					Path:  tObj.Path,
					Value: tObj.Value,
				},
				retryIfValueEquals:    tObj.retryIfValueEquals,
				retryIfValueNotEquals: tObj.retryIfValueNotEquals,
				result:                &types.PathCheckResult{},
			}
			got, err := pc.assertValueString(tObj.State)
			if tObj.IsError && err == nil {
				t.Fatalf("Expected error got none")
			}
			if !tObj.IsError && err != nil {
				t.Fatalf("Expected no error got %s", err.Error())
			}
			if tObj.IsError {
				return
			}
			if got != tObj.ExpectedAssert {
				t.Fatalf("Expected assert = %t got %t", tObj.ExpectedAssert, got)
			}

		})
	}
}
