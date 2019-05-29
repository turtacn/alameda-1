package apiserver

const (
	defaultAddress  = "127.0.0.1:50055"
	defaultUsername = "admin"
	defaultPassword = "password"
)

// Configuration of API server data source
type Config struct {
	Address  string `mapstructure:"address"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// NewDefaultConfig Provide default configuration
func NewDefaultConfig() Config {
	var config = Config{
		Address:  defaultAddress,
		Username: defaultUsername,
		Password: defaultPassword,
	}
	return config
}

// Validate Confirm the configuration is validate
func (c *Config) Validate() error {
	return nil
}
