package aiserver

import (
	grpcutils "github.com/containers-ai/alameda/operator/pkg/utils/grpc"
)

type Config struct {
	Address string `mapstructure:"address"`
}

func NewConfig() *Config {
	c := Config{}
	c.init()
	return &c
}

func (c *Config) init() {
	c.Address = grpcutils.GetAIServiceAddress()
}

func (c *Config) Validate() error {
	return nil
}
