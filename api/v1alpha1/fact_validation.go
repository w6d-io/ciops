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
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func ValidateFact(_ string, fact FactSpec) (admission.Warnings, field.ErrorList) {
	warns := admission.Warnings{}
	errs := field.ErrorList{}
	fact.ValidateFactSpec(&warns, &errs)
	fact.ValidateTrigger(&warns, &errs)
	fact.ValidatePipeline(&warns, &errs)
	return warns, errs
}

func (in *FactSpec) ValidateFactSpec(warns *admission.Warnings, errs *field.ErrorList) {
	if in.EventID == nil {
		*errs = append(*errs, field.Invalid(field.NewPath("spec").Child("eventId"), in.EventID, "it should be present"))
	}
	if len(in.PipelineRef) == 0 {
		*errs = append(*errs, field.Invalid(field.NewPath("spec").Child("pipelineRef"), in.EventID, "it should be present"))
	}
	if len(in.ProjectName) == 0 {
		*errs = append(*errs, field.Invalid(field.NewPath("spec").Child("projectName"), in.EventID, "it should be present"))
	}
	if len(in.ProjectURL) == 0 {
		*errs = append(*errs, field.Invalid(field.NewPath("spec").Child("projectUrl"), in.EventID, "it should be present"))
	}
	if len(in.Ref) == 0 {
		*errs = append(*errs, field.Invalid(field.NewPath("spec").Child("ref"), in.EventID, "it should be present"))
	}
	if len(in.CommitMessage) == 0 {
		*warns = append(*warns, "commitMessage is missing")
	}
	if len(in.Commit) == 0 {
		*errs = append(*errs, field.Invalid(field.NewPath("spec").Child("commit"), in.EventID, "it should be present"))
	}
	if len(in.BeforeSha) == 0 {
		*errs = append(*errs, field.Invalid(field.NewPath("spec").Child("beforeSha"), in.EventID, "it should be present"))
	}
	if len(in.UserId) == 0 {
		*errs = append(*errs, field.Invalid(field.NewPath("spec").Child("userId"), in.EventID, "it should be present"))
	}
	return
}

func (in *FactSpec) ValidateTrigger(_ *admission.Warnings, errs *field.ErrorList) {
	trigger := in.Trigger
	if trigger == nil {
		*errs = append(*errs, field.Invalid(field.NewPath("spec").Child("trigger"), in.EventID, "it should be present"))
		return
	}
	if len(trigger.ID) == 0 {
		*errs = append(*errs, field.Invalid(field.NewPath("spec").Child("trigger").Child("id"), trigger.ID, "it should be present"))
	}
	if len(trigger.Ref) == 0 {
		*errs = append(*errs, field.Invalid(field.NewPath("spec").Child("trigger").Child("ref"), trigger.Ref, "it should be present"))
	}
}

func (in *FactSpec) ValidatePipeline(_ *admission.Warnings, errs *field.ErrorList) {
	if in.PipelineSource == nil || len(in.PipelineSource.Name) == 0 {
		*errs = append(*errs, field.Invalid(field.NewPath("spec").Child("pipeline"), in.PipelineSource, "it should be present"))
	}
	return
}
