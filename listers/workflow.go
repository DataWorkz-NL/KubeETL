package listers

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/dataworkz/kubeetl/api/v1alpha1"
)

//WorkflowLister lists and finds Workflows
type WorkflowLister interface {
	List(ctx context.Context, namespace string) (*v1alpha1.WorkflowList, error)
	Find(ctx context.Context, namespace string, name string) (*v1alpha1.Workflow, error)
}

type workflowLister struct {
	client client.Client
}

func NewWorkflowLister(client client.Client) WorkflowLister {
	return &workflowLister{
		client: client,
	}
}

// List returns a WorkflowList in the given namespace.
func (l *workflowLister) List(ctx context.Context, namespace string) (*v1alpha1.WorkflowList, error) {
	wfList := &v1alpha1.WorkflowList{}
	if err := l.client.List(ctx, wfList, &client.ListOptions{Namespace: namespace}); err != nil {
		return nil, fmt.Errorf("unable to list Workflows: %w", err)
	}

	return wfList, nil
}

func (l *workflowLister) Find(ctx context.Context, namespace string, name string) (*v1alpha1.Workflow, error) {
	wfList, err := l.List(ctx, namespace)
	if err != nil {
		return nil, err
	}

	var res *v1alpha1.Workflow
	for _, wf := range wfList.Items {
		if wf.Name == name {
			res = &wf
		}
	}

	return res, nil
}
