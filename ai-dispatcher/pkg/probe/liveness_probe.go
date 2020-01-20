package probe

import (
	"fmt"

	"github.com/streadway/amqp"
)

type LivenessProbeConfig struct {
	QueueURL string
}

func checkRabbitmqNotBlock(url string) error {
	conn, err := amqp.Dial(url)
	if conn != nil {
		defer conn.Close()
	}

	if err != nil {
		fmt.Println(err)
		return err
	}
	ch, err := conn.Channel()
	q, err := ch.QueueDeclare(
		"test_queue", // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		amqp.Table{
			"x-message-deduplication": true,
		}, // arguments
	)
	if err != nil {
		return err
	}
	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType:  "text/plain",
			Body:         []byte("{'test': '123'}"),
			DeliveryMode: 2, // 2 means persistent
			Headers: amqp.Table{
				//"x-deduplication-header": sender.getMessageHash(msgID),
				"x-deduplication-header": "1000",
			},
		})
	if err != nil {
		return err
	}
	return nil
}
