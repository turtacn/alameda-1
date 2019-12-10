package responses

import (
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	"github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

type NodeExtended struct {
	*types.Node
}

func (p *NodeExtended) ProduceNode() *resources.Node {
	node := resources.Node{}
	node.ObjectMeta = NewObjectMeta(p.ObjectMeta)
	node.StartTime = p.CreateTime
	node.Capacity = NewCapacity(p.Capacity)
	node.AlamedaNodeSpec = NewAlamedaNodeSpec(p.AlamedaNodeSpec)
	return &node
}
