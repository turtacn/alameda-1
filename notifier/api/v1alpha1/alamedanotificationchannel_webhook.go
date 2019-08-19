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
	"context"

	b64 "encoding/base64"

	"github.com/containers-ai/alameda/pkg/utils/log"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var channelWhScope = log.RegisterScope("channel_webhook_logic", "channel webhook logic", 0)

func (r *AlamedaNotificationChannel) SetupWebhookWithManager(mgr ctrl.Manager) error {
	r.mgr = mgr
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-notifying-containers-ai-v1alpha1-alamedanotificationchannel,mutating=true,failurePolicy=fail,groups=notifying.containers.ai,resources=alamedanotificationchannels,verbs=create;update,versions=v1alpha1,name=malamedanotificationchannel.containers.ai

var _ webhook.Defaulter = &AlamedaNotificationChannel{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *AlamedaNotificationChannel) Default() {
	channelWhScope.Infof("default webhook for channel: %s", r.Name)

	if r.Spec.Email.Encryption == "" {
		r.Spec.Email.Encryption = "tls"
	}

	k8sClnt := r.mgr.GetClient()
	oldChannel := &AlamedaNotificationChannel{}
	k8sClnt.Get(context.TODO(), client.ObjectKey{
		Namespace: r.GetNamespace(),
		Name:      r.GetName(),
	}, oldChannel)

	// if username or password modified, encode them again
	if oldChannel.Spec.Email.Username != r.Spec.Email.Username {
		r.Spec.Email.Username = b64.StdEncoding.EncodeToString([]byte(r.Spec.Email.Username))
	}
	if oldChannel.Spec.Email.Password != r.Spec.Email.Password {
		r.Spec.Email.Password = b64.StdEncoding.EncodeToString([]byte(r.Spec.Email.Password))
	}

	annotations := r.GetAnnotations()
	testVal, ok := annotations["notifying.containers.ai/test-channel"]
	if !ok || testVal != "start" {
		annotations["notifying.containers.ai/test-channel"] = "done"
		r.SetAnnotations(annotations)
	}
}

// +kubebuilder:webhook:path=/validate-notifying-containers-ai-v1alpha1-alamedanotificationchannel,mutating=false,failurePolicy=fail,groups=notifying.containers.ai,resources=alamedanotificationchannels,verbs=create;update,versions=v1alpha1,name=valamedanotificationchannel.containers.ai

var _ webhook.Validator = &AlamedaNotificationChannel{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *AlamedaNotificationChannel) ValidateCreate() error {
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *AlamedaNotificationChannel) ValidateDelete() error {
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *AlamedaNotificationChannel) ValidateUpdate(old runtime.Object) error {
	return nil
}
