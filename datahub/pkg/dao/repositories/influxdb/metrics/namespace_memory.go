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

type NamespaceMemoryRepository struct {
	influxDB *InternalInflux.InfluxClient
}

func NewNamespaceMemoryRepositoryWithConfig(influxDBCfg InternalInflux.Config) *NamespaceMemoryRepository {
	return &NamespaceMemoryRepository{
		influxDB: &InternalInflux.InfluxClient{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

func (r *NamespaceMemoryRepository) CreateMetrics(ctx context.Context, metrics []DaoMetricTypes.NamespaceMetricSample) error {

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
				string(EntityInfluxMetric.NamespaceName):        metric.ObjectMeta.Name,
				string(EntityInfluxMetric.NamespaceClusterName): metric.ObjectMeta.ClusterName,
				string(EntityInfluxMetric.NamespaceUID):         metric.ObjectMeta.Uid,
			}

			// Pack influx fields
			fields := map[string]interface{}{
				string(EntityInfluxMetric.NamespaceValue): valueInFloat64,
			}

			// Add to influx point list
			point, err := InfluxClient.NewPoint(string(NamespaceMemory), tags, fields, sample.Timestamp)
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

func (r *NamespaceMemoryRepository) GetNamespaceMetricMap(ctx context.Context, request DaoMetricTypes.ListNamespaceMetricsRequest) (DaoMetricTypes.NamespaceMetricMap, error) {

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

func (r *NamespaceMemoryRepository) read(ctx context.Context, request DaoMetricTypes.ListNamespaceMetricsRequest) (DaoMetricTypes.NamespaceMetricMap, error) {

	statement := InternalInflux.Statement{
		Measurement:    NamespaceMemory,
		QueryCondition: &request.QueryCondition,
		GroupByTags: []string{
			string(EntityInfluxMetric.NamespaceName), string(EntityInfluxMetric.NamespaceClusterName),
			string(EntityInfluxMetric.NamespaceUID),
		},
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
		return DaoMetricTypes.NamespaceMetricMap{}, errors.Wrap(err, "query influxdb failed")
	}

	metricMap := DaoMetricTypes.NewNamespaceMetricMap()
	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			m := DaoMetricTypes.NewNamespaceMetric()
			m.ObjectMeta.Name = group.Tags[string(EntityInfluxMetric.NamespaceName)]
			m.ObjectMeta.ClusterName = group.Tags[string(EntityInfluxMetric.NamespaceClusterName)]
			m.ObjectMeta.Uid = group.Tags[string(EntityInfluxMetric.NamespaceUID)]
			for j := 0; j < group.GetRowNum(); j++ {
				row := group.GetRow(j)
				if row["value"] != "" {
					entity := EntityInfluxMetric.NewNamespaceEntityFromMap(group.GetRow(j))
					sample := FormatTypes.Sample{Timestamp: entity.Time, Value: strconv.FormatFloat(*entity.Value, 'f', -1, 64)}
					m.AddSample(FormatEnum.MetricTypeMemoryUsageBytes, sample)
				}
			}
			metricMap.AddNamespaceMetric(m)
		}
	}

	return metricMap, nil
}

func (r *NamespaceMemoryRepository) steps(ctx context.Context, request DaoMetricTypes.ListNamespaceMetricsRequest) (DaoMetricTypes.NamespaceMetricMap, error) {

	groupByTime := fmt.Sprintf("%s(%ds)", EntityInfluxMetric.NamespaceTime, int(request.StepTime.Seconds()))

	statement := InternalInflux.Statement{
		QueryCondition: &request.QueryCondition,
		Measurement:    NamespaceMemory,
		SelectedFields: []string{string(EntityInfluxMetric.NamespaceValue)},
		GroupByTags: []string{
			string(EntityInfluxMetric.NamespaceName), string(EntityInfluxMetric.NamespaceClusterName),
			string(EntityInfluxMetric.NamespaceUID), groupByTime,
		},
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
		return DaoMetricTypes.NamespaceMetricMap{}, errors.Errorf(`not supported aggregate function "%d"`, request.AggregateOverTimeFunction)
	}
	statement.SetFunction(InternalInflux.Select, f, string(EntityInfluxMetric.NamespaceValue))
	cmd := statement.BuildQueryCmd()

	scope.Debugf("Query inlfuxdb: cmd: %s", cmd)
	response, err := r.influxDB.QueryDB(cmd, string(RepoInflux.Metric))
	if err != nil {
		return DaoMetricTypes.NamespaceMetricMap{}, errors.Wrap(err, "query influxdb failed")
	}

	metricMap := DaoMetricTypes.NewNamespaceMetricMap()
	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			m := DaoMetricTypes.NewNamespaceMetric()
			m.ObjectMeta.Name = group.Tags[string(EntityInfluxMetric.NamespaceName)]
			m.ObjectMeta.ClusterName = group.Tags[string(EntityInfluxMetric.NamespaceClusterName)]
			m.ObjectMeta.Uid = group.Tags[string(EntityInfluxMetric.NamespaceUID)]
			for j := 0; j < group.GetRowNum(); j++ {
				row := group.GetRow(j)
				if row["value"] != "" {
					entity := EntityInfluxMetric.NewNamespaceEntityFromMap(group.GetRow(j))
					sample := FormatTypes.Sample{Timestamp: entity.Time, Value: strconv.FormatFloat(*entity.Value, 'f', -1, 64)}
					m.AddSample(FormatEnum.MetricTypeMemoryUsageBytes, sample)
				}
			}
			metricMap.AddNamespaceMetric(m)
		}
	}

	return metricMap, nil
}
