package clusterstatus

import (
	"github.com/containers-ai/alameda/datahub/pkg/utils"
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"strconv"
	"time"
)

const (
	ControllerName                         influxdb.Tag   = "name"
	ControllerNamespace                    influxdb.Tag   = "namespace"
	ControllerClusterName                  influxdb.Tag   = "cluster_name"
	ControllerUid                          influxdb.Tag   = "uid"
	ControllerKind                         influxdb.Tag   = "kind"
	ControllerAlamedaSpecScalerName        influxdb.Tag   = "alameda_scaler_name"
	ControllerAlamedaSpecScalerScalingTool influxdb.Tag   = "alameda_scaler_scaling_tool"
	ControllerReplicas                     influxdb.Field = "replicas"
	ControllerSpecReplicas                 influxdb.Field = "spec_replicas"
	ControllerAlamedaSpecPolicy            influxdb.Field = "policy"
	ControllerAlamedaSpecEnableExecution   influxdb.Field = "enable_execution"
)

var (
	// ControllerTags is list of tags of alameda_controller_recommendation measurement
	ControllerTags = []influxdb.Tag{
		ControllerName,
		ControllerNamespace,
		ControllerClusterName,
		ControllerUid,
		ControllerKind,
		ControllerAlamedaSpecScalerName,
		ControllerAlamedaSpecScalerScalingTool,
	}

	// ControllerFields is list of fields of alameda_controller_recommendation measurement
	ControllerFields = []influxdb.Field{
		ControllerReplicas,
		ControllerSpecReplicas,
		ControllerAlamedaSpecPolicy,
		ControllerAlamedaSpecEnableExecution,
	}

	ControllerColumns = []string{
		string(ControllerName),
		string(ControllerNamespace),
		string(ControllerClusterName),
		string(ControllerUid),
		string(ControllerKind),
		string(ControllerAlamedaSpecScalerName),
		string(ControllerAlamedaSpecScalerScalingTool),
		string(ControllerReplicas),
		string(ControllerSpecReplicas),
		string(ControllerAlamedaSpecPolicy),
		string(ControllerAlamedaSpecEnableExecution),
	}
)

type ControllerEntity struct {
	// InfluxDB tags
	Time                   time.Time
	Name                   string
	Namespace              string
	ClusterName            string
	Uid                    string
	Kind                   string
	AlamedaSpecScalerName  string
	AlamedaSpecScalingTool string

	// InfluxDB fields
	Replicas                   int32
	SpecReplicas               int32
	AlamedaSpecPolicy          string
	AlamedaSpecEnableExecution string
}

func NewControllerEntity(data map[string]string) *ControllerEntity {
	entity := ControllerEntity{}

	tempTimestamp, _ := utils.ParseTime(data["time"])
	entity.Time = tempTimestamp

	// InfluxDB tags
	if value, exist := data[string(ControllerName)]; exist {
		entity.Name = value
	}
	if value, exist := data[string(ControllerNamespace)]; exist {
		entity.Namespace = value
	}
	if value, exist := data[string(ControllerClusterName)]; exist {
		entity.ClusterName = value
	}
	if value, exist := data[string(ControllerUid)]; exist {
		entity.Uid = value
	}
	if value, exist := data[string(ControllerKind)]; exist {
		entity.Kind = value
	}
	if value, exist := data[string(ControllerAlamedaSpecScalerName)]; exist {
		entity.AlamedaSpecScalerName = value
	}
	if value, exist := data[string(ControllerAlamedaSpecScalerScalingTool)]; exist {
		entity.AlamedaSpecScalingTool = value
	}

	// InfluxDB fields
	if value, exist := data[string(ControllerReplicas)]; exist {
		valueInt64, _ := strconv.ParseInt(value, 10, 64)
		entity.Replicas = int32(valueInt64)
	}
	if value, exist := data[string(ControllerSpecReplicas)]; exist {
		valueInt64, _ := strconv.ParseInt(value, 10, 64)
		entity.SpecReplicas = int32(valueInt64)
	}
	if value, exist := data[string(ControllerAlamedaSpecPolicy)]; exist {
		entity.AlamedaSpecPolicy = value
	}
	if value, exist := data[string(ControllerAlamedaSpecEnableExecution)]; exist {
		entity.AlamedaSpecEnableExecution = value
	}

	return &entity
}

func (p *ControllerEntity) BuildInfluxPoint(measurement string) (*InfluxClient.Point, error) {
	// Pack influx tags
	tags := map[string]string{
		string(ControllerName):                         p.Name,
		string(ControllerNamespace):                    p.Namespace,
		string(ControllerClusterName):                  p.ClusterName,
		string(ControllerUid):                          p.Uid,
		string(ControllerKind):                         p.Kind,
		string(ControllerAlamedaSpecScalerName):        p.AlamedaSpecScalerName,
		string(ControllerAlamedaSpecScalerScalingTool): p.AlamedaSpecScalingTool,
	}

	// Pack influx fields
	fields := map[string]interface{}{
		string(ControllerReplicas):                   p.Replicas,
		string(ControllerSpecReplicas):               p.SpecReplicas,
		string(ControllerAlamedaSpecPolicy):          p.AlamedaSpecPolicy,
		string(ControllerAlamedaSpecEnableExecution): p.AlamedaSpecEnableExecution,
	}

	return InfluxClient.NewPoint(measurement, tags, fields, p.Time)
}
