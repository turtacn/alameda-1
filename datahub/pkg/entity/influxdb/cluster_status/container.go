package clusterstatus

import (
	"strconv"
	"time"

	"github.com/containers-ai/alameda/datahub/pkg/utils"
	influxdb_client "github.com/influxdata/influxdb/client/v2"
)

type containerTag = string
type containerField = string

const (
	// ContainerTime is the time that container information is saved to the measurement
	ContainerTime containerTag = "time"
	// ContainerNamespace is the container namespace
	ContainerNamespace containerTag = "namespace"
	// ContainerPodName is the name of pod that container is running in
	ContainerPodName containerTag = "pod_name"
	// ContainerAlamedaScalerNamespace is the namespace of AlamedaScaler that container belongs to
	ContainerAlamedaScalerNamespace containerTag = "alameda_scaler_namespace"
	// ContainerAlamedaScalerName is the name of AlamedaScaler that container belongs to
	ContainerAlamedaScalerName containerTag = "alameda_scaler_name"
	// ContainerNodeName is the name of node that container is running in
	ContainerNodeName containerTag = "node_name"
	// ContainerName is the container name
	ContainerName containerTag = "name"

	// ContainerResourceRequestCPU is CPU request of the container
	ContainerResourceRequestCPU containerField = "resource_request_cpu"
	// ContainerResourceRequestMemory is memory request of the container
	ContainerResourceRequestMemory containerField = "resource_request_memroy"
	// ContainerResourceLimitCPU is CPU limit of the container
	ContainerResourceLimitCPU containerField = "resource_limit_cpu"
	// ContainerResourceLimitMemory is memory limit of the container
	ContainerResourceLimitMemory containerField = "resource_limit_memory"
	// ContainerIsAlameda is the state that container is predicted or not
	ContainerIsAlameda containerField = "is_alameda"
	// ContainerIsDeleted is the state that container is deleted or not
	ContainerIsDeleted containerField = "is_deleted"
	// ContainerPolicy is the prediction policy of container
	ContainerPolicy containerField = "policy"
	// ContainerPodCreateTime is the creation time of pod
	ContainerPodCreateTime containerField = "pod_create_time"
	// ContainerResourceLink is the resource link of pod
	ContainerResourceLink containerField = "resource_link"
	// ContainerTopControllerName is top controller name of the pod
	ContainerTopControllerName containerField = "top_controller_name"
	// ContainerTopControllerKind is top controller kind of the pod
	ContainerTopControllerKind containerField = "top_controller_kind"
	// ContainerUsedRecommendationID is the recommendation id that the pod applied
	ContainerUsedRecommendationID containerField = "used_recommendation_id"
)

var (
	// ContainerTags is the list of container measurement tags
	ContainerTags = []containerTag{
		ContainerTime, ContainerNamespace, ContainerPodName,
		ContainerAlamedaScalerNamespace, ContainerAlamedaScalerName,
		ContainerNodeName, ContainerName,
	}
	// ContainerFields is the list of container measurement fields
	ContainerFields = []containerField{
		ContainerResourceRequestCPU, ContainerResourceRequestMemory,
		ContainerResourceLimitCPU, ContainerResourceLimitMemory,
		ContainerIsAlameda, ContainerIsDeleted, ContainerPolicy,
		ContainerPodCreateTime, ContainerResourceLink, ContainerTopControllerName,
		ContainerTopControllerKind,
	}
)

// ContainerEntity Entity in database
type ContainerEntity struct {
	Time                   time.Time
	Namespace              *string
	PodName                *string
	AlamedaScalerNamespace *string
	AlamedaScalerName      *string
	NodeName               *string
	Name                   *string
	ResourceRequestCPU     *float64
	ResourceRequestMemory  *int64
	ResourceLimitCPU       *float64
	ResourceLimitMemory    *int64
	IsAlameda              *bool
	IsDeleted              *bool
	Policy                 *string
	PodCreatedTime         *int64
	ResourceLink           *string
	TopControllerName      *string
	TopControllerKind      *string
	UsedRecommendationID   *string
}

// NewContainerEntityFromMap Build entity from map
func NewContainerEntityFromMap(data map[string]string) ContainerEntity {

	// TODO: log error
	tempTimestamp, _ := utils.ParseTime(data[ContainerTime])

	entity := ContainerEntity{
		Time: tempTimestamp,
	}

	if namespace, exist := data[ContainerNamespace]; exist {
		entity.Namespace = &namespace
	}
	if podName, exist := data[ContainerPodName]; exist {
		entity.PodName = &podName
	}
	if alamedaScalerNamespace, exist := data[ContainerAlamedaScalerNamespace]; exist {
		entity.AlamedaScalerNamespace = &alamedaScalerNamespace
	}
	if alamedaScalerName, exist := data[ContainerAlamedaScalerName]; exist {
		entity.AlamedaScalerName = &alamedaScalerName
	}
	if nodeName, exist := data[ContainerNodeName]; exist {
		entity.NodeName = &nodeName
	}
	if name, exist := data[ContainerName]; exist {
		entity.Name = &name
	}
	if resourceRequestCPU, exist := data[ContainerResourceRequestCPU]; exist {
		value, _ := strconv.ParseFloat(resourceRequestCPU, 64)
		entity.ResourceRequestCPU = &value
	}
	if resourceRequestMemory, exist := data[ContainerResourceRequestMemory]; exist {
		value, _ := strconv.ParseInt(resourceRequestMemory, 10, 64)
		entity.ResourceRequestMemory = &value
	}
	if resourceLimitCPU, exist := data[ContainerResourceLimitCPU]; exist {
		value, _ := strconv.ParseFloat(resourceLimitCPU, 64)
		entity.ResourceLimitCPU = &value
	}
	if resourceLimitMemory, exist := data[ContainerResourceLimitMemory]; exist {
		value, _ := strconv.ParseInt(resourceLimitMemory, 10, 64)
		entity.ResourceLimitMemory = &value
	}
	if isAlameda, exist := data[ContainerIsAlameda]; exist {
		value, _ := strconv.ParseBool(isAlameda)
		entity.IsAlameda = &value
	}
	if isDeleted, exist := data[ContainerIsDeleted]; exist {
		value, _ := strconv.ParseBool(isDeleted)
		entity.IsDeleted = &value
	}
	if policy, exist := data[ContainerPolicy]; exist {
		entity.Policy = &policy
	}
	if podCreatedTime, exist := data[ContainerPodCreateTime]; exist {
		value, _ := strconv.ParseInt(podCreatedTime, 10, 64)
		entity.PodCreatedTime = &value
	}
	if resourceLink, exist := data[ContainerResourceLink]; exist {
		entity.ResourceLink = &resourceLink
	}
	if topControllerName, exist := data[ContainerTopControllerName]; exist {
		entity.TopControllerName = &topControllerName
	}
	if topControllerKind, exist := data[ContainerTopControllerKind]; exist {
		entity.TopControllerKind = &topControllerKind
	}
	if usedRecommendationID, exist := data[ContainerUsedRecommendationID]; exist {
		entity.UsedRecommendationID = &usedRecommendationID
	}

	return entity
}

func (e ContainerEntity) InfluxDBPoint(measurementName string) (*influxdb_client.Point, error) {

	tags := map[string]string{}
	if e.Namespace != nil {
		tags[ContainerNamespace] = *e.Namespace
	}
	if e.PodName != nil {
		tags[ContainerPodName] = *e.PodName
	}
	if e.NodeName != nil {
		tags[ContainerNodeName] = *e.NodeName
	}
	if e.Name != nil {
		tags[ContainerName] = *e.Name
	}
	if e.AlamedaScalerNamespace != nil {
		tags[ContainerAlamedaScalerNamespace] = *e.AlamedaScalerNamespace
	}
	if e.AlamedaScalerName != nil {
		tags[ContainerAlamedaScalerName] = *e.AlamedaScalerName
	}

	fields := map[string]interface{}{}
	if e.IsDeleted != nil {
		fields[ContainerIsDeleted] = *e.IsDeleted
	}
	if e.IsAlameda != nil {
		fields[ContainerIsAlameda] = *e.IsAlameda
	}
	if e.Policy != nil {
		fields[ContainerPolicy] = *e.Policy
	}
	if e.ResourceRequestCPU != nil {
		fields[ContainerResourceRequestCPU] = *e.ResourceRequestCPU
	}
	if e.ResourceRequestMemory != nil {
		fields[ContainerResourceRequestMemory] = *e.ResourceRequestMemory
	}
	if e.ResourceLimitCPU != nil {
		fields[ContainerResourceLimitCPU] = *e.ResourceLimitCPU
	}
	if e.ResourceLimitMemory != nil {
		fields[ContainerResourceLimitMemory] = *e.ResourceLimitMemory
	}
	if e.PodCreatedTime != nil {
		fields[ContainerPodCreateTime] = *e.PodCreatedTime
	}
	if e.ResourceLink != nil {
		fields[ContainerResourceLink] = *e.ResourceLink
	}
	if e.TopControllerName != nil {
		fields[ContainerTopControllerName] = *e.TopControllerName
	}
	if e.TopControllerKind != nil {
		fields[ContainerTopControllerKind] = *e.TopControllerKind
	}
	if e.UsedRecommendationID != nil {
		fields[ContainerUsedRecommendationID] = *e.UsedRecommendationID
	}

	return influxdb_client.NewPoint(measurementName, tags, fields, e.Time)
}
