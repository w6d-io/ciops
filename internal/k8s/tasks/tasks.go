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

package tasks

import (
	"context"
	"fmt"
	"strings"
	"time"

	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	apis "github.com/w6d-io/apis/pipeline/v1alpha1"
	"github.com/w6d-io/ciops/api/v1alpha1"
	"github.com/w6d-io/ciops/internal/k8s/steps"
	"github.com/w6d-io/ciops/internal/toolx"
	"github.com/w6d-io/x/logx"
)

// Parse Build is a function that build tasks from a stage
// TODO parameters
// TODO env var
func (t *Tasks) Parse(ctx context.Context, r client.Client, e *v1alpha1.PipelineSource) error {
	log := logx.WithName(ctx, "tasks.Parse")
	p := &e.Spec

	log.V(1).Info("parse task", "pipeline", p.ID)
	var runAfter []string
	//stageMap := make(map[string][]string, len(p.Stages))
	for _, stage := range p.Stages {
		for i := range stage.Tasks {
			var condition string
			task := &stage.Tasks[i]
			if len(task.Actions) == 0 {
				log.V(2).Info("no action found", "task_id", task.ID, "task_name", task.Name)
				continue
			}
			t.SetCondition(ctx, task, int64(p.ProjectID), runAfter, &condition)
			_steps := &steps.Steps{}
			if err := _steps.GetPre(ctx); err != nil {
				return err
			}
			if err := _steps.GetActions(ctx, r, task.Actions); err != nil {
				return err
			}
			if err := _steps.GetPost(ctx); err != nil {
				return err
			}
			t.Add(_steps, task, condition, p.ProjectID.String(), runAfter, stage)
		}
		for _, task := range stage.Tasks {
			runAfter = append(runAfter, task.ID)
			//stageMap[stage.ID] = append(stageMap[stage.ID], task.ID)
		}
	}
	log.V(2).Info("parse task done")
	return nil
}

func (t *Tasks) Add(s *steps.Steps, task *apis.Task, condition, projectId string, runAfter []string, stage apis.Stage) {
	taskTekton := Task{
		task: pipelinev1.Task{
			Spec: pipelinev1.TaskSpec{
				Params: toolx.DeDuplicateParams(s.GetTektonParams()),
				Steps:  s.GetTektonStep(),
			},
		},
	}
	for _, step := range *s {
		taskTekton.workspaces = append(taskTekton.workspaces, step.Workspaces...)
	}
	taskTekton.workspaces = append(taskTekton.workspaces, Workspace.WD...)
	taskTekton.name = task.ID

	taskTekton.namespace = fmt.Sprintf("%s-%v", prefixNamespace, projectId)
	taskTekton.annotations = make(map[string]string)
	taskTekton.labels = make(map[string]string)
	taskTekton.labels["task.w6d.io/type"] = "task"
	taskTekton.labels["task.w6d.io/name"] = strings.ReplaceAll(task.Name, " ", "_")
	taskTekton.labels["task.w6d.io/id"] = strings.ReplaceAll(task.ID, " ", "_")
	taskTekton.labels["stage.w6d.io/id"] = strings.ReplaceAll(stage.ID, " ", "_")
	taskTekton.labels["stage.w6d.io/name"] = strings.ReplaceAll(stage.Name, " ", "_")
	//taskTekton.labels["stage.w6d.io/name"] = stage.Name
	taskTekton.runAfter = runAfter
	if condition != "" {
		//taskTekton.runAfter = append(taskTekton.runAfter, condition)
		taskTekton.when = &pipelinev1.WhenExpression{
			Input:    fmt.Sprintf("$(tasks.%s.results.check)", condition),
			Operator: selection.In,
			Values:   []string{"succeeded"},
		}
	}
	t.Tasks = append(t.Tasks, taskTekton)
	t.Resources = append(t.Resources, taskTekton.Build)
}

// Build create the task into kubernetes cluster
func (t *Task) Build(ctx context.Context, r client.Client, e *v1alpha1.PipelineSource) error {
	log := logx.WithName(ctx, "tasks.Build").WithValues("task", t.name, "namespace", t.namespace)
	log.V(1).Info("build task")
	resource := &pipelinev1.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name:        t.name,
			Namespace:   t.namespace,
			Annotations: t.annotations,
			Labels:      t.labels,
		},
	}
	resource.Annotations[v1alpha1.AnnotationSchedule] = time.Now().Format(time.RFC3339)
	if err := controllerutil.SetControllerReference(e, resource, r.Scheme()); err != nil {
		return err
	}
	op, err := controllerutil.CreateOrUpdate(ctx, r, resource, func() error {
		resource.Spec = pipelinev1.TaskSpec{
			Workspaces: t.workspaces,
			Params:     t.task.Spec.Params,
			Steps:      t.task.Spec.Steps,
			Results:    t.task.Spec.Results,
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

// GetName returns the task name
func (t *Task) GetName() string {
	return t.name
}

// GetRunAfter return the list of tasks to be running before
func (t *Task) GetRunAfter() []string {
	return t.runAfter
}

func (t *Task) GetWhen() *pipelinev1.WhenExpression {
	return t.when
}

func (t *Task) GetParams() []pipelinev1.ParamSpec {
	return t.task.Spec.Params
}
