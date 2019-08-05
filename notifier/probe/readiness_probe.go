package probe

import (
	"github.com/streadway/amqp"
)

type ReadinessProbeConfig struct{}

func connQueue(url string) error {
	_, err := amqp.Dial(url)
	if err != nil {
		return err
	}
	return nil
}
