package v1alpha1

import (
	// "errors"

	"fmt"

	wfv1 "github.com/argoproj/argo/v2/pkg/apis/workflow/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	InjectableValueTypeFile InjectableValueType = "File"
	InjectableValueTypeEnv  InjectableValueType = "Env"
)

// +kubebuilder:object:root=true

// WorkflowList contains a list of ConnectionTypes
type WorkflowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workflow `json:"items"`
}

// +kubebuilder:object:root:=true
type Workflow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              WorkflowSpec `json:"spec"`
}

type WorkflowSpec struct {
	ArgoWorkflowSpec wfv1.WorkflowSpec          `json:",inline"`
	InjectableValues InjectableValues `json:"injectable,"`
	Templates        []Template                 `json:"templates"`
}

type InjectableValues []InjectableValue


// Creates []metav1.OwnerReference for objects owned by this Workflow
func (wf *Workflow) CreateOwnerReferences() []metav1.OwnerReference {
	ref := metav1.OwnerReference{
		APIVersion: wf.APIVersion,
		Kind: wf.Kind,
		Name: wf.Name,
		UID: wf.UID,
	}
	return []metav1.OwnerReference { ref }
}

func (wf *Workflow) GetInjectionSecret() corev1.LocalObjectReference {
	return corev1.LocalObjectReference{Name: fmt.Sprintf("%s-injection", wf.Name)}
}

func (wf *Workflow) GetInjectionSecretName() string {
	return fmt.Sprintf("%s-injections", wf.Name)
}

// Embedding type for Argo Template with Injections added
type Template struct {
	wfv1.Template  `json:",inline"`
	InjectedValues []string `json:"inject,omitempty"`
}

type InjectableValue struct {
	// Name of this InjectableValue
	// +required
	Name string `json:"name"`
	// FIXME: ConnectionRef with Name field?
	// Name of the `Connection` that is being injected here
	//+required
	ConnectionRef corev1.LocalObjectReference `json:"connectionRef,"`

	// If true, all the InjectDefinitions will be applied to all ContainerTemplates in this workflow.
	// If false, consuming templates must specifically request this ConnectionInjection
	//+optional
	Global bool `json:"global,omitempty"`

	// FIXME: meh
	// Name of the injected environment variable
	//+optional
	EnvName string `json:"envName,omitempty"`

	// Path where value will be mounted as a file
	//+optional
	MountPath string `json:"mountPath,omitempty"`

	// Go template that will be rendered using the connection fields as data
	// Example: mysql://{{.user}}:{{.password}}@{{.host}}:{{.port}}/{{.database}}
	//+required
	Content string `json:"content,"`
}

func (iv *InjectableValue) InjectionContainerArgument() string {
	return fmt.Sprintf("%s=%s:%s", iv.Name, iv.ConnectionRef.Name, iv.Content)
}

type InjectableValueType string

func (iv *InjectableValue) GetType() InjectableValueType {
	if iv.EnvName != "" {
		return InjectableValueTypeEnv
	}
	if iv.MountPath != "" {
		return InjectableValueTypeFile
	}

	// FIXME: not sure about this
	return ""
}
