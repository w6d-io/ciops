/*
Copyright 2020 WILDCARD SA.

Licensed under the WILDCARD SA License, Version 1.0 (the "License");
WILDCARD SA is register in french corporation.
You may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.w6d.io/licenses/LICENSE-1.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is prohibited.
Created on 23/09/2022
*/

package pipelineruns

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/types"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/pod"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	notification "github.com/w6d-io/apis/notification/v1alpha1"
	"github.com/w6d-io/ciops/api/v1alpha1"
	"github.com/w6d-io/hook"
	"github.com/w6d-io/x/errorx"
	"github.com/w6d-io/x/logx"
)

type LocalConfig struct {
	Template          *pod.Template                 `json:"podTemplate" mapstructure:"podTemplate"`
	WB                []pipelinev1.WorkspaceBinding `json:"workspaces"  mapstructure:"workspaces"`
	PipelinerunPrefix string                        `json:"prefix"      mapstructure:"prefix"`
}

var (
	LC LocalConfig
)

func Build(ctx context.Context, r client.Client, e *v1alpha1.Fact) error {
	namespace := e.Namespace
	eSpec := e.Spec
	ps := new(v1alpha1.PipelineSource)
	if e.Spec.PipelineSource == nil {
		return errorx.New("spec.pipeline must not be empty")
	}
	if err := r.Get(ctx, types.NamespacedName{
		Name:      e.Spec.PipelineSource.Name,
		Namespace: namespace,
	}, ps); err != nil {

	}
	log := logx.WithName(ctx, "pipelinerun.Build").WithValues("pipelinerun", GetPipelinerunName(*eSpec.EventID), "namespace", e.Namespace)
	log.V(1).Info("build pipelinerun")
	params := []pipelinev1.Param{
		{
			Name: "repoUrl",
			Value: pipelinev1.ParamValue{
				Type:      pipelinev1.ParamTypeString,
				StringVal: eSpec.ProjectURL,
			},
		},
		{
			Name: "revision",
			Value: pipelinev1.ParamValue{
				Type:      pipelinev1.ParamTypeString,
				StringVal: eSpec.Commit,
			},
		},
		{
			Name: "projectName",
			Value: pipelinev1.ParamValue{
				Type:      pipelinev1.ParamTypeString,
				StringVal: eSpec.ProjectName,
			},
		},
		{
			Name: "beforeSha",
			Value: pipelinev1.ParamValue{
				Type:      pipelinev1.ParamTypeString,
				StringVal: eSpec.BeforeSha,
			},
		},
		{
			Name: "userId",
			Value: pipelinev1.ParamValue{
				Type:      pipelinev1.ParamTypeString,
				StringVal: eSpec.UserId,
			},
		},
		{
			Name: "W6D_CI_PIPELINERUN_ID",
			Value: pipelinev1.ParamValue{
				Type:      pipelinev1.ParamTypeString,
				StringVal: GetPipelinerunName(*eSpec.EventID),
			},
		},
		{
			Name: "W6D_CI_EVENT_ID",
			Value: pipelinev1.ParamValue{
				Type:      pipelinev1.ParamTypeString,
				StringVal: fmt.Sprintf("%d", *eSpec.EventID),
			},
		},
		{
			Name: eSpec.Trigger.Ref,
			Value: pipelinev1.ParamValue{
				Type:      pipelinev1.ParamTypeString,
				StringVal: "success",
			},
		},
	}
	if eSpec.Added != nil && len(eSpec.Added) > 0 {
		params = append(params, pipelinev1.Param{
			Name: "added",
			Value: pipelinev1.ParamValue{
				Type:     pipelinev1.ParamTypeArray,
				ArrayVal: eSpec.Added,
			},
		})
	}
	if eSpec.Removed != nil && len(eSpec.Removed) > 0 {
		params = append(params, pipelinev1.Param{
			Name: "removed",
			Value: pipelinev1.ParamValue{
				Type:     pipelinev1.ParamTypeArray,
				ArrayVal: eSpec.Removed,
			},
		})
	}
	if eSpec.Modified != nil && len(eSpec.Modified) > 0 {
		params = append(params, pipelinev1.Param{
			Name: "modified",
			Value: pipelinev1.ParamValue{
				Type:     pipelinev1.ParamTypeArray,
				ArrayVal: eSpec.Modified,
			},
		})
	}

	resource := &pipelinev1.PipelineRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:        GetPipelinerunName(*eSpec.EventID),
			Namespace:   e.Namespace,
			Annotations: map[string]string{},
			Labels: map[string]string{
				"pipeline.w6d.io/event_id":    fmt.Sprintf("%d", *eSpec.EventID),
				"pipeline.w6d.io/name":        fmt.Sprintf("pipelinerun-%d", *eSpec.EventID),
				"pipeline.w6d.io/trigger_id":  eSpec.Trigger.ID,
				"pipeline.w6d.io/provider_id": e.Spec.ProviderId,
				"pipeline.w6d.io/project_id":  fmt.Sprintf("%d", e.Spec.ProjectID),
				"pipeline.w6d.io/type":        eSpec.Trigger.Type,
			},
		},
	}
	resource.Annotations[v1alpha1.AnnotationSchedule] = time.Now().Format(time.RFC3339)

	if err := controllerutil.SetControllerReference(e, resource, r.Scheme()); err != nil {
		return err
	}
	op, err := controllerutil.CreateOrUpdate(ctx, r, resource, func() error {
		//if resource.CreationTimestamp.IsZero() {
		//    log.Info("")
		//}
		resource.Spec = pipelinev1.PipelineRunSpec{
			PipelineRef: &pipelinev1.PipelineRef{
				Name: eSpec.PipelineRef,
			},
			Params: params,
			TaskRunTemplate: pipelinev1.PipelineTaskRunTemplate{
				PodTemplate:        LC.Template,
				ServiceAccountName: fmt.Sprintf("sa-%d", e.Spec.ProjectID),
			},
			Workspaces: LC.WB,
		}
		return nil
	})
	log.V(2).Info(resource.Kind, "content", fmt.Sprintf("%v",
		getObjectContain(resource)))
	if err != nil {
		log.Error(err, "create or update failed", "operation", op)
		return err
	}
	payload := &notification.Notification{
		Id:      ps.Spec.ProjectID.String(),
		Type:    "notification",
		Kind:    "project",
		Scope:   []string{"*"},
		Message: fmt.Sprintf("pipeline run created, eventId = %v", e.Spec.EventID),
		Time:    time.Now().UnixMilli(),
	}
	var action string
	if op == controllerutil.OperationResultCreated {
		action = "created"

	} else {
		action = "updated"
	}
	payload.Message = fmt.Sprintf("pipeline run %s, eventId = %v", action, e.Spec.EventID)
	_ = hook.Send(ctx, payload, fmt.Sprintf("notification.pipelinerun.%s", action))
	_ = hook.Send(ctx, &ps, fmt.Sprintf("ci-status.pipelinerun.%s", action))

	log.Info("resource successfully reconciled", "operation", op)
	return nil
}
