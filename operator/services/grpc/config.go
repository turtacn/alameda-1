package grpc

import (
	"strconv"

	"github.com/containers-ai/alameda/operator/pkg/kubernetes/metrics/prometheus"
	grpcutils "github.com/containers-ai/alameda/operator/pkg/utils/grpc"
)

type Config struct {
	BindAddress string
	Prometheus  *prometheus.Config
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
