package rawdata

import (
	APIServerConfig "github.com/containers-ai/alameda/apiserver/pkg/config"
	Log "github.com/containers-ai/alameda/pkg/utils/log"
	Datahub "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	Rawdata "github.com/containers-ai/federatorai-api/apiserver/rawdata"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
)

var (
	scope = Log.RegisterScope("apiserver", "apiserver log", 0)
)

type ServiceRawdata struct {
	Config *APIServerConfig.Config
}

func NewServiceRawdata(cfg *APIServerConfig.Config) *ServiceRawdata {
	service := ServiceRawdata{}
	service.Config = cfg
	return &service
}

func (c *ServiceRawdata) ReadRawdata(ctx context.Context, in *Rawdata.ReadRawdataRequest) (*Rawdata.ReadRawdataResponse, error) {
	scope.Debug("Request received from ReadRawdata grpc function")

	response := Rawdata.ReadRawdataResponse{}

	// Create connection to datahub
	address := c.Config.Datahub.Address
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	defer conn.Close()
	if err != nil {
		panic(err)
	}

	// Instance rawdata service of datahub
	client := Datahub.NewDatahubServiceClient(conn)

	// Rebuild read rawdata request for datahub
	request := &Datahub.ReadRawdataRequest{}
	request.DatabaseType = in.GetDatabaseType()
	for _, query := range in.GetQueries() {
		request.Queries = append(request.Queries, query)
	}

	// Read rawdata from datahub
	if result, err := client.ReadRawdata(context.Background(), request); err != nil {
		scope.Errorf("apiserver ReadRawdata failed: %v", err)
		response.Status = &status.Status{Code: int32(code.Code_INTERNAL)}
	} else {
		response.Status = &status.Status{Code: int32(code.Code_OK)}
		response.Rawdata = result.Rawdata
	}

	return &response, nil
}

func (c *ServiceRawdata) WriteRawdata(ctx context.Context, in *Rawdata.WriteRawdataRequest) (*status.Status, error) {
	scope.Debug("Request received from WriteRawdata grpc function")

	// Create connection to datahub
	address := c.Config.Datahub.Address
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	defer conn.Close()
	if err != nil {
		panic(err)
	}

	// Instance rawdata service of datahub
	client := Datahub.NewDatahubServiceClient(conn)

	// Rebuild write rawdata request for datahub
	request := &Datahub.WriteRawdataRequest{}
	request.DatabaseType = in.GetDatabaseType()
	for _, rdata := range in.GetRawdata() {
		request.Rawdata = append(request.Rawdata, rdata)
	}

	// Write rawdata to datahub
	response, err := client.WriteRawdata(context.Background(), request)

	return response, err
}
