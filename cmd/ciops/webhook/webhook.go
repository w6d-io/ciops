/*
Copyright 2020 WILDCARD SA.

Licensed under the WILDCARD SA License, Version 1.0 (the "License");
WILDCARD SA is register in french corporation.
You may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.w6d.io/licenses/LICENSE-1.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is prohibited.
Created on 27/01/2024
*/

package webhook

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	tknv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	civ1alpha1 "github.com/w6d-io/ciops/api/v1alpha1"
	"github.com/w6d-io/ciops/internal/config"
	"github.com/w6d-io/ciops/internal/toolx"
	"github.com/w6d-io/x/logx"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	Cmd = &cobra.Command{
		Use:    "webhook",
		Short:  "Run the webhook CI Operator",
		PreRun: config.InitWebhook,
		RunE:   wh,
	}
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(civ1alpha1.AddToScheme(scheme))
	utilruntime.Must(tknv1.AddToScheme(scheme))

}

func wh(_ *cobra.Command, _ []string) error {
	log := logx.WithName(context.Background(), "webhook.command")
	if viper.ConfigFileUsed() == "" {
		log.Info("no configuration file set")
	}
	toolx.ShowVersion(log, config.Version)

	ws := webhook.NewServer(webhook.Options{
		Host: viper.GetString(config.ViperKeyWebhookHost),
		Port: viper.GetInt(config.ViperKeyWebhookPort),
	})
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     viper.GetString(config.ViperKeyMetricsListen),
		HealthProbeBindAddress: viper.GetString(config.ViperKeyProbeListen),
		LeaderElection:         false,
		WebhookServer:          ws,
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
	if err := mgr.AddReadyzCheck("readyz", ws.StartedChecker()); err != nil {
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
