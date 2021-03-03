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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DataSetSpec defines the desired state of DataSet
type DataSetSpec struct {
	// Connection defines the connection to use to retrieve this dataset
	// +optional
	Connection ConnectionFrom `json:"connection,omitempty"`

	// Type defines the type of the DataSet (e.g. MySQL table)
	Type string `json:"type"`

	// StorageType defines whether the DataSet is persisted or ephemeral
	StorageType StorageType `json:"storageType"`

	// Metadata contains any additional information that would be required
	// to fetch the DataSet from the connection, such as a file name
	// or a table name.
	// +optional
	Metadata Credentials `json:"metadata,omitemepty"`
}

type ConnectionFrom struct {
	ConnectionFrom *ConnectionRef `json:"connectionFrom,omitempty"`
}

// StorageType defines whether the Storage is persisted
// in a remote storage or whether it is ephemeral.
// +kubebuilder:validation:Enum=Persistent;Ephemeral
type StorageType string

const (
	// PersistentType defines DataSets which are persisted in a data store.
	PersistentType StorageType = "Persistent"

	// EphemeralType defines DataSets which should be recreated from a job.
	EphemeralType StorageType = "Ephemeral"
)

// DataSetStatus defines the observed state of DataSet
type DataSetStatus struct{}

// +kubebuilder:object:root=true

// DataSet is the Schema for the datasets API
type DataSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DataSetSpec   `json:"spec,omitempty"`
	Status DataSetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DataSetList contains a list of DataSet
type DataSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DataSet `json:"items"`
}

// +kubebuilder:object:root=true

// DataSetType defines the structure of a DataSet
type DataSetType struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	//+required
	Spec DataSetTypeSpec `json:"spec,omitempty"`
}

type DataSetTypeSpec struct {
	// MetadataFields defines the structure of the metadata for the DataSet
	MetadataFields MetadataValidation `json:"metadata,omitempty"`
}

// +kubebuilder:object:root=true

// DataSetTypeList contains a list of DataSetTypes
type DataSetTypeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DataSetType `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DataSet{}, &DataSetList{}, &DataSetType{}, &DataSetTypeList{})
}
