package v1alpha1

import (
	DaoPlanning "github.com/containers-ai/alameda/datahub/pkg/dao/planning"
	DaoPlanningImpl "github.com/containers-ai/alameda/datahub/pkg/dao/planning/impl"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	DatahubV1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

// CreatePodPlannings add pod plannings information to database
func (s *ServiceV1alpha1) CreatePodPlannings(ctx context.Context, in *DatahubV1alpha1.CreatePodPlanningsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreatePodPlannings grpc function: " + AlamedaUtils.InterfaceToString(in))

	var containerDAO DaoPlanning.ContainerOperation = &DaoPlanningImpl.Container{
		InfluxDBConfig: *s.Config.InfluxDB,
	}

	if err := containerDAO.AddPodPlannings(in); err != nil {
		scope.Error(err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, err
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

// CreateControllerPlannings add controller plannings information to database
func (s *ServiceV1alpha1) CreateControllerPlannings(ctx context.Context, in *DatahubV1alpha1.CreateControllerPlanningsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateControllerPlannings grpc function: " + AlamedaUtils.InterfaceToString(in))

	controllerDAO := DaoPlanningImpl.Controller{
		InfluxDBConfig: *s.Config.InfluxDB,
	}

	controllerPlanningList := in.GetControllerPlannings()
	err := controllerDAO.AddControllerPlannings(controllerPlanningList)

	if err != nil {
		scope.Error(err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, err
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

// ListPodPlannings list pod plannings
func (s *ServiceV1alpha1) ListPodPlannings(ctx context.Context, in *DatahubV1alpha1.ListPodPlanningsRequest) (*DatahubV1alpha1.ListPodPlanningsResponse, error) {
	scope.Debug("Request received from ListPodPlannings grpc function: " + AlamedaUtils.InterfaceToString(in))

	var containerDAO DaoPlanning.ContainerOperation = &DaoPlanningImpl.Container{
		InfluxDBConfig: *s.Config.InfluxDB,
	}

	podPlannings, err := containerDAO.ListPodPlannings(in)
	if err != nil {
		scope.Error(err.Error())
		return &DatahubV1alpha1.ListPodPlanningsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	res := &DatahubV1alpha1.ListPodPlanningsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		PodPlannings: podPlannings,
	}
	scope.Debug("Response sent from ListPodPlannings grpc function: " + AlamedaUtils.InterfaceToString(res))
	return res, nil
}

// ListControllerPlannings list controller plannings
func (s *ServiceV1alpha1) ListControllerPlannings(ctx context.Context, in *DatahubV1alpha1.ListControllerPlanningsRequest) (*DatahubV1alpha1.ListControllerPlanningsResponse, error) {
	scope.Debug("Request received from ListControllerPlannings grpc function: " + AlamedaUtils.InterfaceToString(in))

	controllerDAO := &DaoPlanningImpl.Controller{
		InfluxDBConfig: *s.Config.InfluxDB,
	}

	controllerPlannings, err := controllerDAO.ListControllerPlannings(in)
	if err != nil {
		scope.Errorf("api ListControllerPlannings failed: %v", err)
		response := &DatahubV1alpha1.ListControllerPlanningsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
			ControllerPlannings: controllerPlannings,
		}
		return response, nil
	}

	response := &DatahubV1alpha1.ListControllerPlanningsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		ControllerPlannings: controllerPlannings,
	}

	scope.Debug("Response sent from ListControllerPlannings grpc function: " + AlamedaUtils.InterfaceToString(response))
	return response, nil
}
