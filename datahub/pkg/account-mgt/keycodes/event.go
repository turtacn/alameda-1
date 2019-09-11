package keycodes

import (
	"fmt"
	EventMgt "github.com/containers-ai/alameda/internal/pkg/event-mgt"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	K8SUtils "github.com/containers-ai/alameda/pkg/utils/kubernetes"
	DatahubV1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes/timestamp"
	"time"
)

func NewKeycodeEvent(level DatahubV1alpha1.EventLevel, message string) *DatahubV1alpha1.Event {
	namespace := K8SUtils.GetRunningNamespace()

	clusterId, err := K8SUtils.GetClusterUID(K8SClient)
	if err != nil {
		scope.Errorf("failed to get cluster id: %s", err.Error())
	}

	source := &DatahubV1alpha1.EventSource{
		Host:      "",
		Component: fmt.Sprintf("%s-datahub", namespace),
	}

	subject := &DatahubV1alpha1.K8SObjectReference{
		Kind:       "Pod",
		Namespace:  namespace,
		Name:       "Federator.ai",
		ApiVersion: "v1",
	}

	event := &DatahubV1alpha1.Event{
		Time:      &timestamp.Timestamp{Seconds: time.Now().Unix()},
		Id:        AlamedaUtils.GenerateUUID(),
		ClusterId: clusterId,
		Source:    source,
		Type:      DatahubV1alpha1.EventType_EVENT_TYPE_LICENSE,
		Version:   DatahubV1alpha1.EventVersion_EVENT_VERSION_V1,
		Level:     level,
		Subject:   subject,
		Message:   message,
	}

	return event
}

func PostEvent(level DatahubV1alpha1.EventLevel, message string) error {
	if level == DatahubV1alpha1.EventLevel_EVENT_LEVEL_INFO {
		scope.Info(message)
	} else {
		scope.Error(message)
	}

	request := &DatahubV1alpha1.CreateEventsRequest{}
	request.Events = append(request.Events, NewKeycodeEvent(level, message))

	return EventMgt.PostEvents(request)
}
