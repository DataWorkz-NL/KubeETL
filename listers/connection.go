package listers

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/dataworkz/kubeetl/api/v1alpha1"
)

//ConnectionLister lists and finds Connections
type ConnectionLister interface {
	List(ctx context.Context, namespace string) (*v1alpha1.ConnectionList, error)
	Find(ctx context.Context, namespace string, name string) (*v1alpha1.Connection, error)
}

type connectionLister struct {
	client client.Client
}

func NewConnectionLister(client client.Client) ConnectionLister {
	return &connectionLister{
		client: client,
	}
}

// List returns a ConnectionList in the given namespace.
func (l *connectionLister) List(ctx context.Context, namespace string) (*v1alpha1.ConnectionList, error) {
	connList := &v1alpha1.ConnectionList{}
	if err := l.client.List(ctx, connList, &client.ListOptions{Namespace: namespace}); err != nil {
		return nil, fmt.Errorf("unable to list Connections: %w", err)
	}

	return connList, nil
}

func (l *connectionLister) Find(ctx context.Context, namespace string, name string) (*v1alpha1.Connection, error) {
	conList, err := l.List(ctx, namespace)
	if err != nil {
		return nil, err
	}

	var res *v1alpha1.Connection
	for _, conn := range conList.Items {
		if conn.Name == name {
			res = &conn
		}
	}

	return res, nil
}
