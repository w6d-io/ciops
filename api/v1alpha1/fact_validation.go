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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	pipeline "github.com/w6d-io/apis/pipeline/v1alpha1"
)

func ValidateFact(name string, fact FactSpec) (admission.Warnings, error) {
	var allErrs field.ErrorList

	allErrs = append(allErrs, ValidateFactSpec(fact)...)
	allErrs = append(allErrs, ValidateTrigger(fact.Trigger)...)
	allErrs = append(allErrs, ValidatePipeline(fact.Pipeline)...)
	if len(allErrs) == 0 {
		return nil, nil
	}
	return nil, apierrors.NewInvalid(schema.GroupKind{
		Group: GroupVersion.String(),
		Kind:  "Fact",
	}, name, allErrs)
}

func ValidateFactSpec(fact FactSpec) (allErrs field.ErrorList) {
	if fact.EventID == nil {
		allErrs = append(
			allErrs,
			field.Required(
				field.NewPath("spec").Child("eventId"), ""),
		)
	}
	if fact.ProjectID == 0 {
		allErrs = append(
			allErrs,
			field.Required(
				field.NewPath("spec").Child("projectId"), ""),
		)
	}
	if len(fact.PipelineRef) == 0 {
		allErrs = append(
			allErrs,
			field.Required(
				field.NewPath("spec").Child("pipelineRef"), ""),
		)
	}
	if len(fact.ProjectName) == 0 {
		allErrs = append(
			allErrs,
			field.Required(
				field.NewPath("spec").Child("projectName"), ""),
		)
	}
	if len(fact.ProjectURL) == 0 {
		allErrs = append(
			allErrs,
			field.Required(
				field.NewPath("spec").Child("projectUrl"), ""),
		)
	}
	if len(fact.Ref) == 0 {
		allErrs = append(
			allErrs,
			field.Required(
				field.NewPath("spec").Child("ref"), ""),
		)
	}
	if len(fact.Commit) == 0 {
		allErrs = append(
			allErrs,
			field.Required(
				field.NewPath("spec").Child("commit"), ""),
		)
	}
	if len(fact.BeforeSha) == 0 {
		allErrs = append(
			allErrs,
			field.Required(
				field.NewPath("spec").Child("beforeSha"), ""),
		)
	}
	if len(fact.CommitMessage) == 0 {
		allErrs = append(
			allErrs,
			field.Required(
				field.NewPath("spec").Child("commitMessage"), ""),
		)
	}

	if len(fact.UserId) == 0 {
		allErrs = append(
			allErrs,
			field.Required(
				field.NewPath("spec").Child("userId"), ""),
		)
	}
	return
}

func ValidateTrigger(trigger *TriggerSpec) (allErrs field.ErrorList) {
	if trigger == nil {
		allErrs = append(
			allErrs,
			field.Required(
				field.NewPath("spec").Child("trigger"), ""),
		)
		return
	}
	if len(trigger.ID) == 0 {
		allErrs = append(
			allErrs,
			field.Required(
				field.NewPath("spec").Child("trigger").Child("id"), ""),
		)
	}
	if len(trigger.Ref) == 0 {
		allErrs = append(
			allErrs,
			field.Required(
				field.NewPath("spec").Child("trigger").Child("ref"), ""),
		)
	}
	return
}

func ValidatePipeline(p *pipeline.Pipeline) (allErrs field.ErrorList) {
	if p == nil {
		allErrs = append(
			allErrs,
			field.Required(
				field.NewPath("spec").Child("pipeline"), ""),
		)
	}
	return
}
