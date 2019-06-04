package apiserver

import (
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"

	"fmt"
	"google.golang.org/grpc"
)

func CreateClient(apiServerAddress string) (*grpc.ClientConn, datahub_v1alpha1.DatahubServiceClient, error) {

	conn, err := grpc.Dial(apiServerAddress, grpc.WithInsecure(), grpc.WithPerRPCCredentials(&loginCreds{Username: "user", Password: "password"}))
	if err != nil {
		fmt.Print(err)
		return nil, nil, err
	}

	client := datahub_v1alpha1.NewDatahubServiceClient(conn)
	return conn, client, nil
}
