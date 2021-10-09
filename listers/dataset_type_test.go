package listers

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/dataworkz/kubeetl/api/v1alpha1"
)

var _ = Describe("DataSetTypeLister", func() {
	var client client.Client
	var dstl DataSetTypeLister
	var ctx context.Context
	BeforeEach(func() {
		s := runtime.NewScheme()
		s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.ConnectionTypeList{}, &v1alpha1.ConnectionType{})
		_ = v1alpha1.AddToScheme(s)
		client = fake.NewClientBuilder().WithScheme(s).Build()
		dstl = NewDataSetTypeLister(client)
		ctx = context.Background()
	})

	It("Should be able to find a DataSetType based on the type name", func() {
		dsType, err := dstl.Find(ctx, "default", "test")
		Expect(err).To(Succeed())
		Expect(dsType).To(BeNil())
		dt := &v1alpha1.DataSetType{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
		}
		err = client.Create(ctx, dt)
		Expect(err).To(Succeed())
		dsType, err = dstl.Find(ctx, "default", "test")
		Expect(err).To(Succeed())
		Expect(dsType).To(Not(BeNil()))
	})
})
