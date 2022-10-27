/*
Copyright 2022 WILDCARD.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	tknv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	//+kubebuilder:scaffold:imports

	civ1alpha1 "github.com/w6d-io/ciops/api/v1alpha1"
	"github.com/w6d-io/ciops/controllers"
	"github.com/w6d-io/ciops/internal/config"
	"github.com/w6d-io/x/cmdx"
	"github.com/w6d-io/x/logx"
	"github.com/w6d-io/x/pflagx"
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
		RunE:  webhook,
	}
	OsExit = os.Exit
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(civ1alpha1.AddToScheme(scheme))
	utilruntime.Must(tknv1beta1.AddToScheme(scheme))
	//utilruntime.Must(tknv1.AddToScheme(scheme))
	//scheme.AddKnownTypes(tknv1.SchemeGroupVersion, &tknv1.PipelineRun{}, &tknv1.PipelineRunList{})
	//+kubebuilder:scaffold:scheme

	cobra.OnInitialize(config.Init)
	pflagx.CallerSkip = -1
	pflagx.Init(s, &config.CfgFile)
	pflagx.Init(wh, &config.CfgFile)
}

func main() {
	log := logx.WithName(context.TODO(), "Main")
	rootCmd.AddCommand(cmdx.Version(&config.Version, &config.Revision, &config.Built))
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
	log.Info("start service", "Version", config.Version, "Built",
		config.Built, "Revision", config.Revision)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                        scheme,
		MetricsBindAddress:            viper.GetString(config.ViperKeyMetricsListen),
		HealthProbeBindAddress:        viper.GetString(config.ViperKeyProbeListen),
		LeaderElection:                viper.GetBool(config.ViperKeyLeaderElect),
		LeaderElectionID:              viper.GetString(config.ViperKeyLeaderName),
		LeaderElectionNamespace:       viper.GetString(config.ViperKeyLeaderNamespace),
		LeaderElectionReleaseOnCancel: true,
		Namespace:                     viper.GetString(config.ViperKeyNamespace),
	})
	if err != nil {
		log.Error(err, "unable to start manager")
		return err
	}

	if err = (&controllers.FactReconciler{
		Client:     mgr.GetClient(),
		FactScheme: scheme,
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

func webhook(_ *cobra.Command, _ []string) error {
	log := logx.WithName(context.TODO(), "setup")
	if viper.ConfigFileUsed() == "" {
		log.Info("no configuration file set")
	}
	log.Info("start webhook", "Version", config.Version, "Built",
		config.Built, "Revision", config.Revision)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     viper.GetString(config.ViperKeyMetricsListen),
		HealthProbeBindAddress: viper.GetString(config.ViperKeyProbeListen),
		LeaderElection:         false,
		Port:                   viper.GetInt(config.ViperKeyWebhookListen),
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
