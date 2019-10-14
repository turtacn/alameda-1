package metric

import (
	"github.com/containers-ai/alameda/datahub/pkg/utils"
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	"strconv"
	"time"
)

const (
	NodeTime influxdb.Tag = "time"
	NodeName influxdb.Tag = "name"

	NodeValue influxdb.Field = "value"
)

var (
	NodeTags    = []influxdb.Tag{NodeName}
	NodeFields  = []influxdb.Field{NodeValue}
	NodeColumns = []string{string(NodeName), string(NodeValue)}
)

type NodeEntity struct {
	Time time.Time
	Name *string

	Value *float64
}

func NewNodeEntityFromMap(data map[string]string) NodeEntity {
	tempTimestamp, _ := utils.ParseTime(data[string(NodeTime)])

	entity := NodeEntity{
		Time: tempTimestamp,
	}

	// InfluxDB tags
	if valueStr, exist := data[string(NodeName)]; exist {
		entity.Name = &valueStr
	}

	// InfluxDB fields
	if valueFloat, exist := data[string(NodeValue)]; exist {
		value, _ := strconv.ParseFloat(valueFloat, 64)
		entity.Value = &value
	}

	return entity
}
