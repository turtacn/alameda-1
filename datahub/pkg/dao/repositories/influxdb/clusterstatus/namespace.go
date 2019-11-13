package clusterstatus

import (
	EntityInfluxCluster "github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/clusterstatus"
	DaoClusterTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InternalInfluxModels "github.com/containers-ai/alameda/internal/pkg/database/influxdb/models"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"github.com/pkg/errors"
	"time"
)

type NamespaceRepository struct {
	influxDB *InternalInflux.InfluxClient
}

func NewNamespaceRepositoryWithConfig(influxDBCfg InternalInflux.Config) *NamespaceRepository {
	return &NamespaceRepository{
		influxDB: &InternalInflux.InfluxClient{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

func (c *NamespaceRepository) CreateNamespaces(namespaces []*DaoClusterTypes.Namespace) error {
	points := make([]*InfluxClient.Point, 0)

	for _, namespace := range namespaces {
		// Pack influx tags
		tags := map[string]string{
			string(EntityInfluxCluster.NamespaceName):        namespace.ObjectMeta.Name,
			string(EntityInfluxCluster.NamespaceClusterName): namespace.ObjectMeta.ClusterName,
			string(EntityInfluxCluster.NamespaceUid):         namespace.ObjectMeta.Uid,
		}

		// Pack influx fields
		fields := map[string]interface{}{
			string(EntityInfluxCluster.NamespaceValue): "0",
		}

		// Add to influx point list
		point, err := InfluxClient.NewPoint(string(Namespace), tags, fields, time.Unix(0, 0))
		if err != nil {
			scope.Error(err.Error())
			return errors.Wrap(err, "failed to instance influxdb data point")
		}
		points = append(points, point)
	}

	// Batch write influxdb data points
	err := c.influxDB.WritePoints(points, InfluxClient.BatchPointsConfig{
		Database: string(RepoInflux.ClusterStatus),
	})
	if err != nil {
		scope.Error(err.Error())
		return errors.Wrap(err, "failed to batch write influxdb data points")
	}

	return nil
}

func (c *NamespaceRepository) ListNamespaces(request DaoClusterTypes.ListNamespacesRequest) ([]*DaoClusterTypes.Namespace, error) {
	namespaces := make([]*DaoClusterTypes.Namespace, 0)

	statement := InternalInflux.Statement{
		QueryCondition: &request.QueryCondition,
		Measurement:    Namespace,
		GroupByTags:    []string{string(EntityInfluxCluster.NamespaceClusterName)},
	}

	for _, objectMeta := range request.ObjectMeta {
		condition := statement.GenerateCondition(objectMeta.GenerateKeyList(), objectMeta.GenerateValueList(), "AND")
		statement.AppendWhereClauseDirectly("OR", condition)
	}
	statement.SetOrderClauseFromQueryCondition()
	statement.SetLimitClauseFromQueryCondition()
	cmd := statement.BuildQueryCmd()

	response, err := c.influxDB.QueryDB(cmd, string(RepoInflux.ClusterStatus))
	if err != nil {
		return make([]*DaoClusterTypes.Namespace, 0), errors.Wrap(err, "failed to list namespaces")
	}

	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			for j := 0; j < group.GetRowNum(); j++ {
				row := group.GetRow(j)
				namespace := DaoClusterTypes.NewNamespace()
				namespace.ObjectMeta.Initialize(row)
				namespaces = append(namespaces, namespace)
			}
		}
	}

	return namespaces, nil
}
