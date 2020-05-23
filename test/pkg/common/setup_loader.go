package common

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

// SetupLoader loads platform setup files
type SetupLoader struct {
	Path string
}

// Load loads all platform setup files and uses populator
// as a callback to operate against these loaded files.
//
// NOTE:
//	Argument 'populator' is invoked for every valid file
// found in the configured path
func (l *SetupLoader) Load(populator func([]byte) error) error {
	klog.V(3).Infof(
		"Will load & populate setup file(s) at path %q", l.Path,
	)

	files, err := ioutil.ReadDir(l.Path)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return errors.Errorf(
			"No setup files(s) were found at %q", l.Path,
		)
	}

	// there can be multiple config files
	for _, file := range files {
		fileName := file.Name()
		if file.IsDir() || file.Mode().IsDir() {
			klog.V(4).Infof(
				"Will skip setup file %q at path %q: Not a file",
				fileName,
				l.Path,
			)
			// we don't want to load directory
			continue
		}
		if !strings.HasSuffix(fileName, ".yaml") &&
			!strings.HasSuffix(fileName, ".json") {
			klog.V(4).Infof(
				"Will skip setup file %q at path %q: Not yaml or json",
				fileName,
				l.Path,
			)
			continue
		}
		// load the file
		fileNameWithPath := filepath.Join(l.Path, fileName)
		content, err := ioutil.ReadFile(fileNameWithPath)
		if err != nil {
			return errors.Wrapf(
				err,
				"Failed to load setup file %q",
				fileNameWithPath,
			)
		}
		// poluate the loaded content using populator callback
		err = populator(content)
		if err != nil {
			return errors.Wrapf(
				err,
				"Failed to populate from setup file %q",
				fileNameWithPath,
			)
		}
		klog.V(3).Infof(
			"Setup file %q was loaded & populated successfully", fileNameWithPath,
		)
	}

	klog.V(3).Infof(
		"Setup files(s) at path %q were loaded & populated successfully",
		l.Path,
	)
	return nil
}
