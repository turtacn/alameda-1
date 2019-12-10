package requests

import (
	DaoClusterTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

type CreateNodesRequestExtended struct {
	ApiResources.CreateNodesRequest
}

type ListNodesRequestExtended struct {
	*ApiResources.ListNodesRequest
}

type DeleteNodesRequestExtended struct {
	*ApiResources.DeleteNodesRequest
}

func NewNode(node *ApiResources.Node) *DaoClusterTypes.Node {
	if node != nil {
		// Normalize request
		objectMeta := NewObjectMeta(node.GetObjectMeta())
		objectMeta.Namespace = ""
		objectMeta.NodeName = ""

		n := DaoClusterTypes.Node{}
		n.ObjectMeta = &objectMeta
		n.CreateTime = node.GetStartTime()
		n.Capacity = NewCapacity(node.GetCapacity())
		n.AlamedaNodeSpec = NewAlamedaNodeSpec(node.GetAlamedaNodeSpec())

		return &n
	}
	return nil
}

func (p *CreateNodesRequestExtended) Validate() error {
	return nil
}

func (p *CreateNodesRequestExtended) ProduceNodes() []*DaoClusterTypes.Node {
	nodes := make([]*DaoClusterTypes.Node, 0)

	for _, node := range p.GetNodes() {
		nodes = append(nodes, NewNode(node))
	}

	return nodes
}

func (p *ListNodesRequestExtended) Validate() error {
	return nil
}

func (p *ListNodesRequestExtended) ProduceRequest() *DaoClusterTypes.ListNodesRequest {
	request := DaoClusterTypes.NewListNodesRequest()
	request.QueryCondition = QueryConditionExtend{p.GetQueryCondition()}.QueryCondition()
	if p.GetObjectMeta() != nil {
		for _, meta := range p.GetObjectMeta() {
			// Normalize request
			objectMeta := NewObjectMeta(meta)
			objectMeta.Namespace = ""
			objectMeta.NodeName = ""

			if objectMeta.IsEmpty() {
				request := DaoClusterTypes.NewListNodesRequest()
				request.QueryCondition = QueryConditionExtend{p.GetQueryCondition()}.QueryCondition()
				return request
			}
			request.ObjectMeta = append(request.ObjectMeta, &objectMeta)
		}
	}
	return request
}

func (p *DeleteNodesRequestExtended) Validate() error {
	return nil
}

func (p *DeleteNodesRequestExtended) ProduceRequest() *DaoClusterTypes.DeleteNodesRequest {
	request := DaoClusterTypes.NewDeleteNodesRequest()
	if p.GetObjectMeta() != nil {
		for _, meta := range p.GetObjectMeta() {
			// Normalize request
			objectMeta := NewObjectMeta(meta)
			objectMeta.Namespace = ""
			objectMeta.NodeName = ""

			if objectMeta.IsEmpty() {
				request := DaoClusterTypes.NewDeleteNodesRequest()
				return request
			}
			request.ObjectMeta = append(request.ObjectMeta, &objectMeta)
		}
	}
	return request
}
