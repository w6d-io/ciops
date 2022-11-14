/*
Copyright 2020 WILDCARD SA.

Licensed under the WILDCARD SA License, Version 1.0 (the "License");
WILDCARD SA is register in french corporation.
You may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.w6d.io/licenses/LICENSE-1.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is prohibited.
Created on 04/11/2022
*/

package pipelines

import (
	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/w6d-io/ciops/internal/tasks"
)

type Wks struct {
	WB []pipelinev1beta1.WorkspacePipelineTaskBinding `json:"workspacePipelineTaskBinding"`
	W  []pipelinev1beta1.WorkspaceBinding             `json:"workspaces"`
}

type pipeline struct {
	tasks       []tasks.Task
	name        string
	namespace   string
	labels      map[string]string
	annotations map[string]string
	triggers    map[string]string
}

var Workspace Wks
