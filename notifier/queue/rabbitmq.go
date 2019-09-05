package queue

import (
	"encoding/json"

	"github.com/containers-ai/alameda/notifier/notifying"
	"github.com/containers-ai/alameda/pkg/utils/log"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
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
	conn, err := amqp.Dial(queueURL)
	if err != nil {
		scope.Errorf("failed to connect to queue %s: %s", queueURL, err.Error())
	}
	return &rabbitmqClient{
		conn:        conn,
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
