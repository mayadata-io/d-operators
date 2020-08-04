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
	"openebs.io/metac/controller/generic"
)

// Finalize implements the idempotent logic that gets executed when
// Recipe instance is deleted. A Recipe instance is associated with
// a dedicated lock in form of a ConfigMap. Finalize logic tries to
// delete this ConfigMap.
//
// NOTE:
// 	When finalize hook is set in the config metac automatically sets
// a finalizer entry against the Recipe metadata's finalizers field .
// This finalizer entry is removed when SyncHookResponse's Finalized
// field is set to true.
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
		// Since no ConfigMap is found it is safe to delete Recipe
		response.Finalized = true
		return nil
	}
	// A single ConfigMap instance is expected in the attachments
	// That needs to be deleted explicitly
	response.ExplicitDeletes = request.Attachments.List()
	return nil
}
