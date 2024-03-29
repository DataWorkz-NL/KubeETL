package webhooks

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	"github.com/dataworkz/kubeetl/api/v1alpha1"
)

var _ = Describe("DataSet validation webhook", func() {

	var dtype *v1alpha1.DataSetType
	BeforeEach(func() {
		dtype = &v1alpha1.DataSetType{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "mysql",
				Namespace: "default",
			},
			Spec: v1alpha1.DataSetTypeSpec{
				MetadataFields: v1alpha1.MetadataValidation{
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
			},
		}
		err := k8sClient.Create(context.Background(), dtype)
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(k8sClient.Delete(context.Background(), dtype)).Should(Succeed())
	})

	It("Should return an error if no DataSetType exists", func() {
		creds := make(v1alpha1.Credentials)
		creds["test_val"] = v1alpha1.Value{
			Value: "foo",
		}

		ds := &v1alpha1.DataSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-dataset",
				Namespace: "default",
			},
			Spec: v1alpha1.DataSetSpec{
				Type:        "unknown",
				StorageType: v1alpha1.PersistentType,
				Metadata:    creds,
			},
		}

		err := k8sClient.Create(context.Background(), ds)
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).To(Equal("admission webhook \"dataset.dataworkz.nl\" denied the request: Unknown DataSetType: unknown"))
	})

	It("Should return an error if a validation failed", func() {
		creds := make(v1alpha1.Credentials)
		creds["test_val"] = v1alpha1.Value{
			Value: "foo",
		}

		ds := &v1alpha1.DataSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-dataset",
				Namespace: "default",
			},
			Spec: v1alpha1.DataSetSpec{
				Type:        "mysql",
				StorageType: v1alpha1.PersistentType,
				Metadata:    creds,
			},
		}

		err := k8sClient.Create(context.Background(), ds)
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).To(Equal("admission webhook \"dataset.dataworkz.nl\" denied the request: spec.metadata.test_val: Invalid value: \"foo\": Value below MinLength"))
	})

	It("Should return no error for a valid DataSet", func() {
		creds := make(v1alpha1.Credentials)
		creds["test_val"] = v1alpha1.Value{
			Value: "foo2",
		}

		ds := &v1alpha1.DataSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-dataset",
				Namespace: "default",
			},
			Spec: v1alpha1.DataSetSpec{
				Type:        "mysql",
				StorageType: v1alpha1.PersistentType,
				Metadata:    creds,
			},
		}

		err := k8sClient.Create(context.Background(), ds)
		Expect(err).ShouldNot(HaveOccurred())
	})

})
