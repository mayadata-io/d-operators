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

package recipe

import (
	"testing"

	gy "github.com/ghodss/yaml"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var crd = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: mayastorpools.openebs.io
spec:
  group: openebs.io
  versions:
    - name: v1alpha1
      served: true
      storage: true
      subresources:
        status: {}
      schema:
        openAPIV3Schema:
          type: object
          properties:
            apiVersion:
              type: string
            kind:
              type: string
            metadata:
              type: object
            spec:
              description: Specification of the mayastor pool.
              type: object
              required:
              - node
              - disks
              properties:
                node:
                  description: Name of the k8s node where the storage pool is located.
                  type: string
                disks:
                  description: Disk devices (paths or URIs) that should be used for the pool.
                  type: array
                  items:
                    type: string
            status:
              description: Status part updated by the pool controller.
              type: object
              properties:
                state:
                  description: Pool state.
                  type: string
                reason:
                  description: Reason for the pool state value if applicable.
                  type: string
                disks:
                  description: Disk device URIs that are actually used for the pool.
                  type: array
                  items:
                    type: string
                capacity:
                  description: Capacity of the pool in bytes.
                  type: integer
                  format: int64
                  minimum: 0
                used:
                  description: How many bytes are used in the pool.
                  type: integer
                  format: int64
                  minimum: 0
      additionalPrinterColumns:
      - name: Node
        type: string
        description: Node where the storage pool is located
        jsonPath: .spec.node
      - name: State
        type: string
        description: State of the storage pool
        jsonPath: .status.state
      - name: Age
        type: date
        jsonPath: .metadata.creationTimestamp
  scope: Namespaced
  names:
    kind: MayastorPool
    listKind: MayastorPoolList
    plural: mayastorpools
    singular: mayastorpool
    shortNames: ["msp"]
`

func TestMSPCRDV1FromStringUnmarshal(t *testing.T) {
	var unstructObj unstructured.Unstructured
	err := gy.Unmarshal([]byte(crd), &unstructObj)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	var crdObj v1.CustomResourceDefinition
	err = UnstructToTyped(&unstructObj, &crdObj)
	if err == nil {
		t.Fatal(
			"Convert to CRD type: Expected error got none",
		)
	}
}
