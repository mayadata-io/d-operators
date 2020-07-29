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

package gvk

const (
	// KindHTTP represents HTTP custom resource
	KindHTTP string = "HTTP"

	// KindHTTPData represents HTTPData custom resource
	KindHTTPData string = "HTTPData"

	// KindDirectorHTTP represents DirectorHTTP custom resource
	KindDirectorHTTP string = "DirectorHTTP"
)

const (
	// KindRecipe represents Recipe custom resource
	KindRecipe string = "Recipe"

	// APIVersionRecipe represent Recipe custom resource's api version
	APIVersionRecipe string = "dope.mayadata.io/v1"
)

const (
	// APIExtensionsK8sIOV1Beta1 represents apiextensions.k8s.io
	// as group & v1beta1 as version
	APIExtensionsK8sIOV1Beta1 string = "apiextensions.k8s.io/v1beta1"

	// GroupDAOMayadataIO represents dao.mayadata.io as
	// group
	GroupDAOMayadataIO string = "dao.mayadata.io"

	// VersionV1Alpha1 represents v1alpha1 version
	VersionV1Alpha1 string = "v1alpha1"

	// DAOMayadataIOV1Alpha1 represents
	// dao.mayadata.io as group & v1alpha1 as version
	DAOMayadataIOV1Alpha1 string = GroupDAOMayadataIO + "/" + VersionV1Alpha1
)
