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
    "gitlab.w6d.io/w6d/ciops/internal/embedx"
    "os"
    "path/filepath"
    "strings"

    "github.com/spf13/viper"

    "github.com/w6d-io/jsonschema"
    "github.com/w6d-io/x/cmdx"
    "github.com/w6d-io/x/logx"
)

var (
    // Version of application
    Version string

    // Revision is the commit of this version
    Revision string

    // Built is the timestamp od this version
    Built string

    // CfgFile contain the path of the config file
    CfgFile string

    // OsExit is hack for unit-test
    OsExit = os.Exit

    // SkipValidation toggling the config validation
    SkipValidation bool
)

func setDefault() {
    viper.SetDefault(ViperKeyMetricsListen, ":8080")
    viper.SetDefault(ViperKeyProbListen, ":8081")
    viper.SetDefault(ViperKeyNamespace, false)
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
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            logx.WithName(context.TODO(), "Config").Error(err, "failed to read config")
            OsExit(1)
            return
        }
        log.Error(err, "config not found")
        return
    }
    var c map[string]interface{}
    cmdx.Should(viper.Unmarshal(&c), "unmarshal config failed")
    if !SkipValidation {
        log.Info("run config validation")
        cmdx.Must(jsonschema.AddSchema(jsonschema.Config, embedx.ConfigSchema), "add config schema failed")
        cmdx.Must(jsonschema.Config.Validate(&c), "config validation failed")
    }
}

func GetPipelinerunPrefix() string {
    return viper.GetString(ViperKeyPipelinerunPrefix)
}
