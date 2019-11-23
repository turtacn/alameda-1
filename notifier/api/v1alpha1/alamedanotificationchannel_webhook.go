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
	"fmt"
	"strings"
	"time"

	"github.com/containers-ai/alameda/pkg/utils"
	"github.com/containers-ai/alameda/pkg/utils/log"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var channelWhScope = log.RegisterScope("channel_webhook_logic", "channel webhook logic", 0)

func (r *AlamedaNotificationChannel) SetupWebhookWithManager(mgr ctrl.Manager) error {
	r.Mgr = mgr
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-notifying-containers-ai-v1alpha1-alamedanotificationchannel,mutating=true,failurePolicy=fail,groups=notifying.containers.ai,resources=alamedanotificationchannels,verbs=create;update,versions=v1alpha1,name=malamedanotificationchannel.containers.ai

var _ webhook.Defaulter = &AlamedaNotificationChannel{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *AlamedaNotificationChannel) Default() {

	needUpdateAnnotation := false
	channelWhScope.Debugf("default webhook for channel: %s", r.Name)
	if r.Spec.Email.Encryption == "" {
		r.Spec.Email.Encryption = "tls"
	}

	k8sClnt := r.Mgr.GetClient()
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
	if !ok || testVal == "" {
		annotations["notifying.containers.ai/test-channel"] = "done"
		needUpdateAnnotation = true
	}
	_, ok = annotations["notifying.containers.ai/webhook-mutation"]
	if !ok {
		annotations["notifying.containers.ai/webhook-mutation"] = "ok"
		needUpdateAnnotation = true
		channelWhScope.Infof("Add annotation \"webhook-mutation\" for AlamedaNotificationChannel CR(%s)",
			r.GetName())
	}
	if needUpdateAnnotation {
		r.SetAnnotations(annotations)
	}
}

// +kubebuilder:webhook:path=/validate-notifying-containers-ai-v1alpha1-alamedanotificationchannel,mutating=false,failurePolicy=fail,groups=notifying.containers.ai,resources=alamedanotificationchannels,verbs=create;update,versions=v1alpha1,name=valamedanotificationchannel.containers.ai

var _ webhook.Validator = &AlamedaNotificationChannel{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *AlamedaNotificationChannel) ValidateCreate() error {
	return r.validateAlamedaNotificationChannel("create")
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *AlamedaNotificationChannel) ValidateDelete() error {
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *AlamedaNotificationChannel) ValidateUpdate(old runtime.Object) error {
	return r.validateAlamedaNotificationChannel("update")
}

func (r *AlamedaNotificationChannel) validateAlamedaNotificationChannel(op string) error {
	var allErrs field.ErrorList
	channelType := r.Spec.Type
	crName := r.GetName()

	channelWhScope.Debugf("validateAlamedaNotificationChannel enter, op=%s", op)
	//annotations := r.GetObjectMeta().GetAnnotations()
	annotations := r.GetAnnotations()

	if op != "create" {
		// do not update annotation when CR creation due to something in CR data may not ready yet
		// update CR may get error "resource name may not be empty"
		val, ok := annotations["notifying.containers.ai/webhook-validation"]
		channelWhScope.Debugf("validateAlamedaNotificationChannel, webhook-validation: %v, val:[%v]", ok, val)
		if !ok {
			channelWhScope.Infof("Add annotation \"webhook-validation\" for AlamedaNotificationChannel CR(%s)",
				crName)
			annotations["notifying.containers.ai/webhook-validation"] = "ok"
			r.SetAnnotations(annotations)
			err := r.updateAnnotationToCR()
			if err != nil {
				channelWhScope.Errorf("validateAlamedaNotificationChannel: failed to update CR(%s) annotations: %s",
					crName, err.Error())
			}
		} else {
			channelWhScope.Debugf("validateAlamedaNotificationChannel, webhook-validation is existing")
		}
	}

	if testChannel, ok := annotations["notifying.containers.ai/test-channel"]; ok {
		if strings.ToLower(testChannel) != "start" && strings.ToLower(testChannel) != "done" {
			allErrs = append(allErrs, field.Invalid(field.NewPath("metadata").Child("annotations").
				Child("notifying.containers.ai/test-channel"), testChannel,
				fmt.Sprintf("annotation notifying.containers.ai/test-channel does not support %s, please use start instead",
					testChannel)))
		}
	}

	if testChannelTo, ok := annotations["notifying.containers.ai/test-channel-to"]; ok {
		if testChannelTo != "" && !utils.IsEmailValid(testChannelTo) {
			allErrs = append(allErrs, field.Invalid(field.NewPath("metadata").Child("annotations").
				Child("notifying.containers.ai/test-channel-to"), testChannelTo,
				fmt.Sprintf("annotation notifying.containers.ai/test-channel-to value %s is not valid email format",
					testChannelTo)))
		}
	}

	if channelType != "" && channelType != "email" {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("type"),
			channelType, fmt.Sprintf("channel type %s is not supported, please email instead", channelType)))
	}
	if channelType == "email" {
		from := r.Spec.Email.From
		encryption := strings.ToLower(r.Spec.Email.Encryption)
		if from != "" && !utils.IsEmailValid(from) {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("email").Child("from"),
				from, fmt.Sprintf("format of email channel from %s is incorrect", from)))
		}
		if encryption != "" && encryption != "ssl" && encryption != "tls" && encryption != "starttls" {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("email").Child("encryption"),
				encryption, fmt.Sprintf("encryption %s is not supported, please use ssl, tls and starttls instead",
					r.Spec.Email.Encryption)))
		}
	}
	if len(allErrs) == 0 {
		return nil
	}
	channelWhScope.Debugf("validateAlamedaNotificationChannel error end: %v", allErrs)
	return apierrors.NewInvalid(
		schema.GroupKind{Group: "notifying.containers.ai", Kind: "AlamedaNotificationChannel"},
		r.Name, allErrs)
}

func (r *AlamedaNotificationChannel) updateAnnotationToCR() error {
	var err error
	retry := int(10)
	crName := r.GetName()
	crNamespace := r.GetNamespace()
	k8sclnt := r.Mgr.GetClient()
	channelWhScope.Debugf("UpdateAnnotationToCR, CR values: %#v", r)
	for i := 0; i < retry; i++ {
		channelWhScope.Debugf("  =>update CR(%s) %d", crName, i)
		currChannel := &AlamedaNotificationChannel{}
		k8sclnt.Get(context.TODO(), client.ObjectKey{
			Namespace: crNamespace,
			Name:      crName,
		}, currChannel)
		// update modified annotation to CR
		currChannel.Annotations = r.GetAnnotations()
		err = k8sclnt.Update(context.TODO(), currChannel)
		if err == nil {
			break
		}
		channelWhScope.Debugf("Failed to update AlamedaNotificationChannel CR(%s) annotation (retry: %d): %s",
			crName, i, err.Error())
		time.Sleep(100 * time.Millisecond)
	}
	if err == nil {
		channelWhScope.Infof("Update AlamedaNotificationChannel CR(%s) annotation successfully", crName)
	} else {
		channelWhScope.Errorf("Failed to update AlamedaNotificationChannel CR(%s) annotation: %s",
			crName, err.Error())
	}
	return err
}
