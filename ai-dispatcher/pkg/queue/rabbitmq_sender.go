package queue

import (
	"crypto/sha1"
	"fmt"
	"time"

	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

type retry struct {
	publishRetryTime       int
	publishRetryIntervalMS int64
	consumeRetryTime       int
	consumeRetryIntervalMS int64
}

type RabbitMQSender struct {
	conn *amqp.Connection
}

func (sender *RabbitMQSender) SendJsonString(queueName, jsonStr, msgID string) error {
	publishRetryTime := sender.getRetry().publishRetryTime
	publishRetryIntervalMS := sender.getRetry().publishRetryIntervalMS
	for retry := 0; retry < publishRetryTime; retry++ {
		queueCH, err := sender.conn.Channel()

		if err != nil {
			if retry == (publishRetryTime - 1) {
				return err
			}
			continue
		}
		defer queueCH.Close()
		q, err := queueCH.QueueDeclare(
			queueName, // name
			true,      // durable
			false,     // delete when unused
			false,     // exclusive
			false,     // no-wait
			amqp.Table{
				"x-message-deduplication": true,
			}, // arguments
		)
		if err != nil {
			if retry == (publishRetryTime - 1) {
				return err
			}
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
				Headers: amqp.Table{
					"x-deduplication-header": sender.getMessageHash(msgID),
				},
			})

		if err != nil {
			if retry == (publishRetryTime - 1) {
				return err
			}
			time.Sleep(time.Duration(publishRetryIntervalMS) * time.Millisecond)
			continue
		} else {
			return nil
		}
	}
	return fmt.Errorf("unknown error to send message to queue %s", queueName)
}

func NewRabbitMQSender(conn *amqp.Connection) *RabbitMQSender {
	sender := &RabbitMQSender{
		conn: conn,
	}
	return sender
}

func (sender *RabbitMQSender) getRetry() *retry {
	publishRetryTime := viper.GetInt("queue.retry.publishTime")
	if publishRetryTime == 0 {
		publishRetryTime = DEFAULT_PUBLISH_RETRY_TIME
	}

	publishRetryIntervalMS := viper.GetInt64("queue.retry.publishIntervalMs")
	if publishRetryIntervalMS == 0 {
		publishRetryIntervalMS = DEFAULT_PUBLISH_RETRY_INTERVAL_MS
	}
	return &retry{
		publishRetryTime:       publishRetryTime,
		publishRetryIntervalMS: publishRetryIntervalMS,
	}
}

func (sender *RabbitMQSender) getMessageHash(msgStr string) string {
	h := sha1.New()
	h.Write([]byte(msgStr))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}
