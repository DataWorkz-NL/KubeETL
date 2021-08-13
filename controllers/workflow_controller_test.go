package controllers

import (
	"context"
	"time"

	wfv1 "github.com/argoproj/argo/v2/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	// corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	api "github.com/dataworkz/kubeetl/api/v1alpha1"
)

var _ = Describe("WorkflowReconciler", func() {
	const timeout = time.Second * 5
	const interval = time.Second * 1
	var connKey types.NamespacedName
	var cmRef v1.LocalObjectReference
	var secretRef v1.LocalObjectReference

	BeforeEach(func() {
		ctx := context.Background()

		connKey = types.NamespacedName{
			Name:      "wf-connection",
			Namespace: "default",
		}

		secretRef = v1.LocalObjectReference{
			Name: "wf-secret",
		}

		cmRef = v1.LocalObjectReference{
			Name: "wf-cm",
		}

		connSpec := api.ConnectionSpec{
			Credentials: api.Credentials{
				"host": api.Value{Value: "localhost"},
				"user": api.Value{
					ValueFrom: &api.ValueSource{
						ConfigMapKeyRef: &v1.ConfigMapKeySelector{
							LocalObjectReference: cmRef,
							Key:                  "user",
						},
					},
				},
				"password": api.Value{
					ValueFrom: &api.ValueSource{
						SecretKeyRef: &v1.SecretKeySelector{
							LocalObjectReference: secretRef,
							Key:                  "password",
						},
					},
				},
			},
		}

		conn := api.Connection{
			ObjectMeta: metav1.ObjectMeta{
				Name:      connKey.Name,
				Namespace: connKey.Namespace,
			},
			Spec: connSpec,
		}

		Expect(k8sClient.Create(ctx, &conn)).Should(Succeed())
	})

	AfterEach(func() {
		var wf api.Connection
		ctx := context.Background()
		Expect(k8sClient.Get(ctx, connKey, &wf)).Should(Succeed())
		Expect(k8sClient.Delete(ctx, &wf)).Should(Succeed())
	})

	Context("Workflow with injections", func() {
		It("Should inject templates in a DAG", func() {
			ctx := context.Background()
			key := types.NamespacedName{
				Name:      "wf-workflow",
				Namespace: "default",
			}

			spec := api.WorkflowSpec{
				InjectInto: []api.TemplateRef{
					api.TemplateRef{
						Name:           "main",
						InjectedValues: []string{"injectable-host"},
					},
				},
				InjectableValues: api.InjectableValues{
					api.InjectableValue{
						Name:          "injectable-host",
						ConnectionRef: v1.LocalObjectReference{Name: "wf-connection"},
						Content:       "{{.Host}}",
						EnvName:       "HOST",
					},
				},
				ArgoWorkflowSpec: wfv1.WorkflowSpec{
					Templates: []wfv1.Template{
						wfv1.Template{
							Name:      "foo",
							Container: &v1.Container{},
						},
						wfv1.Template{
							Name: "bar",
							Script: &wfv1.ScriptTemplate{
								Container: v1.Container{},
							},
						},
						wfv1.Template{
							Name: "main",
							DAG: &wfv1.DAGTemplate{
								Tasks: []wfv1.DAGTask{
									wfv1.DAGTask{Name: "footask", Template: "foo"},
									wfv1.DAGTask{Name: "bartask", Template: "bar"},
								},
							},
						},
					},
				},
			}

			created := api.Workflow{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: spec,
			}

			Expect(k8sClient.Create(ctx, &created)).Should(Succeed())
			Eventually(func() bool {
				res := &wfv1.Workflow{}
				err := k8sClient.Get(ctx, key, res)
				if err != nil {
					return false
				}

				foo := res.GetTemplateByName("foo")
				if foo == nil {
					return false
				}

				bar := res.GetTemplateByName("foo")
				if bar == nil {
					return false
				}

				containers := []v1.Container{*foo.Container, bar.Script.Container}
				iv, err := created.GetInjectableValueByName("injectable-host")
				if err != nil {
					return false
				}

				for _, c := range containers {
					isInjected := envContainsInjectableValue(c.Env, *iv, created.ConnectionSecretName().Name)
					if !isInjected {
						return false
					}
				}
				return true
			}, timeout, interval).Should(BeTrue())

			Expect(k8sClient.Delete(ctx, &created)).Should(Succeed())

		})

		It("Should", func() {
		})
	})
})

func envContainsInjectableValue(env []v1.EnvVar, iv api.InjectableValue, connectionSecretName string) bool {
	for _, e := range env {
		if e.Name == iv.EnvName {
			vf := e.ValueFrom
			return vf != nil &&
				vf.SecretKeyRef != nil &&
				vf.SecretKeyRef.Key == iv.Name &&
				vf.SecretKeyRef.LocalObjectReference.Name == connectionSecretName
		}
	}
	return false
}
