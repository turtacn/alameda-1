package notifying

import (
	"context"
	"fmt"
	"strings"
	"time"

	notifyingv1alpha1 "github.com/containers-ai/alameda/notifier/api/v1alpha1"
	"github.com/containers-ai/alameda/notifier/channel"
	"github.com/containers-ai/alameda/notifier/event"
	notifier_utils "github.com/containers-ai/alameda/notifier/utils"
	"github.com/containers-ai/alameda/pkg/utils"
	"github.com/containers-ai/alameda/pkg/utils/log"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/spf13/viper"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var scope = log.RegisterScope("notifier", "notifier", 0)

type notifier struct {
	k8sClient     client.Client
	datahubClient datahub_v1alpha1.DatahubServiceClient
	clusterId     string
}

func NewNotifier(mgr manager.Manager,
	datahubClient datahub_v1alpha1.DatahubServiceClient) *notifier {

	clusterId, err := notifier_utils.GetClusterUID(mgr.GetClient())
	if err != nil {
		scope.Errorf("Get cluster id failed: %s", err.Error())
	}
	return &notifier{
		k8sClient:     mgr.GetClient(),
		clusterId:     clusterId,
		datahubClient: datahubClient,
	}
}

func (notifier *notifier) NotifyEvents(evts []*datahub_v1alpha1.Event) {
	alamedaNotificationTopicList := &notifyingv1alpha1.AlamedaNotificationTopicList{}
	err := notifier.k8sClient.List(context.TODO(), alamedaNotificationTopicList)
	if err != nil {
		scope.Errorf(err.Error())
		return
	}

	for _, topic := range alamedaNotificationTopicList.Items {
		for _, evt := range evts {
			notifier.sendEvtBaseOnTopic(evt, &topic)
		}
	}
}

func (notifier *notifier) sendEvtBaseOnTopic(evt *datahub_v1alpha1.Event,
	notificationTopic *notifyingv1alpha1.AlamedaNotificationTopic) {
	if notificationTopic.Spec.Disabled {
		return
	}

	toSend := false
	for specTopicIdx, specTopic := range notificationTopic.Spec.Topics {
		subMatched := (specTopic.Subject == nil || len(specTopic.Subject) == 0)
		for _, sub := range specTopic.Subject {
			if (sub.Namespace == "" || sub.Namespace == evt.Subject.Namespace) &&
				(sub.Name == "" || sub.Name == evt.Subject.Name) &&
				(sub.Kind == "" || sub.Kind == evt.Subject.Kind) &&
				(sub.APIVersion == "" || sub.APIVersion == evt.Subject.ApiVersion) {
				subMatched = true
				break
			}
		}
		typeMatched := (specTopic.Type == nil || len(specTopic.Type) == 0)
		for _, ty := range specTopic.Type {
			if ty == "" || event.EventTypeYamlKeyToIntMap(ty) == int32(evt.Type) {
				typeMatched = true
				break
			}
		}
		lvlMatched := (specTopic.Level == nil || len(specTopic.Level) == 0)
		for _, lvl := range specTopic.Level {
			if lvl == "" || event.EventLevelYamlKeyToIntMap(lvl) == int32(evt.Level) {
				lvlMatched = true
				break
			}
		}
		srcMatched := (specTopic.Source == nil || len(specTopic.Source) == 0)
		for _, src := range specTopic.Source {
			if (src.Host == "" || src.Host == evt.Source.Host) &&
				(src.Component == "" || src.Component == evt.Source.Component) {
				srcMatched = true
			}
		}

		scope.Debugf("topic %s (%d/%d) subject matched: %t, type matched: %t, level matched: %t, source matched: %t",
			notificationTopic.Name, specTopicIdx+1, len(notificationTopic.Spec.Topics),
			subMatched, typeMatched, lvlMatched, srcMatched)
		if subMatched && typeMatched && lvlMatched && srcMatched {
			toSend = true
			break
		}
	}

	if !toSend {
		return
	}

	channelConditions := []*notifyingv1alpha1.AlamedaChannelCondition{}
	for _, emailChannel := range notificationTopic.Spec.Channel.Emails {
		err := notifier.sendEvtByEmails(evt, emailChannel)
		channelCondition := &notifyingv1alpha1.AlamedaChannelCondition{
			Type:    "email",
			Name:    emailChannel.Name,
			Success: err == nil,
			Time:    time.Now().Format(time.RFC3339),
		}

		if err != nil {
			channelCondition.Message = fmt.Sprintf(
				"topic %s failed to send message with email channel %s. %s",
				notificationTopic.Name, emailChannel.Name, err.Error())
		}
		channelConditions = append(channelConditions, channelCondition)
	}

	topicEventResendTime := viper.GetInt64("topicEventResendTime")
	errMsg := ""
	toSendEvt := false
	for _, newCd := range channelConditions {
		isNewErrCd := true
		for _, oldCd := range notificationTopic.Status.ChannelCondictions {
			if oldCd.Type == newCd.Type && oldCd.Name == newCd.Name && oldCd.Time != "" {
				oldTimeSec, oldErr := time.Parse(time.RFC3339, oldCd.Time)
				newTimeSec, newErr := time.Parse(time.RFC3339, newCd.Time)
				if oldErr == nil && newErr == nil && oldCd.Message == newCd.Message &&
					newTimeSec.Unix()-oldTimeSec.Unix() < topicEventResendTime &&
					!newCd.Success {
					isNewErrCd = false
					break
				}
			}
		}
		if isNewErrCd && !newCd.Success {
			toSendEvt = true
			errMsg = fmt.Sprintf("%s %s.", errMsg, newCd.Message)
		}
	}

	errMsg = strings.Trim(errMsg, " ")
	if toSendEvt {
		evtSender := event.NewEventSender(notifier.datahubClient)
		podName := utils.GetRunningPodName()
		evtSender.SendEvents([]*datahub_v1alpha1.Event{
			event.GetEmailNotificationEvent(errMsg, podName, notifier.clusterId),
		})
	}

	latestNotificationTopic := &notifyingv1alpha1.AlamedaNotificationTopic{}
	getErr := notifier.k8sClient.Get(context.Background(), client.ObjectKey{
		Name: notificationTopic.GetName(),
	}, latestNotificationTopic)
	if getErr == nil {
		latestNotificationTopic.Status.ChannelCondictions = channelConditions
		if updateErr := notifier.k8sClient.Update(context.Background(),
			latestNotificationTopic); updateErr != nil {
			scope.Errorf("update topic %s condition status failed: %s",
				latestNotificationTopic.GetName(), updateErr.Error())
		}
	} else {
		scope.Errorf("get topic %s to update condition status failed: %s",
			notificationTopic.GetName(), getErr.Error())
	}
}

func (notifier *notifier) sendEvtByEmails(evt *datahub_v1alpha1.Event,
	emailChannel *notifyingv1alpha1.AlamedaEmailChannel) error {
	alamedaNotificationChannel := &notifyingv1alpha1.AlamedaNotificationChannel{}
	err := notifier.k8sClient.Get(context.TODO(), client.ObjectKey{
		Name: emailChannel.Name,
	}, alamedaNotificationChannel)

	if err != nil {
		return err
	}
	emailNotificationChannel, err := channel.NewEmailClient(
		alamedaNotificationChannel, emailChannel)
	if err != nil {
		return err
	}
	return emailNotificationChannel.SendEvent(evt)
}
