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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Status int

const (

	// Creating means that tekton resource creation is in progress
	Creating State = "Creating"

	// Queued means that the PipelineRun not applied yet due to limitation
	Queued State = "Queued"

	// Running means at least on Step of the Task is running
	Running State = "Running"

	// Failed means at least on Step of the Task is failed
	Failed State = "Failed"

	// Succeeded means that all Task is success
	Succeeded State = "Succeeded"

	// Cancelled means that a TaskRun or PipelineRun has been cancelled
	Cancelled State = "Cancelled"

	// Errored means that at least one tekton resource couldn't be created
	Errored State = "Errored"
)

var (
	// GroupFactKind is group kind for Fact
	GroupFactKind = schema.GroupKind{Group: "ci.w6d.io", Kind: "Fact"}
)

const (
	AnnotationSchedule = "ciops.ci.w6d.io/scheduled-at"
	FactFinalizer      = "fact.ci.w6d.io/finalizer"
)

// TriggerSpec defines the trigger
type TriggerSpec struct {
	ID   string `json:"id,omitempty"`
	Ref  string `json:"ref,omitempty"`
	Type string `json:"type,omitempty"`
}

// FactSpec defines the desired state of Fact
type FactSpec struct {

	// EventID id of the fact
	EventID *int64 `json:"eventId,omitempty"`

	// PipelineRef is the id pipeline resource name
	PipelineRef string `json:"pipelineRef"`

	// ProjectName name of project
	ProjectName string `json:"projectName,omitempty"`

	// ProjectURL url of the project
	ProjectURL string `json:"projectUrl,omitempty"`

	// ProjectID identifier of the project
	ProjectID int64 `json:"projectID,omitempty"`

	// Ref is project reference for this fact
	Ref string `json:"ref,omitempty"`

	// Commit project for this fact
	Commit string `json:"commit,omitempty"`

	// BeforeSha is the previous commit sha for this fact
	BeforeSha string `json:"beforeSha,omitempty"`

	// CommitMessage is the message of this commit fact
	CommitMessage string `json:"commitMessage,omitempty"`

	// UserId is the user id from the repository
	UserId string `json:"userId,omitempty"`

	// Added is the list of files that have been added in this commit
	Added []string `json:"added,omitempty"`

	// Removed is the list of files that have been removed in this commit
	Removed []string `json:"removed,omitempty"`

	// Modified is the list of files that have been modified in this commit
	Modified []string `json:"modified,omitempty"`

	// ProviderId is the id of the provider that send this fact
	ProviderId string `json:"providerId,omitempty"`

	// Trigger
	Trigger *TriggerSpec `json:"trigger,omitempty"`

	// PipelineSource is the pipeline payload
	PipelineSource *corev1.LocalObjectReference `json:"pipeline,omitempty"`

	// TODO to delete token for cloning
	// Deprecated
	Token string `json:"token,omitempty"`
}

// State type
type State string

// FactStatus defines the observed state of Fact
type FactStatus struct {
	// PipelineRunName contains the pipeline run name created by play
	// +optional
	PipelineRunName string `json:"pipelineRunName,omitempty"`

	// State contains the current state of this Play resource.
	// States Running, Failed, Succeeded, Errored
	// +optional
	State State `json:"state,omitempty"`

	// Message contains the pipeline message
	// +optional
	Message string `json:"message,omitempty"`

	// Conditions represents the latest available observations of play
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"
//+kubebuilder:printcolumn:name="PipelineRun",type="string",priority=1,JSONPath=".status.pipelineRunName"
//+kubebuilder:printcolumn:name="Message",type="string",priority=1,JSONPath=".status.message"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC."

// Fact is the Schema for the facts API
type Fact struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FactSpec   `json:"spec,omitempty"`
	Status FactStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FactList contains a list of Fact
type FactList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Fact `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Fact{}, &FactList{})
}
