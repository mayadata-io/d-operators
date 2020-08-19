/*
Copyright 2020 The MayaData Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package schema

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

const (
	// InvalidPathMsg provides the invalid path information
	InvalidPathMsg string = "Invalid path: %q"

	// SupportedPathMsg provides all the supported paths
	// information
	SupportedPathMsg string = "Supported paths: %s"
)

// ErrorMessage is used to form the validation failure message
type ErrorMessage struct {
	Error  string `json:"error"`
	Remedy string `json:"remedy,omitempty"`
}

// FieldPathValidation enables validating the fields in
// yaml like structures
type FieldPathValidation struct {
	// Unstructured object whose fields will be validated
	// against its supported schema
	Target map[string]interface{}

	// Nested field paths that are supported by this
	// unstructured object's schema
	//
	// NOTE:
	//	Each field path set here should represent its absolute
	// field path
	SupportedAbsolutePaths []string

	// The path prefixes that can be ignored if it fails
	// validation
	//
	// For Example:
	//	'metadata.labels' & 'metadata.annotations'
	// can have any nested field name(s). In other words,
	// metadata.labels.app is valid & so is metadata.labels.xyz
	// since app as well as xyz are not managed by the schema.
	//
	// NOTE:
	//	Each prefix set here must end with a dot i.e. `.`
	UserAllowedPathPrefixes []string

	// map of field to supported nested paths
	fieldToSupportedPaths map[string][]string

	// all the validation failures
	failures []ErrorMessage

	// additional info that can help in debugging
	verbose []string
}

// FieldPathValidationStatus conveys the status of field path
// based validation
type FieldPathValidationStatus string

const (
	// FieldPathValidationStatusValid conveys a successful validation
	FieldPathValidationStatusValid FieldPathValidationStatus = "Valid"

	// FieldPathValidationStatusInvalid conveys a failed validation
	FieldPathValidationStatusInvalid FieldPathValidationStatus = "Invalid"
)

// FieldPathValidationResult is used to convey the results of
// validation
type FieldPathValidationResult struct {
	Status   FieldPathValidationStatus `json:"status"`
	Failures []ErrorMessage            `json:"failures,omitempty"`
	Verbose  []string                  `json:"verbose,omitempty"`
}

func (v *FieldPathValidationResult) Error() string {
	if len(v.Failures) == 0 {
		return ""
	}
	raw, err := json.MarshalIndent(
		*v,
		" ",
		".",
	)
	if err != nil {
		panic(err)
	}
	return string(raw)
}

// isSupportedPath returns true if the provided field path
// is supported as part of the schema
func (v *FieldPathValidation) isSupportedPath(fieldPath string) bool {
	if fieldPath == "" {
		return false
	}

	// Regex replace the digits with *
	//
	// NOTE:
	//  One example is to replace field path patterns like
	// - spec.employees.[1].name with spec.employees.[*].name
	// - spec.employees.[10].name with spec.employees.[*].name
	re := regexp.MustCompile(`\[([0-9]+)\]`)
	targetFieldPath := re.ReplaceAllString(fieldPath, "[*]")

	ok := func() bool {
		for _, allowedPathPrefix := range v.UserAllowedPathPrefixes {
			if strings.HasPrefix(targetFieldPath, allowedPathPrefix) {
				// return true if any allowed fieldpath prefix is a prefix
				// of target fieldpath
				return true
			}
		}
		for _, absolutePath := range v.SupportedAbsolutePaths {
			// An EXACT match implies logic is evaluating the given
			// fieldpath as the leaf field (i.e. field that does not
			// nest further)
			//
			// _OR_
			//
			// a PREFIX based match, i.e. the given fieldpath is part
			// of the absolute path
			if absolutePath == targetFieldPath ||
				strings.HasPrefix(
					absolutePath,
					fmt.Sprintf("%s.", targetFieldPath), // ends with a dot
				) {
				// return true if target fieldpath matches the prefix of any
				// supported fieldpaths
				return true
			}
		}
		return false
	}()
	if !ok && fieldPath != targetFieldPath {
		v.verbose = append(
			v.verbose,
			fmt.Sprintf(
				"FieldPath %q was regex-ed as %q",
				fieldPath,
				targetFieldPath,
			),
		)
	}

	return ok
}

func (v *FieldPathValidation) getSupportedPathsForField(fieldName string) []string {
	if fieldName == "" {
		return nil
	}
	if v.fieldToSupportedPaths == nil {
		// one time initialization
		v.fieldToSupportedPaths = make(map[string][]string)
		for _, supportedPath := range v.SupportedAbsolutePaths {
			if strings.Contains(supportedPath, fieldName) {
				v.fieldToSupportedPaths[fieldName] = append(
					v.fieldToSupportedPaths[fieldName],
					supportedPath,
				)
			}
		}
	}
	return v.fieldToSupportedPaths[fieldName]
}

func (v *FieldPathValidation) getRemedyMsgForField(fieldPath, fieldName string) string {
	paths := v.getSupportedPathsForField(fieldName)
	if len(paths) == 0 {
		return fmt.Sprintf("Path %q is not part of schema", fieldPath)
	}
	return fmt.Sprintf(SupportedPathMsg, strings.Join(paths, ", "))
}

func (v *FieldPathValidation) validateFieldPaths(fieldPath string) {
	// we are interested only for invalid paths
	if !v.isSupportedPath(fieldPath) {
		fields := strings.Split(fieldPath, ".")
		// Remedey message is built from the last field name
		lastField := fields[len(fields)-1]
		// construct the validation failure message
		v.failures = append(
			v.failures,
			ErrorMessage{
				Error:  fmt.Sprintf(InvalidPathMsg, fieldPath),
				Remedy: v.getRemedyMsgForField(fieldPath, lastField),
			},
		)
	}
}

func (v *FieldPathValidation) isListAMap(list []interface{}) bool {
	if len(list) == 0 {
		// Can not determine the datatype of list item
		// Hence, return false
		return false
	}
	// Loop over the given list
	for _, item := range list {
		// verify if all items are of type map
		_, ok := item.(map[string]interface{})
		if !ok {
			// No need to proceed since this is not a
			// list of maps
			return false
		}
	}
	return true
}

// makeMapFromList converts the given list to its map equivalent
//
// NOTE:
// 	The provided list of items is converted to a map where
// each map pair is keyed by the list item's index
//
// For example, given the following list:
//
// - name: dope1
//   age: 1
// - name: dope2
//   age: 2
//
// Above list gets converted into below map:
//
// [0]:
//   name: dope1
//   age: 1
// [1]:
//   name: dope2
//   age: 2
func (v *FieldPathValidation) makeMapFromList(list []interface{}) map[string]interface{} {
	if len(list) == 0 {
		return nil
	}
	result := make(map[string]interface{}, len(list))
	for idx, item := range list {
		result[fmt.Sprintf("[%d]", idx)] = item
	}
	return result
}

func (v *FieldPathValidation) validateFieldPathsOfMap(fieldPath string, given map[string]interface{}) {
	var pathPrefix string
	if fieldPath != "" {
		pathPrefix = fmt.Sprintf("%s.", fieldPath)
	}
	for key, val := range given {
		// Nested path is formed by placing a dot i.e. '.' with
		// every field traversed
		//
		// NOTE:
		//	Dot is placed only if provided fieldPath is not empty
		newPath := fmt.Sprintf("%s%s", pathPrefix, key)
		// This leads to recursive calls for nested field path
		v.validate(newPath, val)
	}
}

func (v *FieldPathValidation) validateFieldPathsOfListViaMap(fieldPath string, given []interface{}) {
	// transform the list to map; list items are
	// transformed into a map with its key(s) as the
	// item's index
	givenAsMap := v.makeMapFromList(given)
	// once in map execute map based validation
	v.validateFieldPathsOfMap(fieldPath, givenAsMap)
}

func (v *FieldPathValidation) validateFieldPathsOfArray(fieldPath string, given []interface{}) {
	// If it looks like a list of map, validate list like a map
	if v.isListAMap(given) {
		v.validateFieldPathsOfListViaMap(fieldPath, given)
		return
	}
	// List is a normal array of scalars
	// Hence, validate the current fieldpath
	v.validateFieldPaths(fieldPath)
}

// validate runs field path validations for specific data types
//
// NOTE:
// It supports following data types:
// - map[string]interface{}
// - []interface{}
// - any scalar i.e. golang primitive type
func (v *FieldPathValidation) validate(fieldPath string, given interface{}) {
	for _, allowedPathPrefix := range v.UserAllowedPathPrefixes {
		// allowed path prefixes should end with dot i.e. '.'
		if !strings.HasSuffix(allowedPathPrefix, ".") {
			v.failures = append(
				v.failures,
				ErrorMessage{
					Error: fmt.Sprintf(
						"Invalid path prefix %q", allowedPathPrefix,
					),
					Remedy: "Path prefix should end with a dot i.e. '.'",
				},
			)
		}
	}
	if len(v.failures) != 0 {
		return
	}
	switch givenVal := given.(type) {
	case map[string]interface{}:
		v.validateFieldPathsOfMap(fieldPath, givenVal)
	case []interface{}:
		v.validateFieldPathsOfArray(fieldPath, givenVal)
	default:
		// given is either a scalar or null.
		//
		// NOTE:
		// - We have traversed to the leaf of the object.
		// - No further traversal needs to be done
		v.validateFieldPaths(fieldPath)
	}
}

// Validate validates the provided object against its supported paths
func (v *FieldPathValidation) Validate() *FieldPathValidationResult {
	v.validate("", v.Target)
	if len(v.failures) == 0 {
		// no validation errors; hence valid schema
		return &FieldPathValidationResult{
			Status: FieldPathValidationStatusValid,
		}
	}
	return &FieldPathValidationResult{
		Status:   FieldPathValidationStatusInvalid,
		Failures: v.failures,
		Verbose:  v.verbose,
	}
}
