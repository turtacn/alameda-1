package event

import (
	"time"

	"github.com/containers-ai/alameda/pkg/utils"
	datahub_events "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/events"
	"github.com/golang/protobuf/ptypes/timestamp"
)

func GetEmailNotificationEvent(msg, podName, clusterId string) *datahub_events.Event {
	return &datahub_events.Event{
		Time: &timestamp.Timestamp{
			Seconds: time.Now().Unix(),
		},
		ClusterId: clusterId,
		Type:      datahub_events.EventType_EVENT_TYPE_EMAIL_NOTIFICATION,
		Version:   datahub_events.EventVersion_EVENT_VERSION_V1,
		Level:     datahub_events.EventLevel_EVENT_LEVEL_WARNING,
		Subject: &datahub_events.K8SObjectReference{
			Kind:      "Pod",
			Namespace: utils.GetRunningNamespace(),
			Name:      podName,
		},
		Message: msg,
	}
}
