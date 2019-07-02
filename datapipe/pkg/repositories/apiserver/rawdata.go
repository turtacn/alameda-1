package apiserver

import (
	"fmt"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	commonAPI "github.com/containers-ai/api/common"
	fedRawdataAPI "github.com/containers-ai/federatorai-api/apiserver/rawdata"
)

func WriteRawdata(apiServerAddress string, databaseType commonAPI.DatabaseType, rowDataList []*commonAPI.WriteRawdata) (*status.Status, error) {
	var (
		stat *status.Status
		request = &fedRawdataAPI.WriteRawdataRequest{DatabaseType: databaseType, Rawdata: rowDataList}
	)

	// Create connection to API server
	conn, err := grpc.Dial(apiServerAddress, grpc.WithInsecure())
	if err != nil {
		fmt.Print(err)
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}
	defer conn.Close()

	client := fedRawdataAPI.NewRawdataServiceClient(conn)

	// Send write rawdata request to API server
	stat, err = client.WriteRawdata(NewContextWithCredential(), request)

	// Check if needs to resend request
	if stat != nil {
		if NeedResendRequest(stat, err) {
			stat, err = client.WriteRawdata(NewContextWithCredential(), request)
		}
	}

	stat, _ = CheckResponse(stat, err)

	return stat, nil
}
