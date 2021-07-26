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
	"os"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	"k8s.io/apimachinery/pkg/api/errors"
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
}

// +kubebuilder:rbac:groups=etl.dataworkz.nl.dataworkz.nl,resources=workflows,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=etl.dataworkz.nl.dataworkz.nl,resources=workflows/status,verbs=get;update;patch

func (r *WorkflowReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("workflow", req.NamespacedName)

	var workflow v1alpha1.Workflow
	if err := r.Get(ctx, req.NamespacedName, &workflow); err != nil {
		if !errors.IsNotFound(err) {
			log.Error(err, "unable to fetch Workflow")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	csn := workflow.ConnectionSecretName()
	cs := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: csn.Namespace,
			Name:      csn.Name,
		},
	}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, &cs, func() error { return r.updateSecret(&workflow, &cs) })
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error creating workflow connection secret: %w", err)
	}

	var awf wfv1.Workflow
	_, err = ctrl.CreateOrUpdate(ctx, r.Client, &awf, func() error { return r.updateWorkflow(&workflow, &awf) })

	return ctrl.Result{}, nil
}

func (r *WorkflowReconciler) updateSecret(workflow *v1alpha1.Workflow, secret *corev1.Secret) error {
	if err := ctrl.SetControllerReference(workflow, secret, r.Scheme); err != nil {
		return fmt.Errorf("error setting owner reference on connection secret: %w", err)
	}
	return nil
}

func (r *WorkflowReconciler) updateWorkflow(workflow *v1alpha1.Workflow, awf *wfv1.Workflow) error {
	awf.Spec = workflow.Spec.ArgoWorkflowSpec
	v := corev1.Volume{
		Name: workflow.ConnectionVolumeName(),
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: workflow.ConnectionSecretName().Name,
			},
		},
	}
	awf.Spec.Volumes = append(awf.Spec.Volumes, v)

	// awf.Spec.Entrypoint
	// awf.Spec.Entrypoint
	injectTmpl := wfv1.Template{
		Name:               "run-injection",
		Daemon:             pointer.BoolPtr(true),
		ServiceAccountName: workflow.Spec.InjectionServiceAccount,
		Container: &corev1.Container{
			Image: os.Getenv("KETL_INJECTION_CONTAINER"),
			Args: []string{
				"--workflow",
				awf.Name,
				"--namespace",
				awf.Namespace,
			},
		},
	}

	steps := wfv1.Template{
		Steps: []wfv1.ParallelSteps{
			wfv1.ParallelSteps{
				Steps: []wfv1.WorkflowStep{
					{
						Name:     injectTmpl.Name,
						Template: injectTmpl.Name,
					},
				},
			},
			wfv1.ParallelSteps{
				[]wfv1.WorkflowStep{
					{
						Name:     awf.Spec.Entrypoint,
						Template: awf.Spec.Entrypoint,
					},
				},
			},
		},
	}

	awf.Spec.Templates = append(awf.Spec.Templates, injectTmpl, steps)
	awf.Spec.Entrypoint = steps.Name

	for _, ii := range workflow.Spec.InjectInto {
		ic, err := newInjectionContext(awf, workflow, ii)
		if err != nil {
			// TODO: log
			return err
		}

		template := awf.GetTemplateByName(ii.Name)
		if template == nil {
			// TODO: log?
			return fmt.Errorf("InjectInto contains missing template: %s", ii.Name)
		}
		inject(template, ic)
	}

	if err := ctrl.SetControllerReference(workflow, awf, r.Scheme); err != nil {
		return fmt.Errorf("error setting owner reference on connection secret: %w", err)
	}

	return nil
}

func newInjectionContext(awf *wfv1.Workflow, wf *v1alpha1.Workflow, injection v1alpha1.TemplateRef) (*injectionContext, error) {
	ic := injectionContext{
		awf:            awf,
		templates:      awf.Spec.Templates,
		injectedValues: make([]v1alpha1.InjectableValue, 0, len(injection.InjectedValues)),
	}

	iv, err := wf.GetInjectableValueByName(injection.Name)
	if err != nil {
		return nil, err
	}
	ic.injectedValues = append(ic.injectedValues, *iv)
	return &ic, nil
}

type injectionContext struct {
	wf             *v1alpha1.Workflow
	awf            *wfv1.Workflow
	templates      []wfv1.Template
	injectedValues []v1alpha1.InjectableValue
}

func inject(template *wfv1.Template, ic *injectionContext) error {
	switch tt := template.GetType(); tt {
	case wfv1.TemplateTypeDAG:
		return injectDAG(template, ic)
		// inject into template.DAG.Tasks
		// recursion is easy here, just call InjectTemplate on each referenced Task
	case wfv1.TemplateTypeSteps:
		return injectSteps(template, ic)
		// inject into template.DAG.Tasks
		// recursion is easy here, just call InjectTemplate on each referenced Task
	case wfv1.TemplateTypeScript:
		return injectContainer(&template.Script.Container, ic)
	case wfv1.TemplateTypeContainer:
		return injectContainer(template.Container, ic)
	case wfv1.TemplateTypeSuspend:
	case wfv1.TemplateTypeResource:
	case wfv1.TemplateTypeUnknown:
		// error can't inject into (templatetype)
	}

	// if we're here, it means this is a regular template
	return nil
}

func (ic *injectionContext) getSecretKeyRef(injectableValue string) corev1.SecretKeySelector {
	sks := corev1.SecretKeySelector{
		// todo: get as workflow method?
		LocalObjectReference: corev1.LocalObjectReference{
			Name: ic.awf.Name,
		},
		Key: injectableValue,
	}
	return sks
}

func injectContainer(container *corev1.Container, ic *injectionContext) error {
	for _, iv := range ic.injectedValues {
		sks := ic.getSecretKeyRef(iv.Name)

		switch iv.GetType() {
		case v1alpha1.InjectableValueTypeEnv:
			ev := corev1.EnvVar{
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &sks,
				},
			}
			addEnvVar(container.Env, ev)
		case v1alpha1.InjectableValueTypeFile:
			vm := corev1.VolumeMount{
				MountPath: iv.MountPath,
				Name:      ic.wf.Name,
				SubPath:   iv.Name,
			}
			addVolumeMount(container.VolumeMounts, vm)
		}
	}
	return nil
}

func addEnvVar(vars []corev1.EnvVar, ev corev1.EnvVar) {
	var found bool
	for _, e := range vars {
		if e.Name == ev.Name {
			found = true
		}
	}
	if !found {
		vars = append(vars, ev)
	}
}

func addVolumeMount(mounts []corev1.VolumeMount, vm corev1.VolumeMount) {
	var found bool
	for _, v := range mounts {
		if v.Name == vm.Name {
			found = true
		}
	}
	if !found {
		mounts = append(mounts, vm)
	}
}

func injectDAG(template *wfv1.Template, ic *injectionContext) error {
	errors := make([]error, 0)
	for _, dagTask := range template.DAG.Tasks {
		target := ic.awf.GetTemplateByName(dagTask.Template)
		// TODO: handle nil target
		err := inject(target, ic)
		if err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		// TODO: tidy
		return fmt.Errorf("error while injecting dag template: %s", template.Name)
	}
	return nil
}

func injectSteps(template *wfv1.Template, ic *injectionContext) error {
	errors := make([]error, 0)
	for _, pstep := range template.Steps {
		for _, step := range pstep.Steps {
			target := ic.awf.GetTemplateByName(step.Name)
			// TODO: handle nil target
			err := inject(target, ic)
			if err != nil {
				errors = append(errors, err)
			}
		}
	}
	if len(errors) > 0 {
		// TODO: tidy
		return fmt.Errorf("error while injecting steps template: %s", template.Name)
	}
	return nil
}

func (r *WorkflowReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Workflow{}).
		Complete(r)
}
