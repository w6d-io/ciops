/*
Copyright 2020 WILDCARD SA.

Licensed under the WILDCARD SA License, Version 1.0 (the "License");
WILDCARD SA is register in french corporation.
You may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.w6d.io/licenses/LICENSE-1.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is prohibited.
Created on 26/01/2024
*/

package ciops

import (
	"context"
	"github.com/w6d-io/ciops/cmd/ciops/webhook"
	"os"
	"runtime/debug"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/w6d-io/ciops/cmd/ciops/server"
	"github.com/w6d-io/ciops/internal/config"
	"github.com/w6d-io/x/cmdx"
	"github.com/w6d-io/x/logx"
	"github.com/w6d-io/x/pflagx"
)

var rootCmd = &cobra.Command{
	Use: "ciops",
}

var OsExit = os.Exit

func init() {
	cobra.OnInitialize(config.Base)
	pflagx.CallerSkip = 0
	if os.Getenv("CALLER_SKIP") != "" {
		i, err := strconv.ParseInt(os.Getenv("CALLER_SKIP"), 10, 64)
		if err == nil {
			pflagx.CallerSkip = int(i)
		}
	}
	pflagx.Init(rootCmd, &config.CfgFile)
}
func Execute() {
	log := logx.WithName(context.Background(), "main.command")
	var (
		commit    string
		buildTime string
	)
	bi, _ := debug.ReadBuildInfo()
	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.time":
			buildTime = s.Value
			break
		case "vcs.revision":
			commit = s.Value
		}
	}
	rootCmd.AddCommand(cmdx.Version(&config.Version, &commit, &buildTime))
	rootCmd.AddCommand(server.Cmd)
	rootCmd.AddCommand(webhook.Cmd)

	if err := rootCmd.Execute(); err != nil {
		log.Error(err, "exec command failed")
		OsExit(1)
	}
}
