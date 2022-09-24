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
    "gitlab.w6d.io/w6d/ciops/internal/pipelineruns"
    "k8s.io/client-go/util/retry"

    "github.com/google/uuid"
    tkn "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
    "k8s.io/apimachinery/pkg/api/errors"
    "k8s.io/apimachinery/pkg/runtime"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/controller"

    "github.com/w6d-io/x/logx"
    "gitlab.w6d.io/w6d/ciops/api/v1alpha1"
)

// EventReconciler reconciles a Event object
type EventReconciler struct {
    client.Client
    Scheme *runtime.Scheme
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

    e := &v1alpha1.Event{}
    if err = r.Get(ctx, req.NamespacedName, e); err != nil {
        if errors.IsNotFound(err) {
            log.Info("Event resource not found, Ignore since object must be deleted")
            return ctrl.Result{}, nil
        }
        log.Error(err, "failed to get Event")
        return ctrl.Result{}, err
    }
    if err = pipelineruns.Build(ctx, r, e); err != nil {
        return ctrl.Result{}, err
    }
    log.V(1).Info("update status")
    if err = r.UpdateStatus(ctx, mdb, sts); err != nil {
        log.Error(err, "update status failed")
        return ctrl.Result{Requeue: true}, err
    }
    return ctrl.Result{}, nil
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

func (r *EventReconciler) UpdateStatus(ctx context.Context, e *v1alpha1.Event) error {
    log := logx.WithName(ctx, "EventReconciler.UpdateStatus")
    var err error
    err = retry.RetryOnConflict(retry.DefaultRetry, func() error {

        return nil
    })
    return nil
}
