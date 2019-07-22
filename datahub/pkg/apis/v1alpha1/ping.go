package v1alpha1

import (
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (s *ServiceV1alpha1) Ping(ctx context.Context, in *empty.Empty) (*status.Status, error) {
	scope.Debug("Request received from Ping grpc function")

	return &status.Status{Code: int32(code.Code_OK)}, nil
}
