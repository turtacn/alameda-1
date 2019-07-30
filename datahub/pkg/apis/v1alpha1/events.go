package v1alpha1

import (
	EventMgt "github.com/containers-ai/alameda/internal/pkg/event-mgt"
	DatahubV1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (s *ServiceV1alpha1) CreateEvents(ctx context.Context, in *DatahubV1alpha1.CreateEventsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateEvents grpc function")

	err := EventMgt.PostEvents(in)
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

func (s *ServiceV1alpha1) ListEvents(ctx context.Context, in *DatahubV1alpha1.ListEventsRequest) (*DatahubV1alpha1.ListEventsResponse, error) {
	scope.Debug("Request received from ListEvents grpc function")

	events, err := EventMgt.ListEvents(in)
	if err != nil {
		scope.Error(err.Error())
		return &DatahubV1alpha1.ListEventsResponse{
			Status: &status.Status{
				Code: int32(code.Code_INTERNAL),
			},
			Events: events,
		}, nil
	}

	response := &DatahubV1alpha1.ListEventsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		Events: events,
	}

	return response, nil
}
