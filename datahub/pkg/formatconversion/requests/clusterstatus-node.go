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

func (r *CreateNodesRequestExtended) Validate() error {
	return nil
}

func (r *CreateNodesRequestExtended) ProduceNodes() []*DaoClusterTypes.Node {
	nodes := make([]*DaoClusterTypes.Node, 0)

	for _, n := range r.GetNodes() {
		// Normalize request
		objectMeta := NewObjectMeta(n.GetObjectMeta())
		objectMeta.Namespace = ""
		objectMeta.NodeName = ""

		node := &DaoClusterTypes.Node{}
		node.ObjectMeta = &objectMeta
		node.CreateTime = n.GetStartTime()
		node.Capacity = NewCapacity(n.GetCapacity())
		node.AlamedaNodeSpec = NewAlamedaNodeSpec(n.GetAlamedaNodeSpec())

		nodes = append(nodes, node)
	}

	return nodes
}

func (r *ListNodesRequestExtended) Validate() error {
	return nil
}

func (r *ListNodesRequestExtended) ProduceRequest() DaoClusterTypes.ListNodesRequest {
	request := DaoClusterTypes.NewListNodesRequest()
	request.QueryCondition = QueryConditionExtend{r.GetQueryCondition()}.QueryCondition()
	if r.GetObjectMeta() != nil {
		for _, meta := range r.GetObjectMeta() {
			// Normalize request
			objectMeta := NewObjectMeta(meta)
			objectMeta.Namespace = ""
			objectMeta.NodeName = ""

			if objectMeta.IsEmpty() {
				request := DaoClusterTypes.NewListNodesRequest()
				request.QueryCondition = QueryConditionExtend{r.GetQueryCondition()}.QueryCondition()
				return request
			}
			request.ObjectMeta = append(request.ObjectMeta, objectMeta)
		}
	}
	return request
}
