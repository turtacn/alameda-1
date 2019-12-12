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
	conn       *amqp.Connection
	queueURL   string
	retryItvMS int64
	connNotify chan *amqp.Error
}

func (sender *RabbitMQSender) SendJsonString(queueName, jsonStr, msgID string) error {
	publishRetryTime := sender.getRetry().publishRetryTime
	publishRetryIntervalMS := sender.getRetry().publishRetryIntervalMS
	for retry := 0; retry < publishRetryTime; retry++ {
		err := sender.sendJob(queueName, jsonStr, msgID)
		if err == nil {
			return nil
		}
		if err != nil && retry == publishRetryTime-1 {
			return err
		}
		time.Sleep(time.Duration(publishRetryIntervalMS) * time.Millisecond)
	}
	return fmt.Errorf("unknown error to send message to queue %s", queueName)
}

func (sender *RabbitMQSender) sendJob(queueName, jsonStr, msgID string) error {
	queueCH, err := sender.conn.Channel()
	defer queueCH.Close()

	if err != nil {
		return err
	}
	notifyClose := make(chan *amqp.Error)
	notifyConfirm := make(chan amqp.Confirmation)
	queueCH.Confirm(false)
	queueCH.NotifyClose(notifyClose)
	queueCH.NotifyPublish(notifyConfirm)

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
		return err
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
				//"x-deduplication-header": sender.getMessageHash(msgID),
				"x-deduplication-header": msgID,
			},
		})

	if err != nil {
		return err
	}

	ackTimeoutSec := viper.GetInt("queue.retry.ackTimeoutSec")
	if ackTimeoutSec <= 0 {
		ackTimeoutSec = DEFAULT_ACK_TIMEOUT_SEC
	}
	ticker := time.NewTicker(time.Duration(ackTimeoutSec) * time.Second)

	for {
		select {
		case confirm := <-notifyConfirm:
			if confirm.Ack {
				return nil
			} else {
				return fmt.Errorf("send job to queue %s failed, server does not receive the job", queueName)
			}
		case confirm := <-notifyClose:
			return fmt.Errorf(confirm.Error())
		case confirm := <-sender.connNotify:
			if !sender.conn.IsClosed() {
				err := sender.conn.Close()
				scope.Errorf(err.Error())
			}
			sender.conn = GetQueueConn(sender.queueURL, sender.retryItvMS)
			return fmt.Errorf(confirm.Error())
		case <-ticker.C:
			return fmt.Errorf("send job to queue %s timeout", queueName)
		}
	}
}

func NewRabbitMQSender(queueURL string, retryItvMS int64) (*RabbitMQSender, *amqp.Connection) {
	conn := GetQueueConn(queueURL, retryItvMS)
	sender := &RabbitMQSender{
		queueURL:   queueURL,
		retryItvMS: retryItvMS,
		conn:       conn,
	}
	sender.connNotify = sender.conn.NotifyClose(make(chan *amqp.Error))
	return sender, sender.conn
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
