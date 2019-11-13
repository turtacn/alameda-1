package responses

import (
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	"github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

type NodeExtended struct {
	*types.Node
}

func (n *NodeExtended) ProduceNode() *resources.Node {
	node := resources.Node{}
	node.ObjectMeta = NewObjectMeta(*n.ObjectMeta)
	node.StartTime = n.CreateTime
	node.Capacity = NewCapacity(n.Capacity)
	node.AlamedaNodeSpec = NewAlamedaNodeSpec(n.AlamedaNodeSpec)
	return &node
}
