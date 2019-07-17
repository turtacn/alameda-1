package event

import (
	"encoding/json"
	RepoInfluxEvent "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb/event"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	"github.com/containers-ai/alameda/internal/pkg/message-queue/rabbitmq"
	InternalRabbitMQ "github.com/containers-ai/alameda/internal/pkg/message-queue/rabbitmq"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
)

type Event struct {
	influxDBConfig *InternalInflux.Config
	rabbitMQConfig *InternalRabbitMQ.Config
}

func NewEventWithConfig(influxConfig *InternalInflux.Config, rabbitMQConfig *InternalRabbitMQ.Config) Event {
	return Event{
		influxDBConfig: influxConfig,
		rabbitMQConfig: rabbitMQConfig,
	}
}

func (e *Event) CreateEvents(in *datahub_v1alpha1.CreateEventsRequest) error {
	eventRepo := RepoInfluxEvent.NewEventRepository(e.influxDBConfig)
	return eventRepo.CreateEvents(in)
}

func (e *Event) ListEvents(in *datahub_v1alpha1.ListEventsRequest) ([]*datahub_v1alpha1.Event, error) {
	eventRepo := RepoInfluxEvent.NewEventRepository(e.influxDBConfig)
	return eventRepo.ListEvents(in)
}

func (e *Event) SendEvents(in *datahub_v1alpha1.CreateEventsRequest) error {
	messageQueue, err := rabbitmq.NewRabbitMQSender(e.rabbitMQConfig)
	if err != nil {
		return err
	}
	defer messageQueue.Close()

	events, err := json.Marshal(in.GetEvents())
	if err != nil {
		return err
	}

	err = messageQueue.SendJsonString("event", string(events))
	return err
}
