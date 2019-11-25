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
	"fmt"
	"time"

	notifyingv1alpha1 "github.com/containers-ai/alameda/notifier/api/v1alpha1"
	"github.com/containers-ai/alameda/notifier/channel"
	notifier_utils "github.com/containers-ai/alameda/notifier/utils"
	k8s_utils "github.com/containers-ai/alameda/pkg/utils/kubernetes"
	logUtil "github.com/containers-ai/alameda/pkg/utils/log"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	channelScope = logUtil.RegisterScope("alamedanotificationchannel_controller", "alamedanotificationchannel controller", 0)
)

// AlamedaNotificationChannelReconciler reconciles a AlamedaNotificationChannel object
type AlamedaNotificationChannelReconciler struct {
	client.Client
}

// +kubebuilder:rbac:groups=notifying.containers.ai,resources=alamedanotificationchannels,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=notifying.containers.ai,resources=alamedanotificationchannels/status,verbs=get;update;patch

func (r *AlamedaNotificationChannelReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	alamedaNotificationChannel := &notifyingv1alpha1.AlamedaNotificationChannel{}
	err := r.Get(ctx, req.NamespacedName, alamedaNotificationChannel)
	if err != nil {
		if !k8sapierrors.IsNotFound(err) {
			channelScope.Errorf(err.Error())
			return ctrl.Result{}, err
		}
	}

	if alamedaNotificationChannel.GetAnnotations()["notifying.containers.ai/test-channel"] == "start" {
		if alamedaNotificationChannel.Spec.Type == "email" {
			channelScope.Infof("start testing email channel %s", req.Name)
			err = r.testEmailChannel(alamedaNotificationChannel)
			if err != nil {
				alamedaNotificationChannel.Status.ChannelTest = &notifyingv1alpha1.AlamedaChannelTest{
					Success: false,
					Time:    time.Now().Format(time.RFC3339),
					Message: err.Error(),
				}
				channelScope.Errorf("test email channel %s failed: %s", req.Name, err.Error())

				annotations := alamedaNotificationChannel.GetAnnotations()
				annotations["notifying.containers.ai/test-channel"] = "done"
				alamedaNotificationChannel.SetAnnotations(annotations)

				if updateErr := r.Update(ctx, alamedaNotificationChannel); updateErr != nil {
					channelScope.Errorf("update test annotation and status for email channel %s failed: %s",
						req.Name, err.Error())
					return ctrl.Result{}, updateErr
				}
				return ctrl.Result{}, err
			}

			channelScope.Infof("test email channel %s successful", req.Name)
			annotations := alamedaNotificationChannel.GetAnnotations()
			annotations["notifying.containers.ai/test-channel"] = "done"
			alamedaNotificationChannel.SetAnnotations(annotations)
			alamedaNotificationChannel.Status.ChannelTest = &notifyingv1alpha1.AlamedaChannelTest{
				Success: true,
				Time:    time.Now().Format(time.RFC3339),
			}
			if updateErr := r.Update(ctx, alamedaNotificationChannel); updateErr != nil {
				channelScope.Errorf("update test annotation and status for email channel %s failed: %s",
					req.Name, err.Error())
				return ctrl.Result{}, updateErr
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *AlamedaNotificationChannelReconciler) testEmailChannel(
	alamedaNotificationChannel *notifyingv1alpha1.AlamedaNotificationChannel) error {
	annotations := alamedaNotificationChannel.GetAnnotations()
	from := alamedaNotificationChannel.Spec.Email.From
	to, ok := annotations["notifying.containers.ai/test-channel-to"]
	if !ok || to == "" {
		errMsg := fmt.Sprintf("no correct annotation \"notifying.containers.ai/test-channel-to\" set to test email channel %s",
			alamedaNotificationChannel.GetName())
		return fmt.Errorf(errMsg)
	}

	clusterInfo, err := notifier_utils.GetClusterInfo(r.Client)
	if err != nil {
		channelScope.Errorf("unable to send test email due to get cluster info fail: %s", err.Error())
		return err
	}
	uid, err := k8s_utils.GetClusterUID(r.Client)
	if err != nil {
		channelScope.Errorf("unable to send test email due to get cluster id fail: %s", err.Error())
		return err
	}
	clusterInfo.UID = uid

	emailChannel := &notifyingv1alpha1.AlamedaEmailChannel{}
	emailClient, err := channel.NewEmailClient(alamedaNotificationChannel, emailChannel, &clusterInfo)
	if err != nil {
		return err
	}
	subject := "Test Email"
	recipients := []string{to}
	ccs := []string{}
	attachments := map[string]string{}
	msgHTML := fmt.Sprintf(`
		<html>
			<body>
			This is a test email for Federator.ai email notification.<br>
			###############################################################<br>
			<table cellspacing="5" cellpadding="0">
				<tr><td align="left">Cluster Id:</td><td>%s</td></tr>
				<tr><td align="left">Master Node Hostname:</td><td>%s</td></tr>
				<tr><td align="left">Master Node IP:</td><td>%s</td></tr>
				<tr><td align="left">Time:</td><td>%s</td></tr>
			</table>
		</body>
	</html>
	`, clusterInfo.UID, clusterInfo.MasterNodeHostname, clusterInfo.MasterNodeIP, time.Now().Format(time.RFC3339))
	err = emailClient.SendEmailBySMTP(subject, from, recipients,
		msgHTML, notifier_utils.RemoveEmptyStr(ccs), attachments)
	if err != nil {
		return err
	}
	return nil
}

func (r *AlamedaNotificationChannelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&notifyingv1alpha1.AlamedaNotificationChannel{}).
		Complete(r)
}
