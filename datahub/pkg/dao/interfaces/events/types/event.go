package types

import (
	ApiEvents "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/events"
)

type EventDAO interface {
	CreateEvents(in *ApiEvents.CreateEventsRequest) error
	ListEvents(in *ApiEvents.ListEventsRequest) ([]*ApiEvents.Event, error)
	SendEvents(in *ApiEvents.CreateEventsRequest) error
}
