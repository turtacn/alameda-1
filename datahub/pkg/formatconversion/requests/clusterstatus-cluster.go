package requests

import (
	DaoClusterTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

type CreateClustersRequestExtended struct {
	ApiResources.CreateClustersRequest
}

type ListClustersRequestExtended struct {
	*ApiResources.ListClustersRequest
}

func (r *CreateClustersRequestExtended) Validate() error {
	return nil
}

func (r *CreateClustersRequestExtended) ProduceClusters() []*DaoClusterTypes.Cluster {
	clusters := make([]*DaoClusterTypes.Cluster, 0)

	for _, clst := range r.GetClusters() {
		// Normalize request
		objectMeta := NewObjectMeta(clst.GetObjectMeta())
		objectMeta.Namespace = ""
		objectMeta.NodeName = ""
		objectMeta.ClusterName = ""

		cluster := DaoClusterTypes.NewCluster()
		cluster.ObjectMeta = objectMeta
		clusters = append(clusters, cluster)
	}

	return clusters
}

func (r *ListClustersRequestExtended) Validate() error {
	return nil
}

func (r *ListClustersRequestExtended) ProduceRequest() DaoClusterTypes.ListClustersRequest {
	request := DaoClusterTypes.NewListClustersRequest()
	if r.GetObjectMeta() != nil {
		for _, meta := range r.GetObjectMeta() {
			// Normalize request
			objectMeta := NewObjectMeta(meta)
			objectMeta.Namespace = ""
			objectMeta.NodeName = ""
			objectMeta.ClusterName = ""

			if objectMeta.IsEmpty() {
				return DaoClusterTypes.NewListClustersRequest()
			}
			request.ObjectMeta = append(request.ObjectMeta, objectMeta)
		}
	}
	return request
}
