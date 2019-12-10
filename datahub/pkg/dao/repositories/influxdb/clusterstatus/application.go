package clusterstatus

import (
	EntityInfluxCluster "github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/clusterstatus"
	DaoClusterTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InternalInfluxModels "github.com/containers-ai/alameda/internal/pkg/database/influxdb/models"
	Log "github.com/containers-ai/alameda/pkg/utils/log"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"github.com/pkg/errors"
)

var (
	scope = Log.RegisterScope("cluster_status_db_measurement", "cluster_status DB measurement", 0)
)

type ApplicationRepository struct {
	influxDB *InternalInflux.InfluxClient
}

func NewApplicationRepository(influxDBCfg InternalInflux.Config) *ApplicationRepository {
	return &ApplicationRepository{
		influxDB: &InternalInflux.InfluxClient{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

func (p *ApplicationRepository) CreateApplications(applications []*DaoClusterTypes.Application) error {
	points := make([]*InfluxClient.Point, 0)

	for _, application := range applications {
		entity := application.BuildEntity()

		// Add to influx point list
		pt, err := entity.BuildInfluxPoint(string(Application))
		if err != nil {
			scope.Error(err.Error())
			return errors.Wrap(err, "failed to instance influxdb data point")
		}
		points = append(points, pt)
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

func (p *ApplicationRepository) ListApplications(request *DaoClusterTypes.ListApplicationsRequest) ([]*DaoClusterTypes.Application, error) {
	applications := make([]*DaoClusterTypes.Application, 0)

	statement := InternalInflux.Statement{
		QueryCondition: &request.QueryCondition,
		Measurement:    Application,
		GroupByTags:    []string{string(EntityInfluxCluster.ApplicationClusterName)},
	}

	// Build influx query command
	for _, applicationObjectMeta := range request.ApplicationObjectMeta {
		keyList := applicationObjectMeta.ObjectMeta.GenerateKeyList()
		valueList := applicationObjectMeta.ObjectMeta.GenerateValueList()

		if applicationObjectMeta.ScalingTool != "" && applicationObjectMeta.ScalingTool != ApiResources.ScalingTool_name[0] {
			keyList = append(keyList, string(EntityInfluxCluster.ApplicationScalingTool))
			valueList = append(valueList, applicationObjectMeta.ScalingTool)
		}

		condition := statement.GenerateCondition(keyList, valueList, "AND")
		statement.AppendWhereClauseDirectly("OR", condition)
	}
	statement.SetOrderClauseFromQueryCondition()
	statement.SetLimitClauseFromQueryCondition()
	cmd := statement.BuildQueryCmd()

	response, err := p.influxDB.QueryDB(cmd, string(RepoInflux.ClusterStatus))
	if err != nil {
		return make([]*DaoClusterTypes.Application, 0), errors.Wrap(err, "failed to list applications")
	}

	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			for j := 0; j < group.GetRowNum(); j++ {
				row := group.GetRow(j)
				application := DaoClusterTypes.NewApplication(EntityInfluxCluster.NewApplicationEntity(row))
				applications = append(applications, application)
			}
		}
	}

	return applications, nil
}

func (p *ApplicationRepository) DeleteApplications(request *DaoClusterTypes.DeleteApplicationsRequest) error {
	statement := InternalInflux.Statement{
		Measurement: Application,
	}

	if !p.influxDB.MeasurementExist(string(RepoInflux.ClusterStatus), string(Application)) {
		return nil
	}

	// Build influx drop command
	for _, applicationObjectMeta := range request.ApplicationObjectMeta {
		keyList := make([]string, 0)
		valueList := make([]string, 0)

		if applicationObjectMeta.ObjectMeta != nil {
			keyList = applicationObjectMeta.ObjectMeta.GenerateKeyList()
			valueList = applicationObjectMeta.ObjectMeta.GenerateValueList()
		}

		if applicationObjectMeta.ScalingTool != "" && applicationObjectMeta.ScalingTool != ApiResources.ScalingTool_name[0] {
			keyList = append(keyList, string(EntityInfluxCluster.ApplicationScalingTool))
			valueList = append(valueList, applicationObjectMeta.ScalingTool)
		}

		condition := statement.GenerateCondition(keyList, valueList, "AND")
		statement.AppendWhereClauseDirectly("OR", condition)
	}
	cmd := statement.BuildDropCmd()

	_, err := p.influxDB.QueryDB(cmd, string(RepoInflux.ClusterStatus))
	if err != nil {
		scope.Error(err.Error())
		return errors.Wrap(err, "failed to delete applications")
	}

	return nil
}
