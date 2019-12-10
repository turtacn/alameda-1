package clusterstatus

import (
	EntityInfluxCluster "github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/clusterstatus"
	DaoClusterTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InternalInfluxModels "github.com/containers-ai/alameda/internal/pkg/database/influxdb/models"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"github.com/pkg/errors"
)

type ClusterRepository struct {
	influxDB *InternalInflux.InfluxClient
}

func NewClusterRepository(influxDBCfg InternalInflux.Config) *ClusterRepository {
	return &ClusterRepository{
		influxDB: &InternalInflux.InfluxClient{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

func (p *ClusterRepository) CreateClusters(clusters []*DaoClusterTypes.Cluster) error {
	points := make([]*InfluxClient.Point, 0)

	for _, cluster := range clusters {
		entity := cluster.BuildEntity()

		// Add to influx point list
		point, err := entity.BuildInfluxPoint(string(Cluster))
		if err != nil {
			scope.Error(err.Error())
			return errors.Wrap(err, "failed to instance influxdb data point")
		}
		points = append(points, point)
	}

	// Batch write influxdb data points
	err := p.influxDB.WritePoints(points, InfluxClient.BatchPointsConfig{
		Database: string(RepoInflux.ClusterStatus),
	})
	if err != nil {
		scope.Error(err.Error())
		return errors.Wrap(err, "failed to batch write influxdb data points")
	}

	return nil
}

func (p *ClusterRepository) ListClusters(request *DaoClusterTypes.ListClustersRequest) ([]*DaoClusterTypes.Cluster, error) {
	clusters := make([]*DaoClusterTypes.Cluster, 0)

	statement := InternalInflux.Statement{
		QueryCondition: &request.QueryCondition,
		Measurement:    Cluster,
	}

	for _, objectMeta := range request.ObjectMeta {
		condition := statement.GenerateCondition(objectMeta.GenerateKeyList(), objectMeta.GenerateValueList(), "AND")
		statement.AppendWhereClauseDirectly("OR", condition)
	}
	statement.SetOrderClauseFromQueryCondition()
	statement.SetLimitClauseFromQueryCondition()
	cmd := statement.BuildQueryCmd()

	response, err := p.influxDB.QueryDB(cmd, string(RepoInflux.ClusterStatus))
	if err != nil {
		return make([]*DaoClusterTypes.Cluster, 0), errors.Wrap(err, "failed to list clusters")
	}

	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			for j := 0; j < group.GetRowNum(); j++ {
				row := group.GetRow(j)
				cluster := DaoClusterTypes.NewCluster(EntityInfluxCluster.NewClusterEntity(row))
				clusters = append(clusters, cluster)
			}
		}
	}

	return clusters, nil
}

func (p *ClusterRepository) DeleteClusters(request *DaoClusterTypes.DeleteClustersRequest) error {
	statement := InternalInflux.Statement{
		Measurement: Cluster,
	}

	if !p.influxDB.MeasurementExist(string(RepoInflux.ClusterStatus), string(Cluster)) {
		return nil
	}

	// Build influx drop command
	for _, objectMeta := range request.ObjectMeta {
		condition := statement.GenerateCondition(objectMeta.GenerateKeyList(), objectMeta.GenerateValueList(), "AND")
		statement.AppendWhereClauseDirectly("OR", condition)
	}
	cmd := statement.BuildDropCmd()

	_, err := p.influxDB.QueryDB(cmd, string(RepoInflux.ClusterStatus))
	if err != nil {
		scope.Error(err.Error())
		return errors.Wrap(err, "failed to delete clusters")
	}

	return nil
}
