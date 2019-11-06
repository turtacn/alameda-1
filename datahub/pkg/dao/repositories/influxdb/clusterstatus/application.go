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

type ApplicationRepository struct {
	influxDB *InternalInflux.InfluxClient
}

func NewApplicationRepositoryWithConfig(influxDBCfg InternalInflux.Config) *ApplicationRepository {
	return &ApplicationRepository{
		influxDB: &InternalInflux.InfluxClient{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

func (c *ApplicationRepository) CreateApplications(applications []*DaoClusterTypes.Application) error {
	points := make([]*InfluxClient.Point, 0)

	for _, application := range applications {
		// Pack influx tags
		tags := map[string]string{
			string(EntityInfluxCluster.ApplicationName):        application.ObjectMeta.Name,
			string(EntityInfluxCluster.ApplicationNamespace):   application.ObjectMeta.Namespace,
			string(EntityInfluxCluster.ApplicationClusterName): application.ObjectMeta.ClusterName,
			string(EntityInfluxCluster.ApplicationUid):         application.ObjectMeta.Uid,
		}

		// Pack influx fields
		fields := map[string]interface{}{
			string(EntityInfluxCluster.ApplicationValue): "0",
		}

		// Add to influx point list
		point, err := InfluxClient.NewPoint(string(Application), tags, fields, time.Unix(0, 0))
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

func (c *ApplicationRepository) ListApplications(request DaoClusterTypes.ListApplicationsRequest) ([]*DaoClusterTypes.Application, error) {
	applications := make([]*DaoClusterTypes.Application, 0)

	statement := InternalInflux.Statement{
		QueryCondition: &request.QueryCondition,
		Measurement:    Application,
		GroupByTags:    []string{string(EntityInfluxCluster.ApplicationClusterName)},
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
		return make([]*DaoClusterTypes.Application, 0), errors.Wrap(err, "failed to list applications")
	}

	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			for j := 0; j < group.GetRowNum(); j++ {
				row := group.GetRow(j)
				application := DaoClusterTypes.NewApplication()
				application.ObjectMeta.Initialize(row)
				applications = append(applications, application)
			}
		}
	}

	return applications, nil
}
