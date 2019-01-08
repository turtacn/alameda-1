package influxdb

const (
	// Time is InfluxDB time tag
	Time string = "time"

	// ClusterStatus is cluster_status database
	ClusterStatus Database = "alameda_cluster_status"
	// Prediction is prediction database
	Prediction Database = "alameda_prediction"
	// Recommendation is recommendation database
	Recommendation Database = "alameda_recommendation"
)
