package predictions

import (
	"github.com/containers-ai/alameda/datahub/pkg/utils"
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	"time"
)

const (
	ApplicationTime        influxdb.Tag = "time"
	ApplicationName        influxdb.Tag = "name"
	ApplicationMetric      influxdb.Tag = "metric"
	ApplicationGranularity influxdb.Tag = "granularity"
	ApplicationKind        influxdb.Tag = "kind"

	ApplicationModelId      influxdb.Field = "model_id"
	ApplicationPredictionId influxdb.Field = "prediction_id"
	ApplicationValue        influxdb.Field = "value"
)

var (
	ApplicationTags   = []influxdb.Tag{ApplicationName, ApplicationMetric, ApplicationGranularity}
	ApplicationFields = []influxdb.Field{ApplicationModelId, ApplicationPredictionId, ApplicationValue}
)

// Entity Container prediction entity in influxDB
type ApplicationEntity struct {
	Time        time.Time
	Name        *string
	Metric      *string
	Granularity *string
	Kind        *string

	ModelId      *string
	PredictionId *string
	Value        *string
}

// NewEntityFromMap Build entity from map
func NewApplicationEntityFromMap(data map[string]string) ApplicationEntity {
	// TODO: log error
	tempTimestamp, _ := utils.ParseTime(data[string(ApplicationTime)])

	entity := ApplicationEntity{
		Time: tempTimestamp,
	}

	// InfluxDB tags
	if name, exist := data[string(ApplicationName)]; exist {
		entity.Name = &name
	}
	if metricData, exist := data[string(ApplicationMetric)]; exist {
		entity.Metric = &metricData
	}
	if valueStr, exist := data[string(ApplicationGranularity)]; exist {
		entity.Granularity = &valueStr
	}
	if valueStr, exist := data[string(ApplicationKind)]; exist {
		entity.Kind = &valueStr
	}

	// InfluxDB fields
	if value, exist := data[string(ApplicationModelId)]; exist {
		entity.ModelId = &value
	}
	if value, exist := data[string(ApplicationPredictionId)]; exist {
		entity.PredictionId = &value
	}
	if value, exist := data[string(ApplicationValue)]; exist {
		entity.Value = &value
	}

	return entity
}
