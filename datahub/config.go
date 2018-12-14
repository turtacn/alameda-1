package datahub

import (
	"errors"

	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metrics/prometheus"
	"github.com/containers-ai/alameda/pkg/utils/log"
)

const (
	defaultBindAddress = ":50050"
)

type Config struct {
	BindAddress string             `mapstructure:"bind-address"`
	Prometheus  *prometheus.Config `mapstructure:"prometheus"`
	Log         *log.Config        `mapstructure:"log"`
}

func NewDefaultConfig() Config {

	var (
		defaultlogConfig        = log.NewDefaultConfig()
		defaultPrometheusConfig = prometheus.NewDefaultConfig()
		config                  = Config{
			BindAddress: defaultBindAddress,
			Prometheus:  &defaultPrometheusConfig,
			Log:         &defaultlogConfig,
		}
	)

	return config
}

func (c *Config) Validate() error {

	var err error

	err = c.Prometheus.Validate()
	if err != nil {
		return errors.New("gRPC config validate failed: " + err.Error())
	}

	return nil
}
