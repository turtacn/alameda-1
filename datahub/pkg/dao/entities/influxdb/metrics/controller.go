package metrics

import (
	"github.com/containers-ai/alameda/datahub/pkg/utils"
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	"strconv"
	"time"
)

const (
	ControllerTime        influxdb.Tag = "time"
	ControllerName        influxdb.Tag = "name"
	ControllerNamespace   influxdb.Tag = "namespace"
	ControllerClusterName influxdb.Tag = "cluster_name"
	ControllerKind        influxdb.Tag = "kind"
	ControllerUID         influxdb.Tag = "uid"

	ControllerValue influxdb.Field = "value"
)

var (
	ControllerTags    = []influxdb.Tag{ControllerName, ControllerNamespace, ControllerClusterName, ControllerKind, ControllerUID}
	ControllerFields  = []influxdb.Field{ControllerValue}
	ControllerColumns = []string{string(ControllerName), string(ControllerNamespace), string(ControllerClusterName),
		string(ControllerKind), string(ControllerUID), string(ControllerValue)}
)

type ControllerEntity struct {
	Time        time.Time
	Name        *string
	Namespace   *string
	ClusterName *string
	Kind        *string
	UID         *string

	Value *float64
}

func NewControllerEntityFromMap(data map[string]string) ControllerEntity {
	tempTimestamp, _ := utils.ParseTime(data[string(ControllerTime)])

	entity := ControllerEntity{
		Time: tempTimestamp,
	}

	// InfluxDB tags
	if valueStr, exist := data[string(ControllerName)]; exist {
		entity.Name = &valueStr
	}
	if valueStr, exist := data[string(ControllerNamespace)]; exist {
		entity.Namespace = &valueStr
	}
	if valueStr, exist := data[string(ControllerClusterName)]; exist {
		entity.ClusterName = &valueStr
	}
	if valueStr, exist := data[string(ControllerKind)]; exist {
		entity.Kind = &valueStr
	}
	if valueStr, exist := data[string(ControllerUID)]; exist {
		entity.UID = &valueStr
	}

	// InfluxDB fields
	if valueFloat, exist := data[string(ControllerValue)]; exist {
		value, _ := strconv.ParseFloat(valueFloat, 64)
		entity.Value = &value
	}

	return entity
}
