package mutators

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Mutator ensures the Create Or Update logic is consistently applied, no unnecessary reconciliations
// are performed and the controller reference is set.
type Mutator interface {
	// Mutate will change the state of the cluster from the existing
	// to the expected state
	Mutate(ctx context.Context, controller, spec client.Object, f MergeFn) error
}

type genericMutator struct {
	client client.Client
	scheme *runtime.Scheme
	log    logr.Logger
}

// MergeFn merges the expected object into the existing object.
// To avoid overriding e.g. status fields or immutable fields
// it is better to copy over individual fields
// as opposed to copying the entire object.
type MergeFn func(existing, expected client.Object) error

// Mutate
func (g *genericMutator) Mutate(ctx context.Context, controller, spec client.Object, f MergeFn) error {
	g.log.WithValues("controller", controller, "spec", spec)
	expected := spec.DeepCopyObject().(client.Object)
	err := g.setControllerReference(controller, spec)
	if err != nil {
		return fmt.Errorf("failed to set controller reference: %w", err)
	}

	op, err := ctrl.CreateOrUpdate(ctx, g.client, spec, func() error {
		return f(spec, expected)
	})
	g.log.WithValues("operation", op).Info("create or update operation")
	if err != nil {
		return fmt.Errorf("could not mutate state: %w", err)
	}

	return nil
}

func (g *genericMutator) setControllerReference(controller, obj runtime.Object) error {
	controllerMo, ok := controller.(metav1.Object)
	if !ok {
		return fmt.Errorf("%T is not a metav1.Object, cannot call setControllerReference", controller)
	}
	objMo, ok := obj.(metav1.Object)
	if !ok {
		return fmt.Errorf("%T is not a metav1.Object, cannot call setControllerReference", obj)
	}
	return controllerutil.SetControllerReference(controllerMo, objMo, g.scheme)
}
