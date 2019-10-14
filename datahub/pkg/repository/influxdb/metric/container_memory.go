package metric

import (
	"fmt"
	DaoMetricTypes "github.com/containers-ai/alameda/datahub/pkg/dao/metric/types"
	EntityInfluxMetric "github.com/containers-ai/alameda/datahub/pkg/entity/influxdb/metric"
	DatahubMetric "github.com/containers-ai/alameda/datahub/pkg/metric"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb"
	DatahubUtils "github.com/containers-ai/alameda/datahub/pkg/utils"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InternalInfluxModels "github.com/containers-ai/alameda/internal/pkg/database/influxdb/models"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"github.com/pkg/errors"
	"strconv"
)

type ContainerMemoryRepository struct {
	influxDB *InternalInflux.InfluxClient
}

func NewContainerMemoryRepositoryWithConfig(influxDBCfg InternalInflux.Config) *ContainerMemoryRepository {
	return &ContainerMemoryRepository{
		influxDB: &InternalInflux.InfluxClient{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

func (r *ContainerMemoryRepository) CreateMetrics(metrics []*DaoMetricTypes.ContainerMetricSample) error {
	points := make([]*InfluxClient.Point, 0)

	for _, metricSample := range metrics {
		for _, metric := range metricSample.Metrics {
			// Parse float string to value
			valueInFloat64, err := DatahubUtils.StringToFloat64(metric.Value)
			if err != nil {
				return errors.Wrap(err, "failed to parse string to float64")
			}

			// Pack influx tags
			tags := map[string]string{
				string(EntityInfluxMetric.ContainerPodNamespace): metricSample.Namespace,
				string(EntityInfluxMetric.ContainerPodName):      metricSample.PodName,
				string(EntityInfluxMetric.ContainerName):         metricSample.ContainerName,
				string(EntityInfluxMetric.ContainerRateRange):    strconv.FormatInt(metricSample.RateRange, 10),
			}

			// Pack influx fields
			fields := map[string]interface{}{
				string(EntityInfluxMetric.ContainerValue): valueInFloat64,
			}

			// Add to influx point list
			point, err := InfluxClient.NewPoint(string(ContainerMemory), tags, fields, metric.Timestamp)
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

func (r *ContainerMemoryRepository) ListMetrics(request DaoMetricTypes.ListPodMetricsRequest) ([]*DaoMetricTypes.ContainerMetric, error) {
	steps := int(request.StepTime.Seconds())
	if steps == 0 || steps == 30 {
		return r.read(request)
	} else {
		return r.steps(request)
	}
}

func (r *ContainerMemoryRepository) read(request DaoMetricTypes.ListPodMetricsRequest) ([]*DaoMetricTypes.ContainerMetric, error) {
	containerMetricList := make([]*DaoMetricTypes.ContainerMetric, 0)

	statement := InternalInflux.Statement{
		QueryCondition: &request.QueryCondition,
		Measurement:    ContainerMemory,
		GroupByTags:    []string{string(EntityInfluxMetric.ContainerPodNamespace), string(EntityInfluxMetric.ContainerPodName), string(EntityInfluxMetric.ContainerName), string(EntityInfluxMetric.ContainerRateRange)},
	}

	statement.AppendWhereClauseFromTimeCondition()
	statement.AppendWhereClause(string(EntityInfluxMetric.ContainerPodNamespace), "=", request.Namespace)
	statement.AppendWhereClause(string(EntityInfluxMetric.ContainerPodName), "=", request.PodName)
	statement.AppendWhereClause(string(EntityInfluxMetric.ContainerRateRange), "=", strconv.FormatInt(request.RateRange, 10))
	statement.SetOrderClauseFromQueryCondition()
	statement.SetLimitClauseFromQueryCondition()
	cmd := statement.BuildQueryCmd()

	response, err := r.influxDB.QueryDB(cmd, string(RepoInflux.Metric))
	if err != nil {
		return make([]*DaoMetricTypes.ContainerMetric, 0), errors.Wrap(err, "failed to list container memory metrics")
	}

	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			containerMetric := DaoMetricTypes.NewContainerMetric()
			containerMetric.Namespace = group.Tags[string(EntityInfluxMetric.ContainerPodNamespace)]
			containerMetric.PodName = group.Tags[string(EntityInfluxMetric.ContainerPodName)]
			containerMetric.ContainerName = group.Tags[string(EntityInfluxMetric.ContainerName)]
			containerMetric.RateRange = request.RateRange
			for j := 0; j < group.GetRowNum(); j++ {
				row := group.GetRow(j)
				if row["value"] != "" {
					entity := EntityInfluxMetric.NewContainerEntityFromMap(group.GetRow(j))
					sample := DatahubMetric.Sample{Timestamp: entity.Time, Value: strconv.FormatFloat(*entity.Value, 'f', -1, 64)}
					containerMetric.AddSample(DatahubMetric.TypeContainerMemoryUsageBytes, sample)
				}
			}
			containerMetricList = append(containerMetricList, containerMetric)
		}
	}

	return containerMetricList, nil
}

func (r *ContainerMemoryRepository) steps(request DaoMetricTypes.ListPodMetricsRequest) ([]*DaoMetricTypes.ContainerMetric, error) {
	containerMetricList := make([]*DaoMetricTypes.ContainerMetric, 0)

	groupByTime := fmt.Sprintf("%s(%ds)", EntityInfluxMetric.ContainerTime, int(request.StepTime.Seconds()))

	statement := InternalInflux.Statement{
		QueryCondition: &request.QueryCondition,
		Measurement:    ContainerMemory,
		SelectedFields: []string{string(EntityInfluxMetric.ContainerValue)},
		GroupByTags:    []string{string(EntityInfluxMetric.ContainerPodNamespace), string(EntityInfluxMetric.ContainerPodName), string(EntityInfluxMetric.ContainerName), groupByTime},
	}

	statement.AppendWhereClauseFromTimeCondition()
	statement.AppendWhereClause(string(EntityInfluxMetric.ContainerPodNamespace), "=", request.Namespace)
	statement.AppendWhereClause(string(EntityInfluxMetric.ContainerPodName), "=", request.PodName)
	statement.AppendWhereClause(string(EntityInfluxMetric.ContainerRateRange), "=", strconv.FormatInt(request.RateRange, 10))
	statement.SetOrderClauseFromQueryCondition()
	statement.SetLimitClauseFromQueryCondition()
	statement.SetFunction(InternalInflux.Select, "MAX", string(EntityInfluxMetric.ContainerValue))
	cmd := statement.BuildQueryCmd()

	response, err := r.influxDB.QueryDB(cmd, string(RepoInflux.Metric))
	if err != nil {
		return make([]*DaoMetricTypes.ContainerMetric, 0), errors.Wrap(err, "failed to list container memory metrics")
	}

	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			containerMetric := DaoMetricTypes.NewContainerMetric()
			containerMetric.Namespace = group.Tags[string(EntityInfluxMetric.ContainerPodNamespace)]
			containerMetric.PodName = group.Tags[string(EntityInfluxMetric.ContainerPodName)]
			containerMetric.ContainerName = group.Tags[string(EntityInfluxMetric.ContainerName)]
			containerMetric.RateRange = request.RateRange
			for j := 0; j < group.GetRowNum(); j++ {
				row := group.GetRow(j)
				if row["value"] != "" {
					entity := EntityInfluxMetric.NewContainerEntityFromMap(group.GetRow(j))
					sample := DatahubMetric.Sample{Timestamp: entity.Time, Value: strconv.FormatFloat(*entity.Value, 'f', -1, 64)}
					containerMetric.AddSample(DatahubMetric.TypeContainerMemoryUsageBytes, sample)
				}
			}
			containerMetricList = append(containerMetricList, containerMetric)
		}
	}

	return containerMetricList, nil
}
