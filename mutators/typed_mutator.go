package mutators

import (
	"context"

	"github.com/go-logr/logr"
	batch "k8s.io/api/batch/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TypedMutator has mutation methods for all common resource kinds
type TypedMutator interface {
	MutateCronJob(ctx context.Context, controller client.Object, obj *batch.CronJob) error
}

type typedMutator struct {
	mutator Mutator
}

// New returns a new TypedMutator
func New(client client.Client, scheme *runtime.Scheme, log logr.Logger) TypedMutator {
	m := &typedMutator{
		&genericMutator{
			client: client,
			scheme: scheme,
			log:    log,
		},
	}

	return m
}

// TODO use a label to maintain last applied config
// Due to all the status fields in the pod spec it is hard to check field by field

func (m *typedMutator) MutateCronJob(ctx context.Context, controller client.Object, obj *batch.CronJob) error {
	err := m.mutator.Mutate(ctx, controller, obj, func(existing, expected client.Object) error {
		// existingObj := existing.(*batch.CronJob)
		// expectedObj := expected.(*batch.CronJob)

		return nil
	})

	return err
}
