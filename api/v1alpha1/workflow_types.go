package v1alpha1

import (
	"bytes"
	"fmt"
	"text/template"

	wfv1 "github.com/argoproj/argo/v2/pkg/apis/workflow/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	InjectableValueTypeFile InjectableValueType = "File"
	InjectableValueTypeEnv  InjectableValueType = "Env"
)

// +kubebuilder:object:root=true

// WorkflowList contains a list of Workflows
type WorkflowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workflow `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Workflow is the schema for the workflows API
type Workflow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkflowSpec   `json:"spec"`
	Status WorkflowStatus `json:"status,omitempty"`
}

// ConnectionStatus defines the observed state of Workflow
type WorkflowStatus struct {
	// ArgoWorkflowRef is a reference to the Argo Workflow created for this Workflow
	ArgoWorkflowRef *corev1.ObjectReference `json:"argoWorkflowRef,omitempty"`
}

// WorkflowSpec defines the desired state of Workflow
type WorkflowSpec struct {
	// ArgoWorkflowSpec is an embedded WorkflowSpec from Argo Workflows
	// +required
	ArgoWorkflowSpec wfv1.WorkflowSpec `json:",inline"`

	// InjectableValues defines a collection of InjectableValues for this Workflow
	// +optional
	InjectableValues InjectableValues `json:"injectable,omitempty"`

	// InjectInto contains the templates for which KubeETL should inject
	// the InjectableValues. This should refer to the defined Templates of the
	// WorkflowSpec.
	// +optional
	InjectInto []TemplateRef `json:"injectInto"`

	// InjectionServiceAccount is the name of the service account used to inject connections.
	// This defaults to the Workflow service account.
	// +optional
	InjectionServiceAccount string `json:"injectionServiceAccount"`
}

type InjectableValues []InjectableValue

// TemplateRef extends an Argo Template with additional functionality
type TemplateRef struct {
	// +required
	Name string `json:"name"`

	// InjectedValues contains a list of InjectableValue names that will be injected in this Template
	// +optional
	InjectedValues []string `json:"inject,omitempty"`
}

type InjectableValue struct {
	// Name of this InjectableValue
	// +required
	Name string `json:"name"`

	// Name of the `Connection` that is being injected here
	// +optional
	ConnectionRef corev1.LocalObjectReference `json:"connectionRef"`

	// Name of the `DataSet` that is being injected here
	// +optional
	DataSetRef corev1.LocalObjectReference `json:"dataSetRef"`

	// If true, all the InjectDefinitions will be applied to all ContainerTemplates in this workflow.
	// If false, consuming templates must specifically request this ConnectionInjection
	// +optional
	Global bool `json:"global,omitempty"`

	// Name of the injected environment variable
	// +optional
	EnvName string `json:"envName,omitempty"`

	// Path where value will be mounted as a file
	// +optional
	MountPath string `json:"mountPath,omitempty"`

	// Go template that will be rendered using the connection/dataset fields as data
	// Example: mysql://{{.user}}:{{.password}}@{{.host}}:{{.port}}/{{.database}}
	// +required
	Content ContentTemplate `json:"content"`
}

type ContentTemplate string

func (ct ContentTemplate) Render(data interface{}) (string, error) {
	tmpl, err := template.New("content").
		Option("missingkey=error").
		Parse(string(ct))
	if err != nil {
		return "", fmt.Errorf("error parsing template: %w", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("error rendering template: %w", err)
	}

	return buf.String(), nil
}

type InjectableValueType string

func (iv *InjectableValue) GetType() InjectableValueType {
	switch {
	case iv.EnvName != "":
		return InjectableValueTypeEnv
	case iv.MountPath != "":
		return InjectableValueTypeFile
	default:
		return ""
	}
}

func (wfs *WorkflowSpec) GetInjectableValueByName(name string) (*InjectableValue, error) {
	for i, iv := range wfs.InjectableValues {
		if iv.Name == name {
			return &wfs.InjectableValues[i], nil
		}
	}
	return nil, fmt.Errorf("no InjectableValue found with name %s", name)
}

func init() {
	SchemeBuilder.Register(&Workflow{}, &WorkflowList{})
}
