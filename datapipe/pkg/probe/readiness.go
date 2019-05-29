package probe

import (
	"context"
	"fmt"
	Rawdata "github.com/containers-ai/api/datapipe/rawdata"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type ReadinessProbeConfig struct {
	BindAddr string
}

func queryApiServer(bindAddr string) error {
	conn, err := grpc.Dial(fmt.Sprintf("localhost%s", bindAddr), grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	rawdata := Rawdata.NewRawdataServiceClient(conn)
	_, err = rawdata.ReadRawdata(context.Background(), &Rawdata.ReadRawdataRequest{})
	if err != nil {
		return errors.Wrap(err, "failed to read rawdata")
	}

	return nil
}
