package models

import (
	Client "github.com/influxdata/influxdb/client/v2"
)

type InfluxResultExtend struct {
	Client.Result
}

func NewInfluxResults(results []Client.Result) []*InfluxResultExtend {
	data := make([]*InfluxResultExtend, 0)

	for i := 0; i < len(results); i++ {
		data = append(data, &InfluxResultExtend{results[0]})
	}

	return data
}

func (s *InfluxResultExtend) GetGroupNum() int {
	return len(s.Series)
}

func (s *InfluxResultExtend) GetGroup(index int) *InfluxGroup {
	group := InfluxGroup{s.Series[index]}
	return &group
}
