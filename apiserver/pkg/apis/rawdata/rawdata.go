package rawdata

import (
	APIServerConfig "github.com/containers-ai/alameda/apiserver/pkg/config"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	Log "github.com/containers-ai/alameda/pkg/utils/log"
	Datahub "github.com/containers-ai/api/datahub/rawdata"
	Rawdata "github.com/containers-ai/federatorai-api/apiserver/rawdata"
	"golang.org/x/net/context"
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
	scope.Debug("Request received from ReadRawdata grpc function: " + AlamedaUtils.InterfaceToString(in))

	out := new(Rawdata.ReadRawdataResponse)
	return out, nil
}

func (c *ServiceRawdata) WriteRawdata(ctx context.Context, in *Rawdata.WriteRawdataRequest) (*status.Status, error) {
	scope.Debug("Request received from WriteRawdata grpc function: " + AlamedaUtils.InterfaceToString(in))

	// Create connection to datahub
	address := c.Config.Datahub.Address
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	defer conn.Close()
	if err != nil {
		panic(err)
	}

	// Instance rawdata service of datahub
	client := Datahub.NewRawdataServiceClient(conn)

	// Rebuild write rawdata request for datahub
	request := &Datahub.WriteRawdataRequest{}
	for _, rdata := range in.Rawdata {
		request.Rawdata = append(request.Rawdata, rdata)
	}

	// Write rawdata to datahub
	response, err := client.WriteRawdata(context.Background(), request)

	return response, err
}
