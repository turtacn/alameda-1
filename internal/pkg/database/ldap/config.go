package ldap

const (
	defaultAddress = "alameda-ldap.federatorai:389"
)

// Configuration of LDAP server data source
type Config struct {
	Address string `mapstructure:"address"`
}

// Provide default configuration for LDAP
func NewDefaultConfig() *Config {
	var config = Config{
		Address: defaultAddress,
	}
	return &config
}

// Confirm the configuration is validated
func (c *Config) Validate() error {
	return nil
}
