package influxdb

import (
	"github.com/containers-ai/alameda/pkg/utils/log"
	"time"
)

var (
	scope = log.RegisterScope("Database", "influxdb", 0)

	// ZeroTime is used as a constant of timestamp
	ZeroTime = time.Unix(0, 0)
)

type Database string
type Measurement string

// InfluxDB client interacts with database
type InfluxClient struct {
	Address                string
	Username               string
	Password               string
	RetentionDuration      string
	RetentionShardDuration string
}

// Instance InfluxDB API client with configuration
func NewClient(influxCfg *Config) *InfluxClient {
	return &InfluxClient{
		Address:                influxCfg.Address,
		Username:               influxCfg.Username,
		Password:               influxCfg.Password,
		RetentionDuration:      influxCfg.RetentionDuration,
		RetentionShardDuration: influxCfg.RetentionShardDuration,
	}
}
