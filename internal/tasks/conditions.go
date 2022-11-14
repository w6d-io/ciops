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
	"github.com/w6d-io/ciops/internal/namespaces"
	corev1 "k8s.io/api/core/v1"
	"strings"

	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	pipelinev1alpha1 "github.com/w6d-io/apis/pipeline/v1alpha1"
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
function check_%v() {
  echo "$(params.%s)" | grep -E "%s" 
}
`

var checkTask = `
function check_%v() {
    "$(task.%s.results.check)" == "%s"
}
`

func GetConditionTask(ctx context.Context, task *pipelinev1alpha1.Task, projectID pipelinev1alpha1.ProjectID) *Task {
	log := logx.WithName(ctx, "tasks.GetConditionTask")
	log.V(2).Info("call")
	var checkFunction string
	var andStatement []string
	params := map[string]string{}
	for _, and := range task.Conditions {
		var orStatement []string
		for name, or := range and {
			//name, _ := goutils.CryptoRandom(10, 0, 0, true, true)
			switch or.Type {
			case pipelinev1alpha1.TASK:
				checkFunction += fmt.Sprintf(checkTask, name, or.Ref, or.When)
				orStatement = append(orStatement, fmt.Sprintf(`check_%v`, name))
			case pipelinev1alpha1.TRIGGER:
				params[or.Ref] = or.When
				checkFunction += fmt.Sprintf(checkTrigger, name, or.Ref, or.When)
				orStatement = append(orStatement, fmt.Sprintf(`check_%v`, name))
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
		Namespace: namespaces.GetName(projectID),
		Name:      fmt.Sprintf("condition-%s", task.ID),
	}
	return &Task{
		task: pipelinev1beta1.Task{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nn.Name,
				Namespace: nn.Namespace,
			},
			Spec: pipelinev1beta1.TaskSpec{
				Params: buildParamsFromMap(params),
				Steps: []pipelinev1beta1.Step{
					{
						Container: corev1.Container{
							Name:  "check",
							Image: "alpine",
						},
						Script: script,
					},
				},
				Results: []pipelinev1beta1.TaskResult{
					{
						Name: "check",
					},
				},
			},
		},
		name:      nn.Name,
		namespace: nn.Namespace,
	}
}

func buildParamsFromMap(params map[string]string) (ps []pipelinev1beta1.ParamSpec) {
	for ref := range params {
		ps = append(ps, pipelinev1beta1.ParamSpec{
			Name: ref,
			Type: pipelinev1beta1.ParamTypeString,
			Default: &pipelinev1beta1.ArrayOrString{
				Type:      pipelinev1beta1.ParamTypeString,
				StringVal: "false",
			},
		})
	}
	return
}
