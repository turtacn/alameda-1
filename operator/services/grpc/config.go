package grpc

import (
	"strconv"

	grpcutils "github.com/containers-ai/alameda/operator/pkg/utils/grpc"
)

type Config struct {
	BindAddress string
}

func NewConfig() *Config {

	c := Config{}
	c.init()
	return &c
}

func (c *Config) init() {

	c.BindAddress = ":" + strconv.Itoa(grpcutils.GetServerPort())
}
