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
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type PipelineSpec struct {
	// PipelineRef contains the pipeline to applies the budget. If empty all pipeline will be affected
	PipelineRef *string `json:"pipelineRef,omitempty"`

	// Concurrent contains the number of pipeline running in concurrency
	Concurrent *int64 `json:"concurrent,omitempty"`
}

// FactBudgetSpec defines the desired state of FactBudget
type FactBudgetSpec struct {
	// Pipeline budget for pipeline
	Pipeline PipelineSpec `json:"pipeline"`
}

// FactBudgetStatus defines the observed state of FactBudget
type FactBudgetStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Concurrency",type="string",priority=1,JSONPath=".spec.pipeline.concurrent"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC."

// FactBudget is the Schema for the factbudgets API
type FactBudget struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FactBudgetSpec   `json:"spec,omitempty"`
	Status FactBudgetStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FactBudgetList contains a list of FactBudget
type FactBudgetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FactBudget `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FactBudget{}, &FactBudgetList{})
}
