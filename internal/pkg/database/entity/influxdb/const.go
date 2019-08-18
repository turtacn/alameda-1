package influxdb

import (
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
)

const (
	// Time is InfluxDB time tag
	Time string = "time"

	// EndTime is InfluxDB time tag
	EndTime string = "end_time"
)

const (
	// ClusterStatus is cluster_status database
	ClusterStatus influxdb.Database = "alameda_cluster_status"

	// Event is event database
	Event influxdb.Database = "alameda_event"
)
