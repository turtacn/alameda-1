package eventmgt

import (
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	//"github.com/containers-ai/alameda/internal/pkg/message-queue/rabbitmq"
	InternalRabbitMQ "github.com/containers-ai/alameda/internal/pkg/message-queue/rabbitmq"
)

// InfluxDB client interacts with database
type EventMgt struct {
	//InfluxConfig   *InternalInflux.Config
	RabbitMQConfig *InternalRabbitMQ.Config

	influxDB *InternalInflux.InfluxClient
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
