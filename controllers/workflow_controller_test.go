package controllers

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	wfv1 "github.com/argoproj/argo/v2/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	// corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/dataworkz/kubeetl/api/v1alpha1"
	api "github.com/dataworkz/kubeetl/api/v1alpha1"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyz"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func randomSuffix(s string) string {
	return fmt.Sprintf("%s-%s", s, RandStringBytes(5))
}

func generateWorkflowName() string {
	return fmt.Sprintf("%s-%s", "default-workflow", RandStringBytes(5))
}

var _ = Describe("WorkflowReconciler", func() {
	const timeout = time.Second * 5
	const interval = time.Second * 1
	var resources workflowTestResources

	BeforeEach(func() {
		ctx := context.Background()
		beforeEachWorkflowTest(ctx, &resources)
	})

	AfterEach(func() {
		ctx := context.Background()
		afterEachWorkflowTest(ctx, &resources)
	})

	Context("Injections mounted as files", func() {
		It("Should mount the secret key in templates", func() {
			ctx := context.Background()
			wfName := generateWorkflowName()

			key := types.NamespacedName{
				Name:      wfName,
				Namespace: "default",
			}

			mountPath := "/mnt/injections/host"

			spec := api.WorkflowSpec{
				InjectInto: []api.TemplateRef{
					api.TemplateRef{
						Name:           "containertemplate",
						InjectedValues: []string{"injectable-host"},
					},
					api.TemplateRef{
						Name:           "scripttemplate",
						InjectedValues: []string{"injectable-host"},
					},
				},
				InjectableValues: api.InjectableValues{
					api.InjectableValue{
						Name:          "injectable-host",
						ConnectionRef: v1.LocalObjectReference{Name: "default-connection"},
						Content:       "{{.Host}}",
						MountPath:     mountPath,
					},
				},
				ArgoWorkflowSpec: wfv1.WorkflowSpec{
					Templates: []wfv1.Template{
						wfv1.Template{
							Name:      "containertemplate",
							Container: &v1.Container{},
						},
						wfv1.Template{
							Name: "scripttemplate",
							Script: &wfv1.ScriptTemplate{
								Container: v1.Container{},
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

			var res wfv1.Workflow
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, key, &res)).Should(Succeed())

				script := res.GetTemplateByName("scripttemplate")
				g.Expect(script).ToNot(BeNil())
				g.Expect(script.Script).ToNot(BeNil())

				container := res.GetTemplateByName("containertemplate")
				g.Expect(container).ToNot(BeNil())
				g.Expect(container.Container).ToNot(BeNil())

				iv, err := created.Spec.GetInjectableValueByName("injectable-host")
				g.Expect(err).ToNot(HaveOccurred())

				containers := []v1.Container{*container.Container, script.Script.Container}

				isInjected := func(container v1.Container) bool {
					for _, m := range container.VolumeMounts {
						if m.Name == v1alpha1.ConnectionVolumeName(created.Name) &&
							m.MountPath == iv.MountPath &&
							m.SubPath == iv.Name {
							return true
						}
					}
					return false
				}

				for _, c := range containers {
					g.Expect(isInjected(c)).To(BeTrue())
				}
			}, timeout, interval).Should(Succeed())

			Expect(k8sClient.Delete(ctx, &created)).To(Succeed())
			Expect(k8sClient.Delete(ctx, &res)).To(Succeed())
		})
	})

	Context("Workflow with injections as environment variables", func() {
		It("Should inject templates in a DAG", func() {
			ctx := context.Background()
			wfName := generateWorkflowName()
			key := types.NamespacedName{
				Name:      wfName,
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
						ConnectionRef: v1.LocalObjectReference{Name: "default-connection"},
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

			var res wfv1.Workflow
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, key, &res)).Should(Succeed())

				foo := res.GetTemplateByName("foo")
				g.Expect(foo).ToNot(BeNil())
				g.Expect(foo.Container).ToNot(BeNil())

				bar := res.GetTemplateByName("bar")
				g.Expect(bar).ToNot(BeNil())
				g.Expect(bar.Script).ToNot(BeNil())

				containers := []v1.Container{*foo.Container, bar.Script.Container}
				iv, err := created.Spec.GetInjectableValueByName("injectable-host")
				g.Expect(err).ToNot(HaveOccurred())

				for _, c := range containers {
					isInjected := envContainsInjectableValue(c.Env, *iv, v1alpha1.ConnectionSecret(created.Name, created.Namespace).Name)
					g.Expect(isInjected).To(BeTrue())
				}
			}, timeout, interval).Should(Succeed())

			Expect(k8sClient.Delete(ctx, &created)).To(Succeed())
			Expect(k8sClient.Delete(ctx, &res)).To(Succeed())
		})

		It("Should add a volume to a workflow", func() {
			ctx := context.Background()
			wfName := generateWorkflowName()
			key := types.NamespacedName{
				Name:      wfName,
				Namespace: "default",
			}
			spec := api.WorkflowSpec{
				InjectableValues: api.InjectableValues{
					api.InjectableValue{
						Name:          "injectable-host",
						ConnectionRef: v1.LocalObjectReference{Name: "default-connection"},
						Content:       "{{.Host}}",
						EnvName:       "HOST",
					},
				},
				ArgoWorkflowSpec: wfv1.WorkflowSpec{},
			}
			created := api.Workflow{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: spec,
			}

			Expect(k8sClient.Create(ctx, &created)).To(Succeed())

			var res wfv1.Workflow
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, key, &res)).To(Succeed())
				g.Expect(len(res.Spec.Volumes)).To(Equal(1))
				v := res.Spec.Volumes[0]
				expected := v1.Volume{
					Name: api.ConnectionVolumeName(created.Name),
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: v1alpha1.ConnectionSecret(created.Name, created.Namespace).Name,
						},
					},
				}
				g.Expect(v).To(Equal(expected))
			}, timeout, interval).Should(Succeed())
			Expect(k8sClient.Delete(ctx, &created)).To(Succeed())
			Expect(k8sClient.Delete(ctx, &res)).To(Succeed())
		})

		It("Should a connection injection setup task to a workflow", func() {
			ctx := context.Background()
			wfName := generateWorkflowName()
			key := types.NamespacedName{
				Name:      wfName,
				Namespace: "default",
			}
			spec := api.WorkflowSpec{
				InjectableValues: api.InjectableValues{
					api.InjectableValue{
						Name:          "injectable-host",
						ConnectionRef: v1.LocalObjectReference{Name: "default-connection"},
						Content:       "{{.Host}}",
						EnvName:       "HOST",
					},
				},
				ArgoWorkflowSpec: wfv1.WorkflowSpec{},
			}
			created := api.Workflow{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: spec,
			}

			Expect(k8sClient.Create(ctx, &created)).To(Succeed())

			var res wfv1.Workflow
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, key, &res)).To(Succeed())
				g.Expect(res.Spec.Entrypoint).To(Equal("entrypoint"))
				ep := res.GetTemplateByName(res.Spec.Entrypoint)
				g.Expect(ep).ToNot(BeNil())
				g.Expect(ep.GetType()).To(Equal(wfv1.TemplateTypeSteps))

				g.Expect(len(ep.Steps)).To(Equal(2))
				g.Expect(ep.Steps[0].Steps[0].Name).To(Equal("run-injection"))
			}, timeout, interval).Should(Succeed())

			Expect(k8sClient.Delete(ctx, &created)).To(Succeed())
			Expect(k8sClient.Delete(ctx, &res)).To(Succeed())
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
				vf.SecretKeyRef.Name == connectionSecretName
		}
	}
	return false
}
