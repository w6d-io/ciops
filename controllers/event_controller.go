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
	"github.com/w6d-io/ciops/internal/pipelineruns"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"strings"

	"github.com/google/uuid"
	tkn "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/w6d-io/ciops/api/v1alpha1"
	"github.com/w6d-io/x/logx"
)

// EventReconciler reconciles a Event object
type EventReconciler struct {
    client.Client
    EventScheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=ci.w6d.io,resources=events,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ci.w6d.io,resources=events/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ci.w6d.io,resources=events/finalizers,verbs=update
//+kubebuilder:rbac:groups=v1beta1.tekton.dev,resources=pipelineruns,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=v1beta1.tekton.dev,resources=pipelineruns/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=v1beta1.tekton.dev,resources=pipelineruns/finalizers,verbs=update

func (r *EventReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    correlationID := uuid.New().String()
    ctx = context.WithValue(ctx, logx.CorrelationID, correlationID)
    log := logx.WithName(ctx, "Reconcile").WithValues("event", req.NamespacedName.String())
    var err error

    e := new(v1alpha1.Event)
    if err = r.Get(ctx, req.NamespacedName, e); err != nil {
        if errors.IsNotFound(err) {
            log.Info("Event resource not found, Ignore since object must be deleted")
            return ctrl.Result{}, nil
        }
        log.Error(err, "failed to get Event")
        return ctrl.Result{}, err
    }
    status := v1alpha1.EventStatus{PipelineRunName: pipelineruns.GetPipelinerunName(*e.Spec.EventID)}
    var childPr tkn.PipelineRun
    err = r.Get(ctx, types.NamespacedName{Namespace: req.Namespace}, &childPr)
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

    if err = r.checkConcurrency(ctx, req.NamespacedName, pipelineruns.GetPipelinerunName(*e.Spec.EventID)); err != nil {
        return ctrl.Result{Requeue: true}, err
    }

    if err = pipelineruns.Build(ctx, r, e); err != nil {
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
    log.V(1).Info("update status", "status", v1alpha1.Pending, "step", "6")
    status.State = v1alpha1.Pending
    if err = r.UpdateStatus(ctx, req.NamespacedName, status); err != nil {
        log.Error(err, "update status failed")
        return ctrl.Result{Requeue: true}, err
    }
    return ctrl.Result{Requeue: false}, nil
}

func (r *EventReconciler) GetStatus(state v1alpha1.State) metav1.ConditionStatus {
    switch state {
    case v1alpha1.Errored, v1alpha1.Cancelled, v1alpha1.Failed:
        return metav1.ConditionFalse
    case v1alpha1.Succeeded:
        return metav1.ConditionTrue
    default:
        return metav1.ConditionUnknown
    }
}

func (r *EventReconciler) UpdateStatus(ctx context.Context, nn types.NamespacedName, status v1alpha1.EventStatus) error {
    log := logx.WithName(ctx, "EventReconciler.UpdateStatus")
    log.V(1).Info("update status")
    err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
        e := &v1alpha1.Event{}
        err := r.Get(ctx, nn, e)
        if err != nil {
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
        err = r.Status().Update(ctx, e)
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
func (r *EventReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&v1alpha1.Event{}).
        Owns(&tkn.PipelineRun{}).
        WithOptions(controller.Options{
            MaxConcurrentReconciles: 10,
        }).
        Complete(r)
}

func (r *EventReconciler) Scheme() *runtime.Scheme {
    return r.EventScheme
}
