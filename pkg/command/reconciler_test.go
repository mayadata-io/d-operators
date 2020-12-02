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

package command

import (
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"openebs.io/metac/dynamic/discovery"
	"openebs.io/metac/server"
	metac "openebs.io/metac/start"

	types "mayadata.io/d-operators/types/command"
)

// func TestReconcilerInitChildJobDetails(t *testing.T) {
// 	var tests = map[string]struct {
// 		ChildJob                *unstructured.Unstructured
// 		ExpectChildJob          bool
// 		ExpectCompletedChildJob bool
// 		IsError                 bool
// 	}{
// 		"no child job": {},
// 		"empty child job": {
// 			ChildJob: &unstructured.Unstructured{},
// 		},
// 		"invalid child job": {
// 			ChildJob: &unstructured.Unstructured{
// 				Object: map[string]interface{}{},
// 			},
// 			IsError: true,
// 		},
// 		"invalid child kind": {
// 			ChildJob: &unstructured.Unstructured{
// 				Object: map[string]interface{}{
// 					"kind":       "MyJob",
// 					"apiVersion": types.JobAPIVersion,
// 				},
// 			},
// 			IsError: true,
// 		},
// 		"invalid child api version": {
// 			ChildJob: &unstructured.Unstructured{
// 				Object: map[string]interface{}{
// 					"kind":       types.KindJob,
// 					"apiVersion": "v12",
// 				},
// 			},
// 			IsError: true,
// 		},
// 		"valid child kind & apiversion": {
// 			ChildJob: &unstructured.Unstructured{
// 				Object: map[string]interface{}{
// 					"kind":       types.KindJob,
// 					"apiVersion": types.JobAPIVersion,
// 				},
// 			},
// 			ExpectChildJob: true,
// 		},
// 		"valid and completed child job": {
// 			ChildJob: &unstructured.Unstructured{
// 				Object: map[string]interface{}{
// 					"kind":       types.KindJob,
// 					"apiVersion": types.JobAPIVersion,
// 					"status": map[string]interface{}{
// 						"phase": types.JobPhaseCompleted,
// 					},
// 				},
// 			},
// 			ExpectChildJob:          true,
// 			ExpectCompletedChildJob: true,
// 		},
// 	}
// 	for name, mock := range tests {
// 		name := name
// 		mock := mock
// 		t.Run(name, func(t *testing.T) {
// 			r := &Reconciliation{
// 				childJob: mock.ChildJob,
// 			}
// 			r.initChildJobDetails()
// 			if mock.IsError && r.err == nil {
// 				t.Fatalf("Expected error got none")
// 			}
// 			if !mock.IsError && r.err != nil {
// 				t.Fatalf("Expected no error got %s", r.err.Error())
// 			}
// 			if mock.IsError {
// 				return
// 			}
// 			if mock.ExpectChildJob != r.isChildJobFound {
// 				t.Fatalf(
// 					"Expected child job %t got %t",
// 					mock.ExpectChildJob,
// 					r.isChildJobFound,
// 				)
// 			}
// 			if mock.ExpectCompletedChildJob != r.isChildJobCompleted {
// 				t.Fatalf(
// 					"Expected child job as completed %t got %t",
// 					mock.ExpectCompletedChildJob,
// 					r.isChildJobCompleted,
// 				)
// 			}
// 		})
// 	}
// }

func TestReconcilerInitCommandDetails(t *testing.T) {
	var tests = map[string]struct {
		Command               types.Command
		ExpectCompletedStatus bool
		ExpectErrorStatus     bool
		ExpectRunOnce         bool
		ExpectRunAlways       bool
		ExpectRunNever        bool
		ExpectRetryOnError    bool
		ExpectRetryOnTimeout  bool
	}{
		"empty command": {
			Command:       types.Command{},
			ExpectRunOnce: true,
		},
		"completed command": {
			Command: types.Command{
				Status: types.CommandStatus{
					Phase: types.CommandPhaseCompleted,
				},
			},
			ExpectCompletedStatus: true,
			ExpectRunOnce:         true,
		},
		"completed & run once command": {
			Command: types.Command{
				Spec: types.CommandSpec{
					Enabled: types.Enabled{
						When: types.EnabledOnce,
					},
				},
				Status: types.CommandStatus{
					Phase: types.CommandPhaseCompleted,
				},
			},
			ExpectCompletedStatus: true,
			ExpectRunOnce:         true,
		},
		"completed & run always command": {
			Command: types.Command{
				Spec: types.CommandSpec{
					Enabled: types.Enabled{
						When: types.EnabledAlways,
					},
				},
				Status: types.CommandStatus{
					Phase: types.CommandPhaseCompleted,
				},
			},
			ExpectCompletedStatus: true,
			ExpectRunAlways:       true,
		},
		"completed & run never command": {
			Command: types.Command{
				Spec: types.CommandSpec{
					Enabled: types.Enabled{
						When: types.EnabledNever,
					},
				},
				Status: types.CommandStatus{
					Phase: types.CommandPhaseCompleted,
				},
			},
			ExpectCompletedStatus: true,
			ExpectRunNever:        true,
		},
		"error-ed command": {
			Command: types.Command{
				Status: types.CommandStatus{
					Phase: types.CommandPhaseError,
				},
			},
			ExpectErrorStatus: true,
			ExpectRunOnce:     true,
		},
		"error-ed & run once command": {
			Command: types.Command{
				Spec: types.CommandSpec{
					Enabled: types.Enabled{
						When: types.EnabledOnce,
					},
				},
				Status: types.CommandStatus{
					Phase: types.CommandPhaseError,
				},
			},
			ExpectErrorStatus: true,
			ExpectRunOnce:     true,
		},
		"error-ed & run always command": {
			Command: types.Command{
				Spec: types.CommandSpec{
					Enabled: types.Enabled{
						When: types.EnabledAlways,
					},
				},
				Status: types.CommandStatus{
					Phase: types.CommandPhaseError,
				},
			},
			ExpectErrorStatus: true,
			ExpectRunAlways:   true,
		},
		"error-ed & run never command": {
			Command: types.Command{
				Spec: types.CommandSpec{
					Enabled: types.Enabled{
						When: types.EnabledNever,
					},
				},
				Status: types.CommandStatus{
					Phase: types.CommandPhaseError,
				},
			},
			ExpectErrorStatus: true,
			ExpectRunNever:    true,
		},
		"run once command": {
			Command: types.Command{
				Spec: types.CommandSpec{
					Enabled: types.Enabled{
						When: types.EnabledOnce,
					},
				},
			},
			ExpectRunOnce: true,
		},
		"run always command": {
			Command: types.Command{
				Spec: types.CommandSpec{
					Enabled: types.Enabled{
						When: types.EnabledAlways,
					},
				},
			},
			ExpectRunAlways: true,
		},
		"run never command": {
			Command: types.Command{
				Spec: types.CommandSpec{
					Enabled: types.Enabled{
						When: types.EnabledNever,
					},
				},
			},
			ExpectRunNever: true,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			r := &Reconciliation{
				command: mock.Command,
			}
			r.initCommandDetails()
			if mock.ExpectCompletedStatus != r.isStatusSetAsCompleted {
				t.Fatalf(
					"Expected completed status %t got %t",
					mock.ExpectCompletedStatus,
					r.isStatusSetAsCompleted,
				)
			}
			if mock.ExpectErrorStatus != r.isStatusSetAsError {
				t.Fatalf(
					"Expected error status %t got %t",
					mock.ExpectErrorStatus,
					r.isStatusSetAsError,
				)
			}
			if mock.ExpectRunOnce != r.isRunOnce {
				t.Fatalf(
					"Expected run once %t got %t",
					mock.ExpectRunOnce,
					r.isRunOnce,
				)
			}
			if mock.ExpectRunAlways != r.isRunAlways {
				t.Fatalf(
					"Expected run always %t got %t",
					mock.ExpectRunAlways,
					r.isRunAlways,
				)
			}
			if mock.ExpectRunNever != r.isRunNever {
				t.Fatalf(
					"Expected run never %t got %t",
					mock.ExpectRunNever,
					r.isRunNever,
				)
			}
		})
	}
}

func TestReconcilerInitLocking(t *testing.T) {
	var tests = map[string]struct {
		Command types.Command
		IsError bool
	}{
		"no command": {},
		"with command": {
			Command: types.Command{
				ObjectMeta: v1.ObjectMeta{
					Name:      "my-cmd",
					Namespace: "ns",
				},
				Spec: types.CommandSpec{
					Enabled: types.Enabled{
						When: types.EnabledNever,
					},
				},
			},
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			// initialize with mock details
			metac.KubeDetails = &server.KubeDetails{
				Config: &rest.Config{},
				GetMetacAPIDiscovery: func() *discovery.APIResourceDiscovery {
					return nil
				},
			}
			r := &Reconciliation{
				command: mock.Command,
			}
			r.initLocking()
			if mock.IsError && r.err == nil {
				t.Fatalf("Expected error got none")
			}
			if !mock.IsError && r.err != nil {
				t.Fatalf("Expected no error got %s", r.err.Error())
			}
		})
	}
}
