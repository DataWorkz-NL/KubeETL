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
	Credentials Credentials `json:"credentials"`
}

type Credentials map[string]Value

// +kubebuilder:object:root=true

// ConnectionType defines the structure, validation and behavior of a connection
type ConnectionType struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	//+required
	Spec ConnectionTypeSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// ConnectionTypeList contains a list of ConnectionTypes
type ConnectionTypeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ConnectionType `json:"items"`
}

// ConnectionTypeSpec defines the desired state of ConnectionType
type ConnectionTypeSpec struct {
	//+required
	Name string `json:"name"`

	// CredentialFields used in this ConnectionTypeSpec. Used to validate input.
	//+optional
	Fields []CredentialFieldSpec `json:"fields,omitempty"`

	// Allow extra fields to be submitted that do not match any CredentialField
	//+optional
	AllowExtraFields bool `json:"allowExtraFields,omitempty"`
}

type CredentialFieldSpec struct {
	// Name for this CredentialField. Used as keys in the Credentials-map
	//+required
	Name string `json:"name"`

	// EnvKey is what the environment variable for this field will be called
	//+required
	EnvKey string `json:"envName"`

	// Whether or not this field must be filled
	//+required
	Required bool `json:"required"`

	// Optional methods of validating the field's value
	//+optional
	Validation *Validation `json:"validation,omitempty"`
}

// Contains optional properties used in validating CredentialFields
type Validation struct {
	// At least one must be selected

	//+optional
	MinLength *int32 `json:"minLength,omitempty"`

	//+optional
	MaxLength *int32 `json:"maxLength,omitempty"`

	// A regex pattern, must conform to RE2 syntax
	//+optional
	Regex *ValidationRegex `json:"regex,omitempty"`
}

// ValidationRegex contains a regex pattern conforming to RE2 syntax
type ValidationRegex string

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
	SchemeBuilder.Register(&Connection{}, &ConnectionList{}, &ConnectionType{}, &ConnectionTypeList{})
}
