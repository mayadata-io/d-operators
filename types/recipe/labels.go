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

package types

const (
	// LblKeyIsRecipeLock is the label key to determine if the
	// resource is used to lock the reconciliation of Recipe
	// resource.
	//
	// NOTE:
	// 	This is used to execute reconciliation by only one
	// controller goroutine at a time.
	//
	// NOTE:
	//	A ConfigMap is used as a lock to achieve above behaviour.
	// This ConfigMap will have its labels set with this label key.
	LblKeyIsRecipeLock string = "recipe.dope.mayadata.io/lock"

	// LblKeyRecipeName is the label key to determine the name
	// of the Recipe that the current resource is associated to
	LblKeyRecipeName string = "recipe.dope.mayadata.io/name"

	// LblKeyRecipePhase is the label key to determine the phase
	// of the Recipe. This offers an additional way to determine
	// the phase of the Recipe apart from Recipe's status.phase
	// field.
	LblKeyRecipePhase string = "recipe.dope.mayadata.io/phase"
)
