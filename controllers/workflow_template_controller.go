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

// WorkflowTemplateReconciler reconciles a WorkflowTemplate object
type WorkflowTemplateReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	// ConnectionInjectionImage is the image of the container that will provide connection injections
	ConnectionInjectionImage string
}

// +kubebuilder:rbac:groups=etl.dataworkz.nl.dataworkz.nl,resources=workflowtemplates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=etl.dataworkz.nl.dataworkz.nl,resources=workflowtemplates/status,verbs=get;update;patch

func (r *WorkflowTemplateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("workflow", req.NamespacedName)

	var wft v1alpha1.WorkflowTemplate
	if err := r.Get(ctx, req.NamespacedName, &wft); err != nil {
		if !errors.IsNotFound(err) {
			log.Error(err, "unable to fetch WorkflowTemplate")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	acwf := wfv1.WorkflowTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      wft.Name,
			Namespace: wft.Namespace,
		},
	}
	_, err := ctrl.CreateOrUpdate(ctx, r.Client, &acwf, func() error { return r.updateWorkflowTemplate(&wft, &acwf) })
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error upserting argo workflow template: %w", err)
	}

	return ctrl.Result{}, nil
}

func (r *WorkflowTemplateReconciler) updateWorkflowTemplate(wft *v1alpha1.WorkflowTemplate, awft *wfv1.WorkflowTemplate) error {
	awfSpec, err := createArgoWorkflowSpec(wft.Spec.WorkflowSpec, awft.Name, r.ConnectionInjectionImage, awft.Namespace)
	if err != nil {
		return fmt.Errorf("error creating argo workflow spec: %w", err)
	}
	awft.Spec.WorkflowSpec = awfSpec
	if err := ctrl.SetControllerReference(wft, awft, r.Scheme); err != nil {
		return fmt.Errorf("error setting owner reference on workflow: %w", err)
	}

	return nil
}

func (r *WorkflowTemplateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.WorkflowTemplate{}).
		Complete(r)
}
