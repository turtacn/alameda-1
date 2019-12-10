package clusterstatus

import (
	"github.com/containers-ai/alameda/datahub/pkg/utils"
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"time"
)

const (
	ClusterName  influxdb.Tag   = "name"
	ClusterUid   influxdb.Tag   = "uid"
	ClusterValue influxdb.Field = "value"
)

var (
	ClusterTags = []influxdb.Tag{
		ClusterName,
		ClusterUid,
	}

	ClusterFields = []influxdb.Field{
		ClusterValue,
	}

	ClusterColumns = []string{
		string(ClusterName),
		string(ClusterUid),
		string(ClusterValue),
	}
)

type ClusterEntity struct {
	Time  time.Time
	Name  string
	Uid   string
	Value string
}

func NewClusterEntity(data map[string]string) *ClusterEntity {
	entity := ClusterEntity{}

	tempTimestamp, _ := utils.ParseTime(data["time"])
	entity.Time = tempTimestamp

	// InfluxDB tags
	if value, exist := data[string(ClusterName)]; exist {
		entity.Name = value
	}
	if value, exist := data[string(ClusterUid)]; exist {
		entity.Uid = value
	}

	// InfluxDB fields
	if value, exist := data[string(ClusterValue)]; exist {
		entity.Value = value
	}

	return &entity
}

func (p *ClusterEntity) BuildInfluxPoint(measurement string) (*InfluxClient.Point, error) {
	// Pack influx tags
	tags := map[string]string{
		string(ClusterName): p.Name,
		string(ClusterUid):  p.Uid,
	}

	// Pack influx fields
	fields := map[string]interface{}{
		string(ClusterValue): p.Value,
	}

	return InfluxClient.NewPoint(measurement, tags, fields, p.Time)
}
