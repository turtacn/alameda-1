package predictions

import (
	"github.com/containers-ai/alameda/datahub/pkg/utils"
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	"time"
)

const (
	ContainerTime        influxdb.Tag = "time"
	ContainerName        influxdb.Tag = "name"
	ContainerPodName     influxdb.Tag = "pod_name"
	ContainerNamespace   influxdb.Tag = "namespace"
	ContainerNodeName    influxdb.Tag = "node_name"
	ContainerClusterName influxdb.Tag = "cluster_name"
	ContainerMetric      influxdb.Tag = "metric"
	ContainerMetricType  influxdb.Tag = "kind"
	ContainerGranularity influxdb.Tag = "granularity"

	ContainerModelId      influxdb.Field = "model_id"
	ContainerPredictionId influxdb.Field = "prediction_id"
	ContainerValue        influxdb.Field = "value"
)

var (
	ContainerTags = []influxdb.Tag{
		ContainerName,
		ContainerPodName,
		ContainerNamespace,
		ContainerNodeName,
		ContainerClusterName,
		ContainerMetric,
		ContainerMetricType,
		ContainerGranularity,
	}

	ContainerFields = []influxdb.Field{
		ContainerModelId,
		ContainerPredictionId,
		ContainerValue,
	}
)

// Entity Container prediction entity in influxDB
type ContainerEntity struct {
	Time        time.Time
	Name        *string
	PodName     *string
	Namespace   *string
	NodeName    *string
	ClusterName *string
	Metric      *string
	MetricType  *string
	Granularity *string

	ModelId      *string
	PredictionId *string
	Value        *string
}

// NewEntityFromMap Build entity from map
func NewContainerEntity(data map[string]string) ContainerEntity {
	entity := ContainerEntity{}

	// TODO: log error
	tempTimestamp, _ := utils.ParseTime(data[string(ContainerTime)])
	entity.Time = tempTimestamp

	// InfluxDB tags
	if value, exist := data[string(ContainerName)]; exist {
		entity.Name = &value
	}
	if value, exist := data[string(ContainerPodName)]; exist {
		entity.PodName = &value
	}
	if value, exist := data[string(ContainerNamespace)]; exist {
		entity.Namespace = &value
	}
	if value, exist := data[string(ContainerNodeName)]; exist {
		entity.NodeName = &value
	}
	if value, exist := data[string(ContainerClusterName)]; exist {
		entity.ClusterName = &value
	}
	if value, exist := data[string(ContainerMetric)]; exist {
		entity.Metric = &value
	}
	if value, exist := data[string(ContainerMetricType)]; exist {
		entity.MetricType = &value
	}
	if granularity, exist := data[string(ContainerGranularity)]; exist {
		entity.Granularity = &granularity
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
