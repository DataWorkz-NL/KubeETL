package webhooks

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	"github.com/dataworkz/kubeetl/api/v1alpha1"
)

var _ = Describe("Connection validation webhook", func() {

	var conType *v1alpha1.ConnectionType
	BeforeEach(func() {
		conType = &v1alpha1.ConnectionType{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "mysql",
				Namespace: "default",
			},
			Spec: v1alpha1.ConnectionTypeSpec{
				Name: "mysql",
				Fields: []v1alpha1.CredentialFieldSpec{
					{
						Name:     "test_val",
						Required: true,
						Validation: &v1alpha1.Validation{
							MinLength: pointer.Int32Ptr(4),
						},
					},
				},
			},
		}
		err := k8sClient.Create(context.Background(), conType)
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		k8sClient.Delete(context.Background(), conType)
	})

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

		err := k8sClient.Create(context.Background(), con)
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).To(Equal("admission webhook \"connection.dataworkz.nl\" denied the request: Unknown ConnectionType: unknown"))
	})

	It("Should return an error if a validation failed", func() {
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
				Type:        "mysql",
				Credentials: creds,
			},
		}

		err := k8sClient.Create(context.Background(), con)
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).To(Equal("admission webhook \"connection.dataworkz.nl\" denied the request: spec.credentials.test_val: Invalid value: \"foo\": Value below MinLength"))
	})

	It("Should return no error for a valid Connection", func() {

		creds := make(v1alpha1.Credentials)
		creds["test_val"] = v1alpha1.Value{
			Value: "foo2",
		}

		con := &v1alpha1.Connection{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-connection",
				Namespace: "default",
			},
			Spec: v1alpha1.ConnectionSpec{
				Type:        "mysql",
				Credentials: creds,
			},
		}

		err := k8sClient.Create(context.Background(), con)
		Expect(err).ShouldNot(HaveOccurred())
	})

})
