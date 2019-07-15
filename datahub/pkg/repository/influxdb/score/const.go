package score

import (
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
)

const (
	// SimulatedSchedulingScore Measurement name of simulated scheduling score in influxdb
	SimulatedSchedulingScore InternalInflux.Measurement = "simulated_scheduling_score"
)
