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

package v1alpha1

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/w6d-io/x/logx"
)

// log is for logging in this package.
var eventlog = logx.WithName(context.Background(), "event-resource")

func (in *Event) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(in).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-ci-w6d-io-v1alpha1-event,mutating=true,failurePolicy=fail,sideEffects=None,groups=ci.w6d.io,resources=events,verbs=create;update,versions=v1alpha1,name=mevent.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Event{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *Event) Default() {
	eventlog.Info("default", "name", in.Name)

	// TODO(user): fill in your defaulting logic.
	// nothing to do
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-ci-w6d-io-v1alpha1-event,mutating=false,failurePolicy=fail,sideEffects=None,groups=ci.w6d.io,resources=events,verbs=create;update,versions=v1alpha1,name=vevent.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Event{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *Event) ValidateCreate() error {
	eventlog.Info("validate create", "name", in.Name)

	return ValidateEvent(in.Name, in.Spec)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *Event) ValidateUpdate(old runtime.Object) error {
	eventlog.Info("validate update", "name", in.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *Event) ValidateDelete() error {
	eventlog.Info("validate delete", "name", in.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
