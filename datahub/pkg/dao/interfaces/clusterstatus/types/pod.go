package types

import (
	//"fmt"
	"github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/clusterstatus"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	"github.com/containers-ai/alameda/internal/pkg/database/common"
	ApiCommon "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	"github.com/golang/protobuf/ptypes/timestamp"
	"strconv"
	"strings"
)

// ContainerOperation provides container measurement operations
type PodDAO interface {
	CreatePods([]*Pod) error
	ListPods(ListPodsRequest) ([]*Pod, error)
	//DeletePods([]*resources.Pod) error
}

type Pod struct {
	ObjectMeta     *metadata.ObjectMeta
	CreateTime     *timestamp.Timestamp
	ResourceLink   string
	AppName        string
	AppPartOf      string
	Containers     []*Container
	TopController  *Controller
	Status         *PodStatus
	AlamedaPodSpec *AlamedaPodSpec
}

type ListPodsRequest struct {
	common.QueryCondition
	ObjectMeta  []metadata.ObjectMeta
	Kind        string // Valid values: POD, DEPLOYMENT, DEPLOYMENTCONFIG, ALAMEDASCALER, STATEFULSET
	ScalingTool string // Valid values: NONE, VPA, HPA
}

type AlamedaPodSpec struct {
	AlamedaScaler          *metadata.ObjectMeta
	Policy                 string
	UsedRecommendationId   string
	AlamedaScalerResources *ResourceRequirements
	ScalingTool            string
}

type PodStatus struct {
	Phase   string
	Message string
	Reason  string
}

func NewPod() *Pod {
	pod := Pod{}
	pod.Containers = make([]*Container, 0)
	pod.TopController = NewController()
	return &pod
}

func NewListPodsRequest() ListPodsRequest {
	request := ListPodsRequest{}
	request.ObjectMeta = make([]metadata.ObjectMeta, 0)
	return request
}

func (p *Pod) Initialize(values map[string]string) {
	p.ObjectMeta = &metadata.ObjectMeta{}
	p.ObjectMeta.Initialize(values)
	if value, ok := values[string(clusterstatus.PodCreateTime)]; ok {
		valueInt64, _ := strconv.ParseInt(value, 10, 64)
		p.CreateTime = &timestamp.Timestamp{Seconds: valueInt64}
	}
	p.ResourceLink = values[string(clusterstatus.PodResourceLink)]
	p.AppName = values[string(clusterstatus.PodAppName)]
	p.AppPartOf = values[string(clusterstatus.PodAppPartOf)]
	p.Containers = make([]*Container, 0)

	// Build top controller
	p.TopController.ObjectMeta.Name = values[string(clusterstatus.PodTopControllerName)]
	p.TopController.Kind = values[string(clusterstatus.PodTopControllerKind)]
	if value, ok := values[string(clusterstatus.PodTopControllerReplicas)]; ok {
		valueInt64, _ := strconv.ParseInt(value, 10, 64)
		p.TopController.Replicas = int32(valueInt64)
	}

	// Build status
	if value, ok := values[string(clusterstatus.PodStatusPhase)]; ok {
		if p.Status == nil {
			p.Status = &PodStatus{}
		}
		p.Status.Phase = value
	}
	if value, ok := values[string(clusterstatus.PodStatusMessage)]; ok {
		if p.Status == nil {
			p.Status = &PodStatus{}
		}
		p.Status.Message = value
	}
	if value, ok := values[string(clusterstatus.PodStatusReason)]; ok {
		if p.Status == nil {
			p.Status = &PodStatus{}
		}
		p.Status.Reason = value
	}

	// Build alameda pod spec
	p.AlamedaPodSpec = &AlamedaPodSpec{}
	p.AlamedaPodSpec.AlamedaScaler = &metadata.ObjectMeta{}
	if value, ok := values[string(clusterstatus.PodAlamedaSpecScalerName)]; ok {
		p.AlamedaPodSpec.AlamedaScaler.Name = value
	}
	if value, ok := values[string(clusterstatus.PodAlamedaSpecScalerNamespace)]; ok {
		p.AlamedaPodSpec.AlamedaScaler.Namespace = value
	}
	if value, ok := values[string(clusterstatus.PodAlamedaSpecScalerClusterName)]; ok {
		p.AlamedaPodSpec.AlamedaScaler.ClusterName = value
	}
	if value, ok := values[string(clusterstatus.PodAlamedaSpecPolicy)]; ok {
		p.AlamedaPodSpec.Policy = value
	}
	if value, ok := values[string(clusterstatus.PodAlamedaSpecUsedRecommendationID)]; ok {
		p.AlamedaPodSpec.UsedRecommendationId = value
	}
	if value, ok := values[string(clusterstatus.PodAlamedaSpecScalingTool)]; ok {
		p.AlamedaPodSpec.ScalingTool = value
	}
	p.AlamedaPodSpec.AlamedaScalerResources = &ResourceRequirements{}
	p.AlamedaPodSpec.AlamedaScalerResources.Limits = make(map[int32]string)
	p.AlamedaPodSpec.AlamedaScalerResources.Requests = make(map[int32]string)
	if value, ok := values[string(clusterstatus.PodAlamedaSpecResourceLimitCPU)]; ok {
		if value != "" {
			p.AlamedaPodSpec.AlamedaScalerResources.Limits[int32(ApiCommon.ResourceName_CPU)] = value
		}
	}
	if value, ok := values[string(clusterstatus.PodAlamedaSpecResourceLimitMemory)]; ok {
		if value != "" {
			p.AlamedaPodSpec.AlamedaScalerResources.Limits[int32(ApiCommon.ResourceName_MEMORY)] = value
		}
	}
	if value, ok := values[string(clusterstatus.PodAlamedaSpecResourceRequestCPU)]; ok {
		if value != "" {
			p.AlamedaPodSpec.AlamedaScalerResources.Requests[int32(ApiCommon.ResourceName_CPU)] = value
		}
	}
	if value, ok := values[string(clusterstatus.PodAlamedaSpecResourceRequestMemory)]; ok {
		if value != "" {
			p.AlamedaPodSpec.AlamedaScalerResources.Requests[int32(ApiCommon.ResourceName_MEMORY)] = value
		}
	}
	if value, ok := values[string(clusterstatus.PodAlamedaSpecScalingTool)]; ok {
		p.AlamedaPodSpec.ScalingTool = value
	}
}

func (p *Pod) ClusterNamespacePodName() string {
	if p.ObjectMeta != nil {
		valueList := make([]string, 0)
		if p.ObjectMeta.ClusterName != "" {
			valueList = append(valueList, p.ObjectMeta.ClusterName)
		}
		if p.ObjectMeta.Namespace != "" {
			valueList = append(valueList, p.ObjectMeta.Namespace)
		}
		if p.ObjectMeta.Name != "" {
			valueList = append(valueList, p.ObjectMeta.Name)
		}
		return strings.Join(valueList, "/")
	}
	return ""
}

func (p *PodStatus) Initialize(values map[string]string) {
}
