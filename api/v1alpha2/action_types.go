/*
Copyright 2020 WILDCARD SA.

Licensed under the WILDCARD SA License, Version 1.0 (the "License");
WILDCARD SA is register in french corporation.
You may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.w6d.io/licenses/LICENSE-1.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is prohibited.
Created on 02/02/2024
*/

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Actions []Action

//+kubebuilder:object:root=true

// Action is the Schema for action API
type Action struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Steps are the steps of the build; each step is run sequentially with the
	// source mounted into /workspace.
	Steps []Step `json:"steps,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:skip

// ActionList contains a list of Action
type ActionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items Actions `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Action{}, &ActionList{})
}
