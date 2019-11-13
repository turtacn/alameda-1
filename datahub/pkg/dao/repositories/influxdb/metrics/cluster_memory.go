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

type ClusterMemoryRepository struct {
	influxDB *InternalInflux.InfluxClient
}

func NewClusterMemoryRepositoryWithConfig(influxDBCfg InternalInflux.Config) *ClusterMemoryRepository {
	return &ClusterMemoryRepository{
		influxDB: &InternalInflux.InfluxClient{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

func (r *ClusterMemoryRepository) CreateMetrics(ctx context.Context, metrics []DaoMetricTypes.ClusterMetricSample) error {

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
				string(EntityInfluxMetric.ClusterName): metric.ObjectMeta.Name,
				string(EntityInfluxMetric.ClusterUID):  metric.ObjectMeta.Uid,
			}

			// Pack influx fields
			fields := map[string]interface{}{
				string(EntityInfluxMetric.ClusterValue): valueInFloat64,
			}

			// Add to influx point list
			point, err := InfluxClient.NewPoint(string(ClusterMemory), tags, fields, sample.Timestamp)
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

func (r *ClusterMemoryRepository) GetClusterMetricMap(ctx context.Context, request DaoMetricTypes.ListClusterMetricsRequest) (DaoMetricTypes.ClusterMetricMap, error) {

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

func (r *ClusterMemoryRepository) read(ctx context.Context, request DaoMetricTypes.ListClusterMetricsRequest) (DaoMetricTypes.ClusterMetricMap, error) {

	statement := InternalInflux.Statement{
		Measurement:    ClusterMemory,
		QueryCondition: &request.QueryCondition,
		GroupByTags:    []string{string(EntityInfluxMetric.ClusterName), string(EntityInfluxMetric.ClusterUID)},
	}

	for _, objectMeta := range request.ObjectMetas {
		condition := statement.GenerateCondition(objectMeta.GenerateKeyList(), objectMeta.GenerateValueList(), "AND")
		statement.AppendWhereClauseDirectly("OR", condition)
	}

	statement.AppendWhereClauseFromTimeCondition()
	statement.SetOrderClauseFromQueryCondition()
	statement.SetLimitClauseFromQueryCondition()

	cmd := statement.BuildQueryCmd()
	response, err := r.influxDB.QueryDB(cmd, string(RepoInflux.Metric))
	if err != nil {
		return DaoMetricTypes.ClusterMetricMap{}, errors.Wrap(err, "query influxdb failed")
	}

	metricMap := DaoMetricTypes.NewClusterMetricMap()
	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			m := DaoMetricTypes.NewClusterMetric()
			m.ObjectMeta.Name = group.Tags[string(EntityInfluxMetric.ClusterName)]
			m.ObjectMeta.Uid = group.Tags[string(EntityInfluxMetric.ClusterUID)]
			for j := 0; j < group.GetRowNum(); j++ {
				row := group.GetRow(j)
				if row["value"] != "" {
					entity := EntityInfluxMetric.NewClusterEntityFromMap(group.GetRow(j))
					sample := FormatTypes.Sample{Timestamp: entity.Time, Value: strconv.FormatFloat(*entity.Value, 'f', -1, 64)}
					m.AddSample(FormatEnum.MetricTypeMemoryUsageBytes, sample)
				}
			}
			metricMap.AddClusterMetric(m)
		}
	}

	return metricMap, nil
}

func (r *ClusterMemoryRepository) steps(ctx context.Context, request DaoMetricTypes.ListClusterMetricsRequest) (DaoMetricTypes.ClusterMetricMap, error) {

	groupByTime := fmt.Sprintf("%s(%ds)", EntityInfluxMetric.ClusterTime, int(request.StepTime.Seconds()))

	statement := InternalInflux.Statement{
		QueryCondition: &request.QueryCondition,
		Measurement:    ClusterMemory,
		SelectedFields: []string{string(EntityInfluxMetric.ClusterValue)},
		GroupByTags:    []string{string(EntityInfluxMetric.ClusterName), string(EntityInfluxMetric.ClusterUID), groupByTime},
	}

	for _, objectMeta := range request.ObjectMetas {
		condition := statement.GenerateCondition(objectMeta.GenerateKeyList(), objectMeta.GenerateValueList(), "AND")
		statement.AppendWhereClauseDirectly("OR", condition)
	}

	statement.AppendWhereClauseFromTimeCondition()
	statement.SetOrderClauseFromQueryCondition()
	statement.SetLimitClauseFromQueryCondition()
	statement.SetFunction(InternalInflux.Select, "MAX", string(EntityInfluxMetric.ClusterValue))
	cmd := statement.BuildQueryCmd()

	response, err := r.influxDB.QueryDB(cmd, string(RepoInflux.Metric))
	if err != nil {
		return DaoMetricTypes.ClusterMetricMap{}, errors.Wrap(err, "query influxdb failed")
	}

	metricMap := DaoMetricTypes.NewClusterMetricMap()
	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			m := DaoMetricTypes.NewClusterMetric()
			m.ObjectMeta.Name = group.Tags[string(EntityInfluxMetric.ClusterName)]
			m.ObjectMeta.Uid = group.Tags[string(EntityInfluxMetric.ClusterUID)]
			for j := 0; j < group.GetRowNum(); j++ {
				row := group.GetRow(j)
				if row["value"] != "" {
					entity := EntityInfluxMetric.NewClusterEntityFromMap(group.GetRow(j))
					sample := FormatTypes.Sample{Timestamp: entity.Time, Value: strconv.FormatFloat(*entity.Value, 'f', -1, 64)}
					m.AddSample(FormatEnum.MetricTypeMemoryUsageBytes, sample)
				}
			}
			metricMap.AddClusterMetric(m)
		}
	}

	return metricMap, nil
}
