package pod

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/containers-ai/alameda/operator/datahub/client"
	datahubutils "github.com/containers-ai/alameda/operator/pkg/utils/datahub"
	logUtil "github.com/containers-ai/alameda/pkg/utils/log"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

var (
	scope = logUtil.RegisterScope("datahub pod repository", "datahub pod repository", 0)
)

// PodRepository creates predicted pod to datahub
type PodRepository struct {
	clusterUID string
}

// NewPodRepository return PodRepository instance
func NewPodRepository(clusterUID string) *PodRepository {
	return &PodRepository{
		clusterUID: clusterUID,
	}
}

func (repo *PodRepository) ListAlamedaPods() ([]*datahub_resources.Pod, error) {
	conn, err := grpc.Dial(datahubutils.GetDatahubAddress(), grpc.WithInsecure())
	defer conn.Close()
	if err != nil {
		return nil, errors.Wrapf(err, "list pods from Datahub failed: %s", err.Error())
	}

	req := datahub_resources.ListPodsRequest{
		ObjectMeta: []*datahub_resources.ObjectMeta{
			&datahub_resources.ObjectMeta{
				ClusterName: repo.clusterUID,
			},
		},
		Kind: datahub_resources.Kind_POD,
	}
	aiServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(conn)
	resp, err := aiServiceClnt.ListPods(context.Background(), &req)
	if err != nil {
		return nil, errors.Wrap(err, "list pods from Datahub failed")
	} else if resp == nil {
		return nil, errors.Errorf("list pods from Datahub failed, receive nil response")
	} else if _, err := client.IsResponseStatusOK(resp.Status); err != nil {
		return nil, errors.Wrap(err, "list pods from Datahub failed")
	}
	return resp.Pods, nil
}

// DeletePods delete pods from datahub
func (repo *PodRepository) DeletePods(arg interface{}) error {
	objMeta := []*datahub_resources.ObjectMeta{}
	if pods, ok := arg.([]*datahub_resources.Pod); ok {
		for _, pod := range pods {
			copyPod := *pod
			objMeta = append(objMeta, copyPod.ObjectMeta)
		}
	}
	if meta, ok := arg.([]*datahub_resources.ObjectMeta); ok {
		objMeta = meta
	}

	conn, err := grpc.Dial(datahubutils.GetDatahubAddress(), grpc.WithInsecure())
	defer conn.Close()
	if err != nil {
		return errors.Wrapf(err, "delete pods from Datahub failed: %s", err.Error())
	}

	req := datahub_resources.DeletePodsRequest{
		ObjectMeta: objMeta,
	}

	aiServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(conn)
	if resp, err := aiServiceClnt.DeletePods(context.Background(), &req); err != nil {
		return errors.Wrap(err, "delete pods from Datahub failed")
	} else if _, err := client.IsResponseStatusOK(resp); err != nil {
		return errors.Wrap(err, "delete pods from Datahub failed")
	}
	return nil
}
