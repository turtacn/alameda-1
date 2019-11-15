package influxdb

import (
	DaoClusterTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	RepoInfluxCluster "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/clusterstatus"
	Metadata "github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
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

	controllerRequest := DaoClusterTypes.NewListControllersRequest()
	for _, application := range applications {
		objectMeta := Metadata.ObjectMeta{}
		objectMeta.Namespace = application.ObjectMeta.Namespace
		objectMeta.ClusterName = application.ObjectMeta.ClusterName
		controllerRequest.ObjectMeta = append(controllerRequest.ObjectMeta, objectMeta)
	}

	controllerRepo := RepoInfluxCluster.NewControllerRepository(&p.InfluxDBConfig)
	controllers, err := controllerRepo.ListControllers(controllerRequest)
	if err != nil {
		scope.Error(err.Error())
		return make([]*DaoClusterTypes.Application, 0), err
	}
	for _, controller := range controllers {
		for _, application := range applications {
			if application.ObjectMeta.Name == controller.AlamedaControllerSpec.AlamedaScaler.Name &&
				application.ObjectMeta.Namespace == controller.AlamedaControllerSpec.AlamedaScaler.Namespace &&
				application.ObjectMeta.ClusterName == controller.ObjectMeta.ClusterName {
				if application.Controllers == nil {
					application.Controllers = make([]*DaoClusterTypes.Controller, 0)
				}
				application.Controllers = append(application.Controllers, controller)
				break
			}
		}
	}

	return applications, nil
}
