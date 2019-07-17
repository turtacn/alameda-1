package rabbitmq

import (
	"github.com/streadway/amqp"
)

type RabbitMQSender struct {
	conn           *amqp.Connection
	rabbitMQConfig *Config
}

func NewRabbitMQSender(rabbitMQConfig *Config) (*RabbitMQSender, error) {
	rabbitMQConn, err := amqp.Dial(rabbitMQConfig.URL)
	if err != nil {
		return nil, err
	}

	sender := &RabbitMQSender{
		conn:           rabbitMQConn,
		rabbitMQConfig: rabbitMQConfig,
	}
	return sender, nil
}
