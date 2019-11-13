package metrics

import (
	"github.com/containers-ai/alameda/datahub/pkg/utils"
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	"strconv"
	"time"
)

const (
	NamespaceTime        influxdb.Tag = "time"
	NamespaceName        influxdb.Tag = "name"
	NamespaceClusterName influxdb.Tag = "cluster_name"
	NamespaceUID         influxdb.Tag = "uid"

	NamespaceValue influxdb.Field = "value"
)

var (
	NamespaceTags    = []influxdb.Tag{NamespaceName, NamespaceClusterName, NamespaceUID}
	NamespaceFields  = []influxdb.Field{NamespaceValue}
	NamespaceColumns = []string{string(NamespaceName), string(NamespaceClusterName), string(NamespaceUID), string(NamespaceValue)}
)

type NamespaceEntity struct {
	Time        time.Time
	Name        *string
	ClusterName *string
	UID         *string

	Value *float64
}

func NewNamespaceEntityFromMap(data map[string]string) NamespaceEntity {
	tempTimestamp, _ := utils.ParseTime(data[string(NamespaceTime)])

	entity := NamespaceEntity{
		Time: tempTimestamp,
	}

	// InfluxDB tags
	if valueStr, exist := data[string(NamespaceName)]; exist {
		entity.Name = &valueStr
	}
	if valueStr, exist := data[string(NamespaceClusterName)]; exist {
		entity.ClusterName = &valueStr
	}
	if valueStr, exist := data[string(NamespaceUID)]; exist {
		entity.UID = &valueStr
	}

	// InfluxDB fields
	if valueFloat, exist := data[string(NamespaceValue)]; exist {
		value, _ := strconv.ParseFloat(valueFloat, 64)
		entity.Value = &value
	}

	return entity
}
