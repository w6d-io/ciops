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

package controllers_test

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/w6d-io/ciops/controllers"
	"github.com/w6d-io/ciops/internal/actions"
	"github.com/w6d-io/ciops/internal/namespaces"
	"github.com/w6d-io/ciops/internal/pipelineruns"
	"github.com/w6d-io/x/logx"
	"os"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	tkn "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	zapraw "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	civ1alpha1 "github.com/w6d-io/ciops/api/v1alpha1"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var ctx context.Context
var k8sClient client.Client
var testEnv *envtest.Environment
var scheme = runtime.NewScheme()

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func() {
	correlationID := uuid.New().String()
	ctx = context.Background()
	ctx = context.WithValue(ctx, logx.CorrelationID, correlationID)

	encoder := zapcore.EncoderConfig{
		// Keys can be anything except the empty string.
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
	}
	opts := zap.Options{
		Encoder:         zapcore.NewConsoleEncoder(encoder),
		Development:     true,
		StacktraceLevel: zapcore.PanicLevel,
		Level:           zapcore.Level(int8(-2)),
	}
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true), zap.UseFlagOptions(&opts), zap.RawZapOpts(zapraw.AddCaller(), zapraw.AddCallerSkip(-1))))
	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "config", "crd", "bases"),
			filepath.Join("..", "third_party", "github.com", "tektoncd", "pipeline", "config"),
		},
		ErrorIfCRDPathMissing: true,
	}

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = civ1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = clientgoscheme.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = tkn.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = civ1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme,
	})
	Expect(err).NotTo(HaveOccurred(), "failed to initiate k8s manager")

	err = (&controllers.FactReconciler{
		Client:      k8sManager.GetClient(),
		LocalScheme: k8sManager.GetScheme(),
	}).SetupWithManager(k8sManager)
	Expect(err).NotTo(HaveOccurred(), "failed to setup manager with Fact")

	err = (&controllers.PipelineSourceReconciler{
		Client:      k8sManager.GetClient(),
		LocalScheme: k8sManager.GetScheme(),
	}).SetupWithManager(k8sManager)
	Expect(err).NotTo(HaveOccurred(), "failed to setup manager with PipelineSource")
	namespaces.Prefix = "p6e-cx"
	pipelineruns.LC = pipelineruns.LocalConfig{
		Template:          nil,
		WB:                nil,
		PipelinerunPrefix: "pipelinerun",
	}
	extraConfig("testdata/default-actions.json", actions.Defaults)
	extraConfig("testdata/actions.json", actions.Actions)
	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(ctx)
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	_ = testEnv.Stop()
	//Expect(err).NotTo(HaveOccurred())
})

func extraConfig(extraFile string, rawVar interface{}) {
	if extraFile != "" {

		data, err := os.ReadFile(extraFile)
		Expect(err).ToNot(HaveOccurred())
		Expect(json.Unmarshal(data, rawVar)).To(Succeed())
	}
}
