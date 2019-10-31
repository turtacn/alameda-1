package v1alpha1

import (
	DaoClusterStatusInflux "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/influxdb"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (s *ServiceV1alpha1) CreateControllers(ctx context.Context, in *ApiResources.CreateControllersRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateControllers grpc function: " + AlamedaUtils.InterfaceToString(in))

	controllerDAO := &DaoClusterStatusInflux.Controller{
		InfluxDBConfig: *s.Config.InfluxDB,
	}

	err := controllerDAO.CreateControllers(in.GetControllers())
	if err != nil {
		scope.Error(err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *ServiceV1alpha1) ListControllers(ctx context.Context, in *ApiResources.ListControllersRequest) (*ApiResources.ListControllersResponse, error) {
	scope.Debug("Request received from ListControllers grpc function: " + AlamedaUtils.InterfaceToString(in))

	controllerDAO := &DaoClusterStatusInflux.Controller{
		InfluxDBConfig: *s.Config.InfluxDB,
	}

	controllers, err := controllerDAO.ListControllers(in)
	if err != nil {
		scope.Error(err.Error())
		return &ApiResources.ListControllersResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	response := ApiResources.ListControllersResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		Controllers: controllers,
	}
	return &response, nil
}

func (s *ServiceV1alpha1) DeleteControllers(ctx context.Context, in *ApiResources.DeleteControllersRequest) (*status.Status, error) {
	scope.Debug("Request received from DeleteControllers grpc function: " + AlamedaUtils.InterfaceToString(in))

	controllerDAO := &DaoClusterStatusInflux.Controller{
		InfluxDBConfig: *s.Config.InfluxDB,
	}

	err := controllerDAO.DeleteControllers(in)
	if err != nil {
		scope.Error(err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}
