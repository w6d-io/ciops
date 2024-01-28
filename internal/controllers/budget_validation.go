/*
Copyright 2020 WILDCARD SA.

Licensed under the WILDCARD SA License, Version 1.0 (the "License");
WILDCARD SA is register in french corporation.
You may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.w6d.io/licenses/LICENSE-1.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is prohibited.
Created on 25/09/2022
*/

package controllers

import (
	"context"

	tkn "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/w6d-io/ciops/api/v1alpha1"
	"github.com/w6d-io/ciops/internal/pipelineruns"
	"github.com/w6d-io/x/logx"
)

func (r *FactReconciler) checkConcurrency(ctx context.Context, nn types.NamespacedName, pipelineName string) error {
	log := logx.WithName(ctx, "checkConcurrency")
	log.V(1).Info("getting all pipeline run")
	status := v1alpha1.FactStatus{PipelineRunName: pipelineName}
	prs := new(tkn.PipelineRunList)
	if err := r.List(ctx, prs, client.InNamespace(nn.Namespace)); IgnoreNotExists(err) != nil {
		log.Error(err, "get list pipelinerun failed")
		log.V(1).Info("update status", "status", v1alpha1.Errored,
			"step", "2")
		status.State = v1alpha1.Errored
		if err := r.UpdateStatus(ctx, nn, status); err != nil {
			return err
		}
		return err
	}
	log.V(1).Info("check pipeline run running")
	var runningPipeline []tkn.PipelineRun
	for _, pr := range prs.Items {
		if pipelineruns.IsPipelineRunning(pr) {
			runningPipeline = append(runningPipeline, pr)
		}
	}
	log.V(1).Info("pipelinerun running", "count", len(runningPipeline))

	log.V(1).Info("get fact budget")

	ebs := new(v1alpha1.FactBudgetList)
	if err := r.List(ctx, ebs, client.InNamespace(nn.Namespace)); IgnoreNotExists(err) != nil {
		log.Error(err, "get fact budget failed")
		return err
	}

	if len(ebs.Items) == 0 {
		log.Info("no fact budget")
	}
	var minItems int64
	for _, eb := range ebs.Items {
		if eb.Spec.Pipeline.PipelineRef != nil {
			if pipelineName != *eb.Spec.Pipeline.PipelineRef {
				continue
			}
		}
		if eb.Spec.Pipeline.Concurrent != nil {
			minItems = minInt64(minItems, *eb.Spec.Pipeline.Concurrent)
		}
	}
	if minItems == 0 {
		log.V(2).Info("no pipeline concurrency")
		return nil
	}
	if minItems <= int64(len(runningPipeline)) {
		log.V(1).Info("hit concurrence pipeline", "action", "queued",
			"minimum", minItems, "count", len(runningPipeline))
		log.V(1).Info("update status", "status", v1alpha1.Queued,
			"step", "4")
		status.State = v1alpha1.Queued
		if err := r.UpdateStatus(ctx, nn, status); err != nil {
			return err
		}
	}
	return nil
}

func minInt64(a, b int64) int64 {
	if a == 0 {
		return b
	}
	if b == 0 {
		return a
	}
	if a < b {
		return a
	}
	return b
}
