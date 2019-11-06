package types

import (
	ApiCommon "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

// Node provides node measurement operations
type NodeDAO interface {
	RegisterAlamedaNodes([]*ApiResources.Node) error
	DeregisterAlamedaNodes([]*ApiResources.Node) error
	ListAlamedaNodes(timeRange *ApiCommon.TimeRange) ([]*ApiResources.Node, error)
	ListNodes(ListNodesRequest) ([]*ApiResources.Node, error)
}

type ListNodesRequest struct {
	NodeNames []string
	InCluster bool
}
