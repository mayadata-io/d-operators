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
	"openebs.io/metac/controller/generic"
)

var (
	defaultDeletionResyncTime = float64(30)
)

// Finalize implements the idempotent logic that gets executed when
// Command instance is deleted. A Command instance may have child job &
// dedicated lock in form of a ConfigMap.
// Finalize logic tries to delete child pod, job & ConfigMap
//
// NOTE:
// 	When finalize hook is set in the config metac automatically sets
// a finalizer entry against the Command metadata's finalizers field .
// This finalizer entry is removed when SyncHookResponse's Finalized
// field is set to true
//
// NOTE:
// 	SyncHookRequest is the payload received as part of finalize
// request. Similarly, SyncHookResponse is the payload sent as a
// response as part of finalize request.
//
// NOTE:
//	Returning error will panic this process. We would rather want this
// controller to run continuously. Hence, the errors are handled.
func Finalize(request *generic.SyncHookRequest, response *generic.SyncHookResponse) error {
	if request.Attachments.IsEmpty() {
		// Since no Dependents found it is safe to delete Command
		response.Finalized = true
		return nil
	}

	response.ResyncAfterSeconds = defaultDeletionResyncTime
	// Observed attachments will get deleted
	response.ExplicitDeletes = request.Attachments.List()
	return nil
}
