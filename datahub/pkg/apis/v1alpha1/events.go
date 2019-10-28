package v1alpha1

import (
	EventMgt "github.com/containers-ai/alameda/internal/pkg/event-mgt"
	ApiEvents "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/events"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (s *ServiceV1alpha1) CreateEvents(ctx context.Context, in *ApiEvents.CreateEventsRequest) (*status.Status, error) {
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

func (s *ServiceV1alpha1) ListEvents(ctx context.Context, in *ApiEvents.ListEventsRequest) (*ApiEvents.ListEventsResponse, error) {
	scope.Debug("Request received from ListEvents grpc function")

	events, err := EventMgt.ListEvents(in)
	if err != nil {
		scope.Error(err.Error())
		return &ApiEvents.ListEventsResponse{
			Status: &status.Status{
				Code: int32(code.Code_INTERNAL),
			},
			Events: events,
		}, nil
	}

	response := &ApiEvents.ListEventsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		Events: events,
	}

	return response, nil
}
