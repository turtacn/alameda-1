package influxdb

import (
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
)

const (
	// InfluxDB tag definition
	Time    string = "time"
	EndTime string = "end_time"

	// InfluxDB database definition
	ClusterStatus  influxdb.Database = "alameda_cluster_status"
	Event          influxdb.Database = "alameda_event"
	Gpu            influxdb.Database = "alameda_gpu"
	GpuPrediction  influxdb.Database = "alameda_gpu_prediction"
	Metric         influxdb.Database = "alameda_metric"
	Planning       influxdb.Database = "alameda_planning"
	Prediction     influxdb.Database = "alameda_prediction"
	Recommendation influxdb.Database = "alameda_recommendation"
	Score          influxdb.Database = "alameda_score"
)
