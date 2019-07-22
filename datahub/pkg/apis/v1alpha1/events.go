package v1alpha1

import (
	DaoEvent "github.com/containers-ai/alameda/datahub/pkg/dao/event"
	DatahubV1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (s *ServiceV1alpha1) CreateEvents(ctx context.Context, in *DatahubV1alpha1.CreateEventsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateEvents grpc function")

	eventDAO := DaoEvent.NewEventWithConfig(s.Config.InfluxDB, s.Config.RabbitMQ)

	err := eventDAO.CreateEvents(in)
	if err != nil {
		scope.Error(err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	err = eventDAO.SendEvents(in)
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

	eventDAO := DaoEvent.NewEventWithConfig(s.Config.InfluxDB, s.Config.RabbitMQ)
	events, err := eventDAO.ListEvents(in)
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
