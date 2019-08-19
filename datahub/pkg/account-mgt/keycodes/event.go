package keycodes

import (
	EventMgt "github.com/containers-ai/alameda/internal/pkg/event-mgt"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	DatahubV1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes/timestamp"
	"time"
)

func NewKeycodeEvent(level DatahubV1alpha1.EventLevel, message string) *DatahubV1alpha1.Event {
	event := &DatahubV1alpha1.Event{
		Time:    &timestamp.Timestamp{Seconds: time.Now().Unix()},
		Id:      AlamedaUtils.GenerateUUID(),
		Version: DatahubV1alpha1.EventVersion_EVENT_VERSION_V1,
		Level:   level,
		Message: message,
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
