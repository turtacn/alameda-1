package client

import (
	"context"

	datahubutils "github.com/containers-ai/alameda/operator/pkg/utils/datahub"
	alamutils "github.com/containers-ai/alameda/pkg/utils"
	logUtil "github.com/containers-ai/alameda/pkg/utils/log"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/grpc"
)

var (
	scope = logUtil.RegisterScope("datahub_client", "datahub client", 0)
)

type K8SResource struct {
}

// NewK8SResource return K8SResource instance
func NewK8SResource() *K8SResource {
	return &K8SResource{}
}

func (repo *K8SResource) CreateAlamedaWatchedResource(resources []*datahub_v1alpha1.Controller) error {
	conn, err := grpc.Dial(datahubutils.GetDatahubAddress(), grpc.WithInsecure())
	if err != nil {
		return errors.Errorf("create controllers to datahub failed: %s", err.Error())
	}

	defer conn.Close()
	datahubServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(conn)	
	req := datahub_v1alpha1.CreateControllersRequest{
		Controllers: resources,
	}
	scope.Debugf("Create controllers to datahub with request %s.", alamutils.InterfaceToString(req))
	resp, err := datahubServiceClnt.CreateControllers(context.Background(), &req)
	if err != nil {
		return errors.Errorf("Create controllers failed: %s", err.Error())
	} else if resp.Code != int32(code.Code_OK) {
		return errors.Errorf("Create controllers failed: receive response: code: %d, message: %s", resp.Code, resp.Message)
	}

	return nil
}
