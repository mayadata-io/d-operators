// +build !integration

package types

import (
	"testing"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// ToTyped transforms the provided unstruct instance
// to target type
func ToTyped(src *unstructured.Unstructured, target interface{}) error {
	if src == nil || src.Object == nil {
		return errors.Errorf(
			"Can't transform unstruct to typed: Nil unstruct content",
		)
	}
	if target == nil {
		return errors.Errorf(
			"Can't transform unstruct to typed: Nil target",
		)
	}
	return runtime.DefaultUnstructuredConverter.FromUnstructured(
		src.UnstructuredContent(),
		target,
	)
}

// ToTyped transforms the provided unstruct instance
func TestToTyped(t *testing.T) {
	var tests = map[string]struct {
		Given *unstructured.Unstructured
	}{
		"unstruct to typed command": {
			Given: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":       "Command",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"namespace": "dope",
						"name":      "one-1",
					},
					"spec": map[string]interface{}{
						"commands": []interface{}{
							map[string]interface{}{
								"cmd": []interface{}{
									"kubectl",
									"get",
									"pods",
									"-n",
									"metal",
								},
								"name": "get pods in metal ns",
							},
						},
					},
					"status": map[string]interface{}{
						"counter": map[string]interface{}{
							"errorCount":   1,
							"timeoutCount": 0,
							"warnCount":    0,
						},
						"outputs": map[string]interface{}{
							"get pods in metal ns": map[string]interface{}{
								"cmd":       "kubectl",
								"completed": false,
								"error":     "exec: \"kubectl\": executable file not found in $PATH",
								"executionTime": map[string]interface{}{
									"readableValue":  "0s",
									"valueInSeconds": float64(0),
								},
								"exit":     -1,
								"pid":      0,
								"stderr":   "",
								"stdout":   "",
								"timedout": false,
							},
						},
						"phase":    "Error",
						"reason":   "1 error(s) found",
						"timedout": false,
						"timetakenInSeconds": map[string]interface{}{
							"readableValue":  "0s",
							"valueInSeconds": float64(0),
						},
					},
				},
			},
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			var finale Command
			err := ToTyped(mock.Given, &finale)
			if err != nil {
				t.Fatalf("Expected no error got %s", err.Error())
			}
		})
	}
}
