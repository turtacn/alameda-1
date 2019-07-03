package queue

import (
	"time"

	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

type retry struct {
	publishRetryTime       int
	publishRetryIntervalMS int64
}

type RabbitMQSender struct {
	conn *amqp.Connection
}

func (sender *RabbitMQSender) SendJsonString(queueName, jsonStr string) error {
	publishRetryTime := sender.getRetry().publishRetryTime
	publishRetryIntervalMS := sender.getRetry().publishRetryIntervalMS
	for retry := 0; retry < publishRetryTime; retry++ {
		queueCH, err := sender.conn.Channel()

		if err != nil {
			continue
		}
		defer queueCH.Close()
		q, err := queueCH.QueueDeclare(
			queueName, // name
			true,      // durable
			false,     // delete when unused
			false,     // exclusive
			false,     // no-wait
			nil,       // arguments
		)
		if err != nil {
			continue
		}

		err = queueCH.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType:  "text/plain",
				Body:         []byte(jsonStr),
				DeliveryMode: 2, // 2 means persistent
			})

		if err == nil {
			break
		}

		if retry == (publishRetryTime - 1) {
			return err
		}
		time.Sleep(time.Duration(publishRetryIntervalMS) * time.Millisecond)
	}
	return nil
}

func NewRabbitMQSender(conn *amqp.Connection) *RabbitMQSender {
	sender := &RabbitMQSender{
		conn: conn,
	}
	return sender
}

func (sender *RabbitMQSender) getRetry() *retry {
	publishRetryTime := viper.GetInt("queue.retry.publish_time")
	if publishRetryTime == 0 {
		publishRetryTime = DEFAULT_PUBLISH_RETRY_TIME
	}

	publishRetryIntervalMS := viper.GetInt64("queue.retry.publish_interval_ms")
	if publishRetryIntervalMS == 0 {
		publishRetryIntervalMS = DEFAULT_PUBLISH_RETRY_INTERVAL_MS
	}
	return &retry{
		publishRetryTime:       publishRetryTime,
		publishRetryIntervalMS: publishRetryIntervalMS,
	}
}
