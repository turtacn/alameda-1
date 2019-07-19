package datahub

import (
	"fmt"
	DatahubV1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"google.golang.org/grpc"
)

func CreateClient(apiServerAddress string) (*grpc.ClientConn, DatahubV1alpha1.DatahubServiceClient, error) {
	// Create connection to datahub
	conn, err := grpc.Dial(apiServerAddress, grpc.WithInsecure())
	if err != nil {
		fmt.Print(err)
		return nil, nil, err
	}

	// Instance service client of datahub
	client := DatahubV1alpha1.NewDatahubServiceClient(conn)
	return conn, client, nil
}
