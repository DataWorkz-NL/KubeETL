package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +kubebuilder:object:root=true

// WorkflowTemplateList contains a list of Workflows
type WorkflowTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkflowTemplate `json:"items"`
}

// +kubebuilder:object:root:=true
// +kubebuilder:subresource:status

// WorkflowTemplate is the schema for the workflows API
type WorkflowTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              WorkflowTemplateSpec `json:"spec"`
}

// WorkflowTemplateSpec is a spec of WorkflowTemplate.
type WorkflowTemplateSpec struct {
	WorkflowSpec `json:",inline" protobuf:"bytes,1,opt,name=workflowSpec"`
	// WorkflowMetadata contains some metadata of the workflow to be refer
	WorkflowMetadata *metav1.ObjectMeta `json:"workflowMetadata,omitempty" protobuf:"bytes,2,opt,name=workflowMeta"`
}
