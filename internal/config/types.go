/*
Copyright 2020 WILDCARD SA.

Licensed under the WILDCARD SA License, Version 1.0 (the "License");
WILDCARD SA is register in french corporation.
You may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.w6d.io/licenses/LICENSE-1.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is prohibited.
Created on 21/09/2022
*/

package config

import v1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"

const (
	ViperKeyMetricsListen          = "listen.metrics"
	ViperKeyWebhookListen          = "listen.webhook"
	ViperKeyProbeListen            = "listen.probe"
	ViperKeyLeaderElect            = "leaderElection.leaderElect"
	ViperKeyLeaderName             = "leaderElection.resourceName"
	ViperKeyLeaderNamespace        = "leaderElection.namespace"
	ViperKeyNamespace              = "namespace"
	ViperKeyHooks                  = "hooks"
	ViperKeyPipelinerunPrefix      = "pipelinerun.prefix"
	ViperKeyPipelinerunPodTemplate = "pipelinerun.podTemplate"
	ViperKeyPipelinerunWorkspaces  = "pipelinerun.workspaces"
)

type Config struct {
	Listen struct {
		Metrics string `json:"metrics,omitempty" mapstructure:"metrics"`
		Probe   string `json:"probe,omitempty" mapstructure:"probe"`
	} `json:"listen,omitempty" mapstructure:"listen"`
	Webhook struct {
		Port int `json:"port,omitempty" mapstructure:"port"`
	} `json:"webhook,omitempty" mapstructure:"webhook"`
	LeaderElection struct {
		LeaderElect  bool   `json:"leaderElect,omitempty" mapstructure:"leaderElect"`
		ResourceName string `json:"resourceName,omitempty" mapstructure:"resourceName"`
		Namespace    string `json:"namespace,omitempty" mapstructure:"namespace"`
	} `json:"leaderElection,omitempty" mapstructure:"leaderElection"`
	Pipelinerun struct {
		Prefix                       string                `json:"prefix,omitempty" mapstructure:"prefix"`
		Workspaces                   []v1.WorkspaceBinding `json:"workspaces,omitempty" mapstructure:"workspaces"`
		WorkspacePipelineTaskBinding []struct {
			Name      string `json:"name,omitempty" mapstructure:"name"`
			SubPath   string `json:"subPath,omitempty" mapstructure:"subPath"`
			Workspace string `json:"workspace,omitempty" mapstructure:"workspace"`
		} `json:"workspacePipelineTaskBinding,omitempty" mapstructure:"workspacePipelineTaskBinding"`
		PodTemplate struct {
			NodeSelector struct {
				Role string `json:"role,omitempty" mapstructure:"role"`
			} `json:"nodeSelector,omitempty" mapstructure:"nodeSelector"`
			Tolerations []struct {
				Effect   string `json:"effect,omitempty" mapstructure:"effect"`
				Key      string `json:"key,omitempty" mapstructure:"key"`
				Operator string `json:"operator,omitempty" mapstructure:"operator"`
				Value    string `json:"value,omitempty" mapstructure:"value"`
			} `json:"tolerations,omitempty" mapstructure:"tolerations"`
		} `json:"podTemplate,omitempty" mapstructure:"podTemplate"`
	} `json:"pipelinerun,omitempty" mapstructure:"pipelinerun"`
	Namespace string `json:"namespace,omitempty" mapstructure:"namespace"`
}
