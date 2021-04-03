package controllers

import (
	"context"
	"fmt"

	wfv1 "github.com/argoproj/argo/v2/pkg/apis/workflow/v1alpha1"
	"github.com/go-logr/logr"
	batch "k8s.io/api/batch/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	api "github.com/dataworkz/kubeetl/api/v1alpha1"
	"github.com/dataworkz/kubeetl/mutators"
)

type DataSetReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	mutators.TypedMutator
}

func (r *DataSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("dataset", req.NamespacedName)

	var dataSet api.DataSet
	if err := r.Get(ctx, req.NamespacedName, &dataSet); err != nil {
		if !errors.IsNotFound(err) {
			log.Error(err, "unable to fetch DataSet")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// TODO refactor this piece of nested crap
	if dataSet.Spec.HealthCheck != nil {
		var workflow api.Workflow
		if err := r.Get(ctx, dataSet.Spec.HealthCheck.GetNamespacedName(), &workflow); err != nil {
			log.Error(err, "unable to fetch Workflow for DataSet")

			// TODO extract dataset status updates into function
			dataSet.Status.Healthy = api.Unknown
			err = r.Status().Update(ctx, &dataSet)
			if err != nil {
				log.Error(err, "unable to update DataSet status")
				return ctrl.Result{}, err
			}
		}

		wfr := workflow.Status.ArgoWorkflowRef
		if wfr != nil {
			var argoWorkflow wfv1.Workflow
			key := types.NamespacedName{
				Name:      wfr.Name,
				Namespace: wfr.Namespace,
			}
			if err := r.Get(ctx, key, &argoWorkflow); err != nil {
				log.Error(err, "unable to fetch ArgoWorkflow for DataSet")
				dataSet.Status.Healthy = api.Unknown

				err = r.Status().Update(ctx, &dataSet)
				if err != nil {
					log.Error(err, "unable to update DataSet status")
					return ctrl.Result{}, err
				}
			}

			if argoWorkflow.Status.Failed() {
				dataSet.Status.Healthy = api.Unhealthy
			} else {
				dataSet.Status.Healthy = api.Healthy
			}

			err := r.Status().Update(ctx, &dataSet)
			if err != nil {
				log.Error(err, "unable to update DataSet status")
				return ctrl.Result{}, err
			}
		}

		dataSet.Status.Healthy = api.Unknown

		err := r.Status().Update(ctx, &dataSet)
		if err != nil {
			log.Error(err, "unable to update DataSet status")
			return ctrl.Result{}, err
		}

	}

	return ctrl.Result{}, nil
}

func (r *DataSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.TypedMutator == nil {
		r.TypedMutator = mutators.New(r.Client, r.Scheme, r.Log)
	}

	// TODO update index
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &batch.CronJob{}, ".metadata.controller", func(rawObj client.Object) []string {
		job := rawObj.(*batch.CronJob)
		owner := metav1.GetControllerOf(job)
		if owner == nil {
			return nil
		}

		if owner.APIVersion != api.GroupVersion.String() || owner.Kind != "DataSet" {
			return nil
		}

		return []string{owner.Name}
	}); err != nil {
		return fmt.Errorf("failed to set up index on CronJob: %w", err)
	}

	// TODO set up watch for api.Workflow status changes so that we can reconcile the DataSet status if the Workflow Failed
	// Figure out whether the api.Workflow status changes are triggered if the underlying argo workflow fails/succeeds. If not we need
	// to ensure this happens
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.DataSet{}).
		Owns(&batch.CronJob{}).
		Complete(r)
}
