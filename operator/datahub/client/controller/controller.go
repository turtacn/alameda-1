package controller

import (
	"context"

	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	appsapi_v1 "github.com/openshift/api/apps/v1"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/grpc"
	appsv1 "k8s.io/api/apps/v1"
)

type ControllerRepository struct {
	conn          *grpc.ClientConn
	datahubClient datahub_v1alpha1.DatahubServiceClient
}

// NewControllerRepository return ControllerRepository instance
func NewControllerRepository(conn *grpc.ClientConn) *ControllerRepository {

	datahubClient := datahub_v1alpha1.NewDatahubServiceClient(conn)

	return &ControllerRepository{
		conn:          conn,
		datahubClient: datahubClient,
	}
}

// CreateControllers creates controllers to datahub
func (repo *ControllerRepository) CreateControllers(arg interface{}) error {
	controllersToCreate := []*datahub_resources.Controller{}
	if controllers, ok := arg.([]appsv1.Deployment); ok {
		for _, controller := range controllers {
			controllersToCreate = append(controllersToCreate, &datahub_resources.Controller{
				ObjectMeta: &datahub_resources.ObjectMeta{
					Name:      controller.GetName(),
					Namespace: controller.GetNamespace(),
				},
				Kind: datahub_resources.Kind_DEPLOYMENT,
			})
		}
	}
	if controllers, ok := arg.([]appsv1.StatefulSet); ok {
		for _, controller := range controllers {
			controllersToCreate = append(controllersToCreate, &datahub_resources.Controller{
				ObjectMeta: &datahub_resources.ObjectMeta{
					Name:      controller.GetName(),
					Namespace: controller.GetNamespace(),
				},
				Kind: datahub_resources.Kind_STATEFULSET,
			})
		}

	}
	if controllers, ok := arg.([]appsapi_v1.DeploymentConfig); ok {
		for _, controller := range controllers {
			controllersToCreate = append(controllersToCreate, &datahub_resources.Controller{
				ObjectMeta: &datahub_resources.ObjectMeta{
					Name:      controller.GetName(),
					Namespace: controller.GetNamespace(),
				},
				Kind: datahub_resources.Kind_DEPLOYMENTCONFIG,
			})
		}
	}
	if controllers, ok := arg.([]*datahub_resources.Controller); ok {
		controllersToCreate = controllers
	}

	req := datahub_resources.CreateControllersRequest{
		Controllers: controllersToCreate,
	}

	if reqRes, err := repo.datahubClient.CreateControllers(
		context.Background(), &req); err != nil {
		return errors.Errorf("create controllers to datahub failed: %s", err.Error())
	} else if reqRes == nil {
		return errors.Errorf("create controllers to datahub failed: receive nil status")
	} else if reqRes.Code != int32(code.Code_OK) {
		return errors.Errorf(
			"create controllers to datahub failed: receive statusCode: %d, message: %s",
			reqRes.Code, reqRes.Message)
	}
	return nil
}

func (repo *ControllerRepository) ListControllers() (
	[]*datahub_resources.Controller, error) {
	controllers := []*datahub_resources.Controller{}
	req := datahub_resources.ListControllersRequest{}
	if reqRes, err := repo.datahubClient.ListControllers(
		context.Background(), &req); err != nil {
		if reqRes.Status != nil {
			return controllers, errors.Errorf(
				"list controllers from Datahub failed: %s", err.Error())
		}
		return controllers, err
	} else {
		controllers = reqRes.GetControllers()
	}
	return controllers, nil
}

// DeleteController delete controllers from datahub
func (repo *ControllerRepository) DeleteControllers(arg interface{},
	kindIf interface{}) error {
	objMeta := []*datahub_resources.ObjectMeta{}
	kind := datahub_resources.Kind_POD

	if controllers, ok := arg.([]*appsv1.Deployment); ok {
		kind = datahub_resources.Kind_DEPLOYMENT
		for _, controller := range controllers {
			objMeta = append(objMeta, &datahub_resources.ObjectMeta{
				Name:      controller.GetName(),
				Namespace: controller.GetNamespace(),
			})
		}
	}
	if controllers, ok := arg.([]*appsv1.StatefulSet); ok {
		kind = datahub_resources.Kind_STATEFULSET
		for _, controller := range controllers {
			objMeta = append(objMeta, &datahub_resources.ObjectMeta{
				Name:      controller.GetName(),
				Namespace: controller.GetNamespace(),
			})
		}
	}
	if controllers, ok := arg.([]*appsapi_v1.DeploymentConfig); ok {
		kind = datahub_resources.Kind_DEPLOYMENTCONFIG
		for _, controller := range controllers {
			objMeta = append(objMeta, &datahub_resources.ObjectMeta{
				Name:      controller.GetName(),
				Namespace: controller.GetNamespace(),
			})
		}
	}
	if controllers, ok := arg.([]*datahub_resources.Controller); ok {
		for _, controller := range controllers {
			kind = controller.GetKind()
			objMeta = append(objMeta, &datahub_resources.ObjectMeta{
				Name:      controller.GetObjectMeta().GetName(),
				Namespace: controller.GetObjectMeta().GetNamespace(),
			})
		}
	}
	if meta, ok := arg.([]*datahub_resources.ObjectMeta); ok {
		if theKind, ok := kindIf.(datahub_resources.Kind); ok {
			kind = theKind
			objMeta = meta
		}
	}

	req := datahub_resources.DeleteControllersRequest{
		ObjectMeta: objMeta,
		Kind:       kind,
	}

	if resp, err := repo.datahubClient.DeleteControllers(
		context.Background(), &req); err != nil {
		return errors.Errorf("delete controller from Datahub failed: %s",
			err.Error())
	} else if resp.Code != int32(code.Code_OK) {
		return errors.Errorf(
			"delete controller from Datahub failed: receive code: %d, message: %s",
			resp.Code, resp.Message)
	}
	return nil
}

func (repo *ControllerRepository) Close() {
	repo.conn.Close()
}
