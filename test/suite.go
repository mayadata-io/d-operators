package main

import (
	"flag"

	"k8s.io/klog/v2"
	"mayadata.io/d-operators/test/framework"
)

// ---------------------------------------
// go run suite.go -v=2
// ---------------------------------------
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

	err := framework.Run()
	if err != nil {
		klog.Exitf("%+v", err)
	}
}
