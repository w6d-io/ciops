/*
Copyright 2020 WILDCARD SA.

Licensed under the WILDCARD SA License, Version 1.0 (the "License");
WILDCARD SA is register in french corporation.
You may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.w6d.io/licenses/LICENSE-1.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is prohibited.
Created on 02/02/2024
*/

package pipelines

import (
	"context"
	"fmt"
	"strings"
	"time"

	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/w6d-io/ciops/api/v1alpha1"
	"github.com/w6d-io/ciops/internal/k8s/tasks"
	"github.com/w6d-io/ciops/internal/toolx"
	"github.com/w6d-io/x/logx"
)

func (p *Pipelines) Parse(ctx context.Context, r client.Client, e *v1alpha1.PipelineSource) error {
	log := logx.WithName(ctx, "pipelines.Parse")
	log.V(1).Info("parse pipeline")

	req := &e.Spec
	ts := tasks.Tasks{}
	err := ts.Parse(ctx, r, e)
	if err != nil {
		log.Error(err, "fail to parse tasks")
		return err
	}
	log.V(2).Info("build pipeline tekton")
	pipelineTekton := &pipeline{
		name:        req.ID,
		namespace:   fmt.Sprintf("%s-%v", "p6e-cx", req.ProjectID),
		labels:      make(map[string]string),
		annotations: make(map[string]string),
		triggers:    make(map[string]string),
	}
	for _, trigger := range req.Triggers {
		pipelineTekton.triggers[trigger.ID] = trigger.Ref
	}

	pipelineTekton.tasks = ts

	pipelineTekton.labels["pipeline.w6d.io/id"] = strings.ReplaceAll(req.ID, " ", "_")
	pipelineTekton.labels["pipeline.w6d.io/name"] = strings.ReplaceAll(req.Name, " ", "_")
	pipelineTekton.labels["pipeline.w6d.io/type"] = strings.ReplaceAll(req.Type, " ", "_")
	pipelineTekton.labels["pipeline.w6d.io/number"] = req.PipelineIDNumber
	pipelineTekton.labels["pipeline.w6d.io/projectId"] = fmt.Sprintf("%d", req.ProjectID)
	p.Resources = append(p.Resources, ts.Resources...)
	p.Resources = append(p.Resources, pipelineTekton.build)
	log.V(2).Info("done")
	return nil
}

func (p *pipeline) build(ctx context.Context, r client.Client, e *v1alpha1.PipelineSource) error {
	log := logx.WithName(ctx, "pipeline.build").WithValues("pipeline", p.name, "namespace", p.namespace)
	log.V(1).Info("build pipeline")

	var params []pipelinev1.ParamSpec

	for _, t := range p.tasks.Tasks {
		params = append(params, t.GetParams()...)
	}
	params = toolx.DeDuplicateParams(params)
	var pt []pipelinev1.PipelineTask
	for _, task := range p.tasks.Tasks {
		var taskParams []pipelinev1.Param
		for _, pp := range task.GetParams() {
			taskParams = append(taskParams, pipelinev1.Param{
				Name: pp.Name,
				Value: pipelinev1.ParamValue{
					Type:      pp.Type,
					StringVal: fmt.Sprintf("$(params.%s)", pp.Name),
				},
			})
		}
		cur := pipelinev1.PipelineTask{
			Name:   task.GetName(),
			Params: taskParams,
			TaskRef: &pipelinev1.TaskRef{
				Name: task.GetName(),
				Kind: pipelinev1.NamespacedTaskKind,
			},
			RunAfter:   task.GetRunAfter(),
			Workspaces: Workspace.WB,
		}
		if w := task.GetWhen(); w != nil {
			cur.When = pipelinev1.WhenExpressions{*w}
		}
		pt = append(pt, cur)
	}
	var wks []pipelinev1.PipelineWorkspaceDeclaration

	for _, w := range Workspace.W {
		wks = append(wks, pipelinev1.PipelineWorkspaceDeclaration{
			Name: w.Name,
		})
	}
	resource := &pipelinev1.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name:        p.name,
			Namespace:   p.namespace,
			Annotations: p.annotations,
			Labels:      p.labels,
		},
	}
	resource.Annotations[v1alpha1.AnnotationSchedule] = time.Now().Format(time.RFC3339)

	if err := controllerutil.SetControllerReference(e, resource, r.Scheme()); err != nil {
		return err
	}
	op, err := controllerutil.CreateOrUpdate(ctx, r, resource, func() error {
		resource.Spec = pipelinev1.PipelineSpec{
			Tasks:      pt,
			Params:     params,
			Workspaces: wks,
		}
		return nil
	})
	log.V(2).Info(resource.Kind, "content", fmt.Sprintf("%v",
		toolx.GetObjectContain(resource)))
	if err != nil {
		log.Error(err, "create or update failed", "operation", op)
		return err
	}
	log.Info("resource successfully reconciled", "operation", op)
	return nil
}
