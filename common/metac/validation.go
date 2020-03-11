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

package metac

import (
	"github.com/pkg/errors"
	"openebs.io/metac/controller/generic"
)

// ValidateGenericControllerArgs validates the given request &
// response
func ValidateGenericControllerArgs(
	request *generic.SyncHookRequest,
	response *generic.SyncHookResponse,
) error {
	if request == nil {
		return errors.Errorf("Invalid gctl args: Nil request")
	}
	if request.Watch == nil || request.Watch.Object == nil {
		return errors.Errorf("Invalid gctl args: Nil watch")
	}
	if response == nil {
		return errors.Errorf("Invalid gctl args: Nil response")
	}
	// this is a valid request & response
	return nil
}
