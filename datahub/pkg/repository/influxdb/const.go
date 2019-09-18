package influxdb

import (
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
)

const (
	// Time is InfluxDB time tag
	Time string = "time"

	// EndTime is InfluxDB time tag
	EndTime string = "end_time"

	// ClusterStatus is cluster_status database
	ClusterStatus influxdb.Database = "alameda_cluster_status"

	// Prediction is prediction database
	Prediction influxdb.Database = "alameda_prediction"

	// Recommendation is recommendation database
	Recommendation influxdb.Database = "alameda_recommendation"

	// Planning is planning database
	Planning influxdb.Database = "alameda_planning"

	// Score is score database
	Score influxdb.Database = "alameda_score"

	// Event is score database
	Event influxdb.Database = "alameda_event"

	// Gpu is gpu database
	Gpu influxdb.Database = "alameda_gpu"

	// GpuPrediction is gpu prediction database
	GpuPrediction influxdb.Database = "alameda_gpu_prediction"
)
