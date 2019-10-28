package v1alpha1

import (
	"fmt"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
	ApiRawdata "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/rawdata"
	Common "github.com/containers-ai/api/common"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

// Read rawdata from database
func (s *ServiceV1alpha1) ReadRawdata(ctx context.Context, in *ApiRawdata.ReadRawdataRequest) (*ApiRawdata.ReadRawdataResponse, error) {
	scope.Debug("Request received from ReadRawdata grpc function")

	var (
		err     error
		rawdata = make([]*Common.ReadRawdata, 0)
	)

	switch in.GetDatabaseType() {
	case Common.DatabaseType_INFLUXDB:
		rawdata, err = InternalInflux.ReadRawdata(s.Config.InfluxDB, in.GetQueries())
	case Common.DatabaseType_PROMETHEUS:
		rawdata, err = InternalPromth.ReadRawdata(s.Config.Prometheus, in.GetQueries())
	default:
		err = errors.New(fmt.Sprintf("database type(%s) is not supported", Common.DatabaseType_name[int32(in.GetDatabaseType())]))
	}

	if err != nil {
		scope.Errorf("api ReadRawdata failed: %v", err)
		response := &ApiRawdata.ReadRawdataResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
			Rawdata: rawdata,
		}
		return response, err
	}

	response := &ApiRawdata.ReadRawdataResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		Rawdata: rawdata,
	}

	return response, nil
}

// Write rawdata to database
func (s *ServiceV1alpha1) WriteRawdata(ctx context.Context, in *ApiRawdata.WriteRawdataRequest) (*status.Status, error) {
	scope.Debug("Request received from WriteRawdata grpc function")

	var (
		err error
	)

	switch in.GetDatabaseType() {
	case Common.DatabaseType_INFLUXDB:
		err = InternalInflux.WriteRawdata(s.Config.InfluxDB, in.GetRawdata())
	case Common.DatabaseType_PROMETHEUS:
		err = errors.New(fmt.Sprintf("database type(%s) is not supported yet", Common.DatabaseType_name[int32(in.GetDatabaseType())]))
	default:
		err = errors.New(fmt.Sprintf("database type(%s) is not supported", Common.DatabaseType_name[int32(in.GetDatabaseType())]))
	}

	if err != nil {
		scope.Errorf("api WriteRawdata failed: %v", err)
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, err
	}

	return &status.Status{Code: int32(code.Code_OK)}, nil
}
