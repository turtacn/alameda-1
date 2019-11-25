package controller

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/containers-ai/alameda/operator/datahub/client"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	appsapi_v1 "github.com/openshift/api/apps/v1"

	appsv1 "k8s.io/api/apps/v1"
)

type ControllerRepository struct {
	conn          *grpc.ClientConn
	datahubClient datahub_v1alpha1.DatahubServiceClient

	clusterUID string
}

// NewControllerRepository return ControllerRepository instance
func NewControllerRepository(conn *grpc.ClientConn, clusterUID string) *ControllerRepository {

	datahubClient := datahub_v1alpha1.NewDatahubServiceClient(conn)

	return &ControllerRepository{
		conn:          conn,
		datahubClient: datahubClient,

		clusterUID: clusterUID,
	}
}

// CreateControllers creates controllers to datahub
func (repo *ControllerRepository) CreateControllers(arg interface{}) error {
	controllersToCreate := []*datahub_resources.Controller{}
	if controllers, ok := arg.([]appsv1.Deployment); ok {
		for _, controller := range controllers {
			controllersToCreate = append(controllersToCreate, &datahub_resources.Controller{
				ObjectMeta: &datahub_resources.ObjectMeta{
					Name:        controller.GetName(),
					Namespace:   controller.GetNamespace(),
					ClusterName: repo.clusterUID,
				},
				Kind: datahub_resources.Kind_DEPLOYMENT,
			})
		}
	}
	if controllers, ok := arg.([]appsv1.StatefulSet); ok {
		for _, controller := range controllers {
			controllersToCreate = append(controllersToCreate, &datahub_resources.Controller{
				ObjectMeta: &datahub_resources.ObjectMeta{
					Name:        controller.GetName(),
					Namespace:   controller.GetNamespace(),
					ClusterName: repo.clusterUID,
				},
				Kind: datahub_resources.Kind_STATEFULSET,
			})
		}

	}
	if controllers, ok := arg.([]appsapi_v1.DeploymentConfig); ok {
		for _, controller := range controllers {
			controllersToCreate = append(controllersToCreate, &datahub_resources.Controller{
				ObjectMeta: &datahub_resources.ObjectMeta{
					Name:        controller.GetName(),
					Namespace:   controller.GetNamespace(),
					ClusterName: repo.clusterUID,
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

	if resp, err := repo.datahubClient.CreateControllers(context.Background(), &req); err != nil {
		return errors.Wrap(err, "create controllers to datahub failed")
	} else if _, err := client.IsResponseStatusOK(resp); err != nil {
		return errors.Wrap(err, "create controllers to datahub failed")
	}
	return nil
}

func (repo *ControllerRepository) ListControllers() ([]*datahub_resources.Controller, error) {
	req := datahub_resources.ListControllersRequest{
		ObjectMeta: []*datahub_resources.ObjectMeta{
			&datahub_resources.ObjectMeta{
				ClusterName: repo.clusterUID,
			},
		},
	}

	resp, err := repo.datahubClient.ListControllers(context.Background(), &req)
	if err != nil {
		return nil, errors.Wrap(err, "list controllers from datahub failed")
	} else if resp == nil {
		return nil, errors.Errorf("list controllers from Datahub failed, receive nil response")
	} else if _, err := client.IsResponseStatusOK(resp.Status); err != nil {
		return nil, errors.Wrap(err, "list controllers from Datahub failed")
	}
	return resp.Controllers, nil
}

func (repo *ControllerRepository) ListControllersByApplication(ctx context.Context, namespace, name string) ([]*datahub_resources.Controller, error) {
	req := datahub_resources.ListControllersRequest{
		ObjectMeta: []*datahub_resources.ObjectMeta{
			&datahub_resources.ObjectMeta{
				Namespace:   namespace,
				ClusterName: repo.clusterUID,
			},
		},
	}

	resp, err := repo.datahubClient.ListControllers(ctx, &req)
	if err != nil {
		return nil, errors.Wrap(err, "list controllers from datahub failed")
	} else if resp == nil {
		return nil, errors.Errorf("list controllers from Datahub failed, receive nil response")
	} else if _, err := client.IsResponseStatusOK(resp.Status); err != nil {
		return nil, errors.Wrap(err, "list controllers from Datahub failed")
	}
	controllers := make([]*datahub_resources.Controller, 0, len(resp.Controllers))
	for _, controller := range resp.Controllers {
		copyController := *controller
		if controller != nil && repo.isControllerHasApplicationInfo(*controller, namespace, name) {
			controllers = append(controllers, &copyController)
		}
	}
	return controllers, nil
}

// DeleteControllers delete controllers from datahub
func (repo *ControllerRepository) DeleteControllers(ctx context.Context, arg interface{}, kindIf interface{}) error {
	objMeta := []*datahub_resources.ObjectMeta{}
	kind := datahub_resources.Kind_KIND_UNDEFINED

	switch v := arg.(type) {
	case []*appsv1.Deployment:
		kind = datahub_resources.Kind_DEPLOYMENT
		for _, controller := range v {
			objMeta = append(objMeta, &datahub_resources.ObjectMeta{
				Name:        controller.GetName(),
				Namespace:   controller.GetNamespace(),
				ClusterName: repo.clusterUID,
			})
		}
	case []*appsv1.StatefulSet:
		kind = datahub_resources.Kind_STATEFULSET
		for _, controller := range v {
			objMeta = append(objMeta, &datahub_resources.ObjectMeta{
				Name:        controller.GetName(),
				Namespace:   controller.GetNamespace(),
				ClusterName: repo.clusterUID,
			})
		}
	case []*appsapi_v1.DeploymentConfig:
		kind = datahub_resources.Kind_DEPLOYMENTCONFIG
		for _, controller := range v {
			objMeta = append(objMeta, &datahub_resources.ObjectMeta{
				Name:        controller.GetName(),
				Namespace:   controller.GetNamespace(),
				ClusterName: repo.clusterUID,
			})
		}
	case []*datahub_resources.Controller:
		for _, controller := range v {
			kind = controller.GetKind()
			objMeta = append(objMeta, &datahub_resources.ObjectMeta{
				Name:        controller.GetObjectMeta().GetName(),
				Namespace:   controller.GetObjectMeta().GetNamespace(),
				ClusterName: repo.clusterUID,
			})
		}
	case []*datahub_resources.ObjectMeta:
		if theKind, ok := kindIf.(datahub_resources.Kind); ok {
			kind = theKind
			objMeta = v
		}
	default:
		return errors.Errorf("not supported type(%T)", v)
	}

	req := datahub_resources.DeleteControllersRequest{
		ObjectMeta: objMeta,
		Kind:       kind,
	}

	resp, err := repo.datahubClient.DeleteControllers(ctx, &req)
	if err != nil {
		return errors.Wrap(err, "delete controllers from Datahub failed")
	} else if _, err := client.IsResponseStatusOK(resp); err != nil {
		return errors.Wrap(err, "delete controllers from Datahub failed")
	}
	return nil
}

func (repo *ControllerRepository) Close() {
	repo.conn.Close()
}

func (repo *ControllerRepository) isControllerHasApplicationInfo(controller datahub_resources.Controller, appNamespace, appName string) bool {

	if controller.AlamedaControllerSpec != nil && controller.AlamedaControllerSpec.AlamedaScaler != nil &&
		controller.AlamedaControllerSpec.AlamedaScaler.Namespace == appNamespace && controller.AlamedaControllerSpec.AlamedaScaler.Name == appName {
		return true
	}

	return false
}
