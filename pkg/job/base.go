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

package job

import types "mayadata.io/d-operators/types/job"

// BaseRunner is the common runner used by all action runners
type BaseRunner struct {
	*Fixture
	TaskIndex    int
	TaskName     string
	Retry        *Retryable
	FailFastRule types.FailFastRule
}

// IsFailFastOnDiscoveryError returns true if logic that leads to
// discovery error should not be re-tried
func (r *BaseRunner) IsFailFastOnDiscoveryError() bool {
	return r.FailFastRule == types.FailFastOnDiscoveryError
}

// IsFailFastOnError returns true if logic that lead to given error
// should not be re-tried
func (r *BaseRunner) IsFailFastOnError(err error) bool {
	if _, discoveryErr := err.(*DiscoveryError); discoveryErr {
		return r.IsFailFastOnDiscoveryError()
	}
	return false
}
