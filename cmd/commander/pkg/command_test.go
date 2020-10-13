package pkg

import (
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ToTyped transforms the provided unstruct instance
func TestToTyped(t *testing.T) {
	var timetakenInt int64
	var timeTaken = time.Duration(timetakenInt).Seconds()
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
									"valueInSeconds": timeTaken,
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
							"valueInSeconds": 0 * 100.0 / 100,
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
