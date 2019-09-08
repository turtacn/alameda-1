package queue

import (
	"encoding/json"
	"os"
	"time"

	"github.com/containers-ai/alameda/notifier/notifying"
	"github.com/containers-ai/alameda/pkg/utils/log"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var scope = log.RegisterScope("queue", "queue", 0)

type rabbitmqClient struct {
	conn        *amqp.Connection
	datahubConn *grpc.ClientConn
	mgr         manager.Manager
}

func NewRabbitMQClient(mgr manager.Manager, queueURL string, datahubConn *grpc.ClientConn) *rabbitmqClient {
	retryConn := viper.GetInt("rabbitmq.connRetry")
	retryConnInterval := viper.GetInt64("rabbitmq.connRetryInterval")
	for retry := 0; retry < retryConn; retry++ {
		conn, err := amqp.Dial(queueURL)
		if err == nil {
			return &rabbitmqClient{
				conn:        conn,
				datahubConn: datahubConn,
				mgr:         mgr,
			}
		}
		if err != nil {
			if retry == retryConn-1 {
				scope.Errorf("failed to connect to queue %s: %s", queueURL, err.Error())
				os.Exit(1)
			} else {
				scope.Errorf("failed to connect to queue %s: %s. Retry after %d seconds",
					queueURL, err.Error(), retryConnInterval)
				time.Sleep(time.Duration(retryConnInterval) * time.Second)
			}
		}

	}

	return &rabbitmqClient{
		datahubConn: datahubConn,
		mgr:         mgr,
	}
}

func (client *rabbitmqClient) Start() {
	ch, err := client.conn.Channel()
	defer ch.Close()
	if err != nil {
		scope.Errorf(err.Error())
	}

	q, err := ch.QueueDeclare(
		"event", // name
		true,    // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)

	if err != nil {
		scope.Errorf(err.Error())
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		scope.Errorf(err.Error())
	}
	forever := make(chan bool)
	datahubClient := datahub_v1alpha1.NewDatahubServiceClient(client.datahubConn)
	notifier := notifying.NewNotifier(client.mgr.GetClient(), datahubClient)
	go func() {
		for d := range msgs {
			scope.Infof("Received events: %s", d.Body)
			evts := &[]*datahub_v1alpha1.Event{}
			decodeErr := json.Unmarshal(d.Body, evts)
			if decodeErr != nil {
				scope.Error(decodeErr.Error())
			} else {
				notifier.NotifyEvents(*evts)
			}
			d.Ack(false)
		}
	}()
	<-forever
}
