package metrics

import (
	"github.com/containers-ai/alameda/datahub/pkg/utils"
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	"strconv"
	"time"
)

const (
	ApplicationTime        influxdb.Tag = "time"
	ApplicationName        influxdb.Tag = "name"
	ApplicationNamespace   influxdb.Tag = "namespace"
	ApplicationClusterName influxdb.Tag = "cluster_name"
	ApplicationUID         influxdb.Tag = "uid"

	ApplicationValue influxdb.Field = "value"
)

var (
	ApplicationTags    = []influxdb.Tag{ApplicationName, ApplicationNamespace, ApplicationClusterName, ApplicationUID}
	ApplicationFields  = []influxdb.Field{ApplicationValue}
	ApplicationColumns = []string{string(ApplicationName), string(ApplicationNamespace), string(ApplicationClusterName), string(ApplicationUID), string(ApplicationValue)}
)

type ApplicationEntity struct {
	Time        time.Time
	Name        *string
	Namespace   *string
	ClusterName *string
	UID         *string

	Value *float64
}

func NewApplicationEntityFromMap(data map[string]string) ApplicationEntity {
	tempTimestamp, _ := utils.ParseTime(data[string(ApplicationTime)])

	entity := ApplicationEntity{
		Time: tempTimestamp,
	}

	// InfluxDB tags
	if valueStr, exist := data[string(ApplicationName)]; exist {
		entity.Name = &valueStr
	}
	if valueStr, exist := data[string(ApplicationNamespace)]; exist {
		entity.Namespace = &valueStr
	}
	if valueStr, exist := data[string(ApplicationClusterName)]; exist {
		entity.ClusterName = &valueStr
	}
	if valueStr, exist := data[string(ApplicationUID)]; exist {
		entity.UID = &valueStr
	}

	// InfluxDB fields
	if valueFloat, exist := data[string(ApplicationValue)]; exist {
		value, _ := strconv.ParseFloat(valueFloat, 64)
		entity.Value = &value
	}

	return entity
}
