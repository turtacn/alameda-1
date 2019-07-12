package weavescope

import (
	"crypto/tls"
	"net/http"
)

type WeaveScopeClient struct {
	URL string
}

func NewClient(config *Config) *WeaveScopeClient {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	return &WeaveScopeClient{
		URL: config.URL,
	}
}
