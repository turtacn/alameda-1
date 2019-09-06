/*
Copyright 2019 The Alameda Authors.

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
	"fmt"
	"strings"

	"github.com/containers-ai/alameda/notifier/event"
	"github.com/containers-ai/alameda/pkg/utils"
	"github.com/containers-ai/alameda/pkg/utils/log"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var topicWhScope = log.RegisterScope("topic_webhook_logic", "topic webhook logic", 0)

func (r *AlamedaNotificationTopic) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-notifying-containers-ai-v1alpha1-alamedanotificationtopic,mutating=true,failurePolicy=fail,groups=notifying.containers.ai,resources=alamedanotificationtopics,verbs=create;update,versions=v1alpha1,name=malamedanotificationtopic.containers.ai

var _ webhook.Defaulter = &AlamedaNotificationTopic{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *AlamedaNotificationTopic) Default() {
}

// +kubebuilder:webhook:path=/validate-notifying-containers-ai-v1alpha1-alamedanotificationtopic,mutating=false,failurePolicy=fail,groups=notifying.containers.ai,resources=alamedanotificationtopics,verbs=create;update,versions=v1alpha1,name=valamedanotificationtopic.containers.ai

var _ webhook.Validator = &AlamedaNotificationTopic{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *AlamedaNotificationTopic) ValidateCreate() error {
	return r.validateAlamedaNotificationTopic()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *AlamedaNotificationTopic) ValidateDelete() error {
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *AlamedaNotificationTopic) ValidateUpdate(old runtime.Object) error {
	return r.validateAlamedaNotificationTopic()
}

func (r *AlamedaNotificationTopic) validateAlamedaNotificationTopic() error {
	var allErrs field.ErrorList
	for emailIdx, email := range r.Spec.Channel.Emails {
		for itoIdx, ito := range email.To {
			if ito != "" && !utils.IsEmailValid(ito) {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("channel").Child("emails").
					Index(emailIdx).Child("to").Index(itoIdx),
					ito, fmt.Sprintf("invalid emails %s found",
						ito)))
			}
		}
		for iccIdx, icc := range email.Cc {
			if icc != "" && !utils.IsEmailValid(icc) {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("channel").Child("emails").
					Index(emailIdx).Child("cc").Index(iccIdx),
					icc, fmt.Sprintf("invalid emails %s found",
						icc)))
			}
		}
	}

	for topicIdx, topic := range r.Spec.Topics {
		for typeIdx, iType := range topic.Type {
			if iType != "" && !event.IsEventTypeYamlKeySupported(iType) {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("topics").
					Index(topicIdx).Child("type").Index(typeIdx),
					iType, fmt.Sprintf("topic type %s is not in support list (%s)",
						iType, strings.Join(event.ListEventTypeYamlKey(), ","))))
			}
		}
		for levelIdx, level := range topic.Level {
			if level != "" && !event.IsEventLevelYamlKeySupported(level) {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("topics").
					Index(topicIdx).Child("level").Index(levelIdx),
					level, fmt.Sprintf("topic level %s is not in support list (%s)",
						level, strings.Join(event.ListEventLevelYamlKey(), ","))))
			}
		}
	}

	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(
		schema.GroupKind{Group: "notifying.containers.ai", Kind: "AlamedaNotificationTopic"},
		r.Name, allErrs)
}
