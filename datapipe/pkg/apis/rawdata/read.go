package rawdata

import (
	"fmt"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	Common "github.com/containers-ai/api/common"
	Rawdata "github.com/containers-ai/api/datapipe/rawdata"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (c *ServiceRawdata) ReadRawdata(ctx context.Context, in *Rawdata.ReadRawdataRequest) (*Rawdata.ReadRawdataResponse, error) {
	scope.Debug("Request received from ReadRawdata grpc function: " + AlamedaUtils.InterfaceToString(in))

	var (
		err     error
		rawdata = make([]*Common.ReadRawdata, 0)
	)

	promthConfig := InternalPromth.Config{}
	promthConfig.URL = c.Config.Prometheus.URL
	promthConfig.BearerTokenFile = c.Config.Prometheus.BearerTokenFile
	promthConfig.TLSConfig = &InternalPromth.TLSConfig{}
	promthConfig.TLSConfig.InsecureSkipVerify = c.Config.Prometheus.TLSConfig.InsecureSkipVerify

	switch in.GetDatabaseType() {
	case Common.DatabaseType_PROMETHEUS:
		rawdata, err = InternalPromth.ReadRawdata(&promthConfig, in.GetQueries())
	default:
		err = errors.New(fmt.Sprintf("database type(%s) is not supported", Common.DatabaseType_name[int32(in.GetDatabaseType())]))
	}

	if err != nil {
		scope.Errorf("failed to read rawdata: %v", err)
		response := &Rawdata.ReadRawdataResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
			Rawdata: rawdata,
		}
		return response, err
	}

	response := &Rawdata.ReadRawdataResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		Rawdata: rawdata,
	}

	return response, nil
}
