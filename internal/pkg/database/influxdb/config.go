package influxdb

import (
	"github.com/pkg/errors"
	"net/url"
)

const (
	defaultAddress                = "https://influxdb.alameda.svc.cluster.local:8086"
	defaultUsername               = "datahub"
	defaultPassword               = "datahub"
	defaultInsecureSkipVerify     = true
	defaultRetentionDuration      = "30d"
	defaultRetentionShardDuration = "1d"
)

// Configuration of InfluxDB data source
type Config struct {
	Address                string `mapstructure:"address"`
	Username               string `mapstructure:"username"`
	Password               string `mapstructure:"password"`
	InsecureSkipVerify     bool   `mapstructure:"insecureSkipVerify"`
	RetentionDuration      string `mapstructure:"retentionDuration"`
	RetentionShardDuration string `mapstructure:"retentionShardDuration"`
}

// Provide default configuration for InfluxDB
func NewDefaultConfig() *Config {
	var config = Config{
		Address:                defaultAddress,
		Username:               defaultUsername,
		Password:               defaultPassword,
		InsecureSkipVerify:     defaultInsecureSkipVerify,
		RetentionDuration:      defaultRetentionDuration,
		RetentionShardDuration: defaultRetentionShardDuration,
	}
	return &config
}

// Confirm the InfluxDB configuration is validated
func (c *Config) Validate() error {
	_, err := url.Parse(c.Address)
	if err != nil {
		return errors.New("failed to validate InfluxDB configuration: " + err.Error())
	}
	return nil
}
