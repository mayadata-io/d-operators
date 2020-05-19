package runner

import (
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
	k8s "openebs.io/metac/third_party/kubernetes"
)

// SetupConfigs defines a list of test setup files
type SetupConfigs []unstructured.Unstructured

// SetupConfigsLoader loads test setup files
type SetupConfigsLoader struct {
	Path string
}

// Load loads all test setup files & converts them
// to unstructured instances
func (l *SetupConfigsLoader) Load() (SetupConfigs, error) {
	klog.V(2).Infof(
		"Will load test setup config(s) from path %s",
		l.Path,
	)

	files, readDirErr := ioutil.ReadDir(l.Path)
	if readDirErr != nil {
		return nil, readDirErr
	}

	if len(files) == 0 {
		return nil, errors.Errorf(
			"No test setup config(s) found at %s",
			l.Path,
		)
	}

	var out SetupConfigs

	// there can be multiple config files
	for _, file := range files {
		fileName := file.Name()
		if file.IsDir() || file.Mode().IsDir() {
			klog.V(3).Infof(
				"Will skip test setup config %s at path %s: Not a file",
				fileName,
				l.Path,
			)
			// we don't want to load directory
			continue
		}
		if !strings.HasSuffix(fileName, ".yaml") && !strings.HasSuffix(fileName, ".json") {
			klog.V(3).Infof(
				"Will skip test setup config %s at path %s: Not yaml or json",
				fileName,
				l.Path,
			)
			// we support either proper yaml or json file only
			continue
		}

		fileNameWithPath := l.Path + fileName

		contents, readFileErr := ioutil.ReadFile(fileNameWithPath)
		if readFileErr != nil {
			return nil, errors.Wrapf(
				readFileErr,
				"Failed to read test setup config %s",
				fileNameWithPath,
			)
		}

		ul, loaderr := k8s.YAMLToUnstructuredSlice(contents)
		if loaderr != nil {
			loaderr = errors.Wrapf(
				loaderr,
				"Failed to load test setup config %s",
				fileNameWithPath,
			)
			return nil, loaderr
		}

		klog.V(2).Infof(
			"Test setup config %s loaded successfully",
			fileNameWithPath,
		)
		out = append(out, ul...)
	}

	klog.V(2).Infof(
		"Test setup config(s) loaded successfully from path %s",
		l.Path,
	)
	return out, nil
}
