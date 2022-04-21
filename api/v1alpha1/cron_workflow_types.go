package v1alpha1

import (
	wfv1 "github.com/argoproj/argo/v2/pkg/apis/workflow/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CronWorkflow is the definition of a scheduled workflow resource
// +kubebuilder:object:root=true
type CronWorkflow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              CronWorkflowSpec `json:"spec"`
}

// CronWorkflowList is list of CronWorkflow resources
// +kubebuilder:object:root=true
type CronWorkflowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []CronWorkflow `json:"items"`
}

// CronWorkflowSpec is the specification of a CronWorkflow
type CronWorkflowSpec struct {
	// WorkflowSpec is the spec of the workflow to be run
	// +required
	WorkflowSpec WorkflowSpec `json:"workflowSpec"`

	// Schedule is a schedule to run the Workflow in Cron format
	Schedule string `json:"schedule" protobuf:"bytes,2,opt,name=schedule"`
	// ConcurrencyPolicy is the K8s-style concurrency policy that will be used
	ConcurrencyPolicy wfv1.ConcurrencyPolicy `json:"concurrencyPolicy,omitempty" protobuf:"bytes,3,opt,name=concurrencyPolicy,casttype=ConcurrencyPolicy"`
	// Suspend is a flag that will stop new CronWorkflows from running if set to true
	Suspend bool `json:"suspend,omitempty" protobuf:"varint,4,opt,name=suspend"`
	// StartingDeadlineSeconds is the K8s-style deadline that will limit the time a CronWorkflow will be run after its
	// original scheduled time if it is missed.
	StartingDeadlineSeconds *int64 `json:"startingDeadlineSeconds,omitempty" protobuf:"varint,5,opt,name=startingDeadlineSeconds"`
	// SuccessfulJobsHistoryLimit is the number of successful jobs to be kept at a time
	SuccessfulJobsHistoryLimit *int32 `json:"successfulJobsHistoryLimit,omitempty" protobuf:"varint,6,opt,name=successfulJobsHistoryLimit"`
	// FailedJobsHistoryLimit is the number of failed jobs to be kept at a time
	FailedJobsHistoryLimit *int32 `json:"failedJobsHistoryLimit,omitempty" protobuf:"varint,7,opt,name=failedJobsHistoryLimit"`
	// Timezone is the timezone against which the cron schedule will be calculated, e.g. "Asia/Tokyo". Default is machine's local time.
	Timezone string `json:"timezone,omitempty" protobuf:"bytes,8,opt,name=timezone"`
	// WorkflowMetadata contains some metadata of the workflow to be run
	WorkflowMetadata *metav1.ObjectMeta `json:"workflowMetadata,omitempty" protobuf:"bytes,9,opt,name=workflowMeta"`
}

func init() {
	SchemeBuilder.Register(&CronWorkflow{}, &CronWorkflowList{})
}
