package controllers

import (
	"fmt"

	wfv1 "github.com/argoproj/argo/v2/pkg/apis/workflow/v1alpha1"
	"github.com/dataworkz/kubeetl/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"
)

// createArgoWorkflowSpec creates an Argo Workflow spec based on the supplied v1alpha1.WorkflowSpec
func createArgoWorkflowSpec(wfs v1alpha1.WorkflowSpec, wfName, connectionInjectionImage, namespace string) (wfv1.WorkflowSpec, error) {
	spec := wfs.ArgoWorkflowSpec
	v := corev1.Volume{
		Name: v1alpha1.NameWithHash(wfName),
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: v1alpha1.NameWithHash(wfName),
			},
		},
	}
	spec.Volumes = append(spec.Volumes, v)

	injectTmpl := wfv1.Template{
		Name:               "run-injection",
		Daemon:             pointer.BoolPtr(true),
		ServiceAccountName: wfs.InjectionServiceAccount,
		Container: &corev1.Container{
			Image: connectionInjectionImage,
			Args: []string{
				"--workflow",
				wfName,
				"--namespace",
				namespace,
			},
		},
	}

	oldEntrypoint := spec.Entrypoint
	steps := wfv1.Template{
		Name: "entrypoint",
		Steps: []wfv1.ParallelSteps{
			wfv1.ParallelSteps{
				Steps: []wfv1.WorkflowStep{
					wfv1.WorkflowStep{
						Name:     injectTmpl.Name,
						Template: injectTmpl.Name,
					},
				},
			},
			wfv1.ParallelSteps{
				Steps: []wfv1.WorkflowStep{
					wfv1.WorkflowStep{
						Name:     oldEntrypoint,
						Template: oldEntrypoint,
					},
				},
			},
		},
	}

	spec.Templates = append(spec.Templates, injectTmpl, steps)
	spec.Entrypoint = steps.Name

	for _, ii := range wfs.InjectInto {
		ic, err := newInjectionContext(&spec, wfs, wfName, ii)
		if err != nil {
			// TODO: log
			return wfv1.WorkflowSpec{}, err
		}

		template := getTemplateByName(&spec, ii.Name)
		if template == nil {
			// TODO: log?
			return wfv1.WorkflowSpec{}, fmt.Errorf("InjectInto contains missing template: %s", ii.Name)
		}
		if err := inject(template, ic); err != nil {
			return wfv1.WorkflowSpec{}, err
		}
	}
	return spec, nil
}

func getTemplateByName(spec *wfv1.WorkflowSpec, name string) *wfv1.Template {
	for _, t := range spec.Templates {
		if t.Name == name {
			return &t
		}
	}
	return nil
}

func newInjectionContext(awfSpec *wfv1.WorkflowSpec, wfSpec v1alpha1.WorkflowSpec, wfName string, injection v1alpha1.TemplateRef) (*injectionContext, error) {
	ic := injectionContext{
		awfSpec:        awfSpec,
		hashedWfName:   v1alpha1.NameWithHash(wfName),
		injectedValues: make([]v1alpha1.InjectableValue, 0, len(injection.InjectedValues)),
	}

	for _, v := range injection.InjectedValues {
		iv, err := wfSpec.GetInjectableValueByName(v)
		if err != nil {
			return nil, err
		}
		ic.injectedValues = append(ic.injectedValues, *iv)
	}
	return &ic, nil
}

type injectionContext struct {
	injectedValues []v1alpha1.InjectableValue
	awfSpec        *wfv1.WorkflowSpec
	hashedWfName   string
}

func inject(template *wfv1.Template, ic *injectionContext) error {
	switch tt := template.GetType(); tt {
	case wfv1.TemplateTypeDAG:
		return injectDAG(template, ic)
	case wfv1.TemplateTypeSteps:
		return injectSteps(template, ic)
	case wfv1.TemplateTypeScript:
		return injectContainer(&template.Script.Container, ic)
	case wfv1.TemplateTypeContainer:
		return injectContainer(template.Container, ic)
	}

	return nil
}

func (ic *injectionContext) getSecretKeyRef(injectableValue string) corev1.SecretKeySelector {
	sks := corev1.SecretKeySelector{
		// todo: get as workflow method?
		LocalObjectReference: corev1.LocalObjectReference{
			Name: ic.hashedWfName,
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
				Name: iv.EnvName,
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &sks,
				},
			}
			container.Env = addEnvVar(container.Env, ev)
		case v1alpha1.InjectableValueTypeFile:
			vm := corev1.VolumeMount{
				MountPath: iv.MountPath,
				Name:      ic.hashedWfName,
				SubPath:   iv.Name,
			}
			container.VolumeMounts = addVolumeMount(container.VolumeMounts, vm)
		}
	}
	return nil
}

func addEnvVar(vars []corev1.EnvVar, ev corev1.EnvVar) []corev1.EnvVar {
	var found bool
	for _, e := range vars {
		if e.Name == ev.Name {
			found = true
		}
	}
	if !found {
		vars = append(vars, ev)
	}
	return vars
}

func addVolumeMount(mounts []corev1.VolumeMount, vm corev1.VolumeMount) []corev1.VolumeMount {
	var found bool
	for _, v := range mounts {
		if v.Name == vm.Name {
			found = true
		}
	}
	if !found {
		mounts = append(mounts, vm)
	}
	return mounts
}

func injectDAG(template *wfv1.Template, ic *injectionContext) error {
	errors := make([]error, 0)
	for _, dagTask := range template.DAG.Tasks {
		target := getTemplateByName(ic.awfSpec, dagTask.Template)
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
			target := getTemplateByName(ic.awfSpec, step.Template)
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
