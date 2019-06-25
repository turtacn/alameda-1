package ldap

const (
	defaultAddress = "127.0.0.1:389"
)

// Configuration of LDAP server data source
type Config struct {
	Address string `mapstructure:"address"`
}

// NewDefaultConfig Provide default configuration
func NewDefaultConfig() Config {
	var config = Config{
		Address: defaultAddress,
	}
	return config
}

// Validate Confirm the configuration is validate
func (c *Config) Validate() error {
	return nil
}
