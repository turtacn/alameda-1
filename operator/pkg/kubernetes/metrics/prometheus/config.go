package prometheus

type Config struct {
	Host      string
	Port      string
	Protocol  string
	TLSConfig *TLSConfig
	// Path to bearer token file.
	BearerTokenFile string

	bearerToken string
}

type TLSConfig struct {
	// Path to CA certificate to validate API server certificate with.
	// CAFile string
	// Certificate and key files for client cert authentication to the server.
	// CertFile           string
	// KeyFile            string
	InsecureSkipVerify bool
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
	c.BearerTokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"
}
