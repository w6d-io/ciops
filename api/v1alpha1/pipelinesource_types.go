/*
Copyright 2022 WILDCARD.

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

	apis "github.com/w6d-io/apis/pipeline/v1alpha1"
)

// PipelineSourceStatus defines the observed state of PipelineSource
type PipelineSourceStatus struct {

	// Tasks contains the list of task created
	// +optional
	Tasks []string `json:"tasks,omitempty"`

	// ConditionTasks contains the list of conditions task created
	// +optional
	ConditionTasks []string `json:"conditionTasks,omitempty"`

	// PipelineName contains the name of pipeline resource created
	// +optional
	PipelineName string `json:"pipelineName,omitempty"`

	// State contains the current state of this Play resource.
	// States Running, Failed, Succeeded, Errored
	// +optional
	State metav1.ConditionStatus `json:"state,omitempty"`

	// Message contains the pipeline message
	// +optional
	Message string `json:"message,omitempty"`

	// Conditions represents the latest available observations of PipelineSource
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName="ps"
//+kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"
//+kubebuilder:printcolumn:name="Pipeline",type="string",priority=1,JSONPath=".status.pipelineName"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC."

// PipelineSource is the Schema for the pipelinesources API
type PipelineSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   apis.Pipeline        `json:"spec,omitempty"`
	Status PipelineSourceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PipelineSourceList contains a list of PipelineSource
type PipelineSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PipelineSource `json:"items"`
}

// Option defines a field list option supported
type Option string

type Validation struct {
	ValueType string `json:"valueType,omitempty"`
	MaxLength int    `json:"maxLength,omitempty"`
}

type Documentation struct {
	Link string `json:"link,omitempty"`
}

type Field struct {
	ID                 string        `json:"id,omitempty"`
	Name               string        `json:"name,omitempty"`
	Description        string        `json:"description,omitempty"`
	ReadOnly           bool          `json:"readOnly,omitempty"`
	Visibility         bool          `json:"visibility,omitempty"`
	Duplicate          bool          `json:"duplicate,omitempty"`
	DefaultValue       string        `json:"defaultValue,omitempty"`
	Options            []Option      `json:"options,omitempty"`
	Value              string        `json:"value,omitempty"`
	RestrictedValues   []string      `json:"restrictedValues,omitempty"`
	AutocompleteValues []string      `json:"autocompleteValues,omitempty"`
	Validations        Validation    `json:"validations,omitempty"`
	Documentations     Documentation `json:"documentations,omitempty"`
}

func init() {
	SchemeBuilder.Register(&PipelineSource{}, &PipelineSourceList{})
}
