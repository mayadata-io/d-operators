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

package command

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	types "mayadata.io/d-operators/types/command"
	dynamicapply "openebs.io/metac/dynamic/apply"
)

// JobBuildingConfig helps create new instances of JobBuilding
type JobBuildingConfig struct {
	Command types.Command
}

// JobBuilding builds Kubernetes Job resource
type JobBuilding struct {
	Command types.Command
}

// NewJobBuilder returns a new instance of JobBuilding
func NewJobBuilder(config JobBuildingConfig) *JobBuilding {
	return &JobBuilding{
		Command: config.Command,
	}
}

func (b *JobBuilding) getDefaultJob() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       types.KindJob,
			"apiVersion": types.JobAPIVersion,
			"metadata": map[string]interface{}{
				"name":      b.Command.GetName(),
				"namespace": b.Command.GetNamespace(),
				"labels": map[string]interface{}{
					types.LblKeyCommandIsController: "true",
					types.LblKeyCommandName:         b.Command.GetName(),
					types.LblKeyCommandUID:          string(b.Command.GetUID()),
				},
			},
			"spec": map[string]interface{}{
				"ttlSecondsAfterFinished": int64(0),
				"template": map[string]interface{}{
					"spec": map[string]interface{}{
						"restartPolicy":      "Never",
						"backoffLimit":       int64(0),
						"serviceAccountName": os.Getenv("DOPE_SERVICE_ACCOUNT"),
						"containers": []interface{}{
							map[string]interface{}{
								"name":            "dcmd",
								"image":           "mayadataio/dcmd",
								"imagePullPolicy": "Always",
								"command": []interface{}{
									"/usr/bin/dcmd",
								},
								"args": []interface{}{
									"-v=1",
									fmt.Sprintf("--command-name=%s", b.Command.GetName()),
									fmt.Sprintf("--command-ns=%s", b.Command.GetNamespace()),
								},
							},
						},
						"imagePullSecrets": []interface{}{
							map[string]interface{}{
								"name": "mayadataio-cred",
							},
						},
					},
				},
			},
		},
	}
}

// Build returns the final job specifications
//
// NOTE:
//	This Job uses image capable of running commands
// specified in the Command resource specs.
func (b *JobBuilding) Build() (*unstructured.Unstructured, error) {
	var defaultJob = b.getDefaultJob()
	// Start off by initialising final Job to default
	final := defaultJob
	if b.Command.Spec.Template.Job != nil {
		// Job specs found in Command is the desired
		desired := b.Command.Spec.Template.Job
		// NOTE:
		// - desired Job spec must use Job kind & api version
		desired.SetKind(types.KindJob)
		desired.SetAPIVersion(types.JobAPIVersion)
		// NOTE:
		// - desired Job spec must use Command name & namespace
		desired.SetName(b.Command.GetName())
		desired.SetNamespace(b.Command.GetNamespace())
		// All other fields will be a 3-way merge between
		// default specifications & desired specifications
		finalObj, err := dynamicapply.Merge(
			defaultJob.UnstructuredContent(), // observed = default
			desired.UnstructuredContent(),    // last applied = desired
			desired.UnstructuredContent(),    // desired
		)
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"Failed to build job spec: 3-way merge failed",
			)
		}
		// Reset the final specifications
		final.Object = finalObj
	}
	return final, nil
}
