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

	"github.com/w6d-io/ciops/internal/namespaces"
	"github.com/w6d-io/ciops/internal/tasks"
	"github.com/w6d-io/x/logx"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

func Parse(ctx context.Context, r client.Client, tasks []tasks.Task, req *v1alpha1.PipelineSource) error {
	log := logx.WithName(ctx, "pipelines.Parse")
	log.V(1).Info("parse pipeline")

	pipelineTekton := &pipeline{
		name:        req.Spec.ID,
		tasks:       tasks,
		namespace:   namespaces.GetName(req.Spec.ProjectID),
		labels:      make(map[string]string),
		annotations: make(map[string]string),
		triggers:    make(map[string]string),
	}
	labels := make(map[string]string)
	labels["pipeline.w6d.io/id"] = strings.ReplaceAll(req.Spec.ID, " ", "_")
	labels["pipeline.w6d.io/name"] = strings.ReplaceAll(req.Spec.Name, " ", "_")
	labels["pipeline.w6d.io/type"] = strings.ReplaceAll(req.Spec.Type, " ", "_")
	labels["pipeline.w6d.io/number"] = req.Spec.PipelineIDNumber
	labels["pipeline.w6d.io/projectId"] = fmt.Sprintf("%d", req.Spec.ProjectID)
	for _, trigger := range req.Spec.Triggers {
		pipelineTekton.triggers[trigger.ID] = trigger.Ref
	}

	if err := pipelineTekton.Build(ctx, r, req); err != nil {
		log.Error(err, "build pipeline failed")
		return err
	}
	return nil
}
