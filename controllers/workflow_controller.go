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
	"errors"
	"fmt"

	wfv1 "github.com/argoproj/argo/v2/pkg/apis/workflow/v1alpha1"
	"github.com/dataworkz/kubeetl/api/v1alpha1"
	"github.com/dataworkz/kubeetl/listers"
	// apitypes "k8s.io/apimachinery/pkg/types"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	injectionTemplateName = "kubeetl-connection-setup"
)

type Config struct {
	// Image for constructing carrier secrets for Workflow connections
	InjectorImage string
}

// WorkflowReconciler reconciles a Workflow object
type WorkflowReconciler struct {
	client.Client
	Scheme             *runtime.Scheme
	connectionLister   listers.ConnectionLister
	workflowLister     listers.WorkflowLister
	argoWorkflowLister listers.ArgoWorkflowLister

	Config Config
}

func (r *WorkflowReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Workflow{}).
		Complete(r)
}

// +kubebilder:rbac:groups="",resources=secrets,verbs=create;delete,get
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

	var awf *wfv1.Workflow
	if err := r.Get(ctx, req.NamespacedName, awf); err != nil {
		log.Error(err, "unable to get Argo Workflow")
		return ctrl.Result{}, err
	}
	if awf == nil {
		awf, err := r.CreateInjectedArgoWorkflow(*wf)
		if err != nil {
			log.Error(err, "unable to create injected argo workflow")
			return ctrl.Result{}, err
		}
		r.Client.Create(ctx, &awf)
		// create workflow
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

type valueInjector struct {
	// Name of the secret that contains the rendered InjectableValues
	SecretName  string
	EnvVar      *corev1.EnvVar
	VolumeMount *corev1.VolumeMount
}

func (vi *valueInjector) Inject(target *wfv1.Template) {
	if vi.EnvVar != nil {
		target.Container.Env = append(target.Container.Env, *vi.EnvVar)
	}
	if vi.VolumeMount != nil {
		target.Container.VolumeMounts = append(target.Container.VolumeMounts, *vi.VolumeMount)
	}
}

func (r *WorkflowReconciler) CreateInjectedArgoWorkflow(wf v1alpha1.Workflow) (wfv1.Workflow, error) {
	awf := &wfv1.Workflow{
		Spec:       wf.Spec.ArgoWorkflowSpec,
		ObjectMeta: metav1.ObjectMeta{Name: wf.Name},
	}

	awfSpec := &awf.Spec
	awfSpec.Templates = make([]wfv1.Template, len(wf.Spec.Templates))

	injectors := getInjectors(wf)

	injectionArgs := []string{
		"--target-secret", wf.GetInjectionSecretName(),
	}

	// create volumes
	for _, iv := range wf.Spec.InjectableValues {
		switch iv.GetType() {
		case v1alpha1.InjectableValueTypeFile:
			vol := corev1.Volume{
				Name: iv.Name,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: wf.GetInjectionSecretName(),
					},
				},
			}
			awfSpec.Volumes = append(awfSpec.Volumes, vol)
		}

		injectionArgs = append(injectionArgs, "--value", iv.InjectionContainerArgument())
	}

	for i, tmpl := range wf.Spec.Templates {
		if tmpl.GetType() == wfv1.TemplateTypeContainer {
			atmpl := &awfSpec.Templates[i]

			for _, key := range tmpl.InjectedValues {
				injector, ok := injectors[key]
				if !ok {
					msg := fmt.Sprintf("couldn't find injectable value with key %s", key)
					return wfv1.Workflow{}, errors.New(msg)
				}

				injector.Inject(atmpl)
			}
		}

		awfSpec.Templates[i] = tmpl.Template
	}

	// creates the secret
	// FIXME: still need to mount the connection fields
	// TODO: naming not consistent: connectionSetup/injection
	connectionSetupTemplate := wfv1.Template{
		Name: injectionTemplateName,
		Container: &corev1.Container{
			Name:  "kubeetl-injector",
			Image: r.Config.InjectorImage,
			Args:  injectionArgs,
		},
	}

	addInjectionDaemon(awf, connectionSetupTemplate)
	awf.ObjectMeta.SetOwnerReferences(wf.CreateOwnerReferences())

	return *awf, nil
}

func addInjectionDaemon(wf *wfv1.Workflow, injectionTemplate wfv1.Template) {
	// TODO: big inefficiency here, we're spawning multiple daemons that watch/update secrets
	// this is necessary to ensure no steps run before secrets are setup
	// but it also means we have multiple redundant processes putting load on the k8s api

	for _, t := range wf.Spec.Templates {
		if t.GetType() == wfv1.TemplateTypeSteps {
			// inject in parallel steps
			for _, parallelStep := range t.Steps {
				step := wfv1.WorkflowStep{
					Name:     injectionTemplate.Name,
					Template: injectionTemplate.Name,
				}

				parallelStep.Steps = append([]wfv1.WorkflowStep{step}, parallelStep.Steps...)
			}
		}
		if t.GetType() == wfv1.TemplateTypeDAG {
			connectionSetupTask := wfv1.DAGTask{
				Name:     injectionTemplateName,
				Template: injectionTemplate.Name,
			}

			for _, task := range t.DAG.Tasks {
				// TODO: we could only make branches that actually use connections depend on the injection task
				// this requires traversing the graph
				if len(task.Dependencies) != 0 {
					task.Dependencies = append(task.Dependencies, connectionSetupTask.Name)
				}
			}

			t.DAG.Tasks = append(t.DAG.Tasks, connectionSetupTask)
		}
	}
}

func getInjectors(wf v1alpha1.Workflow) map[string]valueInjector {
	injectors := make(map[string]valueInjector)

	for _, iv := range wf.Spec.InjectableValues {
		var ev *corev1.EnvVar
		var vm *corev1.VolumeMount

		switch iv.GetType() {
		case v1alpha1.InjectableValueTypeEnv:
			ev = &corev1.EnvVar{
				Name: iv.EnvName,
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: wf.GetInjectionSecret(),
						Key:                  iv.Name,
					},
				},
			}
			break
		case v1alpha1.InjectableValueTypeFile:
			vm = &corev1.VolumeMount{
				MountPath: iv.MountPath,
				Name:      iv.Name,
				ReadOnly:  true,
			}
			break
		}

		injectors[iv.Name] = valueInjector{EnvVar: ev, VolumeMount: vm, SecretName: wf.GetInjectionSecretName()}
	}

	return injectors
}
