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
	"github.com/containers-ai/alameda/pkg/utils/log"
	"k8s.io/apimachinery/pkg/runtime"
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
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *AlamedaNotificationTopic) ValidateUpdate(old runtime.Object) error {
	return nil
}
