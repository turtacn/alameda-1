package prometheus

type Config struct {
	Host            string     `mapstructure:"host"`
	Port            string     `mapstructure:"port"`
	Protocol        string     `mapstructure:"protocol"`
	BearerTokenFile string     `mapstructure:"bearer-token-file"`
	TLSConfig       *TLSConfig `mapstructure:"tls-config"`
	// Path to bearer token file.

	bearerToken string
}

type TLSConfig struct {
	// Path to CA certificate to validate API server certificate with.
	// CAFile string
	// Certificate and key files for client cert authentication to the server.
	// CertFile           string
	// KeyFile            string
	InsecureSkipVerify bool `mapstructure:"insecure-skip-verify"`
}

func NewConfig() Config {

	c := &Config{}
	c.init()
	return *c
}

func (c *Config) init() {
	c.Host = "prometheus-k8s.openshift-monitoring"
	c.Port = "9091"
	c.Protocol = "https"
	c.TLSConfig = &TLSConfig{
		InsecureSkipVerify: true,
	}
}
