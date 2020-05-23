package common

import (
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// CMDArgs defines a command along with its arguments
type CMDArgs struct {
	CMD  string
	Args []string
}

// UnstructList defines an array of unstructured instances
type UnstructList []unstructured.Unstructured

// ToTyped transforms the provided unstruct instance
// to target type
func ToTyped(src *unstructured.Unstructured, target interface{}) error {
	if src == nil || src.Object == nil {
		return errors.Errorf(
			"Can't transform unstruct to typed: Nil unstruct content",
		)
	}
	if target == nil {
		return errors.Errorf(
			"Can't transform unstruct to typed: Nil target",
		)
	}
	return runtime.DefaultUnstructuredConverter.FromUnstructured(
		src.UnstructuredContent(),
		target,
	)
}
