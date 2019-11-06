package influxdb

import (
	DaoClusterTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	RepoInfluxCluster "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/clusterstatus"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	Log "github.com/containers-ai/alameda/pkg/utils/log"
)

var (
	scope = Log.RegisterScope("dao_implement", "dao implement", 0)
)

type Application struct {
	InfluxDBConfig InternalInflux.Config
}

func NewApplicationWithConfig(config InternalInflux.Config) DaoClusterTypes.ApplicationDAO {
	return &Application{InfluxDBConfig: config}
}

func (p *Application) CreateApplications(applications []*DaoClusterTypes.Application) error {
	applicationRepo := RepoInfluxCluster.NewApplicationRepositoryWithConfig(p.InfluxDBConfig)
	err := applicationRepo.CreateApplications(applications)
	if err != nil {
		scope.Error(err.Error())
		return err
	}
	return nil
}

func (p *Application) ListApplications(request DaoClusterTypes.ListApplicationsRequest) ([]*DaoClusterTypes.Application, error) {
	applicationRepo := RepoInfluxCluster.NewApplicationRepositoryWithConfig(p.InfluxDBConfig)
	applications, err := applicationRepo.ListApplications(request)
	if err != nil {
		scope.Error(err.Error())
		return make([]*DaoClusterTypes.Application, 0), err
	}
	return applications, nil
}
