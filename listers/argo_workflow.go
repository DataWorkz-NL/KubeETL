package listers

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
	wfv1 "github.com/argoproj/argo/v2/pkg/apis/workflow/v1alpha1"
)

//ArgoWorkflowLister lists and finds ArgoWorkflows
type ArgoWorkflowLister interface {
	List(ctx context.Context, namespace string) (*wfv1.WorkflowList, error)
	Find(ctx context.Context, namespace string, name string) (*wfv1.Workflow, error)
}

type argoWorkflowLister struct {
	client client.Client
}

func NewArgoWorkflowLister(client client.Client) ArgoWorkflowLister {
	return &argoWorkflowLister{
		client: client,
	}
}

// List returns a ArgoWorkflowList in the given namespace.
func (l *argoWorkflowLister) List(ctx context.Context, namespace string) (*wfv1.WorkflowList, error) {
	wfList := &wfv1.WorkflowList{}
	if err := l.client.List(ctx, wfList, &client.ListOptions{Namespace: namespace}); err != nil {
		return nil, fmt.Errorf("unable to list ArgoWorkflows: %w", err)
	}

	return wfList, nil
}

func (l *argoWorkflowLister) Find(ctx context.Context, namespace string, name string) (*wfv1.Workflow, error) {
	wfList, err := l.List(ctx, namespace)
	if err != nil {
		return nil, err
	}

	var res *wfv1.Workflow
	for _, wf := range wfList.Items {
		if wf.Name == name {
			res = &wf
		}
	}

	return res, nil
}
