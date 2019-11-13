package types

import (
	//"fmt"
	"github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/clusterstatus"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	"github.com/containers-ai/alameda/internal/pkg/database/common"
	ApiCommon "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	"github.com/golang/protobuf/ptypes/timestamp"
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
	ScalingTool            string
	AlamedaScalerResources *ResourceRequirements
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

func (p *Pod) Initialize(entity *clusterstatus.PodEntity) {
	p.ObjectMeta = &metadata.ObjectMeta{}
	p.ObjectMeta.Name = entity.Name
	p.ObjectMeta.Namespace = entity.Namespace
	p.ObjectMeta.NodeName = entity.NodeName
	p.ObjectMeta.ClusterName = entity.ClusterName
	p.ObjectMeta.Uid = entity.Uid
	p.CreateTime = &timestamp.Timestamp{Seconds: entity.CreateTime}
	p.ResourceLink = entity.ResourceLink
	p.AppName = entity.AppName
	p.AppPartOf = entity.AppPartOf

	// Build TopController
	p.TopController = &Controller{}
	p.TopController.ObjectMeta.Name = entity.TopControllerName
	p.TopController.Kind = entity.TopControllerKind
	p.TopController.Replicas = entity.TopControllerReplicas

	// Build Status
	p.Status = &PodStatus{}
	p.Status.Phase = entity.StatusPhase
	p.Status.Message = entity.StatusMessage
	p.Status.Reason = entity.StatusReason

	// Build AlamedaPodSpec
	p.AlamedaPodSpec = &AlamedaPodSpec{}
	p.AlamedaPodSpec.AlamedaScaler = &metadata.ObjectMeta{}
	p.AlamedaPodSpec.AlamedaScaler.Name = entity.AlamedaSpecScalerName
	p.AlamedaPodSpec.AlamedaScaler.Namespace = entity.AlamedaSpecScalerNamespace
	p.AlamedaPodSpec.AlamedaScaler.ClusterName = entity.AlamedaSpecScalerClusterName
	p.AlamedaPodSpec.ScalingTool = entity.AlamedaSpecScalingTool
	p.AlamedaPodSpec.Policy = entity.AlamedaSpecPolicy
	p.AlamedaPodSpec.UsedRecommendationId = entity.AlamedaSpecUsedRecommendationID
	p.AlamedaPodSpec.AlamedaScalerResources = &ResourceRequirements{}
	p.AlamedaPodSpec.AlamedaScalerResources.Limits = make(map[int32]string)
	p.AlamedaPodSpec.AlamedaScalerResources.Requests = make(map[int32]string)
	if entity.AlamedaSpecResourceLimitCPU != "" {
		p.AlamedaPodSpec.AlamedaScalerResources.Limits[int32(ApiCommon.ResourceName_CPU)] = entity.AlamedaSpecResourceLimitCPU
	}
	if entity.AlamedaSpecResourceLimitMemory != "" {
		p.AlamedaPodSpec.AlamedaScalerResources.Limits[int32(ApiCommon.ResourceName_MEMORY)] = entity.AlamedaSpecResourceLimitMemory
	}
	if entity.AlamedaSpecResourceRequestCPU != "" {
		p.AlamedaPodSpec.AlamedaScalerResources.Requests[int32(ApiCommon.ResourceName_CPU)] = entity.AlamedaSpecResourceRequestCPU
	}
	if entity.AlamedaSpecResourceRequestMemory != "" {
		p.AlamedaPodSpec.AlamedaScalerResources.Requests[int32(ApiCommon.ResourceName_MEMORY)] = entity.AlamedaSpecResourceRequestMemory
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
