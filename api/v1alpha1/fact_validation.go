/*
Copyright 2020 WILDCARD SA.

Licensed under the WILDCARD SA License, Version 1.0 (the "License");
WILDCARD SA is register in french corporation.
You may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.w6d.io/licenses/LICENSE-1.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is prohibited.
Created on 23/10/2022
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func ValidateFact(_ string, fact FactSpec) (admission.Warnings, *apis.FieldError) {
	var errs *apis.FieldError
	errs = ValidateFactSpec(fact)
	errs = errs.Also(ValidateTrigger(fact.Trigger))
	errs = errs.Also(ValidatePipeline(fact.PipelineSource))
	return nil, errs
}

func ValidateFactSpec(fact FactSpec) (errs *apis.FieldError) {
	if fact.EventID == nil {
		errs = errs.Also(apis.ErrMissingField("spec", "eventId"))
	}
	if len(fact.PipelineRef) == 0 {
		errs = errs.Also(apis.ErrMissingField("spec", "pipelineRef"))
	}
	if len(fact.ProjectName) == 0 {
		errs = errs.Also(apis.ErrMissingField("spec", "projectName"))
	}
	if len(fact.ProjectURL) == 0 {
		errs = errs.Also(apis.ErrMissingField("spec", "projectUrl"))
	}
	if len(fact.Ref) == 0 {
		errs = errs.Also(apis.ErrMissingField("spec", "ref"))
	}
	if len(fact.Commit) == 0 {
		errs = errs.Also(apis.ErrMissingField("spec", "commit"))
	}
	if len(fact.BeforeSha) == 0 {
		errs = errs.Also(apis.ErrMissingField("spec", "beforeSha"))
	}
	if len(fact.CommitMessage) == 0 {
		errs = errs.Also(apis.ErrMissingField("spec", "commitMessage"))
	}

	if len(fact.UserId) == 0 {
		errs = errs.Also(apis.ErrMissingField("spec", "userId"))
	}
	return
}

func ValidateTrigger(trigger *TriggerSpec) (errs *apis.FieldError) {
	if trigger == nil {
		errs = errs.Also(apis.ErrMissingField("spec", "trigger"))
		return
	}
	if len(trigger.ID) == 0 {
		errs = errs.Also(apis.ErrMissingField("spec", "trigger", "id"))
	}
	if len(trigger.Ref) == 0 {
		errs = errs.Also(apis.ErrMissingField("spec", "trigger", "ref"))
	}
	return
}

func ValidatePipeline(p *corev1.LocalObjectReference) (errs *apis.FieldError) {
	if p == nil || len(p.Name) == 0 {
		errs = errs.Also(apis.ErrMissingField("spec", "pipeline"))
	}
	return
}
