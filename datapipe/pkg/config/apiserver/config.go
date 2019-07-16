package apiserver

const (
	DefaultAddress  = "127.0.0.1:50055"
	DefaultUsername = "admin"
	DefaultPassword = "password"
)

// Configuration of API server data source
type Config struct {
	Address  string `mapstructure:"address"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// Provide default configuration for api-server
func NewDefaultConfig() *Config {
	var config = Config{
		Address:  DefaultAddress,
		Username: DefaultUsername,
		Password: DefaultPassword,
	}
	return &config
}

// Confirm the api-server configuration is validated
func (c *Config) Validate() error {
	return nil
}
