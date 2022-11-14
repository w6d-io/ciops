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

package actions

import (
	"context"
	"github.com/w6d-io/ciops/api/v1alpha1"
	"github.com/w6d-io/x/errorx"
	"github.com/w6d-io/x/logx"
	"net/http"
)

func (c *Component) GetStep(ctx context.Context, id string) (v1alpha1.Step, error) {
	log := logx.WithName(ctx, "Action.Get")
	for _, a := range c.Actions {

		if a.ID == id {
			log.V(2).Info("found action", "id", id)
			return a, nil
		}
	}
	log.Error(nil, "action not found", "id", id)
	return v1alpha1.Step{}, &errorx.Error{Code: "cicd_action_not_found", Message: "action not found", StatusCode: http.StatusNotFound}
}
func Get(_ context.Context) *Component {
	c := Actions
	c.Actions = append(c.Actions, Defaults.Actions...)
	return c
}
