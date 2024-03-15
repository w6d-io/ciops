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

package tasks

import (
	"context"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"github.com/w6d-io/ciops/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	prefixBuild = "build"

	// TODO add to config
	prefixNamespace = "p6e-cx"
)

type Wks struct {
	WD []pipelinev1.WorkspaceDeclaration `json:"workspacePipelineTaskBinding"`
}

// Task ...
type Task struct {
	task        pipelinev1.Task
	name        string
	namespace   string
	runAfter    []string
	labels      map[string]string
	annotations map[string]string
	when        *pipelinev1.WhenExpression
	workspaces  []pipelinev1.WorkspaceDeclaration
}

// Tasks ...
type Tasks struct {
	Resources []func(context.Context, client.Client, *v1alpha1.PipelineSource) error
	Tasks     []Task
}

var (
	Workspace Wks
)
