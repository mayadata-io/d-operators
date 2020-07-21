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
	"flag"

	"k8s.io/klog/v2"
	"openebs.io/metac/controller/generic"
	"openebs.io/metac/start"

	"mayadata.io/d-operators/controller/doperator"
	"mayadata.io/d-operators/controller/http"
	"mayadata.io/d-operators/controller/recipe"
	"mayadata.io/d-operators/controller/run"
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
	flag.Set("alsologtostderr", "true")
	flag.Parse()

	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(klogFlags)

	// Sync the glog and klog flags.
	flag.CommandLine.VisitAll(func(f1 *flag.Flag) {
		f2 := klogFlags.Lookup(f1.Name)
		if f2 != nil {
			value := f1.Value.String()
			f2.Value.Set(value)
		}
	})
	defer klog.Flush()

	// controller name & corresponding controller reconcile function
	var controllers = map[string]generic.InlineInvokeFn{
		"sync/recipe":    recipe.Sync,
		"sync/http":      http.Sync,
		"sync/doperator": doperator.Sync,
		"sync/run":       run.Sync,
	}
	for name, ctrl := range controllers {
		generic.AddToInlineRegistry(name, ctrl)
	}
	start.Start()
}
