package influxdb

import (
	DaoClusterTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	RepoInfluxCluster "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/clusterstatus"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
)

type Cluster struct {
	InfluxDBConfig InternalInflux.Config
}

func NewClusterWithConfig(config InternalInflux.Config) DaoClusterTypes.ClusterDAO {
	return &Cluster{InfluxDBConfig: config}
}

func (p *Cluster) CreateClusters(clusters []*DaoClusterTypes.Cluster) error {
	clusterRepo := RepoInfluxCluster.NewClusterRepositoryWithConfig(p.InfluxDBConfig)
	err := clusterRepo.CreateClusters(clusters)
	if err != nil {
		scope.Error(err.Error())
		return err
	}
	return nil
}

func (p *Cluster) ListClusters(request DaoClusterTypes.ListClustersRequest) ([]*DaoClusterTypes.Cluster, error) {
	clusterRepo := RepoInfluxCluster.NewClusterRepositoryWithConfig(p.InfluxDBConfig)
	clusters, err := clusterRepo.ListClusters(request)
	if err != nil {
		scope.Error(err.Error())
		return make([]*DaoClusterTypes.Cluster, 0), err
	}
	return clusters, nil
}
