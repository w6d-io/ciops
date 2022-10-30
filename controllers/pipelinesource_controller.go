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
    "github.com/w6d-io/x/logx"
    "k8s.io/apimachinery/pkg/api/errors"

    civ1alpha1 "github.com/w6d-io/ciops/api/v1alpha1"
    "k8s.io/apimachinery/pkg/runtime"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
)

// PipelineSourceReconciler reconciles a PipelineSource object
type PipelineSourceReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=ci.w6d.io,resources=pipelinesources,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ci.w6d.io,resources=pipelinesources/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ci.w6d.io,resources=pipelinesources/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PipelineSource object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *PipelineSourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    correlationID := uuid.New().String()
    ctx = context.WithValue(ctx, logx.CorrelationID, correlationID)
    log := logx.WithName(ctx, "Reconcile").WithValues("pipelineSource", req.NamespacedName.String())
    var err error

    e := new(civ1alpha1.PipelineSource)
    if err = r.Get(ctx, req.NamespacedName, e); err != nil {
        if errors.IsNotFound(err) {
            log.Info("pipeline source resource not found, Ignore since object must be deleted")
            return ctrl.Result{}, nil
        }
        log.Error(err, "failed to get pipeline source")
        return ctrl.Result{}, err
    }
    return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PipelineSourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&civ1alpha1.PipelineSource{}).
        Complete(r)
}
