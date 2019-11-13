package types

import (
	"github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/clusterstatus"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	"github.com/containers-ai/alameda/internal/pkg/database/common"
	//ApiCommon "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"github.com/golang/protobuf/ptypes/timestamp"
	"strconv"
)

// Node provides node measurement operations
type NodeDAO interface {
	CreateNodes([]*Node) error
	ListNodes(ListNodesRequest) ([]*Node, error)
	DeleteNodes([]*ApiResources.Node) error
}

type Node struct {
	ObjectMeta      *metadata.ObjectMeta
	CreateTime      *timestamp.Timestamp
	Capacity        *Capacity
	AlamedaNodeSpec *AlamedaNodeSpec
}

type ListNodesRequest struct {
	common.QueryCondition
	ObjectMeta []metadata.ObjectMeta
}

type Capacity struct {
	CpuCores                 int64
	MemoryBytes              int64
	NetworkMegabitsPerSecond int64
}

type AlamedaNodeSpec struct {
	Provider *Provider
}

type Provider struct {
	Provider     string
	InstanceType string
	Region       string
	Zone         string
	Os           string
	Role         string
	InstanceId   string
	StorageSize  int64
}

func NewNode() *Node {
	node := Node{}
	node.ObjectMeta = &metadata.ObjectMeta{}
	node.Capacity = NewCapacity()
	node.AlamedaNodeSpec = NewAlamedaNodeSpec()
	return &node
}

func NewListNodesRequest() ListNodesRequest {
	request := ListNodesRequest{}
	request.ObjectMeta = make([]metadata.ObjectMeta, 0)
	return request
}

func NewCapacity() *Capacity {
	capacity := Capacity{}
	return &capacity
}

func NewAlamedaNodeSpec() *AlamedaNodeSpec {
	nodeSpec := AlamedaNodeSpec{}
	nodeSpec.Provider = &Provider{}
	return &nodeSpec
}

func (p *Node) Initialize(values map[string]string) {
	p.ObjectMeta.Initialize(values)
	if value, ok := values[string(clusterstatus.NodeCreateTime)]; ok {
		valueInt64, _ := strconv.ParseInt(value, 10, 64)
		p.CreateTime = &timestamp.Timestamp{Seconds: valueInt64}
	}
	p.Capacity.Initialize(values)
	p.AlamedaNodeSpec.Initialize(values)
}

func (p *Capacity) Initialize(values map[string]string) {
	if value, ok := values[string(clusterstatus.NodeCPUCores)]; ok {
		valueInt64, _ := strconv.ParseInt(value, 10, 64)
		p.CpuCores = valueInt64
	}
	if value, ok := values[string(clusterstatus.NodeMemoryBytes)]; ok {
		valueInt64, _ := strconv.ParseInt(value, 10, 64)
		p.MemoryBytes = valueInt64
	}
	if value, ok := values[string(clusterstatus.NodeMemoryBytes)]; ok {
		valueInt64, _ := strconv.ParseInt(value, 10, 64)
		p.NetworkMegabitsPerSecond = valueInt64
	}
}

func (p *AlamedaNodeSpec) Initialize(values map[string]string) {
	if value, ok := values[string(clusterstatus.NodeIOProvider)]; ok {
		p.Provider.Provider = value
	}
	if value, ok := values[string(clusterstatus.NodeIOInstanceType)]; ok {
		p.Provider.InstanceType = value
	}
	if value, ok := values[string(clusterstatus.NodeIORegion)]; ok {
		p.Provider.Region = value
	}
	if value, ok := values[string(clusterstatus.NodeIOZone)]; ok {
		p.Provider.Zone = value
	}
	if value, ok := values[string(clusterstatus.NodeIOOS)]; ok {
		p.Provider.Os = value
	}
	if value, ok := values[string(clusterstatus.NodeIORole)]; ok {
		p.Provider.Role = value
	}
	if value, ok := values[string(clusterstatus.NodeIOInstanceID)]; ok {
		p.Provider.InstanceId = value
	}
	if value, ok := values[string(clusterstatus.NodeIOStorageSize)]; ok {
		valueInt64, _ := strconv.ParseInt(value, 10, 64)
		p.Provider.StorageSize = valueInt64
	}
}
