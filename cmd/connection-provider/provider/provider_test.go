package provider

import (
	"context"
	"encoding/base64"

	"github.com/dataworkz/kubeetl/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Connection Provider", func() {

	var backingSecret *corev1.Secret
	var workflowSecret *corev1.Secret
	var backingConfigMap *corev1.ConfigMap
	var workflow *v1alpha1.Workflow
	var connection *v1alpha1.Connection

	var workflowSecretKey = types.NamespacedName{
		Namespace: "default",
		Name:      "test-workflow",
	}

	BeforeEach(func() {
		ctx := context.Background()

		workflowSecret = &corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-workflow",
				Namespace: "default",
			},
		}

		backingSecret = &corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:      "backing-secret",
				Namespace: "default",
			},
			StringData: map[string]string{
				"secret-key": "secret-value",
			},
		}

		backingConfigMap = &corev1.ConfigMap{
			ObjectMeta: v1.ObjectMeta{
				Name:      "backing-configmap",
				Namespace: "default",
			},
			Data: map[string]string{
				"cm-key": "cm-value",
			},
		}

		connection = &v1alpha1.Connection{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-connection",
				Namespace: "default",
			},
			Spec: v1alpha1.ConnectionSpec{
				Credentials: v1alpha1.Credentials{
					"configmapRef": v1alpha1.Value{ValueFrom: &v1alpha1.ValueSource{
						ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: backingConfigMap.Name,
							},
							Key: "cm-key",
						},
					}},
					"secretRef": v1alpha1.Value{
						ValueFrom: &v1alpha1.ValueSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: backingSecret.Name,
								},
								Key: "secret-key",
							},
						},
					},
					"inline": v1alpha1.Value{
						Value: "inline-value",
					},
				},
			},
		}

		connectionRef := corev1.LocalObjectReference{Name: connection.Name}

		workflow = &v1alpha1.Workflow{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-workflow",
				Namespace: "default",
			},
			Spec: v1alpha1.WorkflowSpec{
				InjectableValues: v1alpha1.InjectableValues{
					v1alpha1.InjectableValue{
						Name:          "inline-content",
						Content:       "{{.inline}}",
						ConnectionRef: connectionRef,
					},
					v1alpha1.InjectableValue{
						Name:          "secret-ref",
						Content:       "{{.secretRef}}",
						ConnectionRef: connectionRef,
					},
					v1alpha1.InjectableValue{
						Name:          "configmap-ref",
						Content:       "{{.configmapRef}}",
						ConnectionRef: connectionRef,
					},
					v1alpha1.InjectableValue{
						Name:          "combination",
						Content:       "{{.inline}} {{.configmapRef}} {{.secretRef}}",
						ConnectionRef: connectionRef,
					},
				},
			},
		}

		Expect(k8sClient.Create(ctx, workflow)).ToNot(HaveOccurred())
		Expect(k8sClient.Create(ctx, connection)).ToNot(HaveOccurred())
		Expect(k8sClient.Create(ctx, workflowSecret)).ToNot(HaveOccurred())
		Expect(k8sClient.Create(ctx, backingSecret)).ToNot(HaveOccurred())
		Expect(k8sClient.Create(ctx, backingConfigMap)).ToNot(HaveOccurred())

	})

	It("Should work", func() {
		ctx := context.Background()
		err := connProvider.ProvideWorkflowSecret(workflow.Name, workflow.Namespace)
		Expect(err).ToNot(HaveOccurred())

		err = k8sClient.Get(ctx, workflowSecretKey, workflowSecret)
		Expect(err).ToNot(HaveOccurred())

		expectedResults := map[string]string{
			"inline-content": "inline-value",
			"configmap-ref":  "cm-value",
			"secret-ref":     "secret-value",
			"combination":    "inline-value cm-value secret-value",
		}

		for key, expected := range expectedResults {
			val := string(workflowSecret.Data[key])
			Expect(val).To(Equal(expected))
		}

	})
})

func readSecretKey(secret corev1.Secret, key string) string {
	var dst []byte
	data := secret.Data[key]
	_, _ = base64.StdEncoding.Decode(data, dst)
	return string(dst)
}

// var _ = Describe("DataSetReconciler", func() {
// 	const timeout = time.Second * 5
// 	const interval = time.Second * 1
// 	var wfKey types.NamespacedName
// 	var argoWfKey types.NamespacedName
// 	var argoWf wfv1.Workflow
// 	BeforeEach(func() {
// 		ctx := context.Background()

// 		wfKey = types.NamespacedName{
// 			Name:      "test-workflow",
// 			Namespace: "default",
// 		}

// 		wfSpec := api.WorkflowSpec{
// 			ArgoWorkflowSpec: wfv1.WorkflowSpec{},
// 		}

// 		wf := api.Workflow{
// 			ObjectMeta: metav1.ObjectMeta{
// 				Name:      wfKey.Name,
// 				Namespace: wfKey.Namespace,
// 			},
// 			Spec: wfSpec,
// 		}

// 		Expect(k8sClient.Create(ctx, &wf)).Should(Succeed())

// 		argoWfKey = types.NamespacedName{
// 			Name:      "default-argo-workflow",
// 			Namespace: "default",
// 		}
// 		argoWfSpec := wfv1.WorkflowSpec{}

// 		argoWf = wfv1.Workflow{
// 			ObjectMeta: metav1.ObjectMeta{
// 				Name:      argoWfKey.Name,
// 				Namespace: argoWfKey.Namespace,
// 			},
// 			Spec: argoWfSpec,
// 		}
// 		Expect(k8sClient.Create(ctx, &argoWf)).Should(Succeed())
// 	})

// 	AfterEach(func() {
// 		var wf api.Workflow
// 		ctx := context.Background()
// 		Expect(k8sClient.Get(ctx, wfKey, &wf)).Should(Succeed())
// 		Expect(k8sClient.Delete(ctx, &wf)).Should(Succeed())
// 		var argoWf wfv1.Workflow
// 		Expect(k8sClient.Get(ctx, argoWfKey, &argoWf)).Should(Succeed())
// 		Expect(k8sClient.Delete(ctx, &argoWf)).Should(Succeed())
// 	})

// 	Context("DataSet with HealthCheck", func() {
// 		It("Should set DataSet health to Unknown for a unknown Workflow", func() {
// 			ctx := context.Background()
// 			key := types.NamespacedName{
// 				Name:      "default-dataset",
// 				Namespace: "default",
// 			}

// 			spec := api.DataSetSpec{
// 				StorageType: api.PersistentType,
// 				Type:        "MySQL DataSet",
// 				HealthCheck: &api.WorkflowReference{
// 					Namespace: "default",
// 					Name:      "unknown-wf",
// 				},
// 			}

// 			created := api.DataSet{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      key.Name,
// 					Namespace: key.Namespace,
// 				},
// 				Spec: spec,
// 			}

// 			Expect(k8sClient.Create(ctx, &created)).Should(Succeed())

// 			Eventually(func() bool {
// 				res := &api.DataSet{}
// 				err := k8sClient.Get(ctx, key, res)
// 				if err != nil {
// 					return false
// 				}

// 				return res.Status.Healthy == api.Unknown
// 			}, timeout, interval).Should(BeTrue())

// 			Expect(k8sClient.Delete(ctx, &created)).Should(Succeed())
// 		})

// 		It("Should use an existing Workflow as DataSet healthcheck indicator", func() {
// 			ctx := context.Background()
// 			key := types.NamespacedName{
// 				Name:      "default-dataset",
// 				Namespace: "default",
// 			}

// 			spec := api.DataSetSpec{
// 				StorageType: api.PersistentType,
// 				Type:        "MySQL DataSet",
// 				HealthCheck: &api.WorkflowReference{
// 					Namespace: wfKey.Namespace,
// 					Name:      wfKey.Name,
// 				},
// 			}

// 			created := api.DataSet{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      key.Name,
// 					Namespace: key.Namespace,
// 				},
// 				Spec: spec,
// 			}

// 			Expect(k8sClient.Create(ctx, &created)).Should(Succeed())

// 			By("Setting the DataSet Healthcheck label on the WorkFlow")
// 			Eventually(func() bool {
// 				res := &api.Workflow{}
// 				err := k8sClient.Get(ctx, wfKey, res)
// 				if err != nil {
// 					return false
// 				}

// 				return labels.HasLabel(res.Labels, healthcheckLabel)
// 			}, timeout, interval)

// 			By("Updating the status if the workflow executed")
// 			// First fake Workflow controller behaviour
// 			argoWf.Status.Phase = wfv1.NodeFailed
// 			Expect(k8sClient.Update(ctx, &argoWf)).Should(Succeed())
// 			wf := &api.Workflow{}
// 			Expect(k8sClient.Get(ctx, wfKey, wf)).Should(Succeed())
// 			wf.Status.ArgoWorkflowRef = &corev1.ObjectReference{
// 				Name:      argoWfKey.Name,
// 				Namespace: argoWfKey.Namespace,
// 			}
// 			Expect(k8sClient.Status().Update(ctx, wf)).Should(Succeed())

// 			Eventually(func() bool {
// 				res := &api.DataSet{}
// 				err := k8sClient.Get(ctx, key, res)
// 				if err != nil {
// 					return false
// 				}

// 				return res.Status.Healthy == api.Unhealthy
// 			}, timeout, interval).Should(BeTrue())

// 			By("Cleaning up the label if the DataSet no longer has a healthcheck")
// 			ds := &api.DataSet{}
// 			Eventually(func() bool {
// 				err := k8sClient.Get(ctx, key, ds)
// 				if err != nil {
// 					return false
// 				}

// 				return true
// 			}, timeout, interval).Should(BeTrue())

// 			ds.Spec.HealthCheck = nil
// 			Eventually(func() bool {
// 				err := k8sClient.Update(ctx, ds)
// 				if err != nil {
// 					return false
// 				}

// 				return true
// 			}, timeout, interval).Should(BeTrue())

// 			Eventually(func() bool {
// 				res := &api.Workflow{}
// 				err := k8sClient.Get(ctx, wfKey, res)
// 				if err != nil {
// 					return false
// 				}

// 				return !labels.HasLabel(res.Labels, healthcheckLabel)
// 			}, timeout, interval).Should(BeTrue())

// 			Expect(k8sClient.Delete(ctx, &created)).Should(Succeed())
// 		})
// 	})
// })
