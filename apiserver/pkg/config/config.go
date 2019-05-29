package config

import (
	"github.com/containers-ai/alameda/apiserver/pkg/repositories/datahub"
	"github.com/containers-ai/alameda/pkg/utils/log"
)

const (
	defaultBindAddress = ":50055"
)

type Config struct {
	BindAddress string          `mapstructure:"bind-address"`
	Datahub     *datahub.Config `mapstructure:"datahub"`
	Log         *log.Config     `mapstructure:"log"`
}

func NewDefaultConfig() Config {
	var (
		defaultLogConfig       = log.NewDefaultConfig()
		defaultAPIServerConfig = datahub.NewDefaultConfig()
		config                 = Config{
			BindAddress: defaultBindAddress,
			Datahub:     &defaultAPIServerConfig,
			Log:         &defaultLogConfig,
		}
	)

	return config
}

func (c *Config) Validate() error {
	return nil
}
