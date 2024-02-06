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
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/w6d-io/ciops/api/v1alpha1"
	"github.com/w6d-io/ciops/internal/toolx"
	"github.com/w6d-io/x/logx"
)

func (s *ServiceAccount) CreateServiceAccount(ctx context.Context, e *v1alpha1.PipelineSource) error {
	log := logx.WithName(ctx, "sa.CreateServiceAccount").WithValues("ServiceAccount.CreateServiceAccount", e.Name, "namespace", e.Namespace)
	log.V(1).Info("build service account")
	projectID := fmt.Sprintf("p6e-cx-%d", e.Spec.ProjectID)
	m := metadata{
		name:      fmt.Sprintf("%s-%d", GitPrefixSecret, e.Spec.ProjectID),
		namespace: e.Namespace,
		labels: map[string]string{
			"projectID": projectID,
		},
	}

	s.Resources = append(s.Resources, m.build)
	return nil
}

func (m *metadata) build(ctx context.Context, r client.Client, e *v1alpha1.PipelineSource) error {
	log := logx.WithName(ctx, "secret.buildServiceAccount").WithValues("action", GitPrefixSecret)
	log.V(1).Info("creating")

	name := fmt.Sprintf("sa-%d", e.Spec.ProjectID)
	resource := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   e.Namespace,
			Annotations: make(map[string]string),
			Labels:      m.labels,
		},
	}
	resource.Annotations[v1alpha1.AnnotationSchedule] = time.Now().Format(time.RFC3339)
	if err := controllerutil.SetControllerReference(e, resource, r.Scheme()); err != nil {
		return err
	}

	op, err := controllerutil.CreateOrUpdate(ctx, r, resource, func() error {
		resource.Secrets = []corev1.ObjectReference{{Name: m.name}}
		return nil
	})
	log.V(2).Info(resource.Kind, "content", fmt.Sprintf("%v",
		toolx.GetObjectContain(resource)))
	if err != nil {
		log.Error(err, "create or update failed", "operation", op)
		return err
	}
	log.Info("resource successfully reconciled", "operation", op)
	return nil
}
