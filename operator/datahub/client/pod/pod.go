package pod

import (
	"context"

	datahubutils "github.com/containers-ai/alameda/operator/pkg/utils/datahub"
	logUtil "github.com/containers-ai/alameda/pkg/utils/log"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/grpc"
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
	alamedaPods := []*datahub_resources.Pod{}
	conn, err := grpc.Dial(datahubutils.GetDatahubAddress(), grpc.WithInsecure())
	defer conn.Close()
	if err != nil {
		return nil, errors.Wrapf(err, "list Alameda pods from Datahub failed: %s", err.Error())
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
	if resp, err := aiServiceClnt.ListPods(context.Background(), &req); err != nil {
		return alamedaPods, errors.Wrapf(err, "list Alameda pods from Datahub failed: %s", err.Error())
	} else if resp.Status != nil && resp.Status.Code != int32(code.Code_OK) {
		return alamedaPods, errors.Errorf("list Alameda pods from Datahub failed: receive code: %d, message: %s", resp.Status.Code, resp.Status.Message)
	} else {
		alamedaPods = resp.GetPods()
	}
	return alamedaPods, nil
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
		return errors.Wrapf(err, "delete pods from Datahub failed: %s", err.Error())
	} else if resp.Code != int32(code.Code_OK) {
		return errors.Errorf("delete pods from Datahub failed: receive code: %d, message: %s", resp.Code, resp.Message)
	}
	return nil
}
