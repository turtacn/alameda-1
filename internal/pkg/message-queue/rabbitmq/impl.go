package rabbitmq

import (
	"github.com/streadway/amqp"
	"time"
)

func (s *RabbitMQSender) SendJsonString(queueName, jsonStr string) error {
	publishRetryTime := s.rabbitMQConfig.Retry.PublishTime
	publishRetryIntervalMS := s.rabbitMQConfig.Retry.PublishIntervalMS
	for retry := 0; retry < publishRetryTime; retry++ {
		queueCH, err := s.conn.Channel()

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

func (s *RabbitMQSender) Close() error {
	if s.conn != nil {
		return s.conn.Close()
	}

	return nil
}
