/*
Copyright 2020 WILDCARD SA.

Licensed under the WILDCARD SA License, Version 1.0 (the "License");
WILDCARD SA is register in french corporation.
You may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.w6d.io/licenses/LICENSE-1.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is prohibited.
Created on 04/11/2022
*/

package pipelines

import (
	"context"
	"fmt"
	"github.com/w6d-io/ciops/api/v1alpha1"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/w6d-io/ciops/internal/util"
	"github.com/w6d-io/x/logx"
)

func (p *pipeline) Build(ctx context.Context, r client.Client, object metav1.Object) error {
	log := logx.WithName(ctx, "pipeline.Build").WithValues("pipeline", p.name, "namespace", p.namespace)
	log.V(1).Info("build pipeline")

	var params []pipelinev1beta1.ParamSpec

	for _, t := range p.tasks {
		params = append(params, t.GetParams()...)
	}
	params = util.DeDuplicateParams(params)
	var pt []pipelinev1beta1.PipelineTask
	for _, task := range p.tasks {
		var taskParams []pipelinev1beta1.Param
		for _, pp := range task.GetParams() {
			taskParams = append(taskParams, pipelinev1beta1.Param{
				Name: pp.Name,
				Value: pipelinev1beta1.ArrayOrString{
					Type:      pp.Type,
					StringVal: fmt.Sprintf("$(params.%s)", pp.Name),
				},
			})
		}
		cur := pipelinev1beta1.PipelineTask{
			Name:   task.GetName(),
			Params: taskParams,
			TaskRef: &pipelinev1beta1.TaskRef{
				Name: task.GetName(),
				Kind: pipelinev1beta1.NamespacedTaskKind,
			},
			RunAfter:   task.GetRunAfter(),
			Workspaces: Workspace.WB,
		}
		if w := task.GetWhen(); w != nil {
			cur.WhenExpressions = pipelinev1beta1.WhenExpressions{*w}
		}
		pt = append(pt, cur)
	}
	var wks []pipelinev1beta1.PipelineWorkspaceDeclaration

	for _, w := range Workspace.W {
		wks = append(wks, pipelinev1beta1.PipelineWorkspaceDeclaration{
			Name: w.Name,
		})
	}
	resource := &pipelinev1beta1.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name:        p.name,
			Namespace:   p.namespace,
			Annotations: p.annotations,
			Labels:      p.labels,
		},
	}
	op, err := controllerutil.CreateOrUpdate(ctx, r, resource, func() error {
		if resource.CreationTimestamp.IsZero() {
			resource.Annotations[v1alpha1.AnnotationSchedule] = time.Now().Format(time.RFC3339)
			if err := controllerutil.SetControllerReference(object, resource, r.Scheme()); err != nil {
				return err
			}
		}
		resource.Spec = pipelinev1beta1.PipelineSpec{
			Tasks:      pt,
			Params:     params,
			Workspaces: wks,
		}
		return nil
	})
	log.V(2).Info(resource.Kind, "content", fmt.Sprintf("%v",
		util.GetObjectContain(resource)))
	if err != nil {
		log.Error(err, "create or update failed", "operation", op)
		return err
	}
	log.Info("resource successfully reconciled", "operation", op)
	return nil
}
