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
		creds := make(v1alpha1.Credentials)
		creds["test_val"] = v1alpha1.Value{
			Value: "foo",
		}

		con := &v1alpha1.Connection{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-connection",
				Namespace: "default",
			},
			Spec: v1alpha1.ConnectionSpec{
				Type:        "unknown",
				Credentials: creds,
			},
		}

		err := k8sClient.Create(ctx, con)
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).To(Equal("admission webhook \"connection.dataworkz.nl\" denied the request: Unknown ConnectionType: unknown"))
	})
})
