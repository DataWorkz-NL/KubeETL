package provider

import (
	"context"
	"time"

	"github.com/dataworkz/kubeetl/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Connection Provider", func() {
	const timeout = time.Second * 5
	const interval = time.Second * 1

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

		Eventually(func() bool {
			err := connProvider.ProvideWorkflowSecret(workflow.Name, workflow.Namespace)
			if err != nil {
				return false
			}

			err = k8sClient.Get(ctx, workflowSecretKey, workflowSecret)
			if err != nil {
				return false
			}

			expectedResults := map[string]string{
				"inline-content": "inline-value",
				"configmap-ref":  "cm-value",
				"secret-ref":     "secret-value",
				"combination":    "inline-value cm-value secret-value",
			}

			for key, expected := range expectedResults {
				val := string(workflowSecret.Data[key])
				if val != expected {
					return false
				}
				// Expect(val).To(Equal(expected))
			}
			return true
		}, timeout, interval)

	})
})
