package metrics

import (
	"fmt"
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
	"strconv"
)

type NodeCpuRepository struct {
	influxDB *InternalInflux.InfluxClient
}

func NewNodeCpuRepositoryWithConfig(influxDBCfg InternalInflux.Config) *NodeCpuRepository {
	return &NodeCpuRepository{
		influxDB: &InternalInflux.InfluxClient{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

func (r *NodeCpuRepository) CreateMetrics(metrics []*DaoMetricTypes.NodeMetricSample) error {
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
				string(EntityInfluxMetric.NodeName):        metricSample.ObjectMeta.Name,
				string(EntityInfluxMetric.NodeClusterName): metricSample.ObjectMeta.ClusterName,
				string(EntityInfluxMetric.NodeUID):         metricSample.ObjectMeta.Uid,
			}

			// Pack influx fields
			fields := map[string]interface{}{
				string(EntityInfluxMetric.NodeValue): valueInFloat64,
			}

			// Add to influx point list
			point, err := InfluxClient.NewPoint(string(NodeCpu), tags, fields, metric.Timestamp)
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

func (r *NodeCpuRepository) ListMetrics(request DaoMetricTypes.ListNodeMetricsRequest) ([]*DaoMetricTypes.NodeMetric, error) {
	steps := int(request.StepTime.Seconds())
	if steps == 0 || steps == 30 {
		return r.read(request)
	} else {
		return r.steps(request)
	}
}

func (r *NodeCpuRepository) read(request DaoMetricTypes.ListNodeMetricsRequest) ([]*DaoMetricTypes.NodeMetric, error) {
	nodeMetricList := make([]*DaoMetricTypes.NodeMetric, 0)

	statement := InternalInflux.Statement{
		QueryCondition: &request.QueryCondition,
		Measurement:    NodeCpu,
		GroupByTags: []string{
			string(EntityInfluxMetric.NodeName), string(EntityInfluxMetric.NodeClusterName),
			string(EntityInfluxMetric.NodeUID),
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
		return make([]*DaoMetricTypes.NodeMetric, 0), errors.Wrap(err, "failed to list node cpu metrics")
	}

	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			nodeMetric := DaoMetricTypes.NewNodeMetric()
			nodeMetric.ObjectMeta.Name = group.Tags[string(EntityInfluxMetric.NodeName)]
			nodeMetric.ObjectMeta.ClusterName = group.Tags[string(EntityInfluxMetric.NodeClusterName)]
			nodeMetric.ObjectMeta.Uid = group.Tags[string(EntityInfluxMetric.NodeUID)]
			for j := 0; j < group.GetRowNum(); j++ {
				row := group.GetRow(j)
				if row["value"] != "" {
					entity := EntityInfluxMetric.NewNodeEntityFromMap(group.GetRow(j))
					sample := FormatTypes.Sample{Timestamp: entity.Time, Value: strconv.FormatFloat(*entity.Value, 'f', -1, 64)}
					nodeMetric.AddSample(FormatEnum.MetricTypeCPUUsageSecondsPercentage, sample)
				}
			}
			nodeMetricList = append(nodeMetricList, nodeMetric)
		}
	}

	return nodeMetricList, nil
}

func (r *NodeCpuRepository) steps(request DaoMetricTypes.ListNodeMetricsRequest) ([]*DaoMetricTypes.NodeMetric, error) {
	nodeMetricList := make([]*DaoMetricTypes.NodeMetric, 0)

	groupByTime := fmt.Sprintf("%s(%ds)", EntityInfluxMetric.NodeTime, int(request.StepTime.Seconds()))

	statement := InternalInflux.Statement{
		QueryCondition: &request.QueryCondition,
		Measurement:    NodeCpu,
		SelectedFields: []string{string(EntityInfluxMetric.NodeValue)},
		GroupByTags: []string{
			string(EntityInfluxMetric.NodeName), string(EntityInfluxMetric.NodeClusterName),
			string(EntityInfluxMetric.NodeUID), groupByTime,
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
		return nil, errors.Errorf(`not supported aggregate function "%d"`, request.AggregateOverTimeFunction)
	}
	statement.SetFunction(InternalInflux.Select, f, string(EntityInfluxMetric.NodeValue))
	cmd := statement.BuildQueryCmd()

	scope.Debugf("Query inlfuxdb: cmd: %s", cmd)
	response, err := r.influxDB.QueryDB(cmd, string(RepoInflux.Metric))
	if err != nil {
		return make([]*DaoMetricTypes.NodeMetric, 0), errors.Wrap(err, "failed to list node cpu metrics")
	}

	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			nodeMetric := DaoMetricTypes.NewNodeMetric()
			nodeMetric.ObjectMeta.Name = group.Tags[string(EntityInfluxMetric.NodeName)]
			nodeMetric.ObjectMeta.ClusterName = group.Tags[string(EntityInfluxMetric.NodeClusterName)]
			nodeMetric.ObjectMeta.Uid = group.Tags[string(EntityInfluxMetric.NodeUID)]
			for j := 0; j < group.GetRowNum(); j++ {
				row := group.GetRow(j)
				if row["value"] != "" {
					entity := EntityInfluxMetric.NewNodeEntityFromMap(group.GetRow(j))
					sample := FormatTypes.Sample{Timestamp: entity.Time, Value: strconv.FormatFloat(*entity.Value, 'f', -1, 64)}
					nodeMetric.AddSample(FormatEnum.MetricTypeCPUUsageSecondsPercentage, sample)
				}
			}
			nodeMetricList = append(nodeMetricList, nodeMetric)
		}
	}

	return nodeMetricList, nil
}
