package datahub

import (
	"strings"

	"github.com/containers-ai/alameda/admission-controller/pkg/validator/controller"
	autoscaling_v1alpha1 "github.com/containers-ai/alameda/operator/api/v1alpha1"
	"github.com/containers-ai/alameda/pkg/utils/log"
	datahub_client "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"github.com/pkg/errors"
	context "golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	scope = log.RegisterScope("conrtoller-validator", "Datahub conrtoller validator", 0)
)

type validator struct {
	datahubServiceClient datahub_client.DatahubServiceClient
	sigsK8SClient        client.Client
	clusterName          string
}

// NewControllerValidator returns controller validator which fetch controller information from containers-ai/alameda Datahub
func NewControllerValidator(datahubServiceClient datahub_client.DatahubServiceClient, sigsK8SClient client.Client, clusterName string) controller.Validator {
	return &validator{
		datahubServiceClient: datahubServiceClient,
		sigsK8SClient:        sigsK8SClient,
		clusterName:          clusterName,
	}
}

func (v *validator) IsControllerEnabledExecution(namespace, name, kind string) (bool, error) {

	datahubKind, exist := datahub_resources.Kind_value[strings.ToUpper(kind)]
	if !exist {
		return false, errors.Errorf("no matched datahub kind for kind: %s", kind)
	}

	ctx := buildDefaultRequestContext()
	req := &datahub_resources.ListControllersRequest{
		ObjectMeta: []*datahub_resources.ObjectMeta{
			&datahub_resources.ObjectMeta{
				ClusterName: v.clusterName,
				Namespace:   namespace,
				Name:        name,
			},
		},
		Kind: datahub_resources.Kind(datahubKind),
	}
	scope.Debugf("query ListControllers to datahub, send request: %+v", req)
	resp, err := v.datahubServiceClient.ListControllers(ctx, req)
	scope.Debugf("query ListControllers to datahub, received response: %+v", resp)
	if err != nil {
		return false, errors.Errorf("query ListControllers to datahub failed: errMsg: %s", err.Error())
	}
	if resp == nil || resp.Status == nil {
		return false, errors.New("receive nil status from datahub")
	} else if resp.Status.Code != int32(code.Code_OK) {
		return false, errors.Errorf("status code not 0: receive status code: %d,message: %s", resp.Status.Code, resp.Status.Message)
	}
	if len(resp.Controllers) != 1 {
		return false, errors.Errorf("length of response.Controller is %d expect 1", len(resp.Controllers))
	}
	controller := resp.Controllers[0]

	if controller.AlamedaControllerSpec == nil || controller.AlamedaControllerSpec.AlamedaScaler == nil {
		return false, errors.Errorf("cannot find matched AlamedaScaler to controller (%s/%s ,kind: %s) from datahub", namespace, name, kind)
	}
	alamedaScaler := autoscaling_v1alpha1.AlamedaScaler{}
	err = v.sigsK8SClient.Get(
		ctx,
		client.ObjectKey{
			Namespace: controller.AlamedaControllerSpec.AlamedaScaler.Namespace,
			Name:      controller.AlamedaControllerSpec.AlamedaScaler.Name,
		},
		&alamedaScaler)
	if err != nil {
		return false, errors.Errorf("get AlamedaScaler from k8s failed: %s", err.Error())
	}
	scope.Debugf(`get monitoring AlamedaScaler for controller, controller:{
		namespace: %s,
		name: %s,
		kind: %s
	}, AlamedaScaler:{
		namespace: %s,
		name: %s
	}`, namespace, name, kind, alamedaScaler.Namespace, alamedaScaler.Name)
	return alamedaScaler.IsEnableExecution() && alamedaScaler.IsScalingToolTypeVPA(), nil
}

func buildDefaultRequestContext() context.Context {
	return context.TODO()
}
