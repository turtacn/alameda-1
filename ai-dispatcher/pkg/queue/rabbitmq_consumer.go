package queue

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

type RabbitMQConsumer struct {
	conn *amqp.Connection
}

func (consumer *RabbitMQConsumer) ConsumeJsonString(queueName string) (<-chan amqp.Delivery, error) {
	consumeRetryTime := consumer.getRetry().consumeRetryTime
	for retry := 0; retry < consumeRetryTime; retry++ {
		queueCH, err := consumer.conn.Channel()

		if err != nil {
			if retry == (consumeRetryTime - 1) {
				queueCH.Close()
				return nil, err
			}
			continue
		}

		return queueCH.Consume(queueName, "", true, false, false, false, nil)
	}
	return nil,
		fmt.Errorf("unknown error to consume message from queue %s", queueName)
}

func (consumer *RabbitMQConsumer) ReceiveJsonString(queueName string) (
	string, bool, error) {
	consumeRetryTime := consumer.getRetry().consumeRetryTime
	consumeRetryIntervalMS := consumer.getRetry().consumeRetryIntervalMS
	for retry := 0; retry < consumeRetryTime; retry++ {
		queueCH, err := consumer.conn.Channel()

		if err != nil {
			if retry == (consumeRetryTime - 1) {
				return "", true, err
			}
			continue
		}
		defer queueCH.Close()

		msg, ok, err := queueCH.Get(queueName, true)

		if ok && err == nil {
			return string(msg.Body), ok, err
		}
		if err != nil {
			if retry == (consumeRetryTime - 1) {
				return "", true, err
			}
		}
		if !ok {
			if retry == (consumeRetryTime - 1) {
				return "", ok, err
			}
		}
		time.Sleep(time.Duration(consumeRetryIntervalMS) * time.Millisecond)
	}
	return "", false,
		fmt.Errorf("unknown error to receive message from queue %s", queueName)
}

func NewRabbitMQConsumer(conn *amqp.Connection) *RabbitMQConsumer {
	sender := &RabbitMQConsumer{
		conn: conn,
	}
	return sender
}

func (consumer *RabbitMQConsumer) getRetry() *retry {
	consumeRetryTime := viper.GetInt("queue.retry.consumeTime")
	if consumeRetryTime == 0 {
		consumeRetryTime = DEFAULT_CONSUME_RETRY_TIME
	}

	consumeRetryIntervalMS := viper.GetInt64("queue.retry.consumeIntervalMs")
	if consumeRetryIntervalMS == 0 {
		consumeRetryIntervalMS = DEFAULT_CONSUME_RETRY_INTERVAL_MS
	}
	return &retry{
		consumeRetryTime:       consumeRetryTime,
		consumeRetryIntervalMS: consumeRetryIntervalMS,
	}
}
