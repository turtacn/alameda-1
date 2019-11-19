package v1alpha1

import (
	DaoRecommendation "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/recommendations"
	AutoScalingV1alpha1 "github.com/containers-ai/alameda/operator/api/v1alpha1"
	ReconcilerAlamedaRecommendation "github.com/containers-ai/alameda/operator/pkg/reconciler/alamedarecommendation"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiRecommendations "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/recommendations"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
	K8sErrors "k8s.io/apimachinery/pkg/api/errors"
	K8sTypes "k8s.io/apimachinery/pkg/types"
)

// CreatePodRecommendations add pod recommendations information to database
func (s *ServiceV1alpha1) CreatePodRecommendations(ctx context.Context, in *ApiRecommendations.CreatePodRecommendationsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreatePodRecommendations grpc function: " + AlamedaUtils.InterfaceToString(in))

	podRecommendations := in.GetPodRecommendations()
	for _, podRecommendation := range podRecommendations {
		podNS := podRecommendation.GetObjectMeta().GetNamespace()
		podName := podRecommendation.GetObjectMeta().GetName()
		alamedaRecommendation := &AutoScalingV1alpha1.AlamedaRecommendation{}

		if err := s.K8SClient.Get(context.TODO(), K8sTypes.NamespacedName{
			Namespace: podNS,
			Name:      podName,
		}, alamedaRecommendation); err == nil {
			alamedarecommendationReconciler := ReconcilerAlamedaRecommendation.NewReconciler(s.K8SClient, alamedaRecommendation)
			if alamedaRecommendation, err = alamedarecommendationReconciler.UpdateResourceRecommendation(podRecommendation); err == nil {
				if err = s.K8SClient.Update(context.TODO(), alamedaRecommendation); err != nil {
					scope.Error(err.Error())
				}
			}
		} else if !K8sErrors.IsNotFound(err) {
			scope.Error(err.Error())
		}
	}

	containerDAO := DaoRecommendation.NewContainerRecommendationsDAO(*s.Config)
	if err := containerDAO.CreatePodRecommendations(in); err != nil {
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

// ListPodRecommendations list pod recommendations
func (s *ServiceV1alpha1) ListPodRecommendations(ctx context.Context, in *ApiRecommendations.ListPodRecommendationsRequest) (*ApiRecommendations.ListPodRecommendationsResponse, error) {
	scope.Debug("Request received from ListPodRecommendations grpc function: " + AlamedaUtils.InterfaceToString(in))

	containerDAO := DaoRecommendation.NewContainerRecommendationsDAO(*s.Config)
	podRecommendations, err := containerDAO.ListPodRecommendations(in)
	if err != nil {
		scope.Error(err.Error())
		return &ApiRecommendations.ListPodRecommendationsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	res := &ApiRecommendations.ListPodRecommendationsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		PodRecommendations: podRecommendations,
	}
	scope.Debug("Response sent from ListPodRecommendations grpc function: " + AlamedaUtils.InterfaceToString(res))
	return res, nil
}

// ListAvailablePodRecommendations list pod recommendations
func (s *ServiceV1alpha1) ListAvailablePodRecommendations(ctx context.Context, in *ApiRecommendations.ListPodRecommendationsRequest) (*ApiRecommendations.ListPodRecommendationsResponse, error) {
	scope.Debug("Request received from ListAvailablePodRecommendations grpc function: " + AlamedaUtils.InterfaceToString(in))

	containerDAO := DaoRecommendation.NewContainerRecommendationsDAO(*s.Config)
	podRecommendations, err := containerDAO.ListAvailablePodRecommendations(in)
	if err != nil {
		scope.Error(err.Error())
		return &ApiRecommendations.ListPodRecommendationsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	res := &ApiRecommendations.ListPodRecommendationsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		PodRecommendations: podRecommendations,
	}
	scope.Debug("Response sent from ListPodRecommendations grpc function: " + AlamedaUtils.InterfaceToString(res))
	return res, nil
}
