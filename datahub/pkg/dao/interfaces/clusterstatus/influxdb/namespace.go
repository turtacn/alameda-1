package influxdb

import (
	DaoClusterTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	RepoInfluxCluster "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/clusterstatus"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
)

type Namespace struct {
	InfluxDBConfig InternalInflux.Config
}

func NewNamespaceWithConfig(config InternalInflux.Config) DaoClusterTypes.NamespaceDAO {
	return &Namespace{InfluxDBConfig: config}
}

func (p *Namespace) CreateNamespaces(namespaces []*DaoClusterTypes.Namespace) error {
	namespaceRepo := RepoInfluxCluster.NewNamespaceRepositoryWithConfig(p.InfluxDBConfig)
	err := namespaceRepo.CreateNamespaces(namespaces)
	if err != nil {
		scope.Error(err.Error())
		return err
	}
	return nil
}

func (p *Namespace) ListNamespaces(request DaoClusterTypes.ListNamespacesRequest) ([]*DaoClusterTypes.Namespace, error) {
	namespaceRepo := RepoInfluxCluster.NewNamespaceRepositoryWithConfig(p.InfluxDBConfig)
	namespaces, err := namespaceRepo.ListNamespaces(request)
	if err != nil {
		scope.Error(err.Error())
		return make([]*DaoClusterTypes.Namespace, 0), err
	}
	return namespaces, nil
}
