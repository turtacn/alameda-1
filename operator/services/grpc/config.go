package grpc

type Config struct {
	BindAddress string
}

func NewConfig() *Config {

	c := Config{}
	c.init()
	return &c
}

func (c *Config) init() {

	c.BindAddress = ":50050"
}
