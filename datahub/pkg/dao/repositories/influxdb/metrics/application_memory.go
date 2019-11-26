package metrics

import (
	"context"
	"fmt"
	"strconv"

	EntityInfluxMetric "github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/metrics"
	DaoMetricTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/types"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb"
	FormatEnum "github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	FormatTypes "github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	DatahubUtils "github.com/containers-ai/alameda/datahub/pkg/utils"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InternalInfluxModels "github.com/containers-ai/alameda/internal/pkg/database/influxdb/models"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"github.com/pkg/errors"
)

type ApplicationMemoryRepository struct {
	influxDB *InternalInflux.InfluxClient
}

func NewApplicationMemoryRepositoryWithConfig(influxDBCfg InternalInflux.Config) *ApplicationMemoryRepository {
	return &ApplicationMemoryRepository{
		influxDB: &InternalInflux.InfluxClient{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

func (r *ApplicationMemoryRepository) CreateMetrics(ctx context.Context, metrics []DaoMetricTypes.AppMetricSample) error {

	points := make([]*InfluxClient.Point, 0)
	for _, metric := range metrics {
		if metric.MetricType != FormatEnum.MetricTypeMemoryUsageBytes {
			return errors.Errorf(`not supported metric type "%s"`, metric.MetricType)
		}

		for _, sample := range metric.Metrics {
			// Parse float string to value
			valueInFloat64, err := DatahubUtils.StringToFloat64(sample.Value)
			if err != nil {
				return errors.Wrap(err, "failed to parse string to float64")
			}

			// Pack influx tags
			tags := map[string]string{
				string(EntityInfluxMetric.ApplicationName):        metric.ObjectMeta.Name,
				string(EntityInfluxMetric.ApplicationNamespace):   metric.ObjectMeta.Namespace,
				string(EntityInfluxMetric.ApplicationClusterName): metric.ObjectMeta.ClusterName,
				string(EntityInfluxMetric.ApplicationUID):         metric.ObjectMeta.Uid,
			}

			// Pack influx fields
			fields := map[string]interface{}{
				string(EntityInfluxMetric.ApplicationValue): valueInFloat64,
			}

			// Add to influx point list
			point, err := InfluxClient.NewPoint(string(ApplicationMemory), tags, fields, sample.Timestamp)
			if err != nil {
				return errors.Wrap(err, "failed to instance influxdb data point")
			}
			points = append(points, point)
		}
	}

	// Batch write influxdb data points
	err := r.influxDB.WritePoints(points, InfluxClient.BatchPointsConfig{
		Database: string(RepoInflux.Metric),
	})
	if err != nil {
		return errors.Wrap(err, "failed to batch write influxdb data points")
	}

	return nil
}

func (r *ApplicationMemoryRepository) GetApplicationMetricMap(ctx context.Context, request DaoMetricTypes.ListAppMetricsRequest) (DaoMetricTypes.AppMetricMap, error) {

	steps := 0
	if request.StepTime != nil {
		steps = int(request.StepTime.Seconds())
	}
	if steps == 0 || steps == 30 {
		return r.read(ctx, request)
	} else {
		return r.steps(ctx, request)
	}
}

func (r *ApplicationMemoryRepository) read(ctx context.Context, request DaoMetricTypes.ListAppMetricsRequest) (DaoMetricTypes.AppMetricMap, error) {

	statement := InternalInflux.Statement{
		Measurement:    ApplicationMemory,
		QueryCondition: &request.QueryCondition,
		GroupByTags: []string{string(EntityInfluxMetric.ApplicationName), string(EntityInfluxMetric.ApplicationNamespace), string(EntityInfluxMetric.ApplicationClusterName),
			string(EntityInfluxMetric.ApplicationUID)},
	}

	for _, objectMeta := range request.ObjectMetas {
		condition := statement.GenerateCondition(objectMeta.GenerateKeyList(), objectMeta.GenerateValueList(), "AND")
		statement.AppendWhereClauseDirectly("OR", condition)
	}

	statement.AppendWhereClauseFromTimeCondition()
	statement.SetOrderClauseFromQueryCondition()
	statement.SetLimitClauseFromQueryCondition()
	cmd := statement.BuildQueryCmd()

	scope.Debugf("Query inlfuxdb: cmd: %s", cmd)
	response, err := r.influxDB.QueryDB(cmd, string(RepoInflux.Metric))
	if err != nil {
		return DaoMetricTypes.AppMetricMap{}, errors.Wrap(err, "query influxdb failed")
	}

	metricMap := DaoMetricTypes.NewAppMetricMap()
	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			m := DaoMetricTypes.NewAppMetric()
			m.ObjectMeta.Name = group.Tags[string(EntityInfluxMetric.ApplicationName)]
			m.ObjectMeta.Namespace = group.Tags[string(EntityInfluxMetric.ApplicationNamespace)]
			m.ObjectMeta.ClusterName = group.Tags[string(EntityInfluxMetric.ApplicationClusterName)]
			m.ObjectMeta.Uid = group.Tags[string(EntityInfluxMetric.ApplicationUID)]
			for j := 0; j < group.GetRowNum(); j++ {
				row := group.GetRow(j)
				if row["value"] != "" {
					entity := EntityInfluxMetric.NewApplicationEntityFromMap(group.GetRow(j))
					sample := FormatTypes.Sample{Timestamp: entity.Time, Value: strconv.FormatFloat(*entity.Value, 'f', -1, 64)}
					m.AddSample(FormatEnum.MetricTypeMemoryUsageBytes, sample)
				}
			}
			metricMap.AddAppMetric(m)
		}
	}

	return metricMap, nil
}

func (r *ApplicationMemoryRepository) steps(ctx context.Context, request DaoMetricTypes.ListAppMetricsRequest) (DaoMetricTypes.AppMetricMap, error) {

	groupByTime := fmt.Sprintf("%s(%ds)", EntityInfluxMetric.ApplicationTime, int(request.StepTime.Seconds()))

	statement := InternalInflux.Statement{
		QueryCondition: &request.QueryCondition,
		Measurement:    ApplicationMemory,
		SelectedFields: []string{string(EntityInfluxMetric.ApplicationValue)},
		GroupByTags: []string{string(EntityInfluxMetric.ApplicationName), string(EntityInfluxMetric.ApplicationNamespace), string(EntityInfluxMetric.ApplicationClusterName),
			string(EntityInfluxMetric.ApplicationUID), groupByTime},
	}

	for _, objectMeta := range request.ObjectMetas {
		condition := statement.GenerateCondition(objectMeta.GenerateKeyList(), objectMeta.GenerateValueList(), "AND")
		statement.AppendWhereClauseDirectly("OR", condition)
	}

	statement.AppendWhereClauseFromTimeCondition()
	statement.SetOrderClauseFromQueryCondition()
	statement.SetLimitClauseFromQueryCondition()
	f, exist := aggregateFuncToInfluxDBFunc[request.AggregateOverTimeFunction]
	if !exist {
		return DaoMetricTypes.AppMetricMap{}, errors.Errorf(`not supported aggregate function "%d"`, request.AggregateOverTimeFunction)
	}
	statement.SetFunction(InternalInflux.Select, f, string(EntityInfluxMetric.ApplicationValue))
	cmd := statement.BuildQueryCmd()

	scope.Debugf("Query inlfuxdb: cmd: %s", cmd)
	response, err := r.influxDB.QueryDB(cmd, string(RepoInflux.Metric))
	if err != nil {
		return DaoMetricTypes.AppMetricMap{}, errors.Wrap(err, "query influxdb failed")
	}

	metricMap := DaoMetricTypes.NewAppMetricMap()
	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			m := DaoMetricTypes.NewAppMetric()
			m.ObjectMeta.Name = group.Tags[string(EntityInfluxMetric.ApplicationName)]
			m.ObjectMeta.Namespace = group.Tags[string(EntityInfluxMetric.ApplicationNamespace)]
			m.ObjectMeta.ClusterName = group.Tags[string(EntityInfluxMetric.ApplicationClusterName)]
			m.ObjectMeta.Uid = group.Tags[string(EntityInfluxMetric.ApplicationUID)]
			for j := 0; j < group.GetRowNum(); j++ {
				row := group.GetRow(j)
				if row["value"] != "" {
					entity := EntityInfluxMetric.NewApplicationEntityFromMap(group.GetRow(j))
					sample := FormatTypes.Sample{Timestamp: entity.Time, Value: strconv.FormatFloat(*entity.Value, 'f', -1, 64)}
					m.AddSample(FormatEnum.MetricTypeMemoryUsageBytes, sample)
				}
			}
			metricMap.AddAppMetric(m)
		}
	}

	return metricMap, nil
}
