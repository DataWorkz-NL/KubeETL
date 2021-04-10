package controllers

import (
	"context"
	"fmt"

	wfv1 "github.com/argoproj/argo/v2/pkg/apis/workflow/v1alpha1"
	"github.com/go-logr/logr"
	batch "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	api "github.com/dataworkz/kubeetl/api/v1alpha1"
	"github.com/dataworkz/kubeetl/labels"
)

type DataSetReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

const (
	healthcheckLabel = "etl.dataworkz.nl/healthcheck"
)

func (r *DataSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("dataset", req.NamespacedName)

	var dataSet api.DataSet
	if err := r.Get(ctx, req.NamespacedName, &dataSet); err != nil {
		if !errors.IsNotFound(err) {
			log.Error(err, "unable to fetch DataSet")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Remove any labels of workflows not currently used as healthcheck
	var wfl api.WorkflowList
	requirement, err := k8slabels.NewRequirement(healthcheckLabel, selection.Exists, []string{})
	if err != nil {
		log.Error(err, "unable to create label selector for workflows")
		return ctrl.Result{}, err
	}

	labelSelector := k8slabels.NewSelector().Add(*requirement)
	if err := r.List(ctx, &wfl, &client.ListOptions{LabelSelector: labelSelector}); err != nil {
		log.Error(err, "unable to list Workflows")
	} else {
		for _, wf := range wfl.Items {
			key := types.NamespacedName{
				Name:      wf.Name,
				Namespace: wf.Namespace,
			}
			// If the workflow is not used as healthcheck for this DataSet, remove the DataSet from label
			if dataSet.Spec.HealthCheck.GetNamespacedName() != key {
				val := labels.GetLabelValue(wf.Labels, healthcheckLabel)
				ss := labels.StringSet(val)
				newSs := ss.Remove(dataSet.Name)
				if newSs.IsEmpty() {
					newLabels := labels.RemoveLabel(wf.Labels, healthcheckLabel)
					wf.Labels = newLabels
				} else {
					// TODO move to labels package
					wf.Labels[healthcheckLabel] = string(newSs)
				}

				err := r.Update(ctx, &wf)
				if err != nil {
					log.Error(err, "unable to update Workflow labels")
					return ctrl.Result{}, err
				}
			}
		}
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

		// TODO check for label vale of healtcheck label, set it if it isn't there
		// otherwise check if the value contains a reference to this dataset and update it
		// if it doesn't
		val := labels.GetLabelValue(workflow.Labels, healthcheckLabel)
		if val == "" {
			workflow.Labels = labels.AddLabel(workflow.Labels, healthcheckLabel, dataSet.Name)
			err := r.Update(ctx, &workflow)
			if err != nil {
				log.Error(err, "unable to update Workflow labels")
				return ctrl.Result{}, err
			}
		}

		failed, err := r.getArgoWorkflowStatus(ctx, workflow.Status.ArgoWorkflowRef)
		if err != nil {
			dataSet.Status.Healthy = api.Unknown
		} else if failed {
			dataSet.Status.Healthy = api.Unhealthy
		}

		err = r.Status().Update(ctx, &dataSet)
		if err != nil {
			log.Error(err, "unable to update DataSet status")
			return ctrl.Result{}, err
		} else {
			dataSet.Status.Healthy = api.Healthy
		}

	}

	return ctrl.Result{}, nil
}

func (r *DataSetReconciler) getArgoWorkflowStatus(ctx context.Context, wfr *corev1.ObjectReference) (bool, error) {
	if wfr != nil {
		var argoWorkflow wfv1.Workflow
		key := types.NamespacedName{
			Name:      wfr.Name,
			Namespace: wfr.Namespace,
		}
		if err := r.Get(ctx, key, &argoWorkflow); err != nil {
			return false, fmt.Errorf("unable to fetch ArgoWorkflow for DataSet")
		}

		return argoWorkflow.Status.Failed(), nil
	}

	return false, fmt.Errorf("No ArgoWorkflow created for Workflow")
}

func (r *DataSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
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
