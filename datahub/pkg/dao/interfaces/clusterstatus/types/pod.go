package types

import (
	"github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/clusterstatus"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	"github.com/containers-ai/alameda/internal/pkg/database/common"
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	ApiCommon "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	"github.com/golang/protobuf/ptypes/timestamp"
	"strings"
)

// ContainerOperation provides container measurement operations
type PodDAO interface {
	CreatePods([]*Pod) error
	ListPods(*ListPodsRequest) ([]*Pod, error)
	DeletePods(*DeletePodsRequest) error
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
	ObjectMeta  []*metadata.ObjectMeta
	Kind        string // Valid values: DEPLOYMENT, DEPLOYMENTCONFIG, STATEFULSET, ALAMEDASCALER
	ScalingTool string // Valid values: NONE, VPA, HPA
}

type DeletePodsRequest struct {
	PodObjectMeta []*PodObjectMeta
}

type PodObjectMeta struct {
	ObjectMeta    *metadata.ObjectMeta
	TopController *metadata.ObjectMeta
	AlamedaScaler *metadata.ObjectMeta
	Kind          string // Valid values: DEPLOYMENT, DEPLOYMENTCONFIG, STATEFULSET, ALAMEDASCALER
	ScalingTool   string // Valid values: NONE, VPA, HPA
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

func NewPod(entity *clusterstatus.PodEntity) *Pod {
	pod := Pod{}
	pod.Containers = make([]*Container, 0)

	// Build ObjectMeta
	pod.ObjectMeta = &metadata.ObjectMeta{}
	pod.ObjectMeta.Name = entity.Name
	pod.ObjectMeta.Namespace = entity.Namespace
	pod.ObjectMeta.NodeName = entity.NodeName
	pod.ObjectMeta.ClusterName = entity.ClusterName
	pod.ObjectMeta.Uid = entity.Uid

	// Build misc info
	pod.CreateTime = &timestamp.Timestamp{Seconds: entity.CreateTime}
	pod.ResourceLink = entity.ResourceLink
	pod.AppName = entity.AppName
	pod.AppPartOf = entity.AppPartOf

	// Build TopController
	pod.TopController = &Controller{}
	pod.TopController.ObjectMeta = &metadata.ObjectMeta{}
	pod.TopController.ObjectMeta.Name = entity.TopControllerName
	pod.TopController.Kind = entity.TopControllerKind
	pod.TopController.Replicas = entity.TopControllerReplicas

	// Build Status
	pod.Status = NewPodStatus(entity)

	// Build AlamedaPodSpec
	pod.AlamedaPodSpec = NewAlamedaPodSpec(entity)

	return &pod
}

func NewListPodsRequest() *ListPodsRequest {
	request := ListPodsRequest{}
	request.ObjectMeta = make([]*metadata.ObjectMeta, 0)
	return &request
}

func NewDeletePodsRequest() *DeletePodsRequest {
	request := DeletePodsRequest{}
	request.PodObjectMeta = make([]*PodObjectMeta, 0)
	return &request
}

func NewPodObjectMeta(objectMeta, topController, alamedaScaler *metadata.ObjectMeta, kind, scalingTool string) *PodObjectMeta {
	podObjectMeta := PodObjectMeta{}
	podObjectMeta.ObjectMeta = objectMeta
	podObjectMeta.TopController = topController
	podObjectMeta.AlamedaScaler = alamedaScaler
	podObjectMeta.Kind = kind
	podObjectMeta.ScalingTool = scalingTool
	return &podObjectMeta
}

func NewAlamedaPodSpec(entity *clusterstatus.PodEntity) *AlamedaPodSpec {
	spec := AlamedaPodSpec{}
	spec.AlamedaScaler = &metadata.ObjectMeta{}
	spec.AlamedaScaler.Name = entity.AlamedaSpecScalerName
	spec.Policy = entity.AlamedaSpecPolicy
	spec.UsedRecommendationId = entity.AlamedaSpecUsedRecommendationID
	spec.ScalingTool = entity.AlamedaSpecScalerScalingTool
	spec.AlamedaScalerResources = NewResourceRequirements(
		entity.AlamedaSpecResourceLimitCPU,
		entity.AlamedaSpecResourceLimitMemory,
		entity.AlamedaSpecResourceRequestCPU,
		entity.AlamedaSpecResourceRequestMemory,
	)
	return &spec
}

func NewPodStatus(entity *clusterstatus.PodEntity) *PodStatus {
	status := PodStatus{}
	status.Phase = entity.StatusPhase
	status.Message = entity.StatusMessage
	status.Reason = entity.StatusReason
	return &status
}

func (p *Pod) BuildEntity() *clusterstatus.PodEntity {
	entity := clusterstatus.PodEntity{}

	entity.Time = influxdb.ZeroTime
	entity.Name = p.ObjectMeta.Name
	entity.Namespace = p.ObjectMeta.Namespace
	entity.NodeName = p.ObjectMeta.NodeName
	entity.ClusterName = p.ObjectMeta.ClusterName
	entity.Uid = p.ObjectMeta.Uid
	entity.CreateTime = p.CreateTime.GetSeconds()
	entity.ResourceLink = p.ResourceLink
	entity.AppName = p.AppName
	entity.AppPartOf = p.AppPartOf

	if p.TopController != nil {
		entity.TopControllerName = p.TopController.ObjectMeta.Name
		entity.TopControllerKind = p.TopController.Kind
		entity.TopControllerReplicas = p.TopController.Replicas
	}

	if p.Status != nil {
		entity.StatusPhase = p.Status.Phase
		entity.StatusMessage = p.Status.Message
		entity.StatusReason = p.Status.Reason
	}

	if p.AlamedaPodSpec != nil {
		entity.AlamedaSpecScalerName = p.AlamedaPodSpec.AlamedaScaler.Name
		entity.AlamedaSpecPolicy = p.AlamedaPodSpec.Policy
		entity.AlamedaSpecUsedRecommendationID = p.AlamedaPodSpec.UsedRecommendationId
		entity.AlamedaSpecScalerScalingTool = p.AlamedaPodSpec.ScalingTool

		if p.AlamedaPodSpec.AlamedaScalerResources != nil {
			if value, exist := p.AlamedaPodSpec.AlamedaScalerResources.Limits[int32(ApiCommon.ResourceName_CPU)]; exist {
				entity.AlamedaSpecResourceLimitCPU = value
			}
			if value, exist := p.AlamedaPodSpec.AlamedaScalerResources.Limits[int32(ApiCommon.ResourceName_MEMORY)]; exist {
				entity.AlamedaSpecResourceLimitMemory = value
			}
			if value, exist := p.AlamedaPodSpec.AlamedaScalerResources.Requests[int32(ApiCommon.ResourceName_CPU)]; exist {
				entity.AlamedaSpecResourceRequestCPU = value
			}
			if value, exist := p.AlamedaPodSpec.AlamedaScalerResources.Requests[int32(ApiCommon.ResourceName_MEMORY)]; exist {
				entity.AlamedaSpecResourceRequestMemory = value
			}
		}
	}

	return &entity
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
