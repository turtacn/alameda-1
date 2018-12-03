package grafanadatasource

type Config struct {
	BindPort uint16 `mapstructure:"bind-port"`
}

func NewConfig() *Config {

	c := Config{}
	c.init()
	return &c
}

func (c *Config) init() {
	c.BindPort = 50055
}

func (c *Config) Validate() error {
	return nil
}
