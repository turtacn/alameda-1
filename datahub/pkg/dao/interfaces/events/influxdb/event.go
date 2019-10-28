package influxdb

import (
	"encoding/json"
	DaoEventTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/events/types"
	RepoInfluxEvent "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/events"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InternalRabbitMQ "github.com/containers-ai/alameda/internal/pkg/message-queue/rabbitmq"
	ApiEvents "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/events"
)

type Event struct {
	InfluxDBConfig *InternalInflux.Config
	RabbitMQConfig *InternalRabbitMQ.Config
}

func NewEventWithConfig(influxConfig *InternalInflux.Config, rabbitMQConfig *InternalRabbitMQ.Config) DaoEventTypes.EventDAO {
	return &Event{InfluxDBConfig: influxConfig, RabbitMQConfig: rabbitMQConfig}
}

func (e *Event) CreateEvents(in *ApiEvents.CreateEventsRequest) error {
	eventRepo := RepoInfluxEvent.NewEventRepository(e.InfluxDBConfig)
	return eventRepo.CreateEvents(in)
}

func (e *Event) ListEvents(in *ApiEvents.ListEventsRequest) ([]*ApiEvents.Event, error) {
	eventRepo := RepoInfluxEvent.NewEventRepository(e.InfluxDBConfig)
	return eventRepo.ListEvents(in)
}

func (e *Event) SendEvents(in *ApiEvents.CreateEventsRequest) error {
	messageQueue, err := InternalRabbitMQ.NewRabbitMQSender(e.RabbitMQConfig)
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
