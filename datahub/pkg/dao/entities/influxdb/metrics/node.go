package metrics

import (
	"github.com/containers-ai/alameda/datahub/pkg/utils"
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	"strconv"
	"time"
)

const (
	NodeTime        influxdb.Tag = "time"
	NodeName        influxdb.Tag = "name"
	NodeClusterName influxdb.Tag = "cluster_name"
	NodeUID         influxdb.Tag = "uid"

	NodeValue influxdb.Field = "value"
)

var (
	NodeTags    = []influxdb.Tag{NodeName, NodeClusterName, NodeUID}
	NodeFields  = []influxdb.Field{NodeValue}
	NodeColumns = []string{string(NodeName), string(NodeClusterName), string(NodeUID), string(NodeValue)}
)

type NodeEntity struct {
	Time        time.Time
	Name        *string
	ClusterName *string
	UID         *string

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
	if valueStr, exist := data[string(NodeClusterName)]; exist {
		entity.ClusterName = &valueStr
	}
	if valueStr, exist := data[string(NodeUID)]; exist {
		entity.UID = &valueStr
	}

	// InfluxDB fields
	if valueFloat, exist := data[string(NodeValue)]; exist {
		value, _ := strconv.ParseFloat(valueFloat, 64)
		entity.Value = &value
	}

	return entity
}
