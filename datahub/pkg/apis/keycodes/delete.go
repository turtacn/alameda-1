package keycodes

import (
	KeycodeMgt "github.com/containers-ai/alameda/datahub/pkg/account-mgt/keycodes"
	Errors "github.com/containers-ai/alameda/internal/pkg/errors"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	Keycodes "github.com/containers-ai/api/datahub/keycodes"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (c *ServiceKeycodes) DeleteKeycode(ctx context.Context, in *Keycodes.DeleteKeycodeRequest) (*status.Status, error) {
	scope.Debug("Request received from DeleteKeycode grpc function: " + AlamedaUtils.InterfaceToString(in))

	keycodeMgt := KeycodeMgt.NewKeycodeMgt()

	// Validate request
	if in.GetKeycode() == "" {
		return &status.Status{
			Code:    int32(code.Code_INVALID_ARGUMENT),
			Message: Errors.GetReason(Errors.ReasonMissingFieldReq, "Keycode"),
		}, nil
	}

	// Delete keycode
	err := keycodeMgt.DeleteKeycode(in.GetKeycode())
	if err != nil {
		scope.Error(err.Error())
		return &status.Status{
			Code:    CategorizeKeycodeErrorId(err.(*IError).ErrorID),
			Message: err.Error(),
		}, nil
	}

	scope.Infof("Successfully to delete keycode(%s)", in.GetKeycode())

	return &status.Status{Code: int32(code.Code_OK)}, nil
}
