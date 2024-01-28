/*
Copyright 2020 WILDCARD SA.

Licensed under the WILDCARD SA License, Version 1.0 (the "License");
WILDCARD SA is register in french corporation.
You may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.w6d.io/licenses/LICENSE-1.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is prohibited.
Created on 06/11/2022
*/

package controllers_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	pipelinev1alpha1 "github.com/w6d-io/apis/pipeline/v1alpha1"
	"github.com/w6d-io/ciops/api/v1alpha1"
)

var _ = Describe("pipeline source controller", func() {
	const (
		FactName                               = "test-fact"
		ProjectId   pipelinev1alpha1.ProjectID = 4242
		PipelineRef                            = "pipeline-4242-1"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)
	Context("", func() {
		BeforeEach(func() {
		})
		AfterEach(func() {
		})
		It("", func() {
			By("create namespace", func() {
				Expect(DoNamespace(ctx, k8sClient, ProjectId)).Should(Succeed())
			})
			By("create pipeline resource", func() {
				//namespace := namespaces.GetName(ProjectId)
				r := &v1alpha1.PipelineSource{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pipeline-source-1",
						Namespace: "p6e-cx-4242",
					},
					Spec: pipelinev1alpha1.Pipeline{
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
										Conditions: pipelinev1alpha1.Conditions{{{
											Id:   "condition-1",
											Ref:  "trigger-4242-1",
											Type: "trigger",
											When: "main",
										}}},
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
					},
				}
				Expect(k8sClient.Create(ctx, r)).To(Succeed())
			})
			By("is tasks exist", func() {
				tasks := &v1.TaskList{}
				Eventually(func() bool {
					var opts []client.ListOption
					opts = append(opts, client.MatchingLabels{
						"pipeline-source/name": "pipeline-source-1",
						"task.w6d.io/type":     "task",
					})
					Expect(k8sClient.List(ctx, tasks, opts...)).To(Succeed())
					Expect(len(tasks.Items)).To(Equal(1))
					return true
				})
			})
			By("is condition exist", func() {
				tasks := &v1.TaskList{}
				Eventually(func() bool {
					var opts []client.ListOption
					opts = append(opts, client.MatchingLabels{
						"pipeline-source":  "pipeline-source-1",
						"task.w6d.io/type": "condition",
					})
					Expect(k8sClient.List(ctx, tasks, opts...)).To(Succeed())
					Expect(len(tasks.Items)).To(Equal(1))
					return true
				})
			})
			//By("is pipeline exist", func() {
			//	pLookupKey := types.NamespacedName{Namespace: GetName(ProjectId), Name: "pipeline-4242-1"}
			//	p := &v1.Pipeline{}
			//	Eventually(func() bool {
			//		if err := k8sClient.Get(ctx, pLookupKey, p); err != nil {
			//			return false
			//		}
			//		return true
			//	}, timeout, interval).Should(BeTrue())
			//})
		})
	})
})
