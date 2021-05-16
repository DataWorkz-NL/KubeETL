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

var _ = Describe("ConnectionLister", func() {
	var client client.Client
	var cl ConnectionLister
	var ctx context.Context
	BeforeEach(func() {
		s := runtime.NewScheme()
		s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.ConnectionList{}, &v1alpha1.Connection{})
		_ = v1alpha1.AddToScheme(s)
		client = fake.NewFakeClientWithScheme(s)
		cl = NewConnectionLister(client)
		ctx = context.Background()
	})

	It("Should be able to find a Connection based on the type name", func() {
		conn, err := cl.Find(ctx, "default", "test")
		Expect(err).To(Succeed())
		Expect(conn).To(BeNil())
		c := &v1alpha1.Connection{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "etl.dataworkz.nl/v1alpha1",
				Kind:       "Connection",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
		}
		err = client.Create(ctx, c)
		Expect(err).To(Succeed())
		conn, err = cl.Find(ctx, "default", "test")
		Expect(err).To(Succeed())
		Expect(conn).To(Not(BeNil()))
	})
})
