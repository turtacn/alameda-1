package config

import (
	"github.com/containers-ai/alameda/datapipe/pkg/repositories/apiserver"
	"github.com/containers-ai/alameda/pkg/utils/log"
)

const (
	defaultBindAddress = ":50060"
)

type Config struct {
	BindAddress string            `mapstructure:"bind-address"`
	APIServer   *apiserver.Config `mapstructure:"apiserver"`
	Log         *log.Config       `mapstructure:"log"`
}

func NewDefaultConfig() Config {
	var (
		defaultLogConfig       = log.NewDefaultConfig()
		defaultAPIServerConfig = apiserver.NewDefaultConfig()
		config                 = Config{
			BindAddress: defaultBindAddress,
			APIServer:   &defaultAPIServerConfig,
			Log:         &defaultLogConfig,
		}
	)

	return config
}

func (c *Config) Validate() error {
	return nil
}
