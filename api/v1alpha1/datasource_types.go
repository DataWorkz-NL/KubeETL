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

// DataSourceSpec defines the desired state of DataSource
type DataSourceSpec struct {
	// Connection defines the connection to use to retrieve this datasource
	// +optional
	Connection ConnectionFrom `json:"connection,omitempty"`

	// Type defines the type of the DataSource (e.g. MySQL table)
	Type string `json:"type"`

	// StorageType defines whether the DataSource is persisted or ephemeral
	StorageType StorageType `json:"storageType"`

	// Metadata contains any additional information that would be required
	// to fetch the DataSource from the connection, such as a file name
	// or a table name.
	// +optional
	Metadata Credentials `json:"metadata,omitemepty"`

	// Schema contains a reference to a connection where the schema of this
	// DataSource can be fetched from.
	// +optional
	Schema ConnectionFrom `json:"schema,omitempty"`
}

type ConnectionFrom struct {
	ConnectionFrom *ConnectionRef `json:"connectionFrom,omitempty"`
}

// StorageType defines whether the Storage is persisted
// in a remote storage or whether it is ephemeral.
// +kubebuilder:validation:Enum=Persistent;Ephemeral
type StorageType string

const (
	// PersistentType defines DataSources which are persisted in a data store.
	PersistentType StorageType = "Persistent"

	// EphemeralType defines DataSources which should be recreated from a job.
	EphemeralType StorageType = "Ephemeral"
)

// DataSourceStatus defines the observed state of DataSource
type DataSourceStatus struct{}

// +kubebuilder:object:root=true

// DataSource is the Schema for the datasources API
type DataSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DataSourceSpec   `json:"spec,omitempty"`
	Status DataSourceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DataSourceList contains a list of DataSource
type DataSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DataSource `json:"items"`
}

// +kubebuilder:object:root=true

// DataSourceType defines the structure of a DataSource
type DataSourceType struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	//+required
	Spec DataSourceTypeSpec `json:"spec,omitempty"`
}

type DataSourceTypeSpec struct {
	// MetadataFields defines the structure of the metadata for the DataSource
	MetadataFields MetadataValidation `json:"metadata,omitempty"`
}

// +kubebuilder:object:root=true

// DataSourceTypeList contains a list of DataSourceTypes
type DataSourceTypeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DataSourceType `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DataSource{}, &DataSourceList{}, &DataSourceType{}, &DataSourceTypeList{})
}
