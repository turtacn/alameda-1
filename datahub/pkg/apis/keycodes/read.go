package keycodes

import (
	KeycodeMgt "github.com/containers-ai/alameda/datahub/pkg/account-mgt/keycodes"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	Keycodes "github.com/containers-ai/api/datahub/keycodes"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (c *ServiceKeycodes) ListKeycodes(ctx context.Context, in *Keycodes.ListKeycodesRequest) (*Keycodes.ListKeycodesResponse, error) {
	scope.Debug("Request received from ListKeycodes grpc function: " + AlamedaUtils.InterfaceToString(in))

	var (
		err      error
		keycodes []*KeycodeMgt.Keycode
		summary  *KeycodeMgt.Keycode
	)

	keycodeMgt := KeycodeMgt.NewKeycodeMgt()

	if len(in.GetKeycodes()) == 0 {
		// Read all keycodes
		keycodes, summary, err = keycodeMgt.GetAllKeycodes()
	} else {
		// Read keycodes
		keycodes, summary, err = keycodeMgt.GetKeycodes(in.GetKeycodes())
	}

	if err != nil {
		scope.Error(err.Error())
		response := &Keycodes.ListKeycodesResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}
		return response, nil
	}

	// Prepare response
	response := &Keycodes.ListKeycodesResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		Keycodes: TransformKeycodeList(keycodes),
		Summary:  TransformKeycode(summary),
	}

	return response, nil
}
