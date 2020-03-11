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

package main

import (
	"openebs.io/metac/controller/generic"
	"openebs.io/metac/start"

	blockdeviceset "mayadata.io/d-operators/controller/blockdevice/set"
	directorhttp "mayadata.io/d-operators/controller/director/http"
	"mayadata.io/d-operators/controller/http"
)

// main function is the entry point of this binary.
//
// This registers various controller (i.e. kubernetes reconciler)
// handler functions. Each handler function gets triggered due
// to any changes (add, update or delete) to configured watch
// resource.
//
// NOTE:
// 	These functions will also be triggered in case this binary
// gets deployed or redeployed (due to restarts, etc.).
//
// NOTE:
//	One can consider each registered function as an independent
// kubernetes controller & this project as the operator.
func main() {
	generic.AddToInlineRegistry("sync/blockdeviceset", blockdeviceset.Sync)
	generic.AddToInlineRegistry("sync/directorhttp", directorhttp.Sync)
	generic.AddToInlineRegistry("sync/http", http.Sync)

	start.Start()
}
