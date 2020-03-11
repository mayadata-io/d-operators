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
	"fmt"
	"strings"

	"openebs.io/metac/controller/generic"
)

// GetDetailsFromRequest returns details of provided
// response in string format
func GetDetailsFromRequest(req *generic.SyncHookRequest) string {
	if req == nil {
		return ""
	}
	var details []string
	if req.Watch == nil || req.Watch.Object == nil {
		details = append(
			details,
			"GCtl request watch = nil:",
		)
	} else {
		details = append(
			details,
			fmt.Sprintf(
				"GCtl request watch %q / %q:",
				req.Watch.GetNamespace(),
				req.Watch.GetName()),
		)
	}
	var allKinds map[string]int = map[string]int{}
	if req.Attachments == nil {
		details = append(
			details,
			"GCtl request attachments = nil",
		)
	} else {
		for _, attachment := range req.Attachments.List() {
			if attachment == nil || attachment.Object == nil {
				continue
			}
			kind := attachment.GetKind()
			if kind == "" {
				kind = "NA"
			}
			count := allKinds[kind]
			allKinds[kind] = count + 1
		}
		if len(allKinds) > 0 {
			details = append(
				details,
				"GCtl request attachments",
			)
		}
		for kind, count := range allKinds {
			details = append(
				details,
				fmt.Sprintf("[%s %d]", kind, count),
			)
		}
	}
	return strings.Join(details, " ")
}

// GetDetailsFromResponse returns details of provided
// response in string format
func GetDetailsFromResponse(resp *generic.SyncHookResponse) string {
	if resp == nil || len(resp.Attachments) == 0 {
		return ""
	}
	var allKinds map[string]int = map[string]int{}
	for _, attachment := range resp.Attachments {
		if attachment == nil || attachment.Object == nil {
			continue
		}
		kind := attachment.GetKind()
		if kind == "" {
			kind = "NA"
		}
		count := allKinds[kind]
		allKinds[kind] = count + 1
	}
	var details []string
	details = append(
		details,
		"GCtl response attachments",
	)
	for kind, count := range allKinds {
		details = append(
			details,
			fmt.Sprintf("[%s %d]", kind, count))
	}
	return strings.Join(details, " ")
}
