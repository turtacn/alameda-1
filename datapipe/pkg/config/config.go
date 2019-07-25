package config

import (
	"github.com/containers-ai/alameda/datapipe/pkg/config/apiserver"
	"github.com/containers-ai/alameda/internal/pkg/database/prometheus"
	"github.com/containers-ai/alameda/pkg/utils/log"
)

const (
	defaultBindAddress = ":50060"
)

// Configuration of datapipe data source
type Config struct {
	BindAddress string             `mapstructure:"bindAddress"`
	APIServer   *apiserver.Config  `mapstructure:"apiserver"`
	Prometheus  *prometheus.Config `mapstructure:"prometheus"`
	Log         *log.Config        `mapstructure:"log"`
}

// Provide default configuration for datapipe
func NewDefaultConfig() Config {
	var (
		defaultLogConfig        = log.NewDefaultConfig()
		defaultAPIServerConfig  = apiserver.NewDefaultConfig()
		defaultPrometheusConfig = prometheus.NewDefaultConfig()
		config                  = Config{
			BindAddress: defaultBindAddress,
			APIServer:   defaultAPIServerConfig,
			Prometheus:  defaultPrometheusConfig,
			Log:         &defaultLogConfig,
		}
	)

	return config
}

// Confirm the datapipe configuration is validated
func (c *Config) Validate() error {
	return nil
}
