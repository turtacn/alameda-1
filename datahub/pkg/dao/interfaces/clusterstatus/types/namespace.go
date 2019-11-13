package types

import (
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	"github.com/containers-ai/alameda/internal/pkg/database/common"
)

type NamespaceDAO interface {
	CreateNamespaces([]*Namespace) error
	ListNamespaces(ListNamespacesRequest) ([]*Namespace, error)
}

type Namespace struct {
	ObjectMeta metadata.ObjectMeta
}

type ListNamespacesRequest struct {
	common.QueryCondition
	ObjectMeta []metadata.ObjectMeta
}

func NewNamespace() *Namespace {
	return &Namespace{}
}

func NewListNamespacesRequest() ListNamespacesRequest {
	request := ListNamespacesRequest{}
	request.ObjectMeta = make([]metadata.ObjectMeta, 0)
	return request
}
