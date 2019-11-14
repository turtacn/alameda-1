package cluster

import (
	"context"

	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/grpc"
)

type ClusterRepository struct {
	conn          *grpc.ClientConn
	datahubClient datahub_v1alpha1.DatahubServiceClient

	clusterUID string
}

// NewClusterRepository return ClusterRepository instance
func NewClusterRepository(conn *grpc.ClientConn, clusterUID string) *ClusterRepository {

	datahubClient := datahub_v1alpha1.NewDatahubServiceClient(conn)

	return &ClusterRepository{
		conn:          conn,
		datahubClient: datahubClient,

		clusterUID: clusterUID,
	}
}

// CreateClusters creates clusters to datahub
func (repo *ClusterRepository) CreateClusters(arg interface{}) error {
	clusters := []*datahub_resources.Cluster{}
	if apps, ok := arg.([]*datahub_resources.Cluster); ok {
		clusters = apps
	}

	req := datahub_resources.CreateClustersRequest{
		Clusters: clusters,
	}

	if reqRes, err := repo.datahubClient.CreateClusters(
		context.Background(), &req); err != nil {
		return errors.Errorf("create clusters to datahub failed: %s", err.Error())
	} else if reqRes == nil {
		return errors.Errorf("create clusters to datahub failed: receive nil status")
	} else if reqRes.Code != int32(code.Code_OK) {
		return errors.Errorf(
			"create clusters to datahub failed: receive statusCode: %d, message: %s",
			reqRes.Code, reqRes.Message)
	}
	return nil
}

func (repo *ClusterRepository) ListClusters() (
	[]*datahub_resources.Cluster, error) {
	clusters := []*datahub_resources.Cluster{}
	req := datahub_resources.ListClustersRequest{
		ObjectMeta: []*datahub_resources.ObjectMeta{
			&datahub_resources.ObjectMeta{
				ClusterName: repo.clusterUID,
			},
		},
	}
	if reqRes, err := repo.datahubClient.ListClusters(
		context.Background(), &req); err != nil {
		if reqRes.Status != nil {
			return clusters, errors.Errorf(
				"list clusters from Datahub failed: %s", err.Error())
		}
		return clusters, err
	} else {
		clusters = reqRes.GetClusters()
	}
	return clusters, nil
}

func (repo *ClusterRepository) Close() {
	repo.conn.Close()
}
