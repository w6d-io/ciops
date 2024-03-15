/*
Copyright 2020 WILDCARD SA.

Licensed under the WILDCARD SA License, Version 1.0 (the "License");
WILDCARD SA is register in french corporation.
You may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.w6d.io/licenses/LICENSE-1.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is prohibited.
Created on 06/02/2024
*/

package sa

import (
	"context"
	"github.com/w6d-io/ciops/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	GitPrefixSecret = "secret-git"
)

// ServiceAccount ...
type ServiceAccount struct {
	Resources []func(context.Context, client.Client, *v1alpha1.PipelineSource) error
}

type metadata struct {
	name      string
	namespace string
	projectID int64
	labels    map[string]string
}
