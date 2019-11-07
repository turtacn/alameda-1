package responses

import (
	DaoClusterTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

type ClusterExtended struct {
	*DaoClusterTypes.Cluster
}

func (n *ClusterExtended) ProduceCluster() *ApiResources.Cluster {
	cluster := &ApiResources.Cluster{
		ObjectMeta: NewObjectMeta(n.ObjectMeta),
	}
	return cluster
}
