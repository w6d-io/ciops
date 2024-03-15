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

package steps

import (
	"context"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	apis "github.com/w6d-io/apis/pipeline/v1alpha1"
	"github.com/w6d-io/ciops/api/v1alpha2"
	"github.com/w6d-io/ciops/internal/k8s/actions"
	"github.com/w6d-io/x/logx"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	refGitSource        = "git-source"
	refArtefactDownload = "artefact-download"
	refArtefactUpload   = "artefact-upload"
)

type Steps []v1alpha2.Step

func (steps *Steps) GetTektonStep() []pipelinev1.Step {
	var tektonSteps []pipelinev1.Step
	for _, step := range *steps {
		tektonSteps = append(tektonSteps, step.Step)
	}
	return tektonSteps
}

func (steps *Steps) GetTektonParams() []pipelinev1.ParamSpec {
	var ps []pipelinev1.ParamSpec
	for _, step := range *steps {
		params := step.Params
		ps = append(ps, params...)
	}
	return ps
}

func (steps *Steps) GetPre(ctx context.Context) error {
	log := logx.WithName(ctx, "steps.GetPre")
	step, err := actions.GetDefault(ctx).GetStep(ctx, refGitSource)
	if err != nil {
		log.Error(err, "fail to get action", "action", refGitSource)
		return err
	}
	log.V(2).Info("get git source step")
	*steps = append(*steps, step)
	step, err = actions.GetDefault(ctx).GetStep(ctx, refArtefactDownload)
	if err != nil {
		log.Error(err, "fail to get action", "action", refArtefactDownload)
		return err
	}
	log.V(2).Info("get artefact download step")
	*steps = append(*steps, step)
	return nil
}

func (steps *Steps) GetPost(ctx context.Context) error {
	log := logx.WithName(ctx, "steps.GetPost")
	step, err := actions.GetDefault(ctx).GetStep(ctx, refArtefactUpload)
	if err != nil {
		log.Error(err, "fail to get action", "action", refArtefactUpload)
		return err
	}
	log.V(2).Info("get artefact download step")
	*steps = append(*steps, step)
	return nil
}

func (steps *Steps) GetActions(ctx context.Context, r client.Client, a []apis.Action) error {
	log := logx.WithName(ctx, "steps.GetActions")
	for _, action := range a {
		log.V(2).Info("get step", "action_id", action.ID)
		s, err := actions.GetSteps(ctx, r, action.Ref)
		if err != nil {
			log.Error(err, "fail to get action", "action", action.Ref)
			return err
		}
		*steps = append(*steps, s...)
	}
	return nil
}
