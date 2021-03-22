package controllers

import (
	"context"

	"github.com/go-logr/logr"
	batch "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/dataworkz/kubeetl/api/v1alpha1"
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

	cronJob := DataSetToCronJob(dataSet)
	if err := r.TypedMutator.MutateCronJob(ctx, &dataSet, &cronJob); err != nil {
		log.Error(err, "Could not create CronJob for DataSet HealthCheck")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func DataSetToCronJob(dataSet v1alpha1.DataSet) batch.CronJob {
	meta := metav1.ObjectMeta{
		Name:      dataSet.Name,
		Namespace: dataSet.Namespace,
	}

	spec := dataSet.Spec.HealthCheck

	// TODO define sane defaults for the HealthCheck, maybe JobTemplate is to much to handle for
	// users
	// Always is not OK
	spec.JobTemplate.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyOnFailure

	return batch.CronJob{
		ObjectMeta: meta,
		Spec:       spec,
	}
}

func (r *DataSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.TypedMutator == nil {
		r.TypedMutator = mutators.New(r.Client, r.Scheme, r.Log)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&api.DataSet{}).
		Complete(r)
}
