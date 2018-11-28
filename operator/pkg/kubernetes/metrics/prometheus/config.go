package prometheus

import (
	"errors"
	"net/url"
)

type Config struct {
	URL             string     `mapstructure:"url"`
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
	c.URL = "https://prometheus-k8s.openshift-monitoring:9091"
	c.TLSConfig = &TLSConfig{
		InsecureSkipVerify: true,
	}
}

func (c *Config) Validate() error {

	var err error

	_, err = url.Parse(c.URL)
	if err != nil {
		return errors.New("prometheus config validate failed: " + err.Error())
	}

	return nil
}
