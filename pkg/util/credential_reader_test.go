package util

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/dataworkz/kubeetl/api/v1alpha1"
)

var credReader CredentialReader
var _ = Describe("Connection validation webhook", func() {

	var objects []client.Object

	BeforeEach(func() {
		var conn = &v1alpha1.Connection{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
			},
			Spec: v1alpha1.ConnectionSpec{
				Credentials: v1alpha1.Credentials{
					"inlinevalue": v1alpha1.Value{
						Value: "foo inline",
					},
					"configmapkey": v1alpha1.Value{
						ValueFrom: &v1alpha1.ValueSource{
							ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "connection-configmap",
								},
								Key: "test",
							},
						},
					},
					"secretkey": v1alpha1.Value{
						ValueFrom: &v1alpha1.ValueSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "connection-secret",
								},
								Key: "test",
							},
						},
					},
					"missingsecret": v1alpha1.Value{
						ValueFrom: &v1alpha1.ValueSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "does-not-exist",
								},
								Key: "test",
							},
						},
					},
					"missingconfigmap": v1alpha1.Value{
						ValueFrom: &v1alpha1.ValueSource{
							ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "does-not-exist",
								},
								Key: "test",
							},
						},
					},
					"missingsecretkey": v1alpha1.Value{
						ValueFrom: &v1alpha1.ValueSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "connection-secret",
								},
								Key: "does-not-exist",
							},
						},
					},
					"missingconfigmapkey": v1alpha1.Value{
						ValueFrom: &v1alpha1.ValueSource{
							ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "connection-configmap",
								},
								Key: "does-not-exist",
							},
						},
					},
				},
			},
		}

		credReader = NewCredentialReader(k8sClient, conn)

		objects = []client.Object{
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "connection-secret",
					Namespace: "default",
				},
				StringData: map[string]string{
					"test": "foo secret",
				},
			},
			&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "connection-configmap",
					Namespace: "default",
				},
				Data: map[string]string{
					"test": "foo configmap",
				},
			},
		}

		for _, obj := range objects {
			err := k8sClient.Create(context.Background(), obj)
			Expect(err).ShouldNot(HaveOccurred())
		}

	})

	AfterEach(func() {
		for _, obj := range objects {
			Expect(k8sClient.Delete(context.Background(), obj)).To(Succeed())
		}
	})

	It("Should return configmap keys", func() {
		val, err := credReader.ReadValue(context.Background(), "configmapkey")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(val).To(BeIdenticalTo("foo configmap"))
	})

	It("Should return secret keys", func() {
		val, err := credReader.ReadValue(context.Background(), "secretkey")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(val).To(BeIdenticalTo("foo secret"))
	})

	It("Should return inline values", func() {
		val, err := credReader.ReadValue(context.Background(), "inlinevalue")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(val).To(BeIdenticalTo("foo inline"))
	})

	It("Should return an error for referencing missing secret keys", func() {
		_, err := credReader.ReadValue(context.Background(), "missingsecretkey")
		Expect(err).Should(HaveOccurred())
	})

	It("Should return an error for referencing missing configmap keys", func() {
		_, err := credReader.ReadValue(context.Background(), "missingconfigmapkey")
		Expect(err).Should(HaveOccurred())
	})

	It("Should return an error for referencing missing secrets", func() {
		_, err := credReader.ReadValue(context.Background(), "missingsecret")
		Expect(err).Should(HaveOccurred())
	})

	It("Should return an error for referencing missing configmaps", func() {
		_, err := credReader.ReadValue(context.Background(), "missingconfigmap")
		Expect(err).Should(HaveOccurred())
	})

	It("Should return an error for referencing a value that doesn't exist", func() {
		_, err := credReader.ReadValue(context.Background(), "does-not-exist")
		Expect(err).Should(HaveOccurred())
	})
})
