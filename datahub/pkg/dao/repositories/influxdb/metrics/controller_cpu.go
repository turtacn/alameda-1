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

type ControllerCPURepository struct {
	influxDB *InternalInflux.InfluxClient
}

func NewControllerCPURepositoryWithConfig(influxDBCfg InternalInflux.Config) *ControllerCPURepository {
	return &ControllerCPURepository{
		influxDB: &InternalInflux.InfluxClient{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

func (r *ControllerCPURepository) CreateMetrics(ctx context.Context, metrics []DaoMetricTypes.ControllerMetricSample) error {

	points := make([]*InfluxClient.Point, 0)
	for _, metric := range metrics {
		if metric.MetricType != FormatEnum.MetricTypeCPUUsageSecondsPercentage {
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
				string(EntityInfluxMetric.ControllerName):        metric.ObjectMeta.Name,
				string(EntityInfluxMetric.ControllerNamespace):   metric.ObjectMeta.Namespace,
				string(EntityInfluxMetric.ControllerClusterName): metric.ObjectMeta.ClusterName,
				string(EntityInfluxMetric.ControllerKind):        metric.ObjectMeta.Kind,
				string(EntityInfluxMetric.ControllerUID):         metric.ObjectMeta.Uid,
			}

			// Pack influx fields
			fields := map[string]interface{}{
				string(EntityInfluxMetric.ControllerValue): valueInFloat64,
			}

			// Add to influx point list
			point, err := InfluxClient.NewPoint(string(ControllerCpu), tags, fields, sample.Timestamp)
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

func (r *ControllerCPURepository) GetControllerMetricMap(ctx context.Context, request DaoMetricTypes.ListControllerMetricsRequest) (DaoMetricTypes.ControllerMetricMap, error) {

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

func (r *ControllerCPURepository) read(ctx context.Context, request DaoMetricTypes.ListControllerMetricsRequest) (DaoMetricTypes.ControllerMetricMap, error) {

	statement := InternalInflux.Statement{
		Measurement:    ControllerCpu,
		QueryCondition: &request.QueryCondition,
		GroupByTags: []string{
			string(EntityInfluxMetric.ControllerName), string(EntityInfluxMetric.ControllerNamespace),
			string(EntityInfluxMetric.ControllerClusterName), string(EntityInfluxMetric.ControllerKind),
			string(EntityInfluxMetric.ControllerUID),
		},
	}

	for _, objectMeta := range request.ObjectMetas {
		keyList := objectMeta.GenerateKeyList()
		keyList = append(keyList, string(EntityInfluxMetric.ControllerKind))

		valueList := objectMeta.GenerateValueList()
		valueList = append(valueList, request.Kind)

		condition := statement.GenerateCondition(keyList, valueList, "AND")
		statement.AppendWhereClauseDirectly("OR", condition)
	}
	if len(request.ObjectMetas) == 0 {
		statement.AppendWhereClause("AND", string(EntityInfluxMetric.ControllerKind), "=", request.Kind)
	}

	statement.AppendWhereClauseFromTimeCondition()
	statement.SetOrderClauseFromQueryCondition()
	statement.SetLimitClauseFromQueryCondition()
	cmd := statement.BuildQueryCmd()

	scope.Debugf("Query inlfuxdb: cmd: %s", cmd)
	response, err := r.influxDB.QueryDB(cmd, string(RepoInflux.Metric))
	if err != nil {
		return DaoMetricTypes.ControllerMetricMap{}, errors.Wrap(err, "query influxdb failed")
	}

	metricMap := DaoMetricTypes.NewControllerMetricMap()
	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			m := DaoMetricTypes.NewControllerMetric()
			m.ObjectMeta.Name = group.Tags[string(EntityInfluxMetric.ControllerName)]
			m.ObjectMeta.Namespace = group.Tags[string(EntityInfluxMetric.ControllerNamespace)]
			m.ObjectMeta.ClusterName = group.Tags[string(EntityInfluxMetric.ControllerClusterName)]
			m.ObjectMeta.Kind = group.Tags[string(EntityInfluxMetric.ControllerKind)]
			m.ObjectMeta.Uid = group.Tags[string(EntityInfluxMetric.ControllerUID)]
			for j := 0; j < group.GetRowNum(); j++ {
				row := group.GetRow(j)
				if row["value"] != "" {
					entity := EntityInfluxMetric.NewControllerEntityFromMap(group.GetRow(j))
					sample := FormatTypes.Sample{Timestamp: entity.Time, Value: strconv.FormatFloat(*entity.Value, 'f', -1, 64)}
					m.AddSample(FormatEnum.MetricTypeCPUUsageSecondsPercentage, sample)
				}
			}
			metricMap.AddControllerMetric(m)
		}
	}

	return metricMap, nil
}

func (r *ControllerCPURepository) steps(ctx context.Context, request DaoMetricTypes.ListControllerMetricsRequest) (DaoMetricTypes.ControllerMetricMap, error) {

	groupByTime := fmt.Sprintf("%s(%ds)", EntityInfluxMetric.ControllerTime, int(request.StepTime.Seconds()))

	statement := InternalInflux.Statement{
		QueryCondition: &request.QueryCondition,
		Measurement:    ControllerCpu,
		SelectedFields: []string{string(EntityInfluxMetric.ControllerValue)},
		GroupByTags: []string{
			string(EntityInfluxMetric.ControllerName), string(EntityInfluxMetric.ControllerNamespace),
			string(EntityInfluxMetric.ControllerClusterName), string(EntityInfluxMetric.ControllerKind),
			string(EntityInfluxMetric.ControllerUID), groupByTime,
		},
	}

	for _, objectMeta := range request.ObjectMetas {
		keyList := objectMeta.GenerateKeyList()
		keyList = append(keyList, string(EntityInfluxMetric.ControllerKind))

		valueList := objectMeta.GenerateValueList()
		valueList = append(valueList, request.Kind)

		condition := statement.GenerateCondition(keyList, valueList, "AND")
		statement.AppendWhereClauseDirectly("OR", condition)
	}
	if len(request.ObjectMetas) == 0 {
		statement.AppendWhereClause("AND", string(EntityInfluxMetric.ControllerKind), "=", request.Kind)
	}

	statement.AppendWhereClauseFromTimeCondition()
	statement.SetOrderClauseFromQueryCondition()
	statement.SetLimitClauseFromQueryCondition()
	f, exist := aggregateFuncToInfluxDBFunc[request.AggregateOverTimeFunction]
	if !exist {
		return DaoMetricTypes.ControllerMetricMap{}, errors.Errorf(`not supported aggregate function "%d"`, request.AggregateOverTimeFunction)
	}
	statement.SetFunction(InternalInflux.Select, f, string(EntityInfluxMetric.ControllerValue))
	cmd := statement.BuildQueryCmd()

	scope.Debugf("Query inlfuxdb: cmd: %s", cmd)
	response, err := r.influxDB.QueryDB(cmd, string(RepoInflux.Metric))
	if err != nil {
		return DaoMetricTypes.ControllerMetricMap{}, errors.Wrap(err, "query influxdb failed")
	}

	metricMap := DaoMetricTypes.NewControllerMetricMap()
	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			m := DaoMetricTypes.NewControllerMetric()
			m.ObjectMeta.Name = group.Tags[string(EntityInfluxMetric.ControllerName)]
			m.ObjectMeta.Namespace = group.Tags[string(EntityInfluxMetric.ControllerNamespace)]
			m.ObjectMeta.ClusterName = group.Tags[string(EntityInfluxMetric.ControllerClusterName)]
			m.ObjectMeta.Kind = group.Tags[string(EntityInfluxMetric.ControllerKind)]
			m.ObjectMeta.Uid = group.Tags[string(EntityInfluxMetric.ControllerUID)]
			for j := 0; j < group.GetRowNum(); j++ {
				row := group.GetRow(j)
				if row["value"] != "" {
					entity := EntityInfluxMetric.NewControllerEntityFromMap(group.GetRow(j))
					sample := FormatTypes.Sample{Timestamp: entity.Time, Value: strconv.FormatFloat(*entity.Value, 'f', -1, 64)}
					m.AddSample(FormatEnum.MetricTypeCPUUsageSecondsPercentage, sample)
				}
			}
			metricMap.AddControllerMetric(m)
		}
	}

	return metricMap, nil
}
