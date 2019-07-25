package rawdata

import (
	RepoApiServer "github.com/containers-ai/alameda/datapipe/pkg/repositories/apiserver"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	Rawdata "github.com/containers-ai/api/datapipe/rawdata"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (c *ServiceRawdata) WriteRawdata(ctx context.Context, in *Rawdata.WriteRawdataRequest) (*status.Status, error) {
	scope.Debug("Request received from WriteRawdata grpc function: " + AlamedaUtils.InterfaceToString(in))

	stat, err := RepoApiServer.WriteRawdata(c.Config.APIServer.Address, in.GetDatabaseType(), in.GetRawdata())
	if err != nil {
		scope.Errorf("failed to write rawdata: %v", err)
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	return stat, nil
}
