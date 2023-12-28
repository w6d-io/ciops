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
	"knative.dev/pkg/apis"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/w6d-io/x/logx"
)

// log is for logging in this package.
var factlog = logx.WithName(context.Background(), "fact-resource")

func (in *Fact) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(in).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-ci-w6d-io-v1alpha1-fact,mutating=true,failurePolicy=fail,sideEffects=None,groups=ci.w6d.io,resources=facts,verbs=create;update,versions=v1alpha1,name=mfact.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Fact{}
var _ apis.Defaultable = &Fact{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *Fact) Default() {
	factlog.Info("default", "name", in.Name)

	// TODO(user): fill in your defaulting logic.
	// nothing to do
}

func (in *Fact) SetDefaults(_ context.Context) {
	//TODO implement me
	// nothing to do
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-ci-w6d-io-v1alpha1-fact,mutating=false,failurePolicy=fail,sideEffects=None,groups=ci.w6d.io,resources=facts,verbs=create;update,versions=v1alpha1,name=vfact.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Fact{}
var _ apis.Validatable = &Fact{}

func (in *Fact) Validate(_ context.Context) *apis.FieldError {
	_, errs := ValidateFact(in.Name, in.Spec)
	return errs
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *Fact) ValidateCreate() (admission.Warnings, error) {
	factlog.Info("validate create", "name", in.Name)

	return ValidateFact(in.Name, in.Spec)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *Fact) ValidateUpdate(_ runtime.Object) (admission.Warnings, error) {
	factlog.Info("validate update", "name", in.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *Fact) ValidateDelete() (admission.Warnings, error) {
	factlog.Info("validate delete", "name", in.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
