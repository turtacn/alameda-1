package types

import (
	"github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/clusterstatus"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	"github.com/containers-ai/alameda/internal/pkg/database/common"
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
)

type ClusterDAO interface {
	CreateClusters([]*Cluster) error
	ListClusters(*ListClustersRequest) ([]*Cluster, error)
	DeleteClusters(*DeleteClustersRequest) error
}

type Cluster struct {
	ObjectMeta *metadata.ObjectMeta
	Value      string
}

type ListClustersRequest struct {
	common.QueryCondition
	ObjectMeta []*metadata.ObjectMeta
}

type DeleteClustersRequest struct {
	ObjectMeta []*metadata.ObjectMeta
}

func NewCluster(entity *clusterstatus.ClusterEntity) *Cluster {
	cluster := Cluster{}
	cluster.ObjectMeta = &metadata.ObjectMeta{}
	cluster.ObjectMeta.Name = entity.Name
	cluster.ObjectMeta.Uid = entity.Uid
	return &cluster
}

func NewListClustersRequest() *ListClustersRequest {
	request := ListClustersRequest{}
	request.ObjectMeta = make([]*metadata.ObjectMeta, 0)
	return &request
}

func NewDeleteClustersRequest() *DeleteClustersRequest {
	request := DeleteClustersRequest{}
	request.ObjectMeta = make([]*metadata.ObjectMeta, 0)
	return &request
}

func (p *Cluster) BuildEntity() *clusterstatus.ClusterEntity {
	entity := clusterstatus.ClusterEntity{
		// InfluxDB tags
		Time: influxdb.ZeroTime,
		Name: p.ObjectMeta.Name,
		Uid:  p.ObjectMeta.Uid,

		// InfluxDB fields
		Value: p.Value,
	}
	return &entity
}
