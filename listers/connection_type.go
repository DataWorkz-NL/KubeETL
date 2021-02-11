package listers

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/dataworkz/kubeetl/api/v1alpha1"
)

//ConnectionTypeLister lists and finds ConnectionTypes
type ConnectionTypeLister interface {
	List(ctx context.Context, namespace string) (*v1alpha1.ConnectionTypeList, error)
	Find(ctx context.Context, namespace string, conType string) (*v1alpha1.ConnectionType, error)
}

type connectionTypeLister struct {
	client client.Client
}

func NewConnectionTypeLister(client client.Client) ConnectionTypeLister {
	return &connectionTypeLister{
		client: client,
	}
}

// List returns a ConnectionTypeList in the given namespace.
func (l *connectionTypeLister) List(ctx context.Context, namespace string) (*v1alpha1.ConnectionTypeList, error) {
	typeList := &v1alpha1.ConnectionTypeList{}
	if err := l.client.List(ctx, typeList, &client.ListOptions{Namespace: namespace}); err != nil {
		return nil, fmt.Errorf("unable to list ConnectionTypes: %w", err)
	}

	return typeList, nil
}

func (l *connectionTypeLister) Find(ctx context.Context, namespace string, conType string) (*v1alpha1.ConnectionType, error) {
	typeList, err := l.List(ctx, namespace)
	if err != nil {
		return nil, err
	}

	var res *v1alpha1.ConnectionType
	for _, ct := range typeList.Items {
		if ct.Name == conType {
			res = &ct
		}
	}

	return res, nil
}
