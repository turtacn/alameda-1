package pod

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/containers-ai/alameda/operator/datahub/client"
	logUtil "github.com/containers-ai/alameda/pkg/utils/log"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

var (
	scope = logUtil.RegisterScope("datahub pod repository", "datahub pod repository", 0)
)

// PodRepository creates predicted pod to datahub
type PodRepository struct {
	conn          *grpc.ClientConn
	datahubClient datahub_v1alpha1.DatahubServiceClient

	clusterUID string
}

// NewPodRepository return PodRepository instance
func NewPodRepository(conn *grpc.ClientConn, clusterUID string) *PodRepository {
	datahubClient := datahub_v1alpha1.NewDatahubServiceClient(conn)
	return &PodRepository{
		conn:          conn,
		datahubClient: datahubClient,

		clusterUID: clusterUID,
	}
}

func (repo *PodRepository) CreatePods(ctx context.Context, pods []*datahub_resources.Pod) error {
	req := datahub_resources.CreatePodsRequest{
		Pods: pods,
	}
	resp, err := repo.datahubClient.CreatePods(context.Background(), &req)
	if err != nil {
		return errors.Wrap(err, "create pods to Datahub failed")
	} else if _, err := client.IsResponseStatusOK(resp); err != nil {
		return errors.Wrap(err, "create pods to Datahub failed")
	}
	return nil
}

func (repo *PodRepository) ListAlamedaPods() ([]*datahub_resources.Pod, error) {
	req := datahub_resources.ListPodsRequest{
		ObjectMeta: []*datahub_resources.ObjectMeta{
			&datahub_resources.ObjectMeta{
				ClusterName: repo.clusterUID,
			},
		},
		Kind: datahub_resources.Kind_POD,
	}
	resp, err := repo.datahubClient.ListPods(context.Background(), &req)
	if err != nil {
		return nil, errors.Wrapf(err, "list pods from Datahub failed: %s", err.Error())
	} else if resp == nil {
		return nil, errors.Errorf("list pods from Datahub failed, receive nil response")
	} else if _, err := client.IsResponseStatusOK(resp.Status); err != nil {
		return nil, errors.Wrap(err, "list pods from Datahub failed")
	}
	return resp.Pods, nil
}

func (repo *PodRepository) ListAlamedaPodsByAlamedaScaler(ctx context.Context, namespace, name string) ([]*datahub_resources.Pod, error) {
	req := datahub_resources.ListPodsRequest{
		ObjectMeta: []*datahub_resources.ObjectMeta{
			&datahub_resources.ObjectMeta{
				Namespace:   namespace,
				Name:        name,
				ClusterName: repo.clusterUID,
			},
		},
		Kind: datahub_resources.Kind_ALAMEDASCALER,
	}
	resp, err := repo.datahubClient.ListPods(context.Background(), &req)
	if err != nil {
		return nil, errors.Wrapf(err, "list pods from Datahub failed: %s", err.Error())
	} else if resp == nil {
		return nil, errors.Errorf("list pods from Datahub failed, receive nil response")
	} else if _, err := client.IsResponseStatusOK(resp.Status); err != nil {
		return nil, errors.Wrap(err, "list pods from Datahub failed")
	}
	return resp.Pods, nil
}

// DeletePods delete pods from datahub
func (repo *PodRepository) DeletePods(ctx context.Context, arg interface{}) error {
	objMeta := []*datahub_resources.ObjectMeta{}
	switch t := arg.(type) {
	case []*datahub_resources.Pod:
		for _, pod := range t {
			copyPod := *pod
			objMeta = append(objMeta, copyPod.ObjectMeta)
		}
	case []*datahub_resources.ObjectMeta:
		objMeta = t
	default:
		return errors.Errorf("not supported type(%T)", t)
	}
	req := datahub_resources.DeletePodsRequest{
		ObjectMeta: objMeta,
	}
	if resp, err := repo.datahubClient.DeletePods(context.Background(), &req); err != nil {
		return errors.Wrap(err, "delete pods from Datahub failed")
	} else if _, err := client.IsResponseStatusOK(resp); err != nil {
		return errors.Wrap(err, "delete pods from Datahub failed")
	}
	return nil
}
