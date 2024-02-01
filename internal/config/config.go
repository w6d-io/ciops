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
	"context"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/w6d-io/ciops/internal/embedx"
	"github.com/w6d-io/ciops/internal/k8s/pipelineruns"
	"github.com/w6d-io/jsonschema"
	"github.com/w6d-io/x/cmdx"
	"github.com/w6d-io/x/logx"
)

func setDefault() {
	viper.SetDefault(ViperKeyMetricsListen, ":8080")
	viper.SetDefault(ViperKeyProbeListen, ":8081")
	viper.SetDefault(ViperKeyLeaderName, "ciops.ci.w6d.io")
	viper.SetDefault(ViperKeyLeaderElect, false)
	viper.SetDefault(ViperKeyWebhookPort, 9443)
	viper.SetDefault(ViperKeyWebhookHost, "")
	viper.SetDefault(ViperKeyPipelinerunPrefix, "pipelinerun")
}

// FileNameWithoutExtension returns the
func FileNameWithoutExtension(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

func Base() {
	log := logx.WithName(context.Background(), "config.Base")
	base := filepath.Base(CfgFile)

	ext := filepath.Ext(CfgFile)
	log.V(2).Info("viper",
		"path", CfgFile,
		"ext", filepath.Ext(CfgFile),
		"type", strings.TrimLeft(ext, "."),
		"configName", FileNameWithoutExtension(base),
		"base", base,
		"dir", filepath.Dir(CfgFile),
	)
	setDefault()
	viper.SetConfigName(FileNameWithoutExtension(base))
	viper.SetConfigType(strings.TrimLeft(ext, "."))
	viper.AddConfigPath(filepath.Dir(CfgFile))
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.ciops")

	viper.SetEnvPrefix("ciops")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			logx.WithName(context.Background(), "config").Error(err, "failed to read config")
			OsExit(1)
			return
		}
		log.Error(err, "config not found")
		return
	}
}

func InitWebhook(_ *cobra.Command, _ []string) {
	log := logx.WithName(nil, "webhook.config.init")
	var w Webhook
	cmdx.Must(viper.Unmarshal(&w), "unmarshal webhook config failed")
	if !SkipValidation {
		logx.WithName(context.Background(), "webhook.config.validation").Info("run webhook validation")
		cmdx.Must(jsonschema.AddSchema(WebhookSchema, embedx.WebhookSchema), "add webhook config schema failed")
		cmdx.Must(WebhookSchema.Validate(&w), "webhook config validation failed")
	}
	log.Info("config loaded", "file", viper.ConfigFileUsed())
}

func Init(_ *cobra.Command, _ []string) {
	log := logx.WithName(context.Background(), "config.Init")
	var c Config
	cmdx.Must(viper.Unmarshal(&c), "unmarshal config failed")
	if !SkipValidation {
		log.Info("run config validation")
		cmdx.Must(jsonschema.AddSchema(Schema, embedx.ConfigSchema), "add config schema failed")
		cmdx.Must(Schema.Validate(&c), "config validation failed")
	}
	cmdx.Must(hookSubscription(), "hook subscription failed")
	cmdx.Should(viper.UnmarshalKey(ViperKeyPipelinerun, &pipelineruns.LC), "failed to record pod template")
	cmdx.Should(viper.UnmarshalKey(ViperKeyWorkspaces, &pipelineruns.WB), "failed to record pod template")
}
