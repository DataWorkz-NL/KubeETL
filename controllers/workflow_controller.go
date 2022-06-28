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

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	wfv1 "github.com/argoproj/argo/v2/pkg/apis/workflow/v1alpha1"
	"github.com/dataworkz/kubeetl/api/v1alpha1"
)

// WorkflowReconciler reconciles a Workflow object
type WorkflowReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	// ConnectionInjectionImage is the image of the container that will provide connection injections
	ConnectionInjectionImage string
}

// +kubebuilder:rbac:groups=etl.dataworkz.nl.dataworkz.nl,resources=workflows,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=etl.dataworkz.nl.dataworkz.nl,resources=workflows/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=argoproj.io,resources=workflows,verbs=get;list;watch;create;update;patch;delete

func (r *WorkflowReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("workflow", req.NamespacedName)

	var workflow v1alpha1.Workflow
	if err := r.Get(ctx, req.NamespacedName, &workflow); err != nil {
		if !errors.IsNotFound(err) {
			log.Error(err, "unable to fetch Workflow")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	awf := wfv1.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: workflow.Namespace,
			Name:      workflow.Name,
		},
	}
	_, err := ctrl.CreateOrUpdate(ctx, r.Client, &awf, func() error { return r.updateWorkflow(&workflow, &awf) })
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error upserting argo workflow: %w", err)
	}

	return ctrl.Result{}, nil
}

func (r *WorkflowReconciler) updateWorkflow(workflow *v1alpha1.Workflow, awf *wfv1.Workflow) error {
	awfSpec, err := createArgoWorkflowSpec(workflow.Spec, awf.Name, r.ConnectionInjectionImage, awf.Namespace)
	if err != nil {
		return fmt.Errorf("error creating argo workflow spec: %w", err)
	}
	awf.Spec = awfSpec
	if err := ctrl.SetControllerReference(workflow, awf, r.Scheme); err != nil {
		return fmt.Errorf("error setting owner reference on workflow: %w", err)
	}

	return nil
}

func (r *WorkflowReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Workflow{}).
		Complete(r)
}
