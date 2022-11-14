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
	civ1alpha1 "github.com/w6d-io/ciops/api/v1alpha1"
	"github.com/w6d-io/ciops/internal/types"
	"github.com/w6d-io/ciops/internal/util"
	"github.com/w6d-io/x/toolx"
	"k8s.io/apimachinery/pkg/api/meta"
	pkgtypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"time"

	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/w6d-io/x/logx"
)

// Build create the task into kubernetes cluster
func (t *Task) Build(ctx context.Context, r client.Client, object metav1.Object) error {
	log := logx.WithName(ctx, "tasks.Build").WithValues("task", t.name, "namespace", t.namespace)
	log.V(1).Info("build task")
	resource := &pipelinev1beta1.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name:        t.name,
			Namespace:   t.namespace,
			Annotations: t.annotations,
			Labels:      t.labels,
		},
	}
	op, err := controllerutil.CreateOrUpdate(ctx, r, resource, func() error {
		if resource.CreationTimestamp.IsZero() {
			resource.Annotations[types.AnnotationSchedule] = time.Now().Format(time.RFC3339)
			if err := controllerutil.SetControllerReference(object, resource, r.Scheme()); err != nil {
				return err
			}
		}
		resource.Spec = pipelinev1beta1.TaskSpec{
			Workspaces: t.workspaces,
			Params:     t.task.Spec.Params,
			Steps:      t.task.Spec.Steps,
			Results:    t.task.Spec.Results,
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

// GetName returns the task name
func (t *Task) GetName() string {
	return t.name
}

// GetRunAfter return the list of tasks to be running before
func (t *Task) GetRunAfter() []string {
	return t.runAfter
}

func (t *Task) GetWhen() *pipelinev1beta1.WhenExpression {
	return t.when
}

func (t *Task) GetParams() []pipelinev1beta1.ParamSpec {
	return t.task.Spec.Params
}

func (t *Task) UpdateStatus(ctx context.Context, r client.Client, nn pkgtypes.NamespacedName, status civ1alpha1.PipelineSourceStatus) error {
	log := logx.WithName(ctx, "FactReconciler.UpdateStatus").WithValues("resource", nn, "status", status)
	log.V(1).Info("update status")
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {

		e := &civ1alpha1.PipelineSource{}
		if err := r.Get(ctx, nn, e); err != nil {
			return err
		}
		e.Status.State = status.State
		e.Status.Message = status.Message
		for _, task := range status.Tasks {
			if toolx.InArray(task, e.Status.Tasks) {
				continue
			}
			e.Status.Tasks = append(e.Status.Tasks, task)
		}
		for _, ct := range status.ConditionTasks {
			if toolx.InArray(ct, e.Status.Conditions) {
				continue
			}
			e.Status.ConditionTasks = append(e.Status.ConditionTasks, ct)
		}
		for _, condition := range status.Conditions {
			meta.SetStatusCondition(&e.Status.Conditions, condition)
		}
		if err := r.Status().Update(ctx, e); err != nil {
			log.Error(err, "update status failed")
		}
		return nil
	})
	return err
}
