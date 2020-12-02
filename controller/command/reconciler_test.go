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
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"mayadata.io/d-operators/common/controller"
	"mayadata.io/d-operators/common/pointer"
	types "mayadata.io/d-operators/types/command"
	"openebs.io/metac/controller/generic"
)

func TestReconcilerEval(t *testing.T) {
	var tests = map[string]struct {
		Watch   *unstructured.Unstructured
		IsError bool
	}{
		// "Nil watch": {
		// 	IsError: true,
		// },
		// "Nil watch object": {
		// 	Watch:   &unstructured.Unstructured{},
		// 	IsError: true,
		// },
		"Empty watch": {
			Watch: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name":      "test",
						"namespace": "test",
					},
				},
			},
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			r := &Reconciler{
				Reconciler: controller.Reconciler{
					HookRequest: &generic.SyncHookRequest{
						Watch: mock.Watch,
					},
				},
			}
			r.eval()
			if mock.IsError && r.Err == nil {
				t.Fatalf("Expected error got none")
			}
			if !mock.IsError && r.Err != nil {
				t.Fatalf("Expected no error got %s", r.Err.Error())
			}
		})
	}
}

func TestReconcilerSetResyncInterval(t *testing.T) {
	var tests = map[string]struct {
		ResyncInterval        *int64
		OnErrorResyncInterval *int64
		Error                 error
		ExpectedInterval      float64
	}{
		"no resync interval + no error": {
			ExpectedInterval: 0,
		},
		"no resync interval + with error": {
			Error:            fmt.Errorf("some err"),
			ExpectedInterval: 0,
		},
		"no error": {
			ResyncInterval:   pointer.Int64(2),
			ExpectedInterval: 2,
		},
		"with error": {
			ResyncInterval:        pointer.Int64(3),
			OnErrorResyncInterval: pointer.Int64(3),
			Error:                 fmt.Errorf("some err"),
			ExpectedInterval:      3,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			r := &Reconciler{
				observedCommand: types.Command{
					Spec: types.CommandSpec{
						Resync: types.Resync{
							IntervalInSeconds:      mock.ResyncInterval,
							OnErrorResyncInSeconds: mock.OnErrorResyncInterval,
						},
					},
				},
				Reconciler: controller.Reconciler{
					Err:          mock.Error,
					HookResponse: &generic.SyncHookResponse{},
				},
			}
			r.setResyncInterval()
			if mock.ExpectedInterval != r.HookResponse.ResyncAfterSeconds {
				t.Fatalf(
					"Expected resync interval %f got %f",
					mock.ExpectedInterval,
					r.HookResponse.ResyncAfterSeconds,
				)
			}
		})
	}
}

func TestReconcilerSetWatchAttributes(t *testing.T) {
	var tests = map[string]struct {
		Status           types.CommandStatus
		Error            error
		IsWatchStatusSet bool
		IsWatchLabelSet  bool
	}{
		"no status no error": {},
		"with skipped status no error": {
			Status: types.CommandStatus{
				Phase: types.CommandPhaseSkipped,
			},
			IsWatchStatusSet: true,
			IsWatchLabelSet:  true,
		},
		"with completed status no error": {
			Status: types.CommandStatus{
				Phase: types.CommandPhaseCompleted,
			},
		},
		"with no status + error": {
			Error:            fmt.Errorf("some err"),
			IsWatchLabelSet:  true,
			IsWatchStatusSet: true,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			r := &Reconciler{
				status: mock.Status,
				Reconciler: controller.Reconciler{
					Err:          mock.Error,
					HookResponse: &generic.SyncHookResponse{},
				},
			}
			r.setWatchAttributes()
			if mock.IsWatchLabelSet && len(r.HookResponse.Labels) == 0 {
				t.Fatalf("Expected watch labels got none")
			}
			if !mock.IsWatchLabelSet && len(r.HookResponse.Labels) != 0 {
				t.Fatalf("Expected no watch labels got %v", r.HookResponse.Labels)
			}
			if mock.IsWatchStatusSet && len(r.HookResponse.Status) == 0 {
				t.Fatalf("Expected watch status got none")
			}
			if !mock.IsWatchStatusSet && len(r.HookResponse.Status) != 0 {
				t.Fatalf("Expected no watch status got %v", r.HookResponse.Status)
			}
		})
	}
}
