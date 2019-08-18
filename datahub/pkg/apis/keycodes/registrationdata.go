package keycodes

import (
	KeycodeMgt "github.com/containers-ai/alameda/datahub/pkg/account-mgt/keycodes"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	Keycodes "github.com/containers-ai/api/datahub/keycodes"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (c *ServiceKeycodes) GenerateRegistrationData(ctx context.Context, in *empty.Empty) (*Keycodes.GenerateRegistrationDataResponse, error) {
	scope.Debug("Request received from GenerateRegistrationData grpc function: " + AlamedaUtils.InterfaceToString(in))

	keycodeMgt := KeycodeMgt.NewKeycodeMgt()

	// Generate registration data
	registrationData, err := keycodeMgt.GetRegistrationData()
	if err != nil {
		scope.Error(err.Error())
		return &Keycodes.GenerateRegistrationDataResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	return &Keycodes.GenerateRegistrationDataResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		Data: registrationData,
	}, nil
}
