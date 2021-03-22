package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	batch "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	api "github.com/dataworkz/kubeetl/api/v1alpha1"
)

var _ = Describe("DataSetReconciler", func() {
	const timeout = time.Second * 30
	const interval = time.Second * 1

	Context("DataSet with HealthCheck", func() {
		It("Should create a CronJob on creation of the DataSet", func() {
			ctx := context.Background()
			key := types.NamespacedName{
				Name:      "default-dataset",
				Namespace: "default",
			}

			spec := api.DataSetSpec{
				StorageType: api.PersistentType,
				Type:        "MySQL DataSet",
				HealthCheck: batch.CronJobSpec{
					Schedule: "0 * * * *",
					JobTemplate: batch.JobTemplateSpec{
						Spec: batchv1.JobSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:  "test",
											Image: "busybox:latest",
										},
									},
								},
							},
						},
					},
				},
			}

			created := api.DataSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: spec,
			}

			Expect(k8sClient.Create(ctx, &created)).Should(Succeed())
			job := batch.CronJob{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, key, &job)
				if err != nil {
					return false
				}

				return true
			}, timeout, interval).Should(BeTrue())
			Expect(job).ToNot(BeNil())
			Expect(job.Spec.Schedule).To(Equal("0 * * * *"))
		})
	})
})
