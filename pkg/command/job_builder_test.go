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

package command

import (
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	types "mayadata.io/d-operators/types/command"
)

func TestBuilderBuild(t *testing.T) {
	var tests = map[string]struct {
		Command                         types.Command
		ExpectedKind                    string
		ExpectedAPIVersion              string
		ExpectedName                    string
		ExpectedNamespace               string
		ExpectedContainerCount          int
		ExpectedImageName               string
		ExpectedImage                   string
		ExpectedImageArgCount           int
		ExpectedRestartPolicy           string
		ExpectedTTLSecondsAfterFinished int64
		ExpectedLabelCount              int
		IsError                         bool
		ExpectedServiceAccount          string
	}{
		"no job in command spec": {
			Command: types.Command{
				ObjectMeta: v1.ObjectMeta{
					Name:      "cmd",
					Namespace: "def",
				},
			},
			ExpectedAPIVersion:              types.JobAPIVersion,
			ExpectedKind:                    types.KindJob,
			ExpectedName:                    "cmd",
			ExpectedNamespace:               "def",
			ExpectedContainerCount:          1,
			ExpectedLabelCount:              3,
			ExpectedRestartPolicy:           "Never",
			ExpectedTTLSecondsAfterFinished: 0,
			ExpectedImage:                   "mayadataio/daction",
			ExpectedImageName:               "daction",
			ExpectedImageArgCount:           3,
			ExpectedServiceAccount:          "",
		},
		"job in command spec": {
			Command: types.Command{
				ObjectMeta: v1.ObjectMeta{
					Name:      "cmd",
					Namespace: "def",
				},
				Spec: types.CommandSpec{
					Template: types.Template{
						Job: &unstructured.Unstructured{
							Object: map[string]interface{}{
								"metadata": map[string]interface{}{
									"name":      "hi",
									"namespace": "hey",
									"labels": map[string]interface{}{
										"hi": "how-do-u-do",
									},
								},
							},
						},
					},
				},
			},
			ExpectedAPIVersion:              types.JobAPIVersion,
			ExpectedKind:                    types.KindJob,
			ExpectedName:                    "cmd",
			ExpectedNamespace:               "def",
			ExpectedContainerCount:          1,
			ExpectedLabelCount:              4, // extra label
			ExpectedRestartPolicy:           "Never",
			ExpectedTTLSecondsAfterFinished: 0,
			ExpectedImage:                   "mayadataio/daction",
			ExpectedImageName:               "daction",
			ExpectedImageArgCount:           3,
			ExpectedServiceAccount:          "",
		},
		"job with sidecar in command spec": {
			Command: types.Command{
				ObjectMeta: v1.ObjectMeta{
					Name:      "cmd",
					Namespace: "def",
				},
				Spec: types.CommandSpec{
					Template: types.Template{
						Job: &unstructured.Unstructured{
							Object: map[string]interface{}{
								"metadata": map[string]interface{}{
									"name":      "hi",
									"namespace": "hey",
								},
								"spec": map[string]interface{}{
									"template": map[string]interface{}{
										"spec": map[string]interface{}{
											"containers": []interface{}{
												map[string]interface{}{
													"name":  "sidecar",
													"image": "sidecar",
													"command": []interface{}{
														"/usr/bin/go",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			ExpectedAPIVersion:              types.JobAPIVersion,
			ExpectedKind:                    types.KindJob,
			ExpectedName:                    "cmd",
			ExpectedNamespace:               "def",
			ExpectedContainerCount:          2, // sidecar + main
			ExpectedLabelCount:              3,
			ExpectedRestartPolicy:           "Never",
			ExpectedTTLSecondsAfterFinished: 0,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			b := NewJobBuilder(JobBuildingConfig{
				Command: mock.Command,
			})
			got, err := b.Build()
			if mock.IsError && err == nil {
				t.Fatalf("Expected error got none")
			}
			if !mock.IsError && err != nil {
				t.Fatalf("Expected no error got %s", err.Error())
			}
			if mock.IsError {
				return
			}
			if got.GetKind() != mock.ExpectedKind {
				t.Fatalf(
					"Expected kind %s got %s",
					mock.ExpectedKind,
					got.GetKind(),
				)
			}
			if got.GetAPIVersion() != mock.ExpectedAPIVersion {
				t.Fatalf(
					"Expected api version %s got %s",
					mock.ExpectedAPIVersion,
					got.GetAPIVersion(),
				)
			}
			if got.GetName() != mock.ExpectedName {
				t.Fatalf(
					"Expected name %s got %s",
					mock.ExpectedName,
					got.GetName(),
				)
			}
			if got.GetNamespace() != mock.ExpectedNamespace {
				t.Fatalf(
					"Expected namespace %s got %s",
					mock.ExpectedNamespace,
					got.GetNamespace(),
				)
			}
			if len(got.GetLabels()) != mock.ExpectedLabelCount {
				t.Fatalf(
					"Expected label count %d got %d",
					mock.ExpectedLabelCount,
					len(got.GetLabels()),
				)
			}
			if gotTTLSecondsAfterFinished, found, _ := unstructured.NestedInt64(
				got.Object,
				"spec",
				"ttlSecondsAfterFinished",
			); found && gotTTLSecondsAfterFinished != mock.ExpectedTTLSecondsAfterFinished {
				t.Fatalf(
					"Expected ttl %d got %d",
					mock.ExpectedTTLSecondsAfterFinished,
					gotTTLSecondsAfterFinished,
				)
			}
			if gotRestartPolicy, found, _ := unstructured.NestedString(
				got.Object,
				"spec",
				"template",
				"spec",
				"restartPolicy",
			); found && gotRestartPolicy != mock.ExpectedRestartPolicy {
				t.Fatalf(
					"Expected restart policy %s got %s",
					mock.ExpectedRestartPolicy,
					gotRestartPolicy,
				)
			}
			if gotSvcAccName, found, _ := unstructured.NestedString(
				got.Object,
				"spec",
				"template",
				"spec",
				"serviceAccountName",
			); found && gotSvcAccName != mock.ExpectedServiceAccount {
				t.Fatalf(
					"Expected restart policy %s got %s",
					mock.ExpectedServiceAccount,
					gotSvcAccName,
				)
			}
			containers, found, _ := unstructured.NestedSlice(
				got.Object,
				"spec",
				"template",
				"spec",
				"containers",
			)
			if found && mock.ExpectedContainerCount != len(containers) {
				t.Fatalf(
					"Expected container count %d got %d",
					mock.ExpectedContainerCount,
					len(containers),
				)
			}
			// verify main container only
			// this will run when sidecars are not included
			if found && len(containers) == 1 {
				con0 := containers[0]
				con, ok := con0.(map[string]interface{})
				if !ok {
					return
				}
				if mock.ExpectedImage != con["image"] {
					t.Fatalf(
						"Expected image %s got %s",
						mock.ExpectedImage,
						con["image"],
					)
				}
				if mock.ExpectedImageName != con["name"] {
					t.Fatalf(
						"Expected image name %s got %s",
						mock.ExpectedImageName,
						con["name"],
					)
				}
				if mock.ExpectedImageArgCount != len(con["args"].([]interface{})) {
					t.Fatalf(
						"Expected arg count %d got %d",
						mock.ExpectedImageArgCount,
						len(con["args"].([]interface{})),
					)
				}
			}
		})
	}
}
