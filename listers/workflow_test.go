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

var _ = Describe("WorkflowLister", func() {
	var client client.Client
	var wfl WorkflowLister
	var ctx context.Context
	BeforeEach(func() {
		s := runtime.NewScheme()
		s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.WorkflowList{}, &v1alpha1.Workflow{})
		_ = v1alpha1.AddToScheme(s)
		client = fake.NewFakeClientWithScheme(s)
		wfl = NewWorkflowLister(client)
		ctx = context.Background()
	})

	It("Should be able to find a Workflow based on the type name", func() {
		workflow, err := wfl.Find(ctx, "default", "test")
		Expect(err).To(Succeed())
		Expect(workflow).To(BeNil())
		wf := &v1alpha1.Workflow{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "etl.dataworkz.nl/v1alpha1",
				Kind:       "Workflow",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
		}
		err = client.Create(ctx, wf)
		Expect(err).To(Succeed())
		workflow, err = wfl.Find(ctx, "default", "test")
		Expect(err).To(Succeed())
		Expect(workflow).To(Not(BeNil()))
	})
})
