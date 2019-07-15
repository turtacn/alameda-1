package datahub

import (
	"errors"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
	InternalWeaveScope "github.com/containers-ai/alameda/internal/pkg/weavescope"
	"github.com/containers-ai/alameda/pkg/utils/log"
)

const (
	defaultBindAddress = ":50050"
)

type Config struct {
	BindAddress string                     `mapstructure:"bind-address"`
	Prometheus  *InternalPromth.Config     `mapstructure:"prometheus"`
	InfluxDB    *InternalInflux.Config     `mapstructure:"influxdb"`
	WeaveScope  *InternalWeaveScope.Config `mapstructure:"weavescope"`
	Log         *log.Config                `mapstructure:"log"`
}

func NewDefaultConfig() Config {

	var (
		defaultlogConfig        = log.NewDefaultConfig()
		defaultPrometheusConfig = InternalPromth.NewDefaultConfig()
		defaultInfluxDBConfig   = InternalInflux.NewDefaultConfig()
		defaultWeaveScopeConfig = InternalWeaveScope.NewDefaultConfig()
		config                  = Config{
			BindAddress: defaultBindAddress,
			Prometheus:  defaultPrometheusConfig,
			InfluxDB:    defaultInfluxDBConfig,
			WeaveScope:  defaultWeaveScopeConfig,
			Log:         &defaultlogConfig,
		}
	)

	return config
}

func (c *Config) Validate() error {

	var err error

	err = c.Prometheus.Validate()
	if err != nil {
		return errors.New("failed to validate gRPC config: " + err.Error())
	}

	return nil
}
