package probe

import (
	"context"
	"fmt"

	DatahubV1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/grpc"
)

type LivenessProbeConfig struct {
	BindAddr string
}

func queryDatahub(bindAddr string) error {
	conn, err := grpc.Dial(fmt.Sprintf("localhost%s", bindAddr), grpc.WithInsecure())
	if conn != nil {
		defer conn.Close()
	}
	if err != nil {
		return err
	}

	client := DatahubV1alpha1.NewDatahubServiceClient(conn)

	status, err := client.Ping(context.Background(), &empty.Empty{})
	if err != nil {
		return errors.Wrap(err, "failed to connect to datahub")
	}

	if status.GetCode() != int32(code.Code_OK) {
		return errors.Wrap(err, status.GetMessage())
	}

	return nil
}
