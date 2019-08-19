package metrics

import (
	"github.com/containers-ai/alameda/pkg/utils/log"
)

var (
	scope     = log.RegisterScope("notifier", "notifier-alerts", 0)
	Notifiers = make([]AlertInterface, 0)
)

type Notifier struct {
	Enabled       bool   `mapstructure:"enabled"`
	Specs         string `mapstructure:"specs"`
	EventInterval string `mapstructure:"eventInterval"`
	EventLevel    string `mapstructure:"eventLevel"`
}

type AlertInterface interface {
	GetName() string
	GetSpecs() string
	GetEnabled() bool
	Validate()
	GenerateCriteria()
	MeetCriteria() bool
}

type AlertMetrics struct {
	name     string
	notifier *Notifier
}

func (c *AlertMetrics) GetName() string {
	return c.name
}

func (c *AlertMetrics) GetSpecs() string {
	return c.notifier.Specs
}

func (c *AlertMetrics) GetEnabled() bool {
	return c.notifier.Enabled
}

func (c *AlertMetrics) Validate() {
}

func (c *AlertMetrics) GenerateCriteria() {

}

func (c *AlertMetrics) MeetCriteria() bool {
	return false
}
