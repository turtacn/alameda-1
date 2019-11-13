package metrics

import (
	"github.com/containers-ai/alameda/datahub/pkg/utils"
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	"strconv"
	"time"
)

const (
	ClusterTime influxdb.Tag = "time"
	ClusterName influxdb.Tag = "name"
	ClusterUID  influxdb.Tag = "uid"

	ClusterValue influxdb.Field = "value"
)

var (
	ClusterTags    = []influxdb.Tag{ClusterName, ClusterUID}
	ClusterFields  = []influxdb.Field{ClusterValue}
	ClusterColumns = []string{string(ClusterName), string(ClusterUID), string(ClusterValue)}
)

type ClusterEntity struct {
	Time time.Time
	Name *string
	UID  *string

	Value *float64
}

func NewClusterEntityFromMap(data map[string]string) ClusterEntity {
	tempTimestamp, _ := utils.ParseTime(data[string(ClusterTime)])

	entity := ClusterEntity{
		Time: tempTimestamp,
	}

	// InfluxDB tags
	if valueStr, exist := data[string(ClusterName)]; exist {
		entity.Name = &valueStr
	}
	if valueStr, exist := data[string(ClusterUID)]; exist {
		entity.UID = &valueStr
	}

	// InfluxDB fields
	if valueFloat, exist := data[string(ClusterValue)]; exist {
		value, _ := strconv.ParseFloat(valueFloat, 64)
		entity.Value = &value
	}

	return entity
}
