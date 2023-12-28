/*
Copyright 2020 WILDCARD SA.

Licensed under the WILDCARD SA License, Version 1.0 (the "License");
WILDCARD SA is register in french corporation.
You may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.w6d.io/licenses/LICENSE-1.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is prohibited.
Created on 26/12/2023
*/

package ciops

import (
	"context"
	"github.com/go-logr/logr"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	tknv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	//+kubebuilder:scaffold:imports

	civ1alpha1 "github.com/w6d-io/ciops/api/v1alpha1"
	"github.com/w6d-io/ciops/controllers"
	"github.com/w6d-io/ciops/internal/config"
	"github.com/w6d-io/x/cmdx"
	"github.com/w6d-io/x/logx"
	"github.com/w6d-io/x/pflagx"
	"github.com/w6d-io/x/toolx"
)

var (
	scheme  = runtime.NewScheme()
	rootCmd = &cobra.Command{
		Use: "ciops",
	}
	s = &cobra.Command{
		Use:   "serve",
		Short: "Run the CI/CD server",
		RunE:  serve,
	}
	wh = &cobra.Command{
		Use:   "webhook",
		Short: "Run the CI/CD webhook",
		RunE:  webhookCmd,
	}
	OsExit = os.Exit
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(civ1alpha1.AddToScheme(scheme))
	utilruntime.Must(tknv1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme

	cobra.OnInitialize(config.Init)
	pflagx.CallerSkip = -1
	pflagx.Init(s, &config.CfgFile)
	pflagx.Init(wh, &config.CfgFile)
}

func Execute() {
	log := logx.WithName(context.TODO(), "Main")
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
	rootCmd.AddCommand(s)
	rootCmd.AddCommand(wh)
	if err := rootCmd.Execute(); err != nil {
		log.Error(err, "exec command failed")
		OsExit(1)
	}
}

func serve(_ *cobra.Command, _ []string) error {
	log := logx.WithName(context.TODO(), "setup")
	if viper.ConfigFileUsed() == "" {
		log.Info("no configuration file set")
	}
	showVersion(log)
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                        scheme,
		MetricsBindAddress:            viper.GetString(config.ViperKeyMetricsListen),
		HealthProbeBindAddress:        viper.GetString(config.ViperKeyProbeListen),
		LeaderElection:                viper.GetBool(config.ViperKeyLeaderElect),
		LeaderElectionID:              viper.GetString(config.ViperKeyLeaderName),
		LeaderElectionNamespace:       viper.GetString(config.ViperKeyLeaderNamespace),
		LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		log.Error(err, "unable to start manager")
		return err
	}

	if err = (&controllers.FactReconciler{
		Client:      mgr.GetClient(),
		LocalScheme: scheme,
	}).SetupWithManager(mgr); err != nil {
		log.Error(err, "unable to create controller", "controller", "Fact")
		return err
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Error(err, "unable to set up health check")
		return err
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Error(err, "unable to set up ready check")
		return err
	}

	log.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Error(err, "problem running manager")
		return err
	}
	return nil
}

func webhookCmd(_ *cobra.Command, _ []string) error {
	log := logx.WithName(context.TODO(), "setup")
	if viper.ConfigFileUsed() == "" {
		log.Info("no configuration file set")
	}
	showVersion(log)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     viper.GetString(config.ViperKeyMetricsListen),
		HealthProbeBindAddress: viper.GetString(config.ViperKeyProbeListen),
		LeaderElection:         false,
		WebhookServer: webhook.NewServer(webhook.Options{
			Host: viper.GetString(config.ViperKeyWebhookHost),
			Port: viper.GetInt(config.ViperKeyWebhookPort),
		}),
	})
	if err != nil {
		log.Error(err, "unable to start manager")
		return err
	}

	if err = (&civ1alpha1.Fact{}).SetupWebhookWithManager(mgr); err != nil {
		log.Error(err, "unable to create webhook", "webhook", "Fact")

		return err
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Error(err, "unable to set up health check")
		return err
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Error(err, "unable to set up ready check")
		return err
	}

	log.Info("starting webhook")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Error(err, "problem running webhook")
		return err
	}
	return nil
}

func showVersion(log logr.Logger) {
	var info []interface{}
	info = append(info, "version", config.Version)
	bi, _ := debug.ReadBuildInfo()
	for _, s := range bi.Settings {
		if toolx.InArray(s.Key, []string{"vcs.time", "vcs.revision"}) {
			info = append(info, s.Key[4:], s.Value)
		}
	}
	log.Info("start service", info...)
}
