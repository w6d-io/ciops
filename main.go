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
	tknv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
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
		Run: ciops,
	}
	OsExit = os.Exit
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(civ1alpha1.AddToScheme(scheme))
	utilruntime.Must(tknv1beta1.AddToScheme(scheme))
	utilruntime.Must(tknv1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme

	cobra.OnInitialize(config.Init)
	pflagx.CallerSkip = -1
	pflagx.Init(rootCmd, &config.CfgFile)
}

func main() {
	log := logx.WithName(context.TODO(), "Main")
	rootCmd.AddCommand(cmdx.Version(&config.Version, &config.Revision, &config.Built))
	if err := rootCmd.Execute(); err != nil {
		log.Error(err, "exec command failed")
		OsExit(1)
	}
}

func ciops(_ *cobra.Command, _ []string) {
	log := logx.WithName(context.TODO(), "setup")
	if viper.ConfigFileUsed() == "" {
		log.Info("no configuration file set")
	}
	log.Info("start service", "Version", config.Version, "Built",
		config.Built, "Revision", config.Revision)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                        scheme,
		MetricsBindAddress:            viper.GetString(config.ViperKeyMetricsListen),
		Port:                          9443,
		HealthProbeBindAddress:        viper.GetString(config.ViperKeyProbListen),
		LeaderElection:                viper.GetBool(config.ViperKeyEnableLeader),
		LeaderElectionID:              "ciops.w6d.io",
		LeaderElectionReleaseOnCancel: true,
		Namespace:                     viper.GetString(config.ViperKeyNamespace),
	})
	if err != nil {
		log.Error(err, "unable to start manager")
		OsExit(1)
	}

	if err = (&controllers.EventReconciler{
		Client:      mgr.GetClient(),
		EventScheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		log.Error(err, "unable to create controller", "controller", "Event")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Error(err, "unable to set up health check")
		OsExit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Error(err, "unable to set up ready check")
		OsExit(1)
	}

	log.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Error(err, "problem running manager")
		OsExit(1)
	}
}
