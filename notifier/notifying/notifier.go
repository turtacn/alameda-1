package notifying

import (
	"context"
	"os"

	notifyingv1alpha1 "github.com/containers-ai/alameda/notifier/api/v1alpha1"
	"github.com/containers-ai/alameda/notifier/channel"
	"github.com/containers-ai/alameda/notifier/event"

	"github.com/containers-ai/alameda/pkg/utils/log"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

var scope = log.RegisterScope("notifier", "notifier", 0)

type notifier struct {
	k8sClient     client.Client
	datahubClient datahub_v1alpha1.DatahubServiceClient
}

func NewNotifier(k8sClient client.Client,
	datahubClient datahub_v1alpha1.DatahubServiceClient) *notifier {
	return &notifier{
		k8sClient:     k8sClient,
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

func (notifier *notifier) sendEvtBaseOnTopic(evt *datahub_v1alpha1.Event, notificationTopic *notifyingv1alpha1.AlamedaNotificationTopic) {
	if notificationTopic.Spec.Disabled {
		return
	}

	toSend := false
	for _, specTopic := range notificationTopic.Spec.Topics {
		subMatched := false
		for _, sub := range specTopic.Subject {
			if (sub.Namespace == "" || sub.Namespace == evt.Subject.Namespace) &&
				(sub.Name == "" || sub.Name == evt.Subject.Name) &&
				(sub.Kind == "" || sub.Kind == evt.Subject.Kind) &&
				(sub.APIVersion == "" || sub.APIVersion == evt.Subject.ApiVersion) {
				subMatched = true
				break
			}
		}
		typeMatched := false
		for _, ty := range specTopic.Type {
			if ty == "" || event.EventTypeYamlKeyToIntMap(ty) == int32(evt.Type) {
				typeMatched = true
				break
			}
		}
		lvlMatched := false
		for _, lvl := range specTopic.Level {
			if lvl == "" || event.EventLevelYamlKeyToIntMap(lvl) == int32(evt.Level) {
				lvlMatched = true
				break
			}
		}
		srcMatched := false
		for _, src := range specTopic.Source {
			if (src.Host == "" || src.Host == evt.Source.Host) &&
				(src.Component == "" || src.Component == evt.Source.Component) {
				srcMatched = true
			}
		}
		if subMatched && typeMatched && lvlMatched && srcMatched {
			toSend = true
			break
		}
	}
	if !toSend {
		return
	}

	for _, emailChannel := range notificationTopic.Spec.Channel.Emails {
		notifier.sendEvtByEmails(evt, emailChannel)
	}
}

func (notifier *notifier) sendEvtByEmails(evt *datahub_v1alpha1.Event, emailChannel *notifyingv1alpha1.AlamedaEmailChannel) {
	alamedaNotificationChannel := &notifyingv1alpha1.AlamedaNotificationChannel{}
	err := notifier.k8sClient.Get(context.TODO(), client.ObjectKey{
		Name: emailChannel.Name,
	}, alamedaNotificationChannel)

	if err != nil {
		scope.Errorf(err.Error())
		evtSender := event.NewEventSender(notifier.datahubClient)
		podName, hostErr := os.Hostname()
		if hostErr != nil {
			scope.Errorf(err.Error())
		}
		evtSender.SendEvents([]*datahub_v1alpha1.Event{
			event.GetEmailNotificationEvent(err.Error(), podName),
		})
		return
	}
	emailNotificationChannel, err := channel.NewEmailClient(alamedaNotificationChannel, emailChannel)
	if err != nil {
		scope.Errorf(err.Error())
		evtSender := event.NewEventSender(notifier.datahubClient)
		podName, hostErr := os.Hostname()
		if hostErr != nil {
			scope.Errorf(err.Error())
		}
		evtSender.SendEvents([]*datahub_v1alpha1.Event{
			event.GetEmailNotificationEvent(err.Error(), podName),
		})
		return
	}
	emailNotificationChannel.SendEvent(evt)
}
