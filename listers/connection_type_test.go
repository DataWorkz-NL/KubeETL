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

var _ = Describe("ConnectionTypeLister", func() {
	var client client.Client
	var ctl ConnectionTypeLister
	var ctx context.Context
	BeforeEach(func() {
		s := runtime.NewScheme()
		s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.ConnectionTypeList{}, &v1alpha1.ConnectionType{})
		_ = v1alpha1.AddToScheme(s)
		client = fake.NewClientBuilder().WithScheme(s).Build()
		ctl = NewConnectionTypeLister(client)
		ctx = context.Background()
	})

	It("Should be able to find a ConnectionType based on the type name", func() {
		conType, err := ctl.Find(ctx, "default", "test")
		Expect(err).To(Succeed())
		Expect(conType).To(BeNil())
		ct := &v1alpha1.ConnectionType{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "etl.dataworkz.nl/v1alpha1",
				Kind:       "ConnectionType",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
		}
		err = client.Create(ctx, ct)
		Expect(err).To(Succeed())
		conType, err = ctl.Find(ctx, "default", "test")
		Expect(err).To(Succeed())
		Expect(conType).To(Not(BeNil()))
	})
})
