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

import "github.com/w6d-io/ciops/api/v1alpha1"

type Component struct {
	Fields  []v1alpha1.Field `mapstructure:"fields"  json:"fields"`
	Actions []v1alpha1.Step  `mapstructure:"actions" json:"actions"`
}

var Actions = new(Component)
var Defaults = new(Component)
