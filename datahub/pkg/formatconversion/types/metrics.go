package types

import (
	"time"
)

// Sample Data struct representing timestamp and metric value of metric data point
type Sample struct {
	Timestamp time.Time
	Value     string
}

type SamplesByAscTimestamp []Sample

func (d SamplesByAscTimestamp) Len() int {
	return len(d)
}
func (d SamplesByAscTimestamp) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}
func (d SamplesByAscTimestamp) Less(i, j int) bool {
	return d[i].Timestamp.Unix() < d[j].Timestamp.Unix()
}

type SamplesByDescTimestamp []Sample

func (d SamplesByDescTimestamp) Len() int {
	return len(d)
}
func (d SamplesByDescTimestamp) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}
func (d SamplesByDescTimestamp) Less(i, j int) bool {
	return d[i].Timestamp.Unix() > d[j].Timestamp.Unix()
}
