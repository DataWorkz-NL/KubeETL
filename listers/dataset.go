package listers

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/dataworkz/kubeetl/api/v1alpha1"
)

//DataSetLister lists and finds DataSets
type DataSetLister interface {
	List(ctx context.Context, namespace string) (*v1alpha1.DataSetList, error)
	Find(ctx context.Context, namespace string, name string) (*v1alpha1.DataSet, error)
}

type dataSetLister struct {
	client client.Client
}

func NewDataSetLister(client client.Client) DataSetLister {
	return &dataSetLister{
		client: client,
	}
}

// List returns a DataSetList in the given namespace.
func (l *dataSetLister) List(ctx context.Context, namespace string) (*v1alpha1.DataSetList, error) {
	dataList := &v1alpha1.DataSetList{}
	if err := l.client.List(ctx, dataList, &client.ListOptions{Namespace: namespace}); err != nil {
		return nil, fmt.Errorf("unable to list DataSets: %w", err)
	}

	return dataList, nil
}

func (l *dataSetLister) Find(ctx context.Context, namespace string, name string) (*v1alpha1.DataSet, error) {
	dataList, err := l.List(ctx, namespace)
	if err != nil {
		return nil, err
	}

	var res *v1alpha1.DataSet
	for _, ds := range dataList.Items {
		if ds.Name == name {
			res = &ds
		}
	}

	return res, nil
}
