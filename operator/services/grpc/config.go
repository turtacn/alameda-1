package grpc

import (
	"errors"
	"strconv"

	"github.com/containers-ai/alameda/operator/pkg/kubernetes/metrics/prometheus"
	grpcutils "github.com/containers-ai/alameda/operator/pkg/utils/grpc"
)

type Config struct {
	BindAddress string             `mapstructure:"bind-address"`
	Prometheus  *prometheus.Config `mapstructure:"prometheus"`
}

func NewConfig() *Config {

	c := Config{}
	c.init()
	return &c
}

func (c *Config) init() {

	c.BindAddress = ":" + strconv.Itoa(grpcutils.GetServerPort())

	promConfig := prometheus.NewConfig()
	c.Prometheus = &promConfig
}

func (c *Config) Validate() error {

	var err error

	err = c.Prometheus.Validate()
	if err != nil {
		return errors.New("gRPC config validate failed: " + err.Error())
	}

	return nil
}
