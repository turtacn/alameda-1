package app

import (
	"fmt"
	"time"

	config "github.com/containers-ai/alameda/rabbitmq"

	"github.com/spf13/cobra"

	"github.com/streadway/amqp"
)

const (
	publishRetryTime = 3
	retrySec = 10
	queueName        = "test_queue"
)

var (
	pushQueue  string
	PublishCmd = &cobra.Command{
		Use:   "publish",
		Short: "Start Publish",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			startPublish()
		},
	}
)

func init() {
	parseProbeFlag()
}

func parseProbeFlag() {
	PublishCmd.Flags().StringVar(&pushQueue, "queue", queueName, "The push queue name for rabbitmq.")
}

func startPublish() {
	config := config.NewRabbitmqConfig("admin", "adminpass", "127.0.0.1", "5672")
	rabbitmqAddress := fmt.Sprintf("amqp://%s:%s@%s:%s/", config.Account, config.Password, config.Address, config.Port)
	fmt.Println(rabbitmqAddress)
	for retry := 0; retry < publishRetryTime; retry++ {
		conn, err := amqp.Dial(rabbitmqAddress)
		defer conn.Close()
		if err != nil {
			fmt.Println(fmt.Sprintf("Connect Fail: %s", err.Error()))
			time.Sleep(time.Duration(retrySec) * time.Millisecond)
			continue
		}
		ch, err := conn.Channel()

		defer ch.Close()
		if err != nil {
			fmt.Println(fmt.Sprintf("Connect Channel Fail: %s", err.Error()))
			time.Sleep(time.Duration(retrySec) * time.Millisecond)
			continue
		}
		q, err := ch.QueueDeclare(
			pushQueue, // name
			true,      // durable
			false,     // delete when unused
			false,     // exclusive
			false,     // no-wait
			amqp.Table{
				"x-message-deduplication": true,
			}, // arguments
		)
		jsonStr := []byte("{'test': '123'}")
		err = ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType:  "text/plain",
				Body:         []byte(jsonStr),
				DeliveryMode: 2, // 2 means persistent
				Headers: amqp.Table{
					"x-deduplication-header": "mess1",
				},
			})

		if err != nil {
			fmt.Println(fmt.Sprintf("Publish Fail: %s", err.Error()))
			time.Sleep(time.Duration(retrySec) * time.Millisecond)
			continue
		} else {
			fmt.Println("Publish Success")
			break
		}
	}
}
