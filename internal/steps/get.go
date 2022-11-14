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

package steps

import (
	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/w6d-io/ciops/api/v1alpha1"
)

type Steps []v1alpha1.Step

func (steps *Steps) GetTektonStep() []pipelinev1beta1.Step {
	var tektonSteps []pipelinev1beta1.Step
	for _, step := range *steps {
		tektonSteps = append(tektonSteps, step.Step)
	}
	return tektonSteps
}

func (steps *Steps) GetTektonParams() []pipelinev1beta1.ParamSpec {
	var tektonParams []pipelinev1beta1.ParamSpec
	for _, step := range *steps {
		tektonParams = append(tektonParams, step.Params...)
	}
	return tektonParams
}
