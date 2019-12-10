package clusterstatus

import (
	"github.com/containers-ai/alameda/datahub/pkg/utils"
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"time"
)

const (
	NamespaceName        influxdb.Tag   = "name"
	NamespaceClusterName influxdb.Tag   = "cluster_name"
	NamespaceUid         influxdb.Tag   = "uid"
	NamespaceValue       influxdb.Field = "value"
)

var (
	NamespaceTags = []influxdb.Tag{
		NamespaceName,
		NamespaceClusterName,
		NamespaceUid,
	}

	NamespaceFields = []influxdb.Field{
		NamespaceValue,
	}

	NamespaceColumns = []string{
		string(NamespaceName),
		string(NamespaceClusterName),
		string(NamespaceUid),
		string(NamespaceValue),
	}
)

type NamespaceEntity struct {
	// InfluxDB tags
	Time        time.Time
	Name        string
	ClusterName string
	Uid         string

	// InfluxDB fields
	Value string
}

func NewNamespaceEntity(data map[string]string) *NamespaceEntity {
	entity := NamespaceEntity{}

	tempTimestamp, _ := utils.ParseTime(data["time"])
	entity.Time = tempTimestamp

	// InfluxDB tags
	if value, exist := data[string(NamespaceName)]; exist {
		entity.Name = value
	}
	if value, exist := data[string(NamespaceClusterName)]; exist {
		entity.ClusterName = value
	}
	if value, exist := data[string(NamespaceUid)]; exist {
		entity.Uid = value
	}

	// InfluxDB fields
	if value, exist := data[string(NamespaceValue)]; exist {
		entity.Value = value
	}

	return &entity
}

func (p *NamespaceEntity) BuildInfluxPoint(measurement string) (*InfluxClient.Point, error) {
	// Pack influx tags
	tags := map[string]string{
		string(NamespaceName):        p.Name,
		string(NamespaceClusterName): p.ClusterName,
		string(NamespaceUid):         p.Uid,
	}

	// Pack influx fields
	fields := map[string]interface{}{
		string(NamespaceValue): p.Value,
	}

	return InfluxClient.NewPoint(measurement, tags, fields, p.Time)
}
