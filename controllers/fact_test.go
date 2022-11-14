/*
Copyright 2020 WILDCARD SA.

Licensed under the WILDCARD SA License, Version 1.0 (the "License");
WILDCARD SA is register in french corporation.
You may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.w6d.io/licenses/LICENSE-1.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is prohibited.
Created on 05/11/2022
*/

package controllers_test

import (
	"errors"
	corev1 "k8s.io/api/core/v1"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"

	pipelinev1alpha1 "github.com/w6d-io/apis/pipeline/v1alpha1"
	"github.com/w6d-io/ciops/api/v1alpha1"
	"github.com/w6d-io/ciops/controllers"
	"github.com/w6d-io/ciops/internal/namespaces"
)

var _ = Describe("fact controller", func() {
	const (
		FactName                                  = "test-fact"
		ProjectId      pipelinev1alpha1.ProjectID = 4242
		PipelineRef                               = "pipeline-4242-1"
		PipelineSource                            = "pipeline-source-1"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)
	var _ = &pipelinev1alpha1.Pipeline{
		ID:               PipelineRef,
		PipelineIDNumber: "1",
		ProjectID:        ProjectId,
		Triggers: []pipelinev1alpha1.Trigger{
			{
				ID:          "trigger-4242-1",
				Ref:         "github_webhook",
				Type:        "git",
				ComponentId: "",
				Data: map[string]string{
					"branch":    "main",
					"eventType": "Push",
					"provider":  "github",
				},
			},
		},
		Stages: []pipelinev1alpha1.Stage{
			{
				ID:   "stage-4242-1-1659424242",
				Name: "Stage 1",
				Tasks: []pipelinev1alpha1.Task{
					{
						ID:            "task-4242-1-1659424242",
						Name:          "leaks",
						SkipOnFailure: false,
						Conditions:    pipelinev1alpha1.Conditions{{{}}},
						Actions: []pipelinev1alpha1.Action{
							{
								ID:          "action-4242-1-1659424242",
								Name:        "leaks",
								ComponentID: "component-ml-4242-1",
								Ref:         "gitleaks",
								Params: map[string]string{
									"skipOnFailure": "true",
								},
								Data:         map[string]string{},
								Environments: map[string]string{},
								Status:       "",
								StartTime:    0,
								EndTime:      0,
							},
						},
						StartTime: 0,
						EndTime:   0,
						Status:    "",
					},
				},
				Status:    pipelinev1alpha1.Pending.ToString(),
				EndTime:   time.Now().UnixMilli(),
				StartTime: 0,
			},
		},
		Status:    "",
		StartTime: 0,
		EndTime:   0,
		LogUri:    "",
		Complete:  true,
		TriggerId: "trigger-4242-1",
		Commit:    pipelinev1alpha1.Commit{},
		EventID:   "1",
	}
	Context("When fact occurred", func() {
		It("should create pipelinerun", func() {
			namespace := namespaces.GetName(ProjectId)
			By("create namespace", func() {
				Expect(namespaces.DoNamespace(ctx, k8sClient, ProjectId)).Should(Succeed())

			})
			By("create Fact", func() {

				r := &v1alpha1.Fact{
					ObjectMeta: metav1.ObjectMeta{
						Name:      FactName,
						Namespace: namespace,
					},
					Spec: v1alpha1.FactSpec{
						EventID:       pointer.Int64(1),
						PipelineRef:   PipelineRef,
						ProjectName:   FactName,
						ProjectURL:    "https://github.com/w6d-test/test-fact",
						Ref:           "refs/heads/main",
						Commit:        "3cae2b18a93e7bf9ae747287cc618900c955f66f",
						BeforeSha:     "0000000000000000000000000000000000000000",
						CommitMessage: "first commit",
						UserId:        "424242",
						Added: []string{
							"Dockerfile",
							"README.md",
							"app.js",
							"package.json",
						},
						Removed:    []string{},
						Modified:   []string{},
						ProviderId: "github",
						Trigger: &v1alpha1.TriggerSpec{
							ID:   "trigger",
							Ref:  "trigger",
							Type: "push",
						},
						PipelineSource: &corev1.LocalObjectReference{
							Name: PipelineSource,
						},
					},
				}
				Expect(k8sClient.Create(ctx, r)).Should(Succeed())
			})

			By("is pipelinerun exist", func() {
				prLookupKey := types.NamespacedName{Namespace: namespaces.GetName(ProjectId), Name: "pipelinerun-1"}
				pr := &pipelinev1beta1.PipelineRun{}
				Eventually(func() bool {
					err := k8sClient.Get(ctx, prLookupKey, pr)
					if err != nil {
						return false
					}
					return true
				}, timeout, interval).Should(BeTrue())
				Expect(len(pr.Spec.Params)).To(Equal(10))
			})

			By("delete fact", func() {
				r := &v1alpha1.Fact{
					ObjectMeta: metav1.ObjectMeta{
						Name:      FactName,
						Namespace: namespace,
					},
				}
				Expect(k8sClient.Delete(ctx, r)).Should(Succeed())
			})

		})
		It("should failed to get fact", func() {
			r := &controllers.FactReconciler{
				Client:      k8sClient,
				LocalScheme: scheme,
			}
			Expect(r.Scheme()).ToNot(BeNil())
		})
		It("should fail with test message", func() {
			Expect(controllers.IgnoreNotExists(errors.New("test"))).To(HaveOccurred())
		})
		It("should fail on update by getting resource", func() {
			r := &controllers.FactReconciler{
				Client:      k8sClient,
				LocalScheme: scheme,
			}
			Expect(r.UpdateStatus(ctx, types.NamespacedName{
				Namespace: "default",
				Name:      "test-1",
			}, v1alpha1.FactStatus{})).To(HaveOccurred())
		})
		It("evaluates all status", func() {
			r := &controllers.FactReconciler{
				Client:      k8sClient,
				LocalScheme: scheme,
			}
			Expect(r.GetStatus(v1alpha1.Errored)).To(Equal(metav1.ConditionFalse))
			Expect(r.GetStatus(v1alpha1.Succeeded)).To(Equal(metav1.ConditionTrue))
			Expect(r.GetStatus(v1alpha1.Running)).To(Equal(metav1.ConditionUnknown))
		})
	})
})
