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

package controllers

import (
	"context"
	"crypto/md5"
	"fmt"

	// wfv1 "github.com/argoproj/argo/v2/pkg/apis/workflow/v1alpha1"
	"github.com/dataworkz/kubeetl/api/v1alpha1"
	"github.com/dataworkz/kubeetl/listers"
	apitypes "k8s.io/apimachinery/pkg/types"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// WorkflowReconciler reconciles a Workflow object
type WorkflowReconciler struct {
	client.Client
	Scheme             *runtime.Scheme
	connectionLister   listers.ConnectionLister
	workflowLister     listers.WorkflowLister
	argoWorkflowLister listers.ArgoWorkflowLister
}

// +kubebilder:rbac:groups="",resources=secrets,verbs=create;delete
// +kubebuilder:rbac:groups=etl.dataworkz.nl,resources=workflows,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=etl.dataworkz.nl,resources=workflows/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=argoproj.io,resources=workflows,verbs=get;update;patch;delete;deletecollection

func (r *WorkflowReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	log := logf.Log.WithName("reconciler").WithName("workflow")

	var wf *v1alpha1.Workflow
	if err := r.Get(ctx, req.NamespacedName, wf); err != nil {
		log.Error(err, "unable to fetch Workflow")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if wf == nil {
		// create workflow
	}

	for _, ci := range wf.Connections {
		hashInput := fmt.Sprintf("%s%s", req.Name, ci.ConnectionName)
		secretName := fmt.Sprintf("%x", md5.Sum([]byte(hashInput)))
		key := apitypes.NamespacedName{Namespace: wf.Namespace, Name: secretName}

		var secret *corev1.Secret

		if err := r.Get(ctx, key, secret); err != nil {
			log.Error(err, "unable to list secrets")
			return ctrl.Result{}, err
		}
		if secret == nil {
			secret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretName,
					Namespace: req.Namespace,
				},
			}

			if err := r.Create(ctx, secret); err != nil {
				log.Error(err, "unable to create secret")
			}
		}
	}

	// reconcile:

	// Conversion:
	// - empty secret for each injected connection
	// - role scoped to new secret with get, update
	// - rolebinding for secretrole and workflow service account (default if not specified)

	// Initcontainer:
	// mount referenced secret-keys as env vars (or files?)
	// create secret with referenced keys rendered in template
	// template language? go template?

	// your logic here

	return ctrl.Result{}, nil
}

func (r *WorkflowReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Workflow{}).
		Complete(r)
}
