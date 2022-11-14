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

package tasks

import (
	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

// Task ...
type Task struct {
	task        pipelinev1beta1.Task
	name        string
	namespace   string
	runAfter    []string
	labels      map[string]string
	annotations map[string]string
	when        *pipelinev1beta1.WhenExpression
	workspaces  []pipelinev1beta1.WorkspaceDeclaration
}

// Tasks ...
type Tasks struct {
	Tasks []Task
}

type Wks struct {
	WD []pipelinev1beta1.WorkspaceDeclaration `json:"workspacePipelineTaskBinding"`
}

var Workspace Wks

const (
	refGitSource        = "git-source"
	refArtefactDownload = "artefact-download"
	refArtefactUpload   = "artefact-upload"
)
