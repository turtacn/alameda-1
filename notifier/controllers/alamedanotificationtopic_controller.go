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

package controllers

import (
	"context"

	logUtil "github.com/containers-ai/alameda/pkg/utils/log"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	notifyingv1alpha1 "github.com/containers-ai/alameda/notifier/api/v1alpha1"
)

var (
	topicScope = logUtil.RegisterScope("alamedanotificationtopic_controller", "alamedanotificationtopic controller", 0)
)

// AlamedaNotificationTopicReconciler reconciles a AlamedaNotificationTopic object
type AlamedaNotificationTopicReconciler struct {
	client.Client
}

// +kubebuilder:rbac:groups=notifying.containers.ai,resources=alamedanotificationtopics,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=notifying.containers.ai,resources=alamedanotificationtopics/status,verbs=get;update;patch

func (r *AlamedaNotificationTopicReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	return ctrl.Result{}, nil
}

func (r *AlamedaNotificationTopicReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&notifyingv1alpha1.AlamedaNotificationTopic{}).
		Complete(r)
}
