/*
Copyright 2020 WILDCARD SA.

Licensed under the WILDCARD SA License, Version 1.0 (the "License");
WILDCARD SA is register in french corporation.
You may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.w6d.io/licenses/LICENSE-1.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is prohibited.
Created on 03/11/2022
*/

package namespaces

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	pipelinev1alpha1 "github.com/w6d-io/apis/pipeline/v1alpha1"
	"github.com/w6d-io/x/logx"
)

var Prefix string

func DoNamespace(ctx context.Context, r client.Client, projectId pipelinev1alpha1.ProjectID) error {
	log := logx.WithName(ctx, "DoNamespace").WithValues("project_id", projectId.String(), "namespace", GetName(projectId))
	log.V(1).Info("build namespace")

	resource := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: GetName(projectId),
			Labels: map[string]string{
				"ciops.ci.w6d.io/project_id": projectId.String(),
			},
		},
	}
	op, err := controllerutil.CreateOrUpdate(ctx, r, resource, func() error {
		return nil
	})
	if err != nil {
		log.Error(err, "create or update failed", "operation", op)
		return err
	}
	log.Info("resource successfully reconciled", "operation", op)
	return nil
}

func GetName(p pipelinev1alpha1.ProjectID) string {
	return fmt.Sprintf("%s-%s", Prefix, p.String())
}
