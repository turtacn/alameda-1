package apiserver

import (
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"

	commonAPI "github.com/containers-ai/api/common"
	fedRawdataAPI "github.com/containers-ai/federatorai-api/apiserver/rawdata"

	fedRawAPI "github.com/containers-ai/federatorai-api/apiserver/rawdata"

	"fmt"
	"google.golang.org/grpc"
	"time"
)

type loginCreds struct {
	Username string
	Password string
}

func (c *loginCreds) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{
		"username": c.Username,
		"password": c.Password,
	}, nil
}
func (c *loginCreds) RequireTransportSecurity() bool {
	return false
}

func WriteRawdata(apiServerAddress string, rowDataList []*commonAPI.WriteRawdata) (*status.Status, error) {
	request := &fedRawdataAPI.WriteRawdataRequest{
		Rawdata: rowDataList,
	}

	conn, err := grpc.Dial(apiServerAddress, grpc.WithInsecure(), grpc.WithPerRPCCredentials(&loginCreds{Username: "user", Password: "password"}))
	if err != nil {
		fmt.Print(err)
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}
	defer conn.Close()
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)

	client := fedRawAPI.NewRawdataServiceClient(conn)
	_, err = client.WriteRawdata(ctx, request)
	if err != nil {
		fmt.Print(err)
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	return &status.Status{Code: int32(code.Code_OK)}, nil
}
