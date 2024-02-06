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

package pipelines

import (
	"context"
	civ1alpha1 "github.com/w6d-io/ciops/api/v1alpha1"

	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"github.com/w6d-io/ciops/internal/k8s/tasks"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Wks struct {
	WB []pipelinev1.WorkspacePipelineTaskBinding `json:"workspacePipelineTaskBinding"`
	W  []pipelinev1.WorkspaceBinding             `json:"workspaces"`
}
type Pipelines struct {
	Resources []func(context.Context, client.Client, *civ1alpha1.PipelineSource) error
}

type pipeline struct {
	tasks       tasks.Tasks
	name        string
	namespace   string
	labels      map[string]string
	annotations map[string]string
	triggers    map[string]string
}

var (
	Workspace Wks
)
