package clusterstatus

import (
	"github.com/containers-ai/alameda/datahub/pkg/utils"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"strconv"
	"time"
)

type nodeField = string
type nodeTag = string

const (
	NodeTime nodeTag = "time" // NodeTime is the time node information is inserted to database
	NodeName nodeTag = "name" // NodeName is the name of node

	NodeGroup       nodeField = "group"             // NodeGroup is node group name
	NodeInCluster   nodeField = "in_cluster"        // NodeInCluster is the state node is in cluster or not
	NodeCPUCores    nodeField = "node_cpu_cores"    // NodeCPUCores is the amount of cores in node
	NodeMemoryBytes nodeField = "node_memory_bytes" // NodeMemoryBytes is the amount of momory bytes in node

	NodeCreateTime nodeField = "create_time"

	NodeIOProvider     nodeField = "io_provider" // Cloud service provider
	NodeIOInstanceType nodeField = "io_instance_type"
	NodeIORegion       nodeField = "io_region"
	NodeIOZone         nodeField = "io_zone"
	NodeIOOS           nodeField = "io_os"
	NodeIORole         nodeField = "io_role"
	NodeIOInstanceID   nodeField = "io_instance_id"
	NodeIOStorageSize  nodeField = "io_storage_size"
)

var (
	// NodeTags list tags of node measurement
	NodeTags = []nodeTag{NodeTime, NodeName}
	// NodeFields list fields of node measurement
	NodeFields = []nodeField{NodeGroup, NodeInCluster, NodeCPUCores, NodeMemoryBytes, NodeCreateTime, NodeIOProvider, NodeIOInstanceType, NodeIORegion, NodeIOZone}
)

// NodeEntity is entity in database
type NodeEntity struct {
	Time           time.Time
	Name           *string
	NodeGroup      *string
	IsInCluster    *bool
	CPUCores       *int64
	MemoryBytes    *int64
	CreatedTime    *int64
	IOProvider     *string
	IOInstanceType *string
	IORegion       *string
	IOZone         *string
	IOOS           *string
	IORole         *string
	IOInstanceID   *string
	IOStorageSize  *int64
}

// NewNodeEntityFromMap Build entity from map
func NewNodeEntityFromMap(data map[string]string) NodeEntity {

	// TODO: log error
	tempTimestamp, _ := utils.ParseTime(data[ContainerTime])

	entity := NodeEntity{
		Time: tempTimestamp,
	}

	if name, exist := data[NodeName]; exist {
		entity.Name = &name
	}
	if nodeGroup, exist := data[NodeGroup]; exist {
		entity.NodeGroup = &nodeGroup
	}
	if isInCluster, exist := data[NodeInCluster]; exist {
		value, _ := strconv.ParseBool(isInCluster)
		entity.IsInCluster = &value
	}
	if cpuCores, exist := data[NodeCPUCores]; exist {
		value, _ := strconv.ParseInt(cpuCores, 10, 64)
		entity.CPUCores = &value
	}
	if memoryBytes, exist := data[NodeMemoryBytes]; exist {
		value, _ := strconv.ParseInt(memoryBytes, 10, 64)
		entity.MemoryBytes = &value
	}
	if ioProvider, exist := data[NodeIOProvider]; exist {
		entity.IOProvider = &ioProvider
	}
	if ioInstanceType, exist := data[NodeIOInstanceType]; exist {
		entity.IOInstanceType = &ioInstanceType
	}
	if ioRegion, exist := data[NodeIORegion]; exist {
		entity.IORegion = &ioRegion
	}
	if ioZone, exist := data[NodeIOZone]; exist {
		entity.IOZone = &ioZone
	}
	if ioOs, exist := data[NodeIOOS]; exist {
		entity.IOOS = &ioOs
	}
	if ioRole, exist := data[NodeIORole]; exist {
		entity.IORole = &ioRole
	}
	if ioInstanceID, exist := data[NodeIOInstanceID]; exist {
		entity.IOInstanceID = &ioInstanceID
	}
	if ioStorageSize, exist := data[NodeIOStorageSize]; exist {
		value, _ := strconv.ParseInt(ioStorageSize, 10, 64)
		entity.IOStorageSize = &value
	}

	return entity
}

func (e NodeEntity) InfluxDBPoint(measurementName string) (*InfluxClient.Point, error) {

	tags := map[string]string{}
	if e.Name != nil {
		tags[NodeName] = *e.Name
	}

	fields := map[string]interface{}{}
	if e.NodeGroup != nil {
		fields[NodeGroup] = *e.NodeGroup
	}
	if e.IsInCluster != nil {
		fields[NodeInCluster] = *e.IsInCluster
	}
	if e.CPUCores != nil {
		fields[NodeCPUCores] = *e.CPUCores
	}
	if e.MemoryBytes != nil {
		fields[NodeMemoryBytes] = *e.MemoryBytes
	}
	if e.CreatedTime != nil {
		fields[NodeCreateTime] = *e.CreatedTime
	}
	if e.IOProvider != nil {
		fields[NodeIOProvider] = *e.IOProvider
	}
	if e.IOInstanceType != nil {
		fields[NodeIOInstanceType] = *e.IOInstanceType
	}
	if e.IORegion != nil {
		fields[NodeIORegion] = *e.IORegion
	}
	if e.IOZone != nil {
		fields[NodeIOZone] = *e.IOZone
	}
	if e.IOOS != nil {
		fields[NodeIOOS] = *e.IOOS
	}
	if e.IORole != nil {
		fields[NodeIORole] = *e.IORole
	}
	if e.IOInstanceID != nil {
		fields[NodeIOInstanceID] = *e.IOInstanceID
	}
	if e.IOStorageSize != nil {
		fields[NodeIOStorageSize] = *e.IOStorageSize
	}

	return InfluxClient.NewPoint(measurementName, tags, fields, e.Time)
}

func (e NodeEntity) BuildDatahubNode() *datahub_v1alpha1.Node {

	node := &datahub_v1alpha1.Node{
		Capacity: &datahub_v1alpha1.Capacity{},
		Provider: &datahub_v1alpha1.Provider{},
	}

	if e.Name != nil {
		node.Name = *e.Name
	}
	if e.CPUCores != nil {
		node.Capacity.CpuCores = *e.CPUCores
	}
	if e.MemoryBytes != nil {
		node.Capacity.MemoryBytes = *e.MemoryBytes
	}
	if e.IOProvider != nil {
		node.Provider.Provider = *e.IOProvider
	}
	if e.IOInstanceType != nil {
		node.Provider.InstanceType = *e.IOInstanceType
	}
	if e.IORegion != nil {
		node.Provider.Region = *e.IORegion
	}
	if e.IOZone != nil {
		node.Provider.Zone = *e.IOZone
	}
	if e.IOOS != nil {
		node.Provider.Os = *e.IOOS
	}
	if e.IORole != nil {
		node.Provider.Role = *e.IORole
	}
	if e.IOInstanceID != nil {
		node.Provider.InstanceId = *e.IOInstanceID
	}
	if e.IOStorageSize != nil {
		node.Provider.StorageSize = *e.IOStorageSize
	}

	return node
}
