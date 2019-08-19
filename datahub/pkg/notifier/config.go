package notifier

import (
	Metrics "github.com/containers-ai/alameda/datahub/pkg/notifier/metrics"
)

type Config struct {
	Keycode *Metrics.Notifier `mapstructure:"keycode"`
}

func NewDefaultConfig() *Config {
	var config = Config{
		Keycode: &Metrics.Notifier{
			Enabled:       Metrics.DefaultKeycodeEnabled,
			Specs:         Metrics.DefaultKeycodeSpecs,
			EventInterval: Metrics.DefaultKeycodeEventInterval,
			EventLevel:    Metrics.DefaultKeycodeEventLevel,
		},
	}
	return &config
}

func (c *Config) Validate() error {
	return nil
}
