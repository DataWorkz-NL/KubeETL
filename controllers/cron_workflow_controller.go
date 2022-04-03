/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/* TODO:
* - rbac + sa voor inject container
 */
package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	wfv1 "github.com/argoproj/argo/v2/pkg/apis/workflow/v1alpha1"
	"github.com/dataworkz/kubeetl/api/v1alpha1"
)

// CronWorkflowReconciler reconciles a CronWorkflow object
type CronWorkflowReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	// ConnectionInjectionImage is the image of the container that will provide connection injections
	ConnectionInjectionImage string
}

// +kubebuilder:rbac:groups=etl.dataworkz.nl.dataworkz.nl,resources=cronworkflows,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=etl.dataworkz.nl.dataworkz.nl,resources=cronworkflows/status,verbs=get;update;patch

func (r *CronWorkflowReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("workflow", req.NamespacedName)

	var cwf v1alpha1.CronWorkflow
	if err := r.Get(ctx, req.NamespacedName, &cwf); err != nil {
		if !errors.IsNotFound(err) {
			log.Error(err, "unable to fetch CronWorkflow")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	cs := v1alpha1.ConnectionSecret(cwf.Name, cwf.Namespace)

	log.Info("creating connection secret", "name", cs.Name, "namespace", cs.Namespace)

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, &cs, func() error { return r.updateSecret(&cwf, &cs) })
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error creating workflow connection secret: %w", err)
	}

	acwf := wfv1.CronWorkflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cwf.Name,
			Namespace: cwf.Namespace,
		},
	}
	_, err = ctrl.CreateOrUpdate(ctx, r.Client, &acwf, func() error { return r.updateCronWorkflow(&cwf, &acwf) })
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error upserting argo workflow: %w", err)
	}

	return ctrl.Result{}, nil
}

func (r *CronWorkflowReconciler) updateSecret(cwf *v1alpha1.CronWorkflow, secret *corev1.Secret) error {
	if err := ctrl.SetControllerReference(cwf, secret, r.Scheme); err != nil {
		return fmt.Errorf("error setting owner reference on connection secret: %w", err)
	}
	return nil
}

func (r *CronWorkflowReconciler) updateCronWorkflow(cwf *v1alpha1.CronWorkflow, acwf *wfv1.CronWorkflow) error {
	awfSpec, err := createArgoWorkflowSpec(cwf.Spec.WorkflowSpec, acwf.Name, r.ConnectionInjectionImage, acwf.Namespace)
	if err != nil {
		return fmt.Errorf("error creating argo workflow spec: %w", err)
	}
	acwf.Spec.WorkflowSpec = awfSpec
	if err := ctrl.SetControllerReference(cwf, acwf, r.Scheme); err != nil {
		return fmt.Errorf("error setting owner reference on workflow: %w", err)
	}

	return nil
}

func (r *CronWorkflowReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.CronWorkflow{}).
		Complete(r)
}
