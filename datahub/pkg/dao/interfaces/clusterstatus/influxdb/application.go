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
	applicationRepo := RepoInfluxCluster.NewApplicationRepository(p.InfluxDBConfig)
	err := applicationRepo.CreateApplications(applications)
	if err != nil {
		scope.Error(err.Error())
		return err
	}
	return nil
}

func (p *Application) ListApplications(request *DaoClusterTypes.ListApplicationsRequest) ([]*DaoClusterTypes.Application, error) {
	listControllersReq := p.genListControllersRequest(request)

	// List controllers
	controllerRepo := RepoInfluxCluster.NewControllerRepository(p.InfluxDBConfig)
	controllers, err := controllerRepo.ListControllers(listControllersReq)
	if err != nil {
		scope.Error(err.Error())
		return make([]*DaoClusterTypes.Application, 0), err
	}

	// List applications
	applicationRepo := RepoInfluxCluster.NewApplicationRepository(p.InfluxDBConfig)
	applications, err := applicationRepo.ListApplications(request)
	if err != nil {
		scope.Error(err.Error())
		return make([]*DaoClusterTypes.Application, 0), err
	}

	// Append controllers into applications
	for _, controller := range controllers {
		for _, application := range applications {
			if application.ObjectMeta.Name == controller.AlamedaControllerSpec.AlamedaScaler.Name &&
				application.ObjectMeta.Namespace == controller.ObjectMeta.Namespace &&
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

func (p *Application) DeleteApplications(request *DaoClusterTypes.DeleteApplicationsRequest) error {
	delControllersReq := p.genDeleteControllersRequest(request)

	// Delete applications
	applicationRepo := RepoInfluxCluster.NewApplicationRepository(p.InfluxDBConfig)
	if err := applicationRepo.DeleteApplications(request); err != nil {
		scope.Error(err.Error())
		return err
	}

	// Delete controllers
	controllerDAO := NewControllerWithConfig(p.InfluxDBConfig)
	if err := controllerDAO.DeleteControllers(delControllersReq); err != nil {
		scope.Error(err.Error())
		return err
	}

	return nil
}

func (p *Application) genListControllersRequest(request *DaoClusterTypes.ListApplicationsRequest) *DaoClusterTypes.ListControllersRequest {
	listControllersReq := DaoClusterTypes.NewListControllersRequest()

	for _, applicationObjectMeta := range request.ApplicationObjectMeta {
		alamedaScaler := &Metadata.ObjectMeta{}

		if applicationObjectMeta.ObjectMeta != nil {
			alamedaScaler.Name = applicationObjectMeta.ObjectMeta.Name
			alamedaScaler.Namespace = applicationObjectMeta.ObjectMeta.Namespace
			alamedaScaler.ClusterName = applicationObjectMeta.ObjectMeta.ClusterName
		} else {
			alamedaScaler = nil
		}

		controllerObjectMeta := DaoClusterTypes.NewControllerObjectMeta(nil, alamedaScaler, "", applicationObjectMeta.ScalingTool)
		listControllersReq.ControllerObjectMeta = append(listControllersReq.ControllerObjectMeta, controllerObjectMeta)
	}

	return listControllersReq
}

func (p *Application) genDeleteControllersRequest(request *DaoClusterTypes.DeleteApplicationsRequest) *DaoClusterTypes.DeleteControllersRequest {
	delControllersReq := DaoClusterTypes.NewDeleteControllersRequest()

	for _, applicationObjectMeta := range request.ApplicationObjectMeta {
		alamedaScaler := &Metadata.ObjectMeta{}

		if applicationObjectMeta.ObjectMeta != nil {
			alamedaScaler.Name = applicationObjectMeta.ObjectMeta.Name
			alamedaScaler.Namespace = applicationObjectMeta.ObjectMeta.Namespace
			alamedaScaler.ClusterName = applicationObjectMeta.ObjectMeta.ClusterName
		} else {
			alamedaScaler = nil
		}

		controllerObjectMeta := DaoClusterTypes.NewControllerObjectMeta(nil, alamedaScaler, "", applicationObjectMeta.ScalingTool)
		delControllersReq.ControllerObjectMeta = append(delControllersReq.ControllerObjectMeta, controllerObjectMeta)
	}

	return delControllersReq
}
