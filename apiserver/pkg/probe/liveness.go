package probe

import (
	"context"
	"fmt"
	Ping "github.com/containers-ai/federatorai-api/apiserver/ping"
	"google.golang.org/grpc"
)

type LivenessProbeConfig struct {
	BindAddr string
}

func pingApiServer(bindAddr string) error {
	conn, err := grpc.Dial(fmt.Sprintf("localhost%s", bindAddr), grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	ping := Ping.NewPingServiceClient(conn)
	_, err = ping.Ping(context.Background(), &Ping.PingRequest{})

	return err
}
