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
	pipelinev1alpha1 "github.com/w6d-io/apis/pipeline/v1alpha1"
	"strings"

	"github.com/google/uuid"
	tkn "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/w6d-io/ciops/api/v1alpha1"
	"github.com/w6d-io/ciops/internal/pipelineruns"
	"github.com/w6d-io/x/logx"
)

// FactReconciler reconciles a Fact object
type FactReconciler struct {
	client.Client
	LocalScheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=ci.w6d.io,resources=facts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ci.w6d.io,resources=facts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ci.w6d.io,resources=facts/finalizers,verbs=update
//+kubebuilder:rbac:groups=ci.w6d.io,resources=factbudgets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ci.w6d.io,resources=factbudgets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ci.w6d.io,resources=factbudgets/finalizers,verbs=update
//+kubebuilder:rbac:groups=tekton.dev,resources=pipelineruns,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tekton.dev,resources=pipelineruns/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tekton.dev,resources=pipelineruns/finalizers,verbs=update

func (r *FactReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	correlationID := uuid.New().String()
	ctx = context.WithValue(ctx, logx.CorrelationID, correlationID)
	log := logx.WithName(ctx, "Reconcile").WithValues("fact", req.NamespacedName.String())
	var err error

	obj := new(v1alpha1.Fact)
	if err = r.Get(ctx, req.NamespacedName, obj); err != nil {
		if errors.IsNotFound(err) {
			log.Info("Fact resource not found, Ignore since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "failed to get Fact")
		return ctrl.Result{}, err
	}
	status := v1alpha1.FactStatus{PipelineRunName: pipelineruns.GetPipelinerunName(*obj.Spec.EventID)}
	var childPr tkn.PipelineRun
	err = r.Get(ctx, types.NamespacedName{Namespace: req.Namespace, Name: status.PipelineRunName}, &childPr)
	if client.IgnoreNotFound(err) != nil {
		log.Error(err, "Unable to get PipelineRun")
		return ctrl.Result{}, err
	}
	if !errors.IsNotFound(err) {
		status.State = pipelineruns.Condition(childPr.Status.Conditions)
		status.Message = pipelineruns.Message(childPr.Status.Conditions)
		if err := r.UpdateStatus(ctx, req.NamespacedName, status); err != nil {
			log.Error(err, "unable to update Play status")
			return ctrl.Result{Requeue: true}, err
		}
		return ctrl.Result{Requeue: false}, nil
	}
	log.V(1).Info("pipelinerun not found")
	log.V(1).Info("getting all pipeline run")

	if err = r.checkConcurrency(ctx, req.NamespacedName, pipelineruns.GetPipelinerunName(*obj.Spec.EventID)); err != nil {
		return ctrl.Result{Requeue: true}, err
	}

	if err = pipelineruns.Build(ctx, r, obj); err != nil {
		log.Error(err, "failed to create pipelinerun")
		log.V(1).Info("update status", "status", v1alpha1.Errored,
			"step", "5")
		status.State = v1alpha1.Errored
		status.Message = err.Error()
		if err := r.UpdateStatus(ctx, req.NamespacedName, status); err != nil {
			return ctrl.Result{Requeue: true}, err
		}
		return ctrl.Result{Requeue: true}, err
	}
	log.V(1).Info("update status", "status", pipelinev1alpha1.Pending.ToString(), "step", "6")
	status.State = v1alpha1.State(pipelinev1alpha1.Pending.ToString())
	if err = r.UpdateStatus(ctx, req.NamespacedName, status); err != nil {
		log.Error(err, "update status failed")
		return ctrl.Result{Requeue: true}, err
	}
	return ctrl.Result{Requeue: false}, nil
}

func (r *FactReconciler) GetStatus(state v1alpha1.State) metav1.ConditionStatus {
	switch state {
	case v1alpha1.Errored, v1alpha1.Cancelled, v1alpha1.Failed:
		return metav1.ConditionFalse
	case v1alpha1.Succeeded:
		return metav1.ConditionTrue
	default:
		return metav1.ConditionUnknown
	}
}

func (r *FactReconciler) UpdateStatus(ctx context.Context, nn types.NamespacedName, status v1alpha1.FactStatus) error {
	log := logx.WithName(ctx, "FactReconciler.UpdateStatus").WithValues("resource", nn, "status", status)
	log.V(1).Info("update status")
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		e := &v1alpha1.Fact{}
		if err := r.Get(ctx, nn, e); err != nil {
			return err
		}
		e.Status.State = status.State
		e.Status.Message = status.Message
		e.Status.PipelineRunName = status.PipelineRunName
		meta.SetStatusCondition(&e.Status.Conditions, metav1.Condition{
			Type:    string(status.State),
			Status:  r.GetStatus(status.State),
			Reason:  string(status.State),
			Message: status.Message,
		})
		if err := r.Status().Update(ctx, e); err != nil {
			log.Error(err, "update status failed")
		}
		return nil
	})
	return err
}

func IgnoreNotExists(err error) error {
	if err == nil ||
		(strings.HasPrefix(err.Error(), "Index with name field:") &&
			strings.HasSuffix(err.Error(), "does not exist")) {
		return nil
	}
	return err
}

// SetupWithManager sets up the controller with the Manager.
func (r *FactReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Fact{}).
		Owns(&tkn.PipelineRun{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 10,
		}).
		Complete(r)
}

func (r *FactReconciler) Scheme() *runtime.Scheme {
	return r.LocalScheme
}
