package server

import (
	"errors"

	"github.com/containers-ai/alameda/operator/pkg/utils/log"
	grafanadatasource "github.com/containers-ai/alameda/operator/services/grafana-datasource"
	"github.com/containers-ai/alameda/operator/services/grpc"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type Config struct {
	GRPC           *grpc.Config              `mapstructure:"gRPC"`
	Log            *log.Config               `mapstructure:"log"`
	GrafanaBackend *grafanadatasource.Config `mapstructure:"grafana-backend"`
	Manager        manager.Manager
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
	c.Log = log.NewConfig()
	c.GrafanaBackend = grafanadatasource.NewConfig()
}

func (c Config) Validate() error {

	var err error

	err = c.GRPC.Validate()
	if err != nil {
		return errors.New("server config validate failed: " + err.Error())
	}

	return nil
}
