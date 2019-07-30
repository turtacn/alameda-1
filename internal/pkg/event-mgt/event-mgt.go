package eventmgt

import (
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InternalRabbitMQ "github.com/containers-ai/alameda/internal/pkg/message-queue/rabbitmq"

	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
)

var (
	gInfluxDBCfg    = InternalInflux.NewDefaultConfig()
	gRabbitMQConfig = InternalRabbitMQ.NewDefaultConfig()
)

type EventMgt struct {
	RabbitMQConfig *InternalRabbitMQ.Config
	influxDB       *InternalInflux.InfluxClient
}

func InitEventMgt(influxDBCfg *InternalInflux.Config, rabbitMQConfig *InternalRabbitMQ.Config) {
	gInfluxDBCfg = influxDBCfg
	gRabbitMQConfig = rabbitMQConfig
}

func NewEventMgt(influxDBCfg *InternalInflux.Config, rabbitMQConfig *InternalRabbitMQ.Config) *EventMgt {
	return &EventMgt{
		influxDB: &InternalInflux.InfluxClient{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
		RabbitMQConfig: rabbitMQConfig,
	}
}

func PostEvents(in *datahub_v1alpha1.CreateEventsRequest) error {
	eventMgt := NewEventMgt(gInfluxDBCfg, gRabbitMQConfig)
	return eventMgt.PostEvents(in)
}

func ListEvents(in *datahub_v1alpha1.ListEventsRequest) ([]*datahub_v1alpha1.Event, error) {
	eventMgt := NewEventMgt(gInfluxDBCfg, gRabbitMQConfig)
	return eventMgt.ListEvents(in)
}
