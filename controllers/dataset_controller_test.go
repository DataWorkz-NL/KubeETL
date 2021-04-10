package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
				HealthCheck: &api.WorkflowReference{
					Namespace: "default",
					Name:      "healthcheck-wf",
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
		})
	})
})
