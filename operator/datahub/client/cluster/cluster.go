package cluster

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/containers-ai/alameda/operator/datahub/client"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
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

	if resp, err := repo.datahubClient.CreateClusters(context.Background(), &req); err != nil {
		return errors.Wrap(err, "create clusters to datahub failed")
	} else if _, err := client.IsResponseStatusOK(resp); err != nil {
		return errors.Wrap(err, "create clusters to datahub failed")
	}
	return nil
}

func (repo *ClusterRepository) ListClusters() ([]*datahub_resources.Cluster, error) {
	req := datahub_resources.ListClustersRequest{
		ObjectMeta: []*datahub_resources.ObjectMeta{
			&datahub_resources.ObjectMeta{
				ClusterName: repo.clusterUID,
			},
		},
	}

	resp, err := repo.datahubClient.ListClusters(context.Background(), &req)
	if err != nil {
		return nil, errors.Wrap(err, "list clusters from Datahub failed")
	} else if resp == nil {
		return nil, errors.Errorf("list clusters from Datahub failed, receive nil response")
	} else if _, err := client.IsResponseStatusOK(resp.Status); err != nil {
		return nil, errors.Wrap(err, "list clusters from Datahub failed")
	}
	return resp.Clusters, nil
}

func (repo *ClusterRepository) Close() {
	repo.conn.Close()
}
