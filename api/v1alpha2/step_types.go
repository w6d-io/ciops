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

import pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"

//+kubebuilder:object:generate=true

// Step marshalling structure
type Step struct {
	pipelinev1.Step `mapstructure:",inline"     json:",inline"`

	DisplayName string `json:"displayName"`
	// ID of this step
	// +required
	ID string `json:"id"`
	// Category of this step
	// +optional
	Category string `json:"category,omitempty"`
	// Description of this step
	// +optional
	Description string `json:"description,omitempty"`
	// Icon of this action
	// +required
	Icon string `json:"icon"`
	// Fields is the list
	// +optional
	// +listType=atomic
	Fields []string `json:"fields,omitempty"`
	// Targets is path where the front get information to associate with
	// +required
	Targets map[string]string `json:"targets,omitempty"`
	// Params declares a list of input parameters that must be supplied when
	// this Pipeline is run.
	// +optional
	Params pipelinev1.ParamSpecs `json:"params,omitempty"`
	// Workspaces are the volumes that this Task requires.
	// +optional
	// +listType=atomic
	Workspaces []pipelinev1.WorkspaceDeclaration `json:"workspaces,omitempty"`
}
