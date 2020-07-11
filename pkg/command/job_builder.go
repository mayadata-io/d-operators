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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	types "mayadata.io/d-operators/types/command"
	dynamicapply "openebs.io/metac/dynamic/apply"
)

// JobBuilderConfig helps create new instances of JobBuilder
type JobBuilderConfig struct {
	Command types.Command
}

// JobBuilder builds Kubernetes Job that runs commands
type JobBuilder struct {
	Command types.Command
}

// NewJobBuilder returns a new instance of JobBuilder
func NewJobBuilder(config JobBuilderConfig) *JobBuilder {
	return &JobBuilder{
		Command: config.Command,
	}
}

func (b *JobBuilder) getDefaultJob() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       "Job",
			"apiVersion": "batch/v1",
			"metadata": map[string]interface{}{
				"name":      b.Command.GetName(),
				"namespace": b.Command.GetNamespace(),
				"labels": map[string]interface{}{
					"command.dope.metacontroller.io/controller":     "true",
					"command.dope.metacontroller.io/controller-uid": string(b.Command.GetUID()),
				},
			},
			"spec": map[string]interface{}{
				"ttlSecondsAfterFinished": 0,
				"template": map[string]interface{}{
					"spec": map[string]interface{}{
						"restartPolicy": "Never",
						"backoffLimit":  0,
						"containers": []interface{}{
							map[string]interface{}{
								"name":  "commander",
								"image": "mayadataio/dope-commander",
								"command": []interface{}{
									"/usr/bin/dope-commander",
								},
								"args": []interface{}{
									"--logtostderr",
									"--run-as-local",
									"-v=1",
								},
							},
						},
					},
				},
			},
		},
	}
}

// Build returns the final job specifications that in turn will run
// the commands
func (b *JobBuilder) Build() (*unstructured.Unstructured, error) {
	var final = b.getDefaultJob()
	if b.Command.Spec.Template != nil && b.Command.Spec.Template.Job != nil {
		desired := b.Command.Spec.Template.Job
		// desired spec must use the command name & namespace
		desired.SetName(b.Command.GetName())
		desired.SetNamespace(b.Command.GetNamespace())
		// 3-way merge
		finalObj, err := dynamicapply.Merge(
			final.UnstructuredContent(),   // observed
			desired.UnstructuredContent(), // last applied
			desired.UnstructuredContent(), // desired
		)
		if err != nil {
			return nil, err
		}
		final.Object = finalObj
	}
	return final, nil
}
