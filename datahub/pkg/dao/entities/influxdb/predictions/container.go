package predictions

import (
	"github.com/containers-ai/alameda/datahub/pkg/utils"
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	"time"
)

const (
	ContainerTime        influxdb.Tag = "time"
	ContainerNamespace   influxdb.Tag = "namespace"
	ContainerPodName     influxdb.Tag = "pod_name"
	ContainerName        influxdb.Tag = "name"
	ContainerMetric      influxdb.Tag = "metric"
	ContainerKind        influxdb.Tag = "kind"
	ContainerGranularity influxdb.Tag = "granularity"

	ContainerModelId      influxdb.Field = "model_id"
	ContainerPredictionId influxdb.Field = "prediction_id"
	ContainerValue        influxdb.Field = "value"
)

var (
	ContainerTags   = []influxdb.Tag{ContainerNamespace, ContainerPodName, ContainerName, ContainerMetric, ContainerKind, ContainerGranularity}
	ContainerFields = []influxdb.Field{ContainerModelId, ContainerPredictionId, ContainerValue}
)

// Entity Container prediction entity in influxDB
type ContainerEntity struct {
	Time        time.Time
	Namespace   *string
	PodName     *string
	Name        *string
	Metric      *string
	Granularity *string
	Kind        *string

	ModelId      *string
	PredictionId *string
	Value        *string
}

// NewEntityFromMap Build entity from map
func NewContainerEntityFromMap(data map[string]string) ContainerEntity {
	// TODO: log error
	tempTimestamp, _ := utils.ParseTime(data[string(ContainerTime)])

	entity := ContainerEntity{
		Time: tempTimestamp,
	}

	// InfluxDB tags
	if namespace, exist := data[string(ContainerNamespace)]; exist {
		entity.Namespace = &namespace
	}
	if podName, exist := data[string(ContainerPodName)]; exist {
		entity.PodName = &podName
	}
	if name, exist := data[string(ContainerName)]; exist {
		entity.Name = &name
	}
	if metricData, exist := data[string(ContainerMetric)]; exist {
		entity.Metric = &metricData
	}
	if granularity, exist := data[string(ContainerGranularity)]; exist {
		entity.Granularity = &granularity
	}
	if kind, exist := data[string(ContainerKind)]; exist {
		entity.Kind = &kind
	}

	// InfluxDB fields
	if value, exist := data[string(ContainerModelId)]; exist {
		entity.ModelId = &value
	}
	if value, exist := data[string(ContainerPredictionId)]; exist {
		entity.PredictionId = &value
	}
	if value, exist := data[string(ContainerValue)]; exist {
		entity.Value = &value
	}

	return entity
}
