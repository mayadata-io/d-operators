package controlplane

import (
	"bytes"
	"io"
	"path"

	"mayadata.io/d-operators/test/pkg/controlplane/internal"
)

// KubeCtl is a wrapper around the kubectl binary.
type KubeCtl struct {
	// Path where the kubectl binary can be found.
	//
	// If this is left empty, we will attempt to locate a binary,
	// by checking for the TEST_ASSET_KUBECTL environment variable,
	// and the default test assets directory. See the "Binaries"
	// section above (in doc.go) for details.
	Path string

	// Opts can be used to configure additional flags which will
	// be used each time the wrapped binary is called.
	//
	// For example, you might want to use this to set the URL of
	// the APIServer to connect to.
	Opts []string

	// Out, Err specify where KubeCtl should write its StdOut &
	// StdErr to.
	//
	// If not specified, the output will be discarded in case of
	// no errors or added to error details in case of errors.
	Out io.Writer
	Err io.Writer

	// out & err buffers are used to provide additional details
	// in case of errors. These buffers are used only if Out & Err
	// are not set by the callers of KubeCtl.
	outBuf *bytes.Buffer
	errBuf *bytes.Buffer
}

// Run executes the wrapped binary with some preconfigured options
// and the arguments given to this method.
func (k *KubeCtl) Run(args ...string) error {
	if k.Path == "" {
		k.Path = internal.BinPathFinder("kubectl")
	}
	// add to existing options if any
	allArgs := append(k.Opts, args...)
	cmd := NewCommand(CommandConfig{
		Err: k.Err,
		Out: k.Out,
	})
	return cmd.Run(k.Path, allArgs...)
}

// RunOp executes the wrapped binary with some preconfigured options
// and the arguments given to this method. It returns the ouput of the
// command as well.
func (k *KubeCtl) RunOp(args ...string) (string, error) {
	if k.Path == "" {
		k.Path = internal.BinPathFinder("kubectl")
	}
	// add to existing options if any
	allArgs := append(k.Opts, args...)
	cmd := NewCommand(CommandConfig{
		Err: k.Err,
		Out: k.Out,
	})
	return cmd.RunOp(k.Path, allArgs...)
}

// ApplyConfig holds kubernetes yaml files in a
// pre-determined order
//
// NOTE:
//	This config is used to apply the mentioned
// files using kubectl
type ApplyConfig struct {
	// location of kubernetes yaml files
	Path      string
	YAMLFiles []string
}

// Apply does kubectl apply of kubernetes yaml files
func (k *KubeCtl) Apply(config ApplyConfig) error {
	if len(config.YAMLFiles) == 0 {
		return k.Run(
			"apply",
			"-f",
			config.Path,
		)
	}
	for _, yml := range config.YAMLFiles {
		err := k.Run(
			"apply",
			"-f",
			path.Join(
				config.Path,
				yml,
			),
		)
		if err != nil {
			return err
		}
	}
	return nil
}
