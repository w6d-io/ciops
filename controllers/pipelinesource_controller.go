/*
Copyright 2022 WILDCARD.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"github.com/google/uuid"
	tkn "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	civ1alpha1 "github.com/w6d-io/ciops/api/v1alpha1"
	"github.com/w6d-io/ciops/internal/namespaces"
	"github.com/w6d-io/ciops/internal/pipelines"
	"github.com/w6d-io/ciops/internal/tasks"
	"github.com/w6d-io/x/logx"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// PipelineSourceReconciler reconciles a PipelineSource object
type PipelineSourceReconciler struct {
	client.Client
	LocalScheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=ci.w6d.io,resources=pipelinesources,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ci.w6d.io,resources=pipelinesources/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ci.w6d.io,resources=pipelinesources/finalizers,verbs=update
//+kubebuilder:rbac:groups=tekton.dev,resources=pipelines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tekton.dev,resources=pipelines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tekton.dev,resources=pipelines/finalizers,verbs=update
//+kubebuilder:rbac:groups=tekton.dev,resources=tasks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tekton.dev,resources=tasks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tekton.dev,resources=tasks/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *PipelineSourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	correlationID := uuid.New().String()
	ctx = context.WithValue(ctx, logx.CorrelationID, correlationID)
	log := logx.WithName(ctx, "Reconcile").WithValues("pipelineSource", req.NamespacedName.String())
	var err error

	e := new(civ1alpha1.PipelineSource)
	if err = r.Get(ctx, req.NamespacedName, e); err != nil {
		if errors.IsNotFound(err) {
			log.Info("pipeline source resource not found, Ignore since object must be deleted")
			return ctrl.Result{
				Requeue: false,
			}, nil
		}
		log.Error(err, "failed to get pipeline source")
		return ctrl.Result{
			Requeue: true,
		}, err
	}
	namespace := namespaces.GetName(e.Spec.ProjectID)
	log = log.WithValues("namespace", namespace)
	// DoNamespace
	if err = namespaces.DoNamespace(ctx, r, e.Spec.ProjectID); err != nil {
		log.Error(err, "failed to make namespace", "action", "requeue")
		return ctrl.Result{Requeue: true}, err
	}
	t := tasks.Tasks{}
	if err = t.Parse(ctx, r, e); err != nil {
		log.Error(err, "failed on tasks,conditions parsing/creating", "action", "requeue")
		return ctrl.Result{Requeue: true}, err
	}
	if err = pipelines.Parse(ctx, r, t.Tasks, e); err != nil {
		log.Error(err, "failed on pipeline parsing/creating", "action", "requeue")
		return ctrl.Result{Requeue: true}, err
	}
	log.Info("resources successfully reconciled")
	return ctrl.Result{
		Requeue: false,
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PipelineSourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&civ1alpha1.PipelineSource{}).
		Owns(&tkn.Pipeline{}).
		Owns(&tkn.Task{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 10,
		}).
		Complete(r)
}

func (r *PipelineSourceReconciler) UpdateStatus(ctx context.Context, nn types.NamespacedName, status civ1alpha1.PipelineSourceStatus) error {
	log := logx.WithName(ctx, "FactReconciler.UpdateStatus").WithValues("resource", nn, "status", status)
	log.V(1).Info("update status")
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		e := &civ1alpha1.PipelineSource{}
		if err := r.Get(ctx, nn, e); err != nil {
			return err
		}
		e.Status.State = status.State
		e.Status.Message = status.Message
		if status.PipelineName != "" {
			e.Status.PipelineName = status.PipelineName
		}
		meta.SetStatusCondition(&e.Status.Conditions, metav1.Condition{
			Type:    string(status.State),
			Status:  status.State,
			Reason:  "",
			Message: status.Message,
		})
		if err := r.Status().Update(ctx, e); err != nil {
			log.Error(err, "update status failed")
		}
		return nil
	})
	return err
}

func (r *PipelineSourceReconciler) Scheme() *runtime.Scheme {
	return r.LocalScheme
}
