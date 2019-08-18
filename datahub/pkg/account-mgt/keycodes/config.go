package keycodes

import (
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InternalLdap "github.com/containers-ai/alameda/internal/pkg/database/ldap"
)

const (
	defaultCliPath         = "/opt/prophetstor/federatorai/bin/license_main"
	defaultRefreshInterval = 180
)

// Configuration of keycode CLI
type Config struct {
	CliPath         string
	RefreshInterval int64
	AesKey          []byte
	InfluxDB        *InternalInflux.Config
	Ldap            *InternalLdap.Config
}

// Provide default configuration for keycode CLI
func NewDefaultConfig() *Config {
	var config = Config{
		CliPath:         defaultCliPath,
		RefreshInterval: defaultRefreshInterval,
		InfluxDB:        InternalInflux.NewDefaultConfig(),
		Ldap:            InternalLdap.NewDefaultConfig(),
	}
	return &config
}

// Confirm the keycode CLI configuration is validated
func (c *Config) Validate() error {
	return nil
}
