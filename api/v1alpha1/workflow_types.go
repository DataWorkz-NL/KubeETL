package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	wfv1 "github.com/argoproj/argo/v2/pkg/apis/workflow/v1alpha1"
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
	WorkflowSpec      `json:"spec"`
}

type WorkflowSpec struct {
	ArgoWorkflowSpec wfv1.WorkflowSpec     `json:",inline"`
	Connections      []ConnectionInjection `json:"connections,"`
	Templates        []Template            `json:"templates"`
}

// Embedding type for Argo Template with Injections added
type Template struct {
	wfv1.Template `json:",inline"`
	Injection     []InjectDefinitionRef `json:"inject,omitempty"`
}

// Contains the required keys to select an InjectDefinition
type InjectDefinitionRef struct {
	// The connection name (or alias if defined)
	//+required
	ConnectionKey string `json:"connection"`

	//+optional
	Name string `json:"name,omitempty"`
}

// Contains a reference to a `Connection` and the desired injections
// If no InjectDefinitions are specified, credentials will be injected as environment variables
//+kubebuilder:object:generate:=true
type ConnectionInjection struct {
	// Name of the `Connection` that is being injected here
	//+required
	ConnectionName string `json:"name,"`

	// Optional alias for consuming templates
	//+optional
	Alias string `json:"alias,omitempty"`

	// If true, all the InjectDefinitions will be applied to all ContainerTemplates in this workflow.
	// If false, consuming templates must specifically request this ConnectionInjection
	//+optional
	Global bool `json:"global,omitempty"`

	// A list of injections
	//+optional
	InjectDefinitions []InjectDefinition `json:"injectable,"`
}

// +kubebuilder:object:generate:=true

// InjectDefinition specifies how a connection will be injected
type InjectDefinition struct {
	// Identifier for this injection in case of selective injections
	//+optional
	Name string `json:"name,omitempty"`

	// Name of the injected environment variable
	//+optional
	Key string `json:"key,omitempty"`

	// Path where value will be mounted as a file
	//+optional
	Path string `json:"path,omitempty"`

	// Go template that will be rendered using the connection fields as data
	// Example: mysql://{{.user}}:{{.password}}@{{.host}}:{{.port}}/{{.database}}
	//+required
	Value string `json:"value,"`
}
