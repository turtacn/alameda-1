package server

import (
	"github.com/containers-ai/alameda/operator/services/grpc"
)

type Config struct {
	GRPC *grpc.Config
}

func NewConfig() Config {

	c := Config{}
	c.init()

	return c
}

func (c *Config) init() {

	c.GRPC = grpc.NewConfig()
}

func (c Config) Validate() error {

	return nil
}
