package listers

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/dataworkz/kubeetl/api/v1alpha1"
)

// DataSetLister lists and finds DataSets
type DataSetTypeLister interface {
	List(ctx context.Context, namespace string) (*v1alpha1.DataSetTypeList, error)
	Find(ctx context.Context, namespace string, conType string) (*v1alpha1.DataSetType, error)
}

type dataSetTypeLister struct {
	client client.Client
}

func NewDataSetTypeLister(client client.Client) DataSetTypeLister {
	return &dataSetTypeLister{
		client: client,
	}
}

// List returns a DataSetTypeList in the given namespace.
func (l *dataSetTypeLister) List(ctx context.Context, namespace string) (*v1alpha1.DataSetTypeList, error) {
	typeList := &v1alpha1.DataSetTypeList{}
	if err := l.client.List(ctx, typeList, &client.ListOptions{Namespace: namespace}); err != nil {
		return nil, fmt.Errorf("unable to list ConnectionTypes: %w", err)
	}

	return typeList, nil
}

func (l *dataSetTypeLister) Find(ctx context.Context, namespace string, dtype string) (*v1alpha1.DataSetType, error) {
	typeList, err := l.List(ctx, namespace)
	if err != nil {
		return nil, err
	}

	var res *v1alpha1.DataSetType
	for _, dt := range typeList.Items {
		if dt.Name == dtype {
			res = &dt
		}
	}

	return res, nil
}
