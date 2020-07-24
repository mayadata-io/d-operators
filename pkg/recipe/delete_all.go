package recipe

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "mayadata.io/d-operators/types/recipe"
	"openebs.io/metac/dynamic/clientset"
)

func (r *TaskRunner) deleteAll() (*types.TaskStatus, error) {
	var message = fmt.Sprintf(
		"Delete: Resource %s %s: GVK %s",
		r.Task.DeleteAll.State.GetNamespace(),
		r.Task.DeleteAll.State.GetAPIVersion(),
		r.Task.DeleteAll.State.GetKind(),
	)
	var client *clientset.ResourceClient
	var err error
	err = r.Retry.Waitf(
		func() (bool, error) {
			client, err = r.GetClientForAPIVersionAndKind(
				r.Task.DeleteAll.State.GetAPIVersion(),
				r.Task.DeleteAll.State.GetKind(),
			)
			if err != nil {
				return r.IsFailFastOnDiscoveryError(), err
			}
			return true, nil
		},
		message,
	)
	if err != nil {
		return nil, err
	}
	err = client.
		Namespace(r.Task.DeleteAll.State.GetNamespace()).
		DeleteCollection(
			&metav1.DeleteOptions{},
			metav1.ListOptions{},
		)
	if err != nil {
		return nil, err
	}
	return &types.TaskStatus{
		Phase:   types.TaskStatusPassed,
		Message: message,
	}, nil
}
