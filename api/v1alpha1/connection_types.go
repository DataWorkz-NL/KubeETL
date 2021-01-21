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

package v1alpha1

import (
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConnectionSpec defines the desired state of Connection
type ConnectionSpec struct {
	// Type contains the type of protocol that should be utilised in this connection.
	// This can be used for a dynamic determination of what source is being connected to.
	//+optional
	Type string `json:"type,omitempty"`

	// All required information to achieve the connection is stored in the credentials
	//+required
	Credentials *Credentials `json:"credentials"`
}

// Credentials store all relevant information to achieve a connection
type Credentials struct {
	// The URL to connect to
	//+optional
	URL *Value `json:"url,omitempty"`
	// Holds the username required to achieve a connection
	//+optional
	Username *Value `json:"username,omitempty"`
	// Holds the password required to achieve a connection
	//+optional
	Password *Value `json:"password,omitempty"`
}

// Value contains either a direct value or a value from a source
type Value struct {
	// +optional
	Value string `json:"value,omitempty"`
	// Source for the value. Cannot be used if Value is already defined
	// +optional
	ValueFrom *ValueSource `json:"valueFrom,omitempty"`
}

// ValueSource holds a reference to either a ConfigMap or a Secret
type ValueSource struct {
	// Select at least one

	// Selects a key of a ConfigMap.
	// +optional
	ConfigMapKeyRef *apiv1.ConfigMapKeySelector `json:"configMapKeyRef,omitempty" protobuf:"bytes,3,opt,name=configMapKeyRef"`
	// Selects a key of a secret in the pod's namespace
	// +optional
	SecretKeyRef *apiv1.SecretKeySelector `json:"secretKeyRef,omitempty" protobuf:"bytes,4,opt,name=secretKeyRef"`
}

// ConnectionStatus defines the observed state of Connection
type ConnectionStatus struct{}

// +kubebuilder:object:root=true

// Connection is the Schema for the connections API
type Connection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConnectionSpec   `json:"spec,omitempty"`
	Status ConnectionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ConnectionList contains a list of Connection
type ConnectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Connection `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Connection{}, &ConnectionList{})
}
