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

package actions

import (
	"context"
	"github.com/w6d-io/ciops/api/v1alpha2"
	"net/http"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/w6d-io/ciops/api/v1alpha1"
	"github.com/w6d-io/x/errorx"
	"github.com/w6d-io/x/logx"
)

func GetDefault(_ context.Context) *Component {
	return Defaults
}

func (c *Component) GetStep(ctx context.Context, id string) (v1alpha2.Step, error) {
	log := logx.WithName(ctx, "action.get")
	for _, a := range c.Actions {

		if a.ID == id {
			log.V(2).Info("found action", "id", id)
			return a, nil
		}
	}
	log.Error(nil, "action not found", "id", id)
	return v1alpha2.Step{}, &errorx.Error{Code: "cicd_action_not_found", Message: "action not found", StatusCode: http.StatusNotFound}
}

func GetSteps(ctx context.Context, r client.Client, id string) ([]v1alpha2.Step, error) {
	log := logx.WithName(ctx, "actions.GetSteps")
	obj := &v1alpha2.ActionList{}

	var opts []client.ListOption
	opts = append(opts, client.InNamespace(""))
	opts = append(opts, client.MatchingLabels{"action.ci.w6d.io/id": id})
	if err := r.List(ctx, obj, opts...); client.IgnoreNotFound(err) != nil {
		log.Error(err, "list action failed")
		return nil, err
	}
	if len(obj.Items) == 0 {
		return nil, apierrors.NewNotFound(v1alpha1.GroupVersion.WithResource("action").GroupResource(), id)
	}
	var steps []v1alpha2.Step
	for _, a := range obj.Items {
		steps = append(steps, a.Steps...)
	}
	return steps, nil
}
