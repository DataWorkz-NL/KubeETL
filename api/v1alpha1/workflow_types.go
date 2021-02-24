package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// corev1 "k8s.io/api/core/v1"
	// "k8s.io/apimachinery/pkg/runtime/schema"
	// wfv1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	// wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
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
	Connections      []ConnectionInjection `json:"connections,"`
	ArgoWorkflowSpec wfv1.WorkflowSpec     `json:",inline"`
}

// +kubebuilder:object:generate:=true
type ConnectionInjection struct {
	ConnectionName string          `json:"name,"`
	InjectionSpecs []InjectionSpec `json:"inject,"`
}

// +kubebuilder:object:generate:=true
type InjectionSpec struct {
	Key   string `json:"key,omitempty"`
	Path  string `json:"path,omitempty"`
	Value string `json:"value,"`
}

