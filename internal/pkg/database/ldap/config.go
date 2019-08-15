package ldap

const (
	defaultAddress = "alameda-ldap.federatorai:389"
	defaultBaseDN  = ""
	defaultAdminID = ""
	defaultAdminPW = ""
)

// Configuration of LDAP server data source
type Config struct {
	Address string `mapstructure:"address"`
	BaseDN  string `mapstructure:"baseDN"`
	AdminID string `mapstructure:"adminID"`
	AdminPW string `mapstructure:"adminPW"`
}

// Provide default configuration for LDAP
func NewDefaultConfig() *Config {
	var config = Config{
		Address: defaultAddress,
		BaseDN:  defaultBaseDN,
		AdminID: defaultAdminID,
		AdminPW: defaultAdminPW,
	}
	return &config
}

// Confirm the configuration is validated
func (c *Config) Validate() error {
	return nil
}
