package v1alpha1

import (
	DaoCluster "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus"
	FormatRequest "github.com/containers-ai/alameda/datahub/pkg/formatconversion/requests"
	FormatResponse "github.com/containers-ai/alameda/datahub/pkg/formatconversion/responses"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

// CreatePods add containers information of pods to database
func (s *ServiceV1alpha1) CreatePods(ctx context.Context, in *ApiResources.CreatePodsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreatePods grpc function: " + AlamedaUtils.InterfaceToString(in))

	requestExtended := FormatRequest.CreatePodsRequestExtended{CreatePodsRequest: *in}
	if err := requestExtended.Validate(); err != nil {
		return &status.Status{
			Code:    int32(code.Code_INVALID_ARGUMENT),
			Message: err.Error(),
		}, nil
	}

	podDAO := DaoCluster.NewPodDAO(*s.Config)
	if err := podDAO.CreatePods(requestExtended.ProducePods()); err != nil {
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

// ListAlamedaPods returns predicted pods
func (s *ServiceV1alpha1) ListPods(ctx context.Context, in *ApiResources.ListPodsRequest) (*ApiResources.ListPodsResponse, error) {
	scope.Debug("Request received from ListAlamedaPods grpc function: " + AlamedaUtils.InterfaceToString(in))

	requestExt := FormatRequest.ListPodsRequestExtended{ListPodsRequest: in}
	if err := requestExt.Validate(); err != nil {
		return &ApiResources.ListPodsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INVALID_ARGUMENT),
				Message: err.Error(),
			},
		}, nil
	}

	podDAO := DaoCluster.NewPodDAO(*s.Config)
	pds, err := podDAO.ListPods(requestExt.ProduceRequest())
	if err != nil {
		scope.Errorf("ListNodes failed: %+v", err)
		return &ApiResources.ListPodsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	pods := make([]*ApiResources.Pod, 0)
	for _, pd := range pds {
		podExtended := FormatResponse.PodExtended{Pod: pd}
		pod := podExtended.ProducePod()
		pods = append(pods, pod)
	}

	response := ApiResources.ListPodsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		Pods: pods,
	}

	return &response, nil
}

// DeletePods update containers information of pods to database
func (s *ServiceV1alpha1) DeletePods(ctx context.Context, in *ApiResources.DeletePodsRequest) (*status.Status, error) {
	scope.Debug("Request received from DeletePods grpc function: " + AlamedaUtils.InterfaceToString(in))

	requestExt := FormatRequest.DeletePodsRequestExtended{DeletePodsRequest: in}
	if err := requestExt.Validate(); err != nil {
		return &status.Status{
			Code:    int32(code.Code_INVALID_ARGUMENT),
			Message: err.Error(),
		}, nil
	}

	podDAO := DaoCluster.NewPodDAO(*s.Config)
	if err := podDAO.DeletePods(requestExt.ProduceRequest()); err != nil {
		scope.Errorf("failed to delete pods: %+v", err)
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}
