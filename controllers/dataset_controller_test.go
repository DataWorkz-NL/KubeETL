package controllers

import (
	"context"
	"time"

	wfv1 "github.com/argoproj/argo/v2/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	api "github.com/dataworkz/kubeetl/api/v1alpha1"
	"github.com/dataworkz/kubeetl/labels"
)

var _ = Describe("DataSetReconciler", func() {
	const timeout = time.Second * 5
	const interval = time.Second * 1
	var wfKey types.NamespacedName
	BeforeEach(func() {
		wfKey = types.NamespacedName{
			Name:      "default-workflow",
			Namespace: "default",
		}

		wfSpec := api.WorkflowSpec{
			ArgoWorkflowSpec: wfv1.WorkflowSpec{},
		}

		wf := api.Workflow{
			ObjectMeta: metav1.ObjectMeta{
				Name:      wfKey.Name,
				Namespace: wfKey.Namespace,
			},
			Spec: wfSpec,
		}

		ctx := context.Background()
		Expect(k8sClient.Create(ctx, &wf)).Should(Succeed())
	})

	AfterEach(func() {
		var wf api.Workflow
		ctx := context.Background()
		Expect(k8sClient.Get(ctx, wfKey, &wf)).Should(Succeed())
		Expect(k8sClient.Delete(ctx, &wf)).Should(Succeed())
	})

	Context("DataSet with HealthCheck", func() {
		It("Should set DataSet health to Unknown for a unknown Workflow", func() {
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
					Name:      "unknown-wf",
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

			Eventually(func() bool {
				res := &api.DataSet{}
				err := k8sClient.Get(ctx, key, res)
				if err != nil {
					return false
				}

				return res.Status.Healthy == api.Unknown
			}, timeout, interval).Should(BeTrue())

			Expect(k8sClient.Delete(ctx, &created)).Should(Succeed())
		})

		It("Should use an existing Workflow as DataSet healthcheck indicator", func() {
			ctx := context.Background()
			key := types.NamespacedName{
				Name:      "default-dataset",
				Namespace: "default",
			}

			spec := api.DataSetSpec{
				StorageType: api.PersistentType,
				Type:        "MySQL DataSet",
				HealthCheck: &api.WorkflowReference{
					Namespace: wfKey.Namespace,
					Name:      wfKey.Name,
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

			By("Setting the DataSet Healthcheck label on the WorkFlow")
			Eventually(func() bool {
				var res *api.Workflow
				err := k8sClient.Get(ctx, wfKey, res)
				if err != nil {
					return false
				}

				return labels.HasLabel(res.Labels, healthcheckLabel)
			}, timeout, interval)

			By("Cleaning up the label if the DataSet no longer has a healthcheck")
			var ds api.DataSet
			Eventually(func() bool {
				err := k8sClient.Get(ctx, key, &ds)
				if err != nil {
					return false
				}

				return true
			}, timeout, interval).Should(BeTrue())

			ds.Spec.HealthCheck = nil
			Eventually(func() bool {
				err := k8sClient.Update(ctx, &ds)
				if err != nil {
					return false
				}

				return true
			}, timeout, interval).Should(BeTrue())

			Eventually(func() bool {
				var res *api.Workflow
				err := k8sClient.Get(ctx, wfKey, res)
				if err != nil {
					return false
				}

				return !labels.HasLabel(res.Labels, healthcheckLabel)
			}, timeout, interval).Should(BeTrue())

			Expect(k8sClient.Delete(ctx, &created)).Should(Succeed())
		})
	})
})
