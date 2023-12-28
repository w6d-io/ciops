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
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/w6d-io/ciops/internal/embedx"
	"github.com/w6d-io/ciops/internal/pipelineruns"
	"github.com/w6d-io/jsonschema"
	"github.com/w6d-io/x/cmdx"
	"github.com/w6d-io/x/logx"
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

func Init() {
	base := filepath.Base(CfgFile)
	log := logx.WithName(nil, "Config.Init")
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
			logx.WithName(context.TODO(), "Config").Error(err, "failed to read config")
			OsExit(1)
			return
		}
		log.Error(err, "config not found")
		return
	}
	var c Config
	cmdx.Should(viper.Unmarshal(&c), "unmarshal config failed")
	if !SkipValidation {
		log.Info("run config validation")
		cmdx.Must(jsonschema.AddSchema(jsonschema.Config, embedx.ConfigSchema), "add config schema failed")
		cmdx.Must(jsonschema.Config.Validate(&c), "config validation failed")
	}
	cmdx.Must(hookSubscription(), "hook subscription failed")
	cmdx.Should(viper.UnmarshalKey(ViperKeyPipelinerun, &pipelineruns.LC), "failed to record pod template")
	//cmdx.Should(viper.UnmarshalKey(ViperKeyPipelinerun, &pipelineruns.Workspace), "failed to record workspace")
	//extraConfigJson(ViperKeyPipelinerunPodTemplate, &pipelineruns.PodTemplate)
	//extraConfigJson(ViperKeyPipelinerunWorkspaces, &pipelineruns.Workspace)
	//pipelineruns.PipelinerunPrefix = viper.GetString(ViperKeyPipelinerunPrefix)
}

func extraConfigJson(key string, rawVar interface{}) {
	extraFile := viper.GetString(key)
	logx.WithName(nil, "extraConfig").V(2).Info("parameters", "key", key, "extraFile", extraFile)
	if extraFile != "" {
		base := filepath.Base(extraFile)
		ext := filepath.Ext(extraFile)
		cfg := viper.New()
		cfg.SetConfigName(FileNameWithoutExtension(base))
		cfg.SetConfigType(strings.TrimLeft(ext, "."))
		cfg.AddConfigPath(filepath.Dir(extraFile))
		if err := cfg.ReadInConfig(); err != nil {
			logx.WithName(context.TODO(), "loading").Info("config", "key", key, "extraFile", extraFile)
			cmdx.Should(err, "fail to read config")
		} else {
			logx.WithName(context.TODO(), "loading").Info("config", "key", key, "extraFile", extraFile)
			cmdx.Should(convert(extraFile, rawVar), "load config failed")
		}
		cfg.WatchConfig()
		cfg.OnConfigChange(func(in fsnotify.Event) {
			logx.WithName(context.TODO(), "reloading").Info("config", "key", key, "extraFile", extraFile)
			cmdx.Should(convert(extraFile, rawVar), "load config failed")
		})
	}
}
func convert(extraFile string, target interface{}) error {
	yamlFile, err := os.ReadFile(extraFile)

	if err != nil {
		return err
	} else {
		var tmp map[string]interface{}
		err = yaml.Unmarshal(yamlFile, &tmp)
		if err != nil {
			return err
		}
		d, err := json.Marshal(&tmp)
		if err != nil {
			return err
		}
		err = json.Unmarshal(d, &target)
		if err != nil {
			return err
		}
	}
	return nil
}
