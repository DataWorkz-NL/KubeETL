package controllers

import (
	"context"
	"fmt"

	wfv1 "github.com/argoproj/argo/v2/pkg/apis/workflow/v1alpha1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

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
			if dataSet.Spec.HealthCheck == nil || dataSet.Spec.HealthCheck.GetNamespacedName() != key {
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
		} else {
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
			} else {
				dataSet.Status.Healthy = api.Healthy
			}

			err = r.Status().Update(ctx, &dataSet)
			if err != nil {
				log.Error(err, "unable to update DataSet status")
				return ctrl.Result{}, err
			}
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
	// TODO set up watch for api.Workflow status changes so that we can reconcile the DataSet status if the Workflow Failed
	// Figure out whether the api.Workflow status changes are triggered if the underlying argo workflow fails/succeeds. If not we need
	// to ensure this happens
	wfKind := &source.Kind{Type: &api.Workflow{}}
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.DataSet{}).
		Watches(wfKind, workflowEventHandler()).
		Complete(r)
}

// workflowEventHandler returns a custom event handler to translate Workflow events into Dataset events.
// If the workflow has an "etl.dataworkz.nl/healthcheck" label, that can be used to translate the label values
// into DataSet. The label value is a comma seperated list of DataSet names. The DataSet is assumed to be
// in the same namespace as the workflow.
func workflowEventHandler() handler.EventHandler {
	mapFn := func(obj client.Object) []reconcile.Request {
		l := obj.GetLabels()
		if labels.HasLabel(l, healthcheckLabel) {
			s := labels.GetLabelValue(l, healthcheckLabel)
			ss := labels.StringSet(s)
			names := ss.Split()
			var requests []reconcile.Request
			for _, name := range names {
				req := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      name,
						Namespace: obj.GetNamespace(),
					},
				}
				requests = append(requests, req)
			}
			return requests
		}

		return []reconcile.Request{}
	}

	return handler.EnqueueRequestsFromMapFunc(mapFn)
}
