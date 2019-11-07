package types

import (
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	"github.com/containers-ai/alameda/internal/pkg/database/common"
)

type ClusterDAO interface {
	CreateClusters([]*Cluster) error
	ListClusters(ListClustersRequest) ([]*Cluster, error)
}

type Cluster struct {
	ObjectMeta metadata.ObjectMeta
}

type ListClustersRequest struct {
	common.QueryCondition
	ObjectMeta []metadata.ObjectMeta
}

func NewCluster() *Cluster {
	return &Cluster{}
}

func NewListClustersRequest() ListClustersRequest {
	request := ListClustersRequest{}
	request.ObjectMeta = make([]metadata.ObjectMeta, 0)
	return request
}
