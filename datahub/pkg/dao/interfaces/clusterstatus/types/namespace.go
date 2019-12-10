package types

import (
	"github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/clusterstatus"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	"github.com/containers-ai/alameda/internal/pkg/database/common"
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
)

type NamespaceDAO interface {
	CreateNamespaces([]*Namespace) error
	ListNamespaces(*ListNamespacesRequest) ([]*Namespace, error)
	DeleteNamespaces(*DeleteNamespacesRequest) error
}

type Namespace struct {
	ObjectMeta *metadata.ObjectMeta
	Value      string
}

type ListNamespacesRequest struct {
	common.QueryCondition
	ObjectMeta []*metadata.ObjectMeta
}

type DeleteNamespacesRequest struct {
	ObjectMeta []*metadata.ObjectMeta
}

func NewNamespace(entity *clusterstatus.NamespaceEntity) *Namespace {
	namespace := Namespace{}
	namespace.ObjectMeta = &metadata.ObjectMeta{}
	namespace.ObjectMeta.Name = entity.Name
	namespace.ObjectMeta.ClusterName = entity.ClusterName
	namespace.ObjectMeta.Uid = entity.Uid
	namespace.Value = entity.Value
	return &namespace
}

func NewListNamespacesRequest() *ListNamespacesRequest {
	request := ListNamespacesRequest{}
	request.ObjectMeta = make([]*metadata.ObjectMeta, 0)
	return &request
}

func NewDeleteNamespacesRequest() *DeleteNamespacesRequest {
	request := DeleteNamespacesRequest{}
	request.ObjectMeta = make([]*metadata.ObjectMeta, 0)
	return &request
}

func (p *Namespace) BuildEntity() *clusterstatus.NamespaceEntity {
	entity := clusterstatus.NamespaceEntity{
		// InfluxDB tags
		Time:        influxdb.ZeroTime,
		Name:        p.ObjectMeta.Name,
		ClusterName: p.ObjectMeta.ClusterName,
		Uid:         p.ObjectMeta.Uid,

		// InfluxDB fields
		Value: p.Value,
	}
	return &entity
}
