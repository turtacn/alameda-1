package clusterstatus

import (
	"time"

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
	ResourceRequestCPU     *string
	ResourceRequestMemory  *string
	ResourceLimitCPU       *string
	ResourceLimitMemory    *string
	IsAlameda              *string
	IsDeleted              *string
	Policy                 *string
}

func NewContainerEntityFromMap(data map[string]string) ContainerEntity {

	tempTimestamp, _ := time.Parse("2006-01-02T15:04:05.999999Z07:00", data[ContainerTime])

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
		entity.ResourceRequestCPU = &resourceRequestCPU
	}
	if resourceRequestMemory, exist := data[ContainerResourceRequestMemory]; exist {
		entity.ResourceRequestMemory = &resourceRequestMemory
	}
	if resourceLimitCPU, exist := data[ContainerResourceLimitCPU]; exist {
		entity.ResourceLimitCPU = &resourceLimitCPU
	}
	if resourceLimitMemory, exist := data[ContainerResourceLimitMemory]; exist {
		entity.ResourceLimitMemory = &resourceLimitMemory
	}
	if isAlameda, exist := data[ContainerIsAlameda]; exist {
		entity.IsAlameda = &isAlameda
	}
	if isDeleted, exist := data[ContainerIsDeleted]; exist {
		entity.IsDeleted = &isDeleted
	}
	if policy, exist := data[ContainerPolicy]; exist {
		entity.Policy = &policy
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

	return influxdb_client.NewPoint(measurementName, tags, fields, e.Time)
}
