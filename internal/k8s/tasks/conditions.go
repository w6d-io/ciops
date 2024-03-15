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

	"github.com/Masterminds/goutils"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	apis "github.com/w6d-io/apis/pipeline/v1alpha1"
	"github.com/w6d-io/x/logx"
)

var checkHeader = `#!/bin/sh

set -o pipefail
`

var check = `
if "%s"
then
  printf "succeeded" > $(results.check.path)
else
  printf "failed" > $(results.check.path)
fi
`

var checkTrigger = `
function check_%s() {
  echo "$(params.%s)" | grep -E "%s" 
}
`

var checkTask = `
function check_%s() {
    "$(task.%s.results.check)" == "%s"
}
`

func GetConditionTask(ctx context.Context, task *apis.Task, projectID int64) *Task {
	log := logx.WithName(ctx, "tasks.GetConditionTask")
	log.V(2).Info("call")
	var checkFunction string
	var andStatement []string
	params := map[string]string{}
	for _, and := range task.Conditions {
		var orStatement []string
		for _, or := range and {
			name, _ := goutils.CryptoRandom(10, 0, 0, true, true)
			switch or.Type {
			case apis.TASK:
				checkFunction += fmt.Sprintf(checkTask, name, or.Ref, or.When)
				orStatement = append(orStatement, fmt.Sprintf(`check_%s`, name))
			case apis.TRIGGER:
				params[or.Ref] = or.When
				checkFunction += fmt.Sprintf(checkTrigger, name, or.Ref, or.When)
				orStatement = append(orStatement, fmt.Sprintf(`check_%s`, name))
			}
		}
		andStatement = append(andStatement, strings.Join(orStatement, " || "))
	}
	ifStatement := strings.Join(andStatement, " && ")
	if ifStatement == "" {
		return nil
	}
	script := checkHeader + checkFunction + fmt.Sprintf(check, ifStatement)
	nn := types.NamespacedName{
		Namespace: fmt.Sprintf("%s-%v", prefixNamespace, projectID),
		Name:      fmt.Sprintf("condition-%s", task.ID),
	}
	return &Task{
		task: pipelinev1.Task{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nn.Name,
				Namespace: nn.Namespace,
			},
			Spec: pipelinev1.TaskSpec{
				Params: buildParamsFromMap(params),
				Steps: []pipelinev1.Step{
					{
						Name:   "check",
						Image:  "alpine",
						Script: script,
					},
				},
				Results: []pipelinev1.TaskResult{
					{
						Name: "check",
					},
				},
			},
		},
		name:      nn.Name,
		namespace: nn.Namespace,
		labels: map[string]string{
			"task.w6d.io/hidden": "true",
			"task.w6d.io/type":   "condition",
		},
		annotations: make(map[string]string),
	}
}

func (t *Tasks) SetCondition(ctx context.Context, task *apis.Task, projectID int64, runAfter []string, name *string) {
	log := logx.WithName(ctx, "tasks.GetConditionTask")
	log.V(2).Info("call")
	var checkFunction string
	var andStatement []string
	params := map[string]string{}
	for _, and := range task.Conditions {
		var orStatement []string
		for _, or := range and {
			name, _ := goutils.CryptoRandom(10, 0, 0, true, true)
			switch or.Type {
			case apis.TASK:
				checkFunction += fmt.Sprintf(checkTask, name, or.Ref, or.When)
				orStatement = append(orStatement, fmt.Sprintf(`check_%s`, name))
			case apis.TRIGGER:
				params[or.Ref] = or.When
				checkFunction += fmt.Sprintf(checkTrigger, name, or.Ref, or.When)
				orStatement = append(orStatement, fmt.Sprintf(`check_%s`, name))
			}
		}
		andStatement = append(andStatement, strings.Join(orStatement, " || "))
	}
	ifStatement := strings.Join(andStatement, " && ")
	if ifStatement == "" {
		return
	}
	script := checkHeader + checkFunction + fmt.Sprintf(check, ifStatement)
	nn := types.NamespacedName{
		Namespace: fmt.Sprintf("%s-%v", prefixNamespace, projectID),
		Name:      fmt.Sprintf("condition-%s", task.ID),
	}
	cond := &Task{
		task: pipelinev1.Task{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nn.Name,
				Namespace: nn.Namespace,
			},
			Spec: pipelinev1.TaskSpec{
				Params: buildParamsFromMap(params),
				Steps: []pipelinev1.Step{
					{
						Name:   "check",
						Image:  "alpine",
						Script: script,
					},
				},
				Results: []pipelinev1.TaskResult{
					{
						Name: "check",
					},
				},
			},
		},
		name:      nn.Name,
		namespace: nn.Namespace,
		runAfter:  runAfter,
		labels: map[string]string{
			"task.w6d.io/hidden": "true",
			"task.w6d.io/type":   "condition",
		},
		annotations: make(map[string]string),
		workspaces:  Workspace.WD,
	}
	*name = cond.name
	t.Tasks = append(t.Tasks, *cond)
	t.Resources = append(t.Resources, cond.Build)
}

func buildParamsFromMap(params map[string]string) (ps []pipelinev1.ParamSpec) {
	for ref := range params {
		ps = append(ps, pipelinev1.ParamSpec{
			Name: ref,
			Type: pipelinev1.ParamTypeString,
			Default: &pipelinev1.ParamValue{
				Type:      pipelinev1.ParamTypeString,
				StringVal: "false",
			},
		})
	}
	return
}
