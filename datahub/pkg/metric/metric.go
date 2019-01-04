package metric

import (
	"time"
)

// Sample Data struct representing timestamp and metric value of metric data point
type Sample struct {
	Timestamp time.Time
	Value     string
}
