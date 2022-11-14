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

package tasks

import (
	"context"
	"fmt"
	notification "github.com/w6d-io/apis/notification/v1alpha1"
	"github.com/w6d-io/hook"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	pkgtypes "k8s.io/apimachinery/pkg/types"
	"strings"
	"time"

	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/w6d-io/ciops/api/v1alpha1"
	"github.com/w6d-io/ciops/internal/actions"
	"github.com/w6d-io/ciops/internal/namespaces"
	"github.com/w6d-io/ciops/internal/steps"
	"github.com/w6d-io/ciops/internal/util"
	"github.com/w6d-io/x/logx"
)

// Parse Build is a function that build tasks from a stage
// TODO parameters
// TODO env var
func (t *Tasks) Parse(ctx context.Context, r client.Client, ps *v1alpha1.PipelineSource) error {
	p := ps.Spec
	log := logx.WithName(ctx, "tasks.Parse")
	log.V(1).Info("parse task", "pipeline-source", ps.Name)

	var runAfter []string
	//stageMap := make(map[string][]string, len(p.Stages))
	for _, stage := range p.Stages {
		for i := range stage.Tasks {
			var condition string
			task := &stage.Tasks[i]
			cond := GetConditionTask(ctx, task, p.ProjectID)
			if cond != nil {
				cond.runAfter = runAfter
				condition = cond.name
				cond.workspaces = append(cond.workspaces, Workspace.WD...)
				t.Tasks = append(t.Tasks, *cond)
				cond.annotations = make(map[string]string)
				cond.labels = make(map[string]string)
				cond.labels["task.w6d.io/type"] = "condition"
				cond.labels["task.w6d.io/hidden"] = "true"
				cond.labels["pipeline-source/name"] = ps.Name
				cond.labels["task.w6d.io/name"] = strings.ReplaceAll(task.Name, " ", "_")
				cond.labels["task.w6d.io/id"] = strings.ReplaceAll(task.ID, " ", "_")
				cond.labels["stage.w6d.io/id"] = strings.ReplaceAll(stage.ID, " ", "_")
				cond.labels["stage.w6d.io/name"] = strings.ReplaceAll(stage.Name, " ", "_")
				if err := cond.Build(ctx, r, ps); err != nil && !errors.IsAlreadyExists(err) {
					_ = hook.Send(ctx, &notification.Notification{
						Id:      ps.Spec.ProjectID.String(),
						Type:    "notification",
						Kind:    "project",
						Scope:   []string{"*"},
						Message: fmt.Sprintf("failed to create condition %s", cond.name),
						Time:    time.Now().UnixMilli(),
					}, "notification.fact.condition.failed")
					log.Error(err, "failed to build condition", "name", cond.name)
					return err
				}
				if err := cond.UpdateStatus(ctx, r, pkgtypes.NamespacedName{
					Namespace: ps.Namespace,
					Name:      ps.Name,
				}, v1alpha1.PipelineSourceStatus{
					ConditionTasks: []string{cond.name},
					State:          "",
					Message:        "",
					Conditions: []metav1.Condition{
						{
							Type:               fmt.Sprintf("ciops.ci.w6d.io/%s", cond.name),
							Status:             "True",
							ObservedGeneration: 0,
							LastTransitionTime: metav1.Time{},
							Reason:             "Created",
							Message:            "condition created",
						},
					},
				}); err != nil {
					log.Error(err, "failed to update status")
				}
				_ = hook.Send(ctx, &notification.Notification{
					Id:      ps.Spec.ProjectID.String(),
					Type:    "notification",
					Kind:    "project",
					Scope:   []string{"*"},
					Message: fmt.Sprintf("condition %s created", task.Name),
					Time:    time.Now().UnixMilli(),
				}, "notification.fact.condition.created")

			}
			if len(task.Actions) == 0 {
				log.V(2).Info("no action found", "task_id", task.ID, "task_name", task.Name)
				continue
			}
			var _steps steps.Steps
			step, err := actions.Get(ctx).GetStep(ctx, refGitSource)
			if err != nil {
				log.Error(err, "fail to get action", "action", refGitSource)
				return err
			}
			log.V(2).Info("get git source step")
			_steps = append(_steps, step)
			step, err = actions.Get(ctx).GetStep(ctx, refArtefactDownload)
			if err != nil {
				log.Error(err, "fail to get action", "action", refArtefactDownload)
				return err
			}
			log.V(2).Info("get artefact download step")
			_steps = append(_steps, step)
			for _, action := range task.Actions {
				log.V(2).Info("get step", "action_id", action.ID)
				step, err := actions.Get(ctx).GetStep(ctx, action.Ref)
				if err != nil {
					log.Error(err, "fail to get action", "action", action.Ref)
					return err
				}
				step.Name = action.ID
				_steps = append(_steps, step)
			}
			step, err = actions.Get(ctx).GetStep(ctx, refArtefactUpload)
			if err != nil {
				log.Error(err, "fail to get action", "action", refArtefactUpload)
				return err
			}
			log.V(2).Info("get artefact download step")
			_steps = append(_steps, step)
			taskTekton := Task{
				task: pipelinev1beta1.Task{
					Spec: pipelinev1beta1.TaskSpec{
						Params: util.DeDuplicateParams(_steps.GetTektonParams()),
						Steps:  _steps.GetTektonStep(),
					},
				},
			}
			for _, step := range _steps {
				taskTekton.workspaces = append(taskTekton.workspaces, step.Workspaces...)
			}
			taskTekton.workspaces = append(taskTekton.workspaces, Workspace.WD...)
			taskTekton.name = task.ID

			taskTekton.namespace = namespaces.GetName(p.ProjectID)
			taskTekton.annotations = make(map[string]string)
			taskTekton.labels = make(map[string]string)
			taskTekton.labels["task.w6d.io/type"] = "task"
			taskTekton.labels["pipeline-source/name"] = ps.Name
			taskTekton.labels["task.w6d.io/name"] = strings.ReplaceAll(task.Name, " ", "_")
			taskTekton.labels["task.w6d.io/id"] = strings.ReplaceAll(task.ID, " ", "_")
			taskTekton.labels["stage.w6d.io/id"] = strings.ReplaceAll(stage.ID, " ", "_")
			taskTekton.labels["stage.w6d.io/name"] = strings.ReplaceAll(stage.Name, " ", "_")
			taskTekton.runAfter = runAfter
			if condition != "" {
				//taskTekton.runAfter = append(taskTekton.runAfter, condition)
				taskTekton.when = &pipelinev1beta1.WhenExpression{
					Input:    fmt.Sprintf("$(tasks.%s.results.check)", condition),
					Operator: selection.In,
					Values:   []string{"succeeded"},
				}
			}
			if err := taskTekton.Build(ctx, r, ps); err != nil && !errors.IsAlreadyExists(err) {
				_ = hook.Send(ctx, &notification.Notification{
					Id:      ps.Spec.ProjectID.String(),
					Type:    "notification",
					Kind:    "project",
					Scope:   []string{"*"},
					Message: fmt.Sprintf("failed to create task %s", task.Name),
					Time:    time.Now().UnixMilli(),
				}, "notification.fact.task.failed")
				return err
			}
			_ = hook.Send(ctx, &notification.Notification{
				Id:      ps.Spec.ProjectID.String(),
				Type:    "notification",
				Kind:    "project",
				Scope:   []string{"*"},
				Message: fmt.Sprintf("task %s created", task.Name),
				Time:    time.Now().UnixMilli(),
			}, "notification.fact.task.created")

			if err := taskTekton.UpdateStatus(ctx, r, pkgtypes.NamespacedName{
				Namespace: ps.Namespace,
				Name:      ps.Name,
			}, v1alpha1.PipelineSourceStatus{
				Tasks:   []string{taskTekton.name},
				State:   "",
				Message: "",
				Conditions: []metav1.Condition{
					{
						Type:               fmt.Sprintf("ciops.ci.w6d.io/%s", taskTekton.name),
						Status:             "True",
						ObservedGeneration: 0,
						LastTransitionTime: metav1.Time{},
						Reason:             "Created",
						Message:            "task created",
					},
				},
			}); err != nil {
				log.Error(err, "failed to update status")
			}
			t.Tasks = append(t.Tasks, taskTekton)
		}
		for _, task := range stage.Tasks {
			runAfter = append(runAfter, task.ID)
			//stageMap[stage.ID] = append(stageMap[stage.ID], task.ID)
		}
	}
	return nil
}
