package controllers

import (
	"context"
	"time"

	wfv1 "github.com/argoproj/argo/v2/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	api "github.com/dataworkz/kubeetl/api/v1alpha1"
	"github.com/dataworkz/kubeetl/labels"
)

var _ = Describe("DataSetReconciler", func() {
	const timeout = time.Second * 5
	const interval = time.Second * 1
	var wfKey types.NamespacedName
	var argoWfKey types.NamespacedName
	var argoWf wfv1.Workflow
	BeforeEach(func() {
		ctx := context.Background()

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

		Expect(k8sClient.Create(ctx, &wf)).Should(Succeed())

		argoWfKey = types.NamespacedName{
			Name:      "default-argo-workflow",
			Namespace: "default",
		}
		argoWfSpec := wfv1.WorkflowSpec{}

		argoWf = wfv1.Workflow{
			ObjectMeta: metav1.ObjectMeta{
				Name:      argoWfKey.Name,
				Namespace: argoWfKey.Namespace,
			},
			Spec: argoWfSpec,
		}
		Expect(k8sClient.Create(ctx, &argoWf)).Should(Succeed())
	})

	AfterEach(func() {
		var wf api.Workflow
		ctx := context.Background()
		Expect(k8sClient.Get(ctx, wfKey, &wf)).Should(Succeed())
		Expect(k8sClient.Delete(ctx, &wf)).Should(Succeed())
		var argoWf wfv1.Workflow
		Expect(k8sClient.Get(ctx, argoWfKey, &argoWf)).Should(Succeed())
		Expect(k8sClient.Delete(ctx, &argoWf)).Should(Succeed())
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
				res := &api.Workflow{}
				err := k8sClient.Get(ctx, wfKey, res)
				if err != nil {
					return false
				}

				return labels.HasLabel(res.Labels, healthcheckLabel)
			}, timeout, interval)

			By("Updating the status if the workflow executed")
			// First fake Workflow controller behaviour
			argoWf.Status.Phase = wfv1.NodeFailed
			Expect(k8sClient.Update(ctx, &argoWf)).Should(Succeed())
			wf := &api.Workflow{}
			Expect(k8sClient.Get(ctx, wfKey, wf)).Should(Succeed())
			wf.Status.ArgoWorkflowRef = &corev1.ObjectReference{
				Name:      argoWfKey.Name,
				Namespace: argoWfKey.Namespace,
			}
			Expect(k8sClient.Status().Update(ctx, wf)).Should(Succeed())

			Eventually(func() bool {
				res := &api.DataSet{}
				err := k8sClient.Get(ctx, key, res)
				if err != nil {
					return false
				}

				return res.Status.Healthy == api.Unhealthy
			}, timeout, interval).Should(BeTrue())

			By("Cleaning up the label if the DataSet no longer has a healthcheck")
			ds := &api.DataSet{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, key, ds)
				if err != nil {
					return false
				}

				return true
			}, timeout, interval).Should(BeTrue())

			ds.Spec.HealthCheck = nil
			Eventually(func() bool {
				err := k8sClient.Update(ctx, ds)
				if err != nil {
					return false
				}

				return true
			}, timeout, interval).Should(BeTrue())

			Eventually(func() bool {
				res := &api.Workflow{}
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