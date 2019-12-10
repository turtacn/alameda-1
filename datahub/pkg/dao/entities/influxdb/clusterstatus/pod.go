package clusterstatus

import (
	"github.com/containers-ai/alameda/datahub/pkg/utils"
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"strconv"
	"time"
)

const (
	PodName                         influxdb.Tag = "name"
	PodNamespace                    influxdb.Tag = "namespace"
	PodNodeName                     influxdb.Tag = "node_name"
	PodClusterName                  influxdb.Tag = "cluster_name"
	PodUid                          influxdb.Tag = "uid"
	PodTopControllerName            influxdb.Tag = "top_controller_name"
	PodTopControllerKind            influxdb.Tag = "top_controller_kind"
	PodAlamedaSpecScalerName        influxdb.Tag = "alameda_scaler_name"
	PodAlamedaSpecScalerScalingTool influxdb.Tag = "alameda_scaler_scaling_tool"
	PodAppName                      influxdb.Tag = "app_name"
	PodAppPartOf                    influxdb.Tag = "app_part_of"

	PodCreateTime                       influxdb.Field = "pod_create_time"
	PodResourceLink                     influxdb.Field = "resource_link"
	PodTopControllerReplicas            influxdb.Field = "top_controller_replicas"
	PodStatusPhase                      influxdb.Field = "pod_phase"
	PodStatusMessage                    influxdb.Field = "pod_message"
	PodStatusReason                     influxdb.Field = "pod_reason"
	PodAlamedaSpecPolicy                influxdb.Field = "policy"
	PodAlamedaSpecUsedRecommendationID  influxdb.Field = "used_recommendation_id"
	PodAlamedaSpecResourceLimitCPU      influxdb.Field = "alameda_scaler_resource_limit_cpu"
	PodAlamedaSpecResourceLimitMemory   influxdb.Field = "alameda_scaler_resource_limit_memory"
	PodAlamedaSpecResourceRequestCPU    influxdb.Field = "alameda_scaler_resource_request_cpu"
	PodAlamedaSpecResourceRequestMemory influxdb.Field = "alameda_scaler_resource_request_memory"
)

var (
	PodTags = []influxdb.Tag{
		PodName,
		PodNamespace,
		PodNodeName,
		PodClusterName,
		PodUid,
		PodTopControllerName,
		PodTopControllerKind,
		PodAlamedaSpecScalerName,
		PodAlamedaSpecScalerScalingTool,
		PodAppName,
		PodAppPartOf,
	}

	PodFields = []influxdb.Field{
		PodCreateTime,
		PodResourceLink,
		PodTopControllerReplicas,
		PodStatusPhase,
		PodStatusMessage,
		PodStatusReason,
		PodAlamedaSpecPolicy,
		PodAlamedaSpecUsedRecommendationID,
		PodAlamedaSpecResourceLimitCPU,
		PodAlamedaSpecResourceLimitMemory,
		PodAlamedaSpecResourceRequestCPU,
		PodAlamedaSpecResourceRequestMemory,
	}
)

type PodEntity struct {
	Time                         time.Time
	Name                         string
	Namespace                    string
	NodeName                     string
	ClusterName                  string
	Uid                          string
	TopControllerName            string
	TopControllerKind            string
	AlamedaSpecScalerName        string
	AlamedaSpecScalerScalingTool string
	AppName                      string
	AppPartOf                    string

	CreateTime                       int64
	ResourceLink                     string
	TopControllerReplicas            int32
	StatusPhase                      string
	StatusMessage                    string
	StatusReason                     string
	AlamedaSpecPolicy                string
	AlamedaSpecUsedRecommendationID  string
	AlamedaSpecResourceLimitCPU      string // TODO: check if type string or float64
	AlamedaSpecResourceLimitMemory   string // TODO: check if type string or float64
	AlamedaSpecResourceRequestCPU    string // TODO: check if type string or float64
	AlamedaSpecResourceRequestMemory string // TODO: check if type string or float64
}

func NewPodEntity(data map[string]string) *PodEntity {
	entity := PodEntity{}

	tempTimestamp, _ := utils.ParseTime(data["time"])
	entity.Time = tempTimestamp

	// InfluxDB tags
	if value, exist := data[string(PodName)]; exist {
		entity.Name = value
	}
	if value, exist := data[string(PodNamespace)]; exist {
		entity.Namespace = value
	}
	if value, exist := data[string(PodNodeName)]; exist {
		entity.NodeName = value
	}
	if value, exist := data[string(PodClusterName)]; exist {
		entity.ClusterName = value
	}
	if value, exist := data[string(PodUid)]; exist {
		entity.Uid = value
	}
	if value, exist := data[string(PodTopControllerName)]; exist {
		entity.TopControllerName = value
	}
	if value, exist := data[string(PodTopControllerKind)]; exist {
		entity.TopControllerKind = value
	}
	if value, exist := data[string(PodAlamedaSpecScalerName)]; exist {
		entity.AlamedaSpecScalerName = value
	}
	if value, exist := data[string(PodAlamedaSpecScalerScalingTool)]; exist {
		entity.AlamedaSpecScalerScalingTool = value
	}
	if value, exist := data[string(PodAppName)]; exist {
		entity.AppName = value
	}
	if value, exist := data[string(PodAppPartOf)]; exist {
		entity.AppPartOf = value
	}

	// InfluxDB fields
	if value, exist := data[string(PodCreateTime)]; exist {
		valueInt64, _ := strconv.ParseInt(value, 10, 64)
		entity.CreateTime = valueInt64
	}
	if value, exist := data[string(PodResourceLink)]; exist {
		entity.ResourceLink = value
	}
	if value, exist := data[string(PodTopControllerReplicas)]; exist {
		valueInt64, _ := strconv.ParseInt(value, 10, 64)
		entity.TopControllerReplicas = int32(valueInt64)
	}
	if value, exist := data[string(PodStatusPhase)]; exist {
		entity.StatusPhase = value
	}
	if value, exist := data[string(PodStatusMessage)]; exist {
		entity.StatusMessage = value
	}
	if value, exist := data[string(PodStatusReason)]; exist {
		entity.StatusReason = value
	}
	if value, exist := data[string(PodAlamedaSpecPolicy)]; exist {
		entity.AlamedaSpecPolicy = value
	}
	if value, exist := data[string(PodAlamedaSpecUsedRecommendationID)]; exist {
		entity.AlamedaSpecUsedRecommendationID = value
	}
	if value, exist := data[string(PodAlamedaSpecResourceLimitCPU)]; exist {
		entity.AlamedaSpecResourceLimitCPU = value
	}
	if value, exist := data[string(PodAlamedaSpecResourceLimitMemory)]; exist {
		entity.AlamedaSpecResourceLimitMemory = value
	}
	if value, exist := data[string(PodAlamedaSpecResourceRequestCPU)]; exist {
		entity.AlamedaSpecResourceRequestCPU = value
	}
	if value, exist := data[string(PodAlamedaSpecResourceRequestMemory)]; exist {
		entity.AlamedaSpecResourceRequestMemory = value
	}

	return &entity
}

func (p *PodEntity) BuildInfluxPoint(measurement string) (*InfluxClient.Point, error) {
	// Pack influx tags
	tags := map[string]string{
		string(PodName):                         p.Name,
		string(PodNamespace):                    p.Namespace,
		string(PodNodeName):                     p.NodeName,
		string(PodClusterName):                  p.ClusterName,
		string(PodUid):                          p.Uid,
		string(PodTopControllerName):            p.TopControllerName,
		string(PodTopControllerKind):            p.TopControllerKind,
		string(PodAlamedaSpecScalerName):        p.AlamedaSpecScalerName,
		string(PodAlamedaSpecScalerScalingTool): p.AlamedaSpecScalerScalingTool,
		string(PodAppName):                      p.AppName,
		string(PodAppPartOf):                    p.AppPartOf,
	}

	// Pack influx fields
	fields := map[string]interface{}{
		string(PodCreateTime):                       p.CreateTime,
		string(PodResourceLink):                     p.ResourceLink,
		string(PodTopControllerReplicas):            p.TopControllerReplicas,
		string(PodStatusPhase):                      p.StatusPhase,
		string(PodStatusMessage):                    p.StatusMessage,
		string(PodStatusReason):                     p.StatusReason,
		string(PodAlamedaSpecPolicy):                p.AlamedaSpecPolicy,
		string(PodAlamedaSpecUsedRecommendationID):  p.AlamedaSpecUsedRecommendationID,
		string(PodAlamedaSpecResourceLimitCPU):      p.AlamedaSpecResourceLimitCPU,
		string(PodAlamedaSpecResourceLimitMemory):   p.AlamedaSpecResourceLimitMemory,
		string(PodAlamedaSpecResourceRequestCPU):    p.AlamedaSpecResourceRequestCPU,
		string(PodAlamedaSpecResourceRequestMemory): p.AlamedaSpecResourceRequestMemory,
	}

	return InfluxClient.NewPoint(measurement, tags, fields, p.Time)
}
