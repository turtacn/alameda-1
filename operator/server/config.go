package server

import (
	"github.com/containers-ai/alameda/operator/services/grpc"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type Config struct {
	GRPC    *grpc.Config
	Manager manager.Manager
}

func NewConfig(manager manager.Manager) Config {

	c := Config{
		Manager: manager,
	}
	c.init()

	return c
}

func (c *Config) init() {

	c.GRPC = grpc.NewConfig()
}

func (c Config) Validate() error {

	return nil
}
