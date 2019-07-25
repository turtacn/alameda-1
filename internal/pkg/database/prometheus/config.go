package prometheus

import (
	"github.com/pkg/errors"
	"net/url"
)

const (
	defaultURL             = "https://prometheus-k8s.openshift-monitoring:9091"
	defaultBearerTokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"
)

// Configuration of prometheus data source
type Config struct {
	URL             string     `mapstructure:"url"`
	BearerTokenFile string     `mapstructure:"bearerTokenFile"`
	TLSConfig       *TLSConfig `mapstructure:"tlsConfig"`
	bearerToken     string
}

// Configuration of tls connection
type TLSConfig struct {
	InsecureSkipVerify bool `mapstructure:"insecureSkipVerify"`
}

// Provide default configuration for prometheus
func NewDefaultConfig() *Config {
	var config = Config{
		URL:             defaultURL,
		BearerTokenFile: defaultBearerTokenFile,
		TLSConfig: &TLSConfig{
			InsecureSkipVerify: true,
		},
	}
	return &config
}

// Confirm the prometheus configuration is validated
func (c *Config) Validate() error {
	_, err := url.Parse(c.URL)
	if err != nil {
		scope.Errorf("invalid URL: %s", err.Error())
		scope.Error("failed to validate prometheus configuration")
		return errors.New("invalid URL")
	}
	return nil
}
