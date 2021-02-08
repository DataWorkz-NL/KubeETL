package webhooks

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/dataworkz/kubeetl/api/v1alpha1"
)

var _ = Describe("Connection validation webhook", func() {
	ctx := context.Background()

	It("Should return an error if no ConnectionType exists", func() {
		con := &v1alpha1.Connection{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-connection",
				Namespace: "default",
			},
			Spec: v1alpha1.ConnectionSpec{
				Type: "unknown",
			},
		}

		err := k8sClient.Create(ctx, con)
		Expect(err).ShouldNot(HaveOccurred())
	})
})
