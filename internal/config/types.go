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

import (
	"os"

	v1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"

	"github.com/w6d-io/jsonschema"
)

var (
	// Version of application
	Version string

	// CfgFile contain the path of the config file
	CfgFile string

	// OsExit is hack for unit-test
	OsExit = os.Exit

	// SkipValidation toggling the config validation
	SkipValidation bool
)

const (
	ViperKeyMetricsListen       = "listen.metrics"
	ViperKeyProbeListen         = "listen.probe"
	ViperKeyWebhookPort         = "webhook.port"
	ViperKeyWebhookHost         = "webhook.host"
	ViperKeyLeaderElect         = "election.enabled"
	ViperKeyLeaderName          = "election.resourceName"
	ViperKeyLeaderNamespace     = "election.namespace"
	ViperKeyHooks               = "hooks"
	ViperKeyPipelinerun         = "pipelinerun"
	ViperKeyPipelinerunPrefix   = "pipelinerun.prefix"
	ViperKeyExtraDefaultActions = "extra.defaultActions"

	ViperKeyWorkspaces = "workspaces"
)

const (
	Schema jsonschema.SchemaType = iota
	WebhookSchema
)

type Config struct {
	Listen struct {
		Probe   string `json:"probe,omitempty" mapstructure:"probe"`
		Metrics string `json:"metrics,omitempty" mapstructure:"metrics"`
	} `json:"listen,omitempty" mapstructure:"listen"`
	Election struct {
		Enabled      bool   `json:"enabled,omitempty" mapstructure:"enabled"`
		ResourceName string `json:"resourceName,omitempty" mapstructure:"resourceName"`
		Namespace    string `json:"namespace,omitempty" mapstructure:"namespace"`
	} `json:"election,omitempty" mapstructure:"election"`
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

type Webhook struct {
	Listen struct {
		Probe   string `mapstructure:"probe"   json:"probe,omitempty"`
		Metrics string `mapstructure:"metrics" json:"metrics,omitempty"`
	} `mapstructure:"listen" json:"listen,omitempty"`
	Webhook struct {
		Host string `mapstructure:"host" json:"host,omitempty"`
		Port int    `mapstructure:"port" json:"port,omitempty"`
	} `mapstructure:"webhook" json:"webhook,omitempty"`
}
