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

func (c *ServiceKeycodes) AddKeycode(ctx context.Context, in *Keycodes.AddKeycodeRequest) (*Keycodes.AddKeycodeResponse, error) {
	scope.Debug("Request received from AddKeycode grpc function: " + AlamedaUtils.InterfaceToString(in))

	keycodeMgt := KeycodeMgt.NewKeycodeMgt()

	// Validate request
	if in.GetKeycode() == "" {
		return &Keycodes.AddKeycodeResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INVALID_ARGUMENT),
				Message: Errors.GetReason(Errors.ReasonMissingFieldReq, "Keycode"),
			},
		}, nil
	}

	// Add keycode
	err := keycodeMgt.AddKeycode(in.GetKeycode())
	if err != nil {
		scope.Error(err.Error())
		return &Keycodes.AddKeycodeResponse{
			Status: &status.Status{
				Code:    CategorizeKeycodeErrorId(err.(*IError).ErrorID),
				Message: err.Error(),
			},
		}, nil
	}

	scope.Infof("Successfully to add keycode(%s)", in.GetKeycode())

	keycode, err := keycodeMgt.GetKeycode(in.GetKeycode())
	return &Keycodes.AddKeycodeResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		Keycode: TransformKeycode(keycode),
	}, nil
}
