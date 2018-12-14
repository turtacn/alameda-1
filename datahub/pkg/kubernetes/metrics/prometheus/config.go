package prometheus

import (
	"errors"
	"net/url"
)

const (
	defaultURL             = "https://prometheus-k8s.openshift-monitoring:9091"
	defaultBearerTokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"
)

type Config struct {
	URL             string     `mapstructure:"url"`
	BearerTokenFile string     `mapstructure:"bearer-token-file"`
	TLSConfig       *TLSConfig `mapstructure:"tls-config"`

	bearerToken string
}

type TLSConfig struct {
	InsecureSkipVerify bool `mapstructure:"insecure-skip-verify"`
}

func NewDefaultConfig() Config {

	var config = Config{
		URL:             defaultURL,
		BearerTokenFile: defaultBearerTokenFile,
		TLSConfig: &TLSConfig{
			InsecureSkipVerify: true,
		},
	}
	return config
}

func (c *Config) Validate() error {

	var err error

	_, err = url.Parse(c.URL)
	if err != nil {
		return errors.New("prometheus config validate failed: " + err.Error())
	}

	return nil
}
