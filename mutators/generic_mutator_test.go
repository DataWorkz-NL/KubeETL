package mutators

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func Test_genericMutator_Mutate(t *testing.T) {
	g := NewGomegaWithT(t)
	controller := createController()
	tests := []struct {
		name       string
		controller client.Object
		existing   *appsv1.Deployment
		spec       *appsv1.Deployment
		mergeFn    MergeFn
		expectFn   func(context.Context, *GomegaWithT, client.Client, *appsv1.Deployment, error)
	}{
		{
			"Can create new deployments if they don't exist yet",
			controller,
			nil,
			&appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "default-deployment",
					Namespace: "default",
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: pointer.Int32Ptr(1),
				},
			},
			func(existing, expected client.Object) error {
				e := existing.(*appsv1.Deployment)
				d := expected.(*appsv1.Deployment)
				e.Spec.Replicas = d.Spec.Replicas
				return nil
			},
			func(ctx context.Context, g *GomegaWithT, c client.Client, d *appsv1.Deployment, err error) {
				g.Expect(err).To(Succeed())

				key := types.NamespacedName{
					Name:      d.GetName(),
					Namespace: d.GetNamespace(),
				}
				var res appsv1.Deployment
				err = c.Get(ctx, key, &res)
				g.Expect(err).To(Succeed())
				g.Expect(res.Spec.Replicas).To(Equal(d.Spec.Replicas))
			},
		},
		{
			"Can update existing deployments",
			controller,
			&appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "default-deployment",
					Namespace: "default",
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: pointer.Int32Ptr(1),
				},
			},
			&appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "default-deployment",
					Namespace: "default",
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: pointer.Int32Ptr(2),
				},
			},
			func(existing, expected client.Object) error {
				e := existing.(*appsv1.Deployment)
				d := expected.(*appsv1.Deployment)
				e.Spec.Replicas = d.Spec.Replicas
				return nil
			},
			func(ctx context.Context, g *GomegaWithT, c client.Client, d *appsv1.Deployment, err error) {
				g.Expect(err).To(Succeed())

				key := types.NamespacedName{
					Name:      d.GetName(),
					Namespace: d.GetNamespace(),
				}
				var res appsv1.Deployment
				err = c.Get(ctx, key, &res)
				g.Expect(err).To(Succeed())

				g.Expect(res.Spec.Replicas).To(Equal(d.Spec.Replicas))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := runtime.NewScheme()
			_ = clientgoscheme.AddToScheme(s)
			c := fake.NewFakeClientWithScheme(s)
			log := zap.New(zap.UseDevMode(true))

			m := &genericMutator{
				client: c,
				scheme: s,
				log:    log,
			}

			ctx := context.Background()

			if tt.existing != nil {
				err := c.Create(ctx, tt.existing)
				g.Expect(err).To(Succeed())
			}

			err := m.Mutate(ctx, tt.controller, tt.spec, tt.mergeFn)
			tt.expectFn(ctx, g, c, tt.spec, err)
		})
	}
}

func createController() *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-deployment-controller",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(1),
		},
	}
}
