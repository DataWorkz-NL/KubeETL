package mutators

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	batch "k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func Test_typedMutator_MutateDeployment(t *testing.T) {
	g := NewGomegaWithT(t)
	controller := createController()
	tests := []struct {
		name       string
		controller client.Object
		existing   *batch.CronJob
		spec       *batch.CronJob
		expectFn   func(context.Context, *GomegaWithT, client.Client, *batch.CronJob, error)
	}{
		{
			"Can create new CronJobs if they don't exist yet",
			controller,
			nil,
			&batch.CronJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "default-job",
					Namespace: "default",
				},
				Spec: batch.CronJobSpec{
					Schedule: "1 * * * *",
				},
			},
			func(ctx context.Context, g *GomegaWithT, c client.Client, d *batch.CronJob, err error) {
				g.Expect(err).To(Succeed())

				key := types.NamespacedName{
					Name:      d.GetName(),
					Namespace: d.GetNamespace(),
				}
				var res batch.CronJob
				err = c.Get(ctx, key, &res)
				g.Expect(err).To(Succeed())

				g.Expect(res.Spec.Schedule).To(Equal(d.Spec.Schedule))
			},
		},
		{
			"Can update existing CronJobs",
			controller,
			&batch.CronJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "default-job",
					Namespace: "default",
				},
				Spec: batch.CronJobSpec{
					Schedule: "1 * * * *",
				},
			},
			&batch.CronJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "default-job",
					Namespace: "default",
				},
				Spec: batch.CronJobSpec{
					Schedule: "2 * * * *",
				},
			},
			func(ctx context.Context, g *GomegaWithT, c client.Client, d *batch.CronJob, err error) {
				g.Expect(err).To(Succeed())

				key := types.NamespacedName{
					Name:      d.GetName(),
					Namespace: d.GetNamespace(),
				}
				var res batch.CronJob
				err = c.Get(ctx, key, &res)
				g.Expect(err).To(Succeed())

				g.Expect(res.Spec.Schedule).To(Equal(d.Spec.Schedule))
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

			tm := &typedMutator{
				mutator: m,
			}
			ctx := context.Background()

			if tt.existing != nil {
				err := c.Create(ctx, tt.existing)
				g.Expect(err).To(Succeed())
			}

			err := tm.MutateCronJob(ctx, tt.controller, tt.spec)
			tt.expectFn(ctx, g, c, tt.spec, err)
		})
	}
}
