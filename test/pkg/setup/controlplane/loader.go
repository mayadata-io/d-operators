package controlplane

import (
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"mayadata.io/d-operators/test/pkg/common"
	"mayadata.io/d-operators/test/pkg/types"
	k8s "openebs.io/metac/third_party/kubernetes"
)

// SetupList defines a list of setup files
// that defines setting up a Kubernetes ControlPlane
type SetupList []types.ControlPlaneSetup

// UnstructListToSetupList transforms the given unstruct
// list to list of ControlPlaneSetups
func UnstructListToSetupList(ul common.UnstructList) (SetupList, error) {
	var out SetupList
	for _, unstruct := range ul {
		var cp types.ControlPlaneSetup
		err := common.ToTyped(&unstruct, &cp)
		if err != nil {
			return nil, err
		}
		out = append(out, cp)
	}
	return out, nil
}

// SetupLoader loads ControlPlaneSetup files
type SetupLoader struct {
	common.SetupLoader
}

// Load loads all test setup files & converts them
// to unstructured instances
func (l *SetupLoader) Load() (SetupList, error) {
	klog.V(2).Infof(
		"Will load control plane setup(s) from path %s",
		l.Path,
	)
	var out common.UnstructList
	// load the control plane setups by passing in a populator
	l.SetupLoader.Load(func(content []byte) error {
		ul, err := k8s.YAMLToUnstructuredSlice(content)
		if err != nil {
			return errors.Wrapf(
				err,
				"Failed to load ControlPlan setup",
			)
		}
		out = append(out, ul...)
		return nil
	})
	return UnstructListToSetupList(out)
}

// LoadControlPlaneSetups loads all control plane setups
// found at a specific path
func LoadControlPlaneSetups() (SetupList, error) {
	l := &SetupLoader{
		common.SetupLoader{
			Path: DefaultControlPlaneSetupFilePath,
		},
	}
	return l.Load()
}
