package clusterstatus

import (
	//"github.com/containers-ai/alameda/datahub/pkg/utils"
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	//ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	//"strconv"
	"time"
)

const (
	PodName                         influxdb.Tag = "name"
	PodNamespace                    influxdb.Tag = "namespace"
	PodNodeName                     influxdb.Tag = "node_name"
	PodClusterName                  influxdb.Tag = "cluster_name"
	PodUid                          influxdb.Tag = "uid"
	PodAlamedaSpecScalerName        influxdb.Tag = "alameda_scaler_name"
	PodAlamedaSpecScalerNamespace   influxdb.Tag = "alameda_scaler_namespace"
	PodAlamedaSpecScalerClusterName influxdb.Tag = "alameda_scaler_cluster_name"
	PodAppName                      influxdb.Tag = "app_name"
	PodAppPartOf                    influxdb.Tag = "app_part_of"

	PodCreateTime                       influxdb.Field = "pod_create_time"
	PodResourceLink                     influxdb.Field = "resource_link"
	PodTopControllerName                influxdb.Field = "top_controller_name"
	PodTopControllerKind                influxdb.Field = "top_controller_kind"
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
	PodAlamedaSpecScalingTool           influxdb.Field = "scaling_tool"
)

var (
	PodTags = []influxdb.Tag{
		PodName,
		PodNamespace,
		PodNodeName,
		PodClusterName,
		PodUid,
		PodAlamedaSpecScalerName,
		PodAlamedaSpecScalerNamespace,
		PodAlamedaSpecScalerClusterName,
		PodAppName,
		PodAppPartOf,
	}

	PodFields = []influxdb.Field{
		PodCreateTime,
		PodResourceLink,
		PodTopControllerName,
		PodTopControllerKind,
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
		PodAlamedaSpecScalingTool,
	}
)

type PodEntity struct {
	Time                         time.Time
	Name                         string
	Namespace                    string
	NodeName                     string
	ClusterName                  string
	Uid                          string
	AlamedaSpecScalerName        string
	AlamedaSpecScalerNamespace   string
	AlamedaSpecScalerClusterName string
	AppName                      string
	AppPartOf                    string

	CreateTime                       int64
	ResourceLink                     string
	TopControllerName                string
	TopControllerKind                string
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
	AlamedaSpecScalingTool           string
}

func (p *PodEntity) BuildInfluxPoint(measurement string) (*InfluxClient.Point, error) {
	// Pack influx tags
	tags := map[string]string{
		string(PodName):                         p.Name,
		string(PodNamespace):                    p.Namespace,
		string(PodNodeName):                     p.NodeName,
		string(PodClusterName):                  p.ClusterName,
		string(PodUid):                          p.Uid,
		string(PodAlamedaSpecScalerName):        p.AlamedaSpecScalerName,
		string(PodAlamedaSpecScalerNamespace):   p.AlamedaSpecScalerNamespace,
		string(PodAlamedaSpecScalerClusterName): p.AlamedaSpecScalerClusterName,
		string(PodAppName):                      p.AppName,
		string(PodAppPartOf):                    p.AppPartOf,
	}

	// Pack influx fields
	fields := map[string]interface{}{
		string(PodCreateTime):                       p.CreateTime,
		string(PodResourceLink):                     p.ResourceLink,
		string(PodTopControllerName):                p.TopControllerName,
		string(PodTopControllerKind):                p.TopControllerKind,
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
		string(PodAlamedaSpecScalingTool):           p.AlamedaSpecScalingTool,
	}

	return InfluxClient.NewPoint(measurement, tags, fields, p.Time)
}
