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

type NamespaceRepository struct {
	influxDB *InternalInflux.InfluxClient
}

func NewNamespaceRepository(influxDBCfg InternalInflux.Config) *NamespaceRepository {
	return &NamespaceRepository{
		influxDB: &InternalInflux.InfluxClient{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

func (p *NamespaceRepository) CreateNamespaces(namespaces []*DaoClusterTypes.Namespace) error {
	points := make([]*InfluxClient.Point, 0)

	for _, namespace := range namespaces {
		entity := namespace.BuildEntity()

		// Add to influx point list
		point, err := entity.BuildInfluxPoint(string(Namespace))
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

func (p *NamespaceRepository) ListNamespaces(request *DaoClusterTypes.ListNamespacesRequest) ([]*DaoClusterTypes.Namespace, error) {
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

	response, err := p.influxDB.QueryDB(cmd, string(RepoInflux.ClusterStatus))
	if err != nil {
		return make([]*DaoClusterTypes.Namespace, 0), errors.Wrap(err, "failed to list namespaces")
	}

	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			for j := 0; j < group.GetRowNum(); j++ {
				row := group.GetRow(j)
				namespace := DaoClusterTypes.NewNamespace(EntityInfluxCluster.NewNamespaceEntity(row))
				namespaces = append(namespaces, namespace)
			}
		}
	}

	return namespaces, nil
}

func (p *NamespaceRepository) DeleteNamespaces(request *DaoClusterTypes.DeleteNamespacesRequest) error {
	statement := InternalInflux.Statement{
		Measurement: Namespace,
	}

	if !p.influxDB.MeasurementExist(string(RepoInflux.ClusterStatus), string(Namespace)) {
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
		return errors.Wrap(err, "failed to delete controllers")
	}

	return nil
}
