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

type DeleteClustersRequestExtended struct {
	*ApiResources.DeleteClustersRequest
}

func NewCluster(cluster *ApiResources.Cluster) *DaoClusterTypes.Cluster {
	if cluster != nil {
		// Normalize request
		objectMeta := NewObjectMeta(cluster.GetObjectMeta())
		objectMeta.Namespace = ""
		objectMeta.NodeName = ""
		objectMeta.ClusterName = ""

		c := DaoClusterTypes.Cluster{}
		c.ObjectMeta = &objectMeta

		return &c
	}
	return nil
}

func (p *CreateClustersRequestExtended) Validate() error {
	return nil
}

func (p *CreateClustersRequestExtended) ProduceClusters() []*DaoClusterTypes.Cluster {
	clusters := make([]*DaoClusterTypes.Cluster, 0)

	for _, cluster := range p.GetClusters() {
		clusters = append(clusters, NewCluster(cluster))
	}

	return clusters
}

func (p *ListClustersRequestExtended) Validate() error {
	return nil
}

func (p *ListClustersRequestExtended) ProduceRequest() *DaoClusterTypes.ListClustersRequest {
	request := DaoClusterTypes.NewListClustersRequest()
	if p.GetObjectMeta() != nil {
		for _, meta := range p.GetObjectMeta() {
			// Normalize request
			objectMeta := NewObjectMeta(meta)
			objectMeta.Namespace = ""
			objectMeta.NodeName = ""
			objectMeta.ClusterName = ""

			if objectMeta.IsEmpty() {
				return DaoClusterTypes.NewListClustersRequest()
			}
			request.ObjectMeta = append(request.ObjectMeta, &objectMeta)
		}
	}
	return request
}

func (p *DeleteClustersRequestExtended) Validate() error {
	return nil
}

func (p *DeleteClustersRequestExtended) ProduceRequest() *DaoClusterTypes.DeleteClustersRequest {
	request := DaoClusterTypes.NewDeleteClustersRequest()
	if p.GetObjectMeta() != nil {
		for _, meta := range p.GetObjectMeta() {
			// Normalize request
			objectMeta := NewObjectMeta(meta)
			objectMeta.Namespace = ""
			objectMeta.NodeName = ""
			objectMeta.ClusterName = ""

			if objectMeta.IsEmpty() {
				request := DaoClusterTypes.NewDeleteClustersRequest()
				return request
			}
			request.ObjectMeta = append(request.ObjectMeta, &objectMeta)
		}
	}
	return request
}
