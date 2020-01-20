package probe

import (
	"context"
	"fmt"
	"os/exec"

	datahub_client "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"google.golang.org/grpc"
)

type ReadinessProbeConfig struct {
	DatahubAddr   string
	AdmCtrSrvPort int32
}

func queryDatahub(datahubAddr string) error {
	conn, err := grpc.Dial(datahubAddr, grpc.WithInsecure())
	if conn != nil {
		defer conn.Close()
	}
	if err != nil {
		return err
	}

	datahubServiceClnt := datahub_client.NewDatahubServiceClient(conn)
	res, err := datahubServiceClnt.ListNodes(context.Background(), &datahub_resources.ListNodesRequest{})
	if err != nil {
		return err
	}

	if len(res.GetNodes()) == 0 {
		return fmt.Errorf("No nodes found in datahub")
	}

	return err
}

func queryWebhookSrv(port int32) error {

	svcURL := fmt.Sprintf("https://localhost:%s", fmt.Sprint(port))
	curlCmd := exec.Command("curl", "-k", svcURL)

	_, err := curlCmd.CombinedOutput()
	if err != nil {
		return err
	}

	return nil
}
