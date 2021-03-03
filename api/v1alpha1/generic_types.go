package v1alpha1

import (
	apiv1 "k8s.io/api/core/v1"
)

type MetadataValidation struct {
	// List of fields specified for validation.
	//+optional
	Fields []CredentialFieldSpec `json:"fields,omitempty"`

	// Allow extra fields to be submitted that do not match any CredentialField
	//+optional
	AllowExtraFields bool `json:"allowExtraFields,omitempty"`
}

type Credentials map[string]Value

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

	// Whether or not this field is sensitive.
	// If a field is sensitive, the only valid ValueSource
	// is a SecretKeyRef. Plain text values and ConfigMapKeyRefs
	// are not allowed.
	//+optional
	Sensitive bool `json:"sensitive"`

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
