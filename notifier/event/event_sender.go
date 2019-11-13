package event

import (
	"context"

	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	datahub_events "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/events"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/rpc/code"
)

type eventSender struct {
	datahubClient datahub_v1alpha1.DatahubServiceClient
}

func NewEventSender(datahubClient datahub_v1alpha1.DatahubServiceClient) *eventSender {
	return &eventSender{
		datahubClient: datahubClient,
	}
}

func (evtSender *eventSender) SendEvents(events []*datahub_events.Event) error {
	if len(events) == 0 {
		return nil
	}

	request := datahub_events.CreateEventsRequest{
		Events: events,
	}
	status, err := evtSender.datahubClient.CreateEvents(context.TODO(), &request)
	if err != nil {
		return errors.Errorf("send events to Datahub failed: %s", err.Error())
	} else if status == nil {
		return errors.Errorf("send events to Datahub failed: receive nil status")
	} else if status.Code != int32(code.Code_OK) {
		return errors.Errorf("send events to Datahub failed: statusCode: %d, message: %s",
			status.Code, status.Message)
	}

	return nil
}
