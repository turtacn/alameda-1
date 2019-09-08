package event

import (
	"time"

	"github.com/containers-ai/alameda/pkg/utils"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes/timestamp"
)

func GetEmailNotificationEvent(msg, podName string) *datahub_v1alpha1.Event {
	return &datahub_v1alpha1.Event{
		Time: &timestamp.Timestamp{
			Seconds: time.Now().Unix(),
		},
		Type:    datahub_v1alpha1.EventType_EVENT_TYPE_EMAIL_NOTIFICATION,
		Version: datahub_v1alpha1.EventVersion_EVENT_VERSION_V1,
		Level:   datahub_v1alpha1.EventLevel_EVENT_LEVEL_WARNING,
		Subject: &datahub_v1alpha1.K8SObjectReference{
			Kind:      "Pod",
			Namespace: utils.GetRunningNamespace(),
			Name:      podName,
		},
		Message: msg,
	}
}
