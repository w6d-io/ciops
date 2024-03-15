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

package server

import (
	"context"
	"github.com/w6d-io/ciops/api/v1alpha2"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	civ1alpha1 "github.com/w6d-io/ciops/api/v1alpha1"
	"github.com/w6d-io/ciops/internal/config"
	"github.com/w6d-io/ciops/internal/controllers"
	"github.com/w6d-io/ciops/internal/toolx"
	"github.com/w6d-io/x/logx"
)

var (
	Cmd = &cobra.Command{
		Use:    "server",
		Short:  "Run the CI operator server",
		RunE:   server,
		PreRun: config.Init,
	}
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(civ1alpha1.AddToScheme(scheme))
	utilruntime.Must(pipelinev1.AddToScheme(scheme))
	utilruntime.Must(v1alpha2.AddToScheme(scheme))

	//+kubebuilder:scaffold:scheme
}

func server(_ *cobra.Command, _ []string) error {
	log := logx.WithName(context.Background(), "server.server")
	if viper.ConfigFileUsed() == "" {
		log.Info("no configuration file set")
	}
	toolx.ShowVersion(log, config.Version)
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
	if err = (&controllers.PipelineSourceReconciler{
		Client:      mgr.GetClient(),
		LocalScheme: scheme,
	}).SetupWithManager(mgr); err != nil {
		log.Error(err, "unable to create controller", "controller", "PipelineSource")
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
