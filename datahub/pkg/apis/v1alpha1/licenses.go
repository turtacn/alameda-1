package v1alpha1

import (
	KeycodeMgt "github.com/containers-ai/alameda/datahub/pkg/account-mgt/keycodes"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiLicenses "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/licenses"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (s *ServiceV1alpha1) GetLicense(ctx context.Context, in *empty.Empty) (*ApiLicenses.GetLicenseResponse, error) {
	scope.Debug("Request received from GetLicense grpc function: " + AlamedaUtils.InterfaceToString(in))

	keycodeMgt := KeycodeMgt.NewKeycodeMgt()
	license := &ApiLicenses.License{Valid: keycodeMgt.IsValid()}

	response := &ApiLicenses.GetLicenseResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		License: license,
	}

	scope.Debug("Response sent from GetLicense grpc function: " + AlamedaUtils.InterfaceToString(response))
	return response, nil
}
