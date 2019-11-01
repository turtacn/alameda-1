package prometheus

import (
	"crypto/tls"
	"github.com/containers-ai/alameda/pkg/utils/log"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	scope = log.RegisterScope("Database", "prometheus", 0)
)

// Prometheus client definition
type Prometheus struct {
	Config    *Config
	Client    *http.Client
	Transport *http.Transport
}

// Instance prometheus API client with configuration
func NewClient(config *Config) (*Prometheus, error) {
	var (
		requestTimeout   = 30 * time.Second
		handShakeTimeout = 5 * time.Second
	)

	// Validate prometheus configuration file
	if err := config.Validate(); err != nil {
		scope.Error("failed to create prometheus instance")
		return nil, err
	}

	// Create http transport
	tr := &http.Transport{
		TLSHandshakeTimeout: handShakeTimeout,
	}
	if config.TLSConfig != nil {
		tr.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: config.TLSConfig.InsecureSkipVerify,
		}
	}

	// Create http client
	client := &http.Client{
		Timeout:   requestTimeout,
		Transport: tr,
	}

	// Read prometheus bearer token file
	if config.BearerTokenFile != "" {
		token, err := ioutil.ReadFile(config.BearerTokenFile)
		if err != nil {
			scope.Errorf("failed to read bearer token file: %s", err.Error())
			scope.Error("failed to create prometheus instance")
			return nil, errors.New("failed to read bearer token file")
		}
		config.bearerToken = string(token)
	}

	return &Prometheus{
		Config:    config,
		Client:    client,
		Transport: tr,
	}, nil
}
