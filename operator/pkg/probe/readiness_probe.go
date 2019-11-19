package probe

import (
	"context"
	"os/exec"

	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"google.golang.org/grpc"
)

type ReadinessProbeConfig struct {
	WHSrvPort   int32
	DatahubAddr string
}

func queryDatahub(datahubAddr string) error {
	conn, err := grpc.Dial(datahubAddr, grpc.WithInsecure())
	if err != nil {
		return err
	}

	defer conn.Close()
	datahubServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(conn)
	_, err = datahubServiceClnt.ListNodes(context.Background(), &datahub_resources.ListNodesRequest{})
	if err != nil {
		return err
	}

	return err
}

func queryWebhookSrv(svcURL string) error {
	curlCmd := exec.Command("curl", "-k", svcURL)

	_, err := curlCmd.CombinedOutput()
	if err != nil {
		return err
	}

	return nil
}
