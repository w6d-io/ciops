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
	"bytes"
	"fmt"
	tkn "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes/scheme"
	"knative.dev/pkg/apis/duck/v1beta1"

	"gitlab.w6d.io/w6d/ciops/api/v1alpha1"
	"gitlab.w6d.io/w6d/ciops/internal/config"
)

func GetPipelinerunName(id int64) string {
	return fmt.Sprintf("%s-%d", config.GetPipelinerunPrefix(), id)
}

func getObjectContain(obj runtime.Object) string {
	s := json.NewSerializerWithOptions(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme, json.SerializerOptions{Yaml: true})
	buf := new(bytes.Buffer)
	if err := s.Encode(obj, buf); err != nil {
		return "<ERROR>\n"
	}
	return buf.String()
}

// Condition returns a kubernetes State
func Condition(c v1beta1.Conditions) (status v1alpha1.State) {
	if len(c) == 0 {
		return "---"
	}

	switch c[0].Status {
	case corev1.ConditionFalse:
		status = v1alpha1.Failed
	case corev1.ConditionTrue:
		status = v1alpha1.Succeeded
	case corev1.ConditionUnknown:
		status = v1alpha1.Running
	}
	if c[0].Reason != "" {
		if c[0].Reason == "PipelineRunCancelled" || c[0].Reason == "TaskRunCancelled" {
			status = v1alpha1.Cancelled
		}
	}
	return
}

// Message returns a kubernetes Message
func Message(c v1beta1.Conditions) string {
	if len(c) == 0 {
		return ""
	}
	return c[0].Message
}

// IsPipelineRunning return whether or not the pipeline is running
func IsPipelineRunning(pr tkn.PipelineRun) bool {

	nonRunningState := map[v1alpha1.State]bool{
		v1alpha1.Failed:    true,
		v1alpha1.Cancelled: true,
		v1alpha1.Succeeded: true,
	}
	if _, ok := nonRunningState[Condition(pr.Status.Conditions)]; ok {
		return false
	}
	return true
}
