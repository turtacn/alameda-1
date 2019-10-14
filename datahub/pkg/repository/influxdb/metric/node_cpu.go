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
	"strings"
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
				string(EntityInfluxMetric.NodeName): metricSample.NodeName,
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
		GroupByTags:    []string{string(EntityInfluxMetric.NodeName)},
	}

	whereClause := ""
	for _, value := range request.NodeNames {
		whereClause += fmt.Sprintf("\"%s\"='%s' OR ", EntityInfluxMetric.NodeName, value)
	}
	whereClause = strings.TrimSuffix(whereClause, "OR ")
	whereClause = "(" + whereClause + ")"

	statement.AppendWhereClauseFromTimeCondition()
	statement.AppendWhereClauseDirectly(whereClause)
	statement.SetOrderClauseFromQueryCondition()
	statement.SetLimitClauseFromQueryCondition()
	cmd := statement.BuildQueryCmd()

	response, err := r.influxDB.QueryDB(cmd, string(RepoInflux.Metric))
	if err != nil {
		return make([]*DaoMetricTypes.NodeMetric, 0), errors.Wrap(err, "failed to list node cpu metrics")
	}

	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			nodeMetric := DaoMetricTypes.NewNodeMetric()
			nodeMetric.NodeName = group.Tags[string(EntityInfluxMetric.NodeName)]
			for j := 0; j < group.GetRowNum(); j++ {
				row := group.GetRow(j)
				if row["value"] != "" {
					entity := EntityInfluxMetric.NewNodeEntityFromMap(group.GetRow(j))
					sample := DatahubMetric.Sample{Timestamp: entity.Time, Value: strconv.FormatFloat(*entity.Value, 'f', -1, 64)}
					nodeMetric.AddSample(DatahubMetric.TypeNodeCPUUsageSecondsPercentage, sample)
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
		GroupByTags:    []string{string(EntityInfluxMetric.NodeName), groupByTime},
	}

	whereClause := ""
	for _, value := range request.NodeNames {
		whereClause += fmt.Sprintf("\"%s\"='%s' OR ", EntityInfluxMetric.NodeName, value)
	}
	whereClause = strings.TrimSuffix(whereClause, "OR ")
	whereClause = "(" + whereClause + ")"

	statement.AppendWhereClauseFromTimeCondition()
	statement.AppendWhereClauseDirectly(whereClause)
	statement.SetOrderClauseFromQueryCondition()
	statement.SetLimitClauseFromQueryCondition()
	statement.SetFunction(InternalInflux.Select, "MAX", string(EntityInfluxMetric.NodeValue))
	cmd := statement.BuildQueryCmd()

	response, err := r.influxDB.QueryDB(cmd, string(RepoInflux.Metric))
	if err != nil {
		return make([]*DaoMetricTypes.NodeMetric, 0), errors.Wrap(err, "failed to list node cpu metrics")
	}

	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			nodeMetric := DaoMetricTypes.NewNodeMetric()
			nodeMetric.NodeName = group.Tags[string(EntityInfluxMetric.NodeName)]
			for j := 0; j < group.GetRowNum(); j++ {
				row := group.GetRow(j)
				if row["value"] != "" {
					entity := EntityInfluxMetric.NewNodeEntityFromMap(group.GetRow(j))
					sample := DatahubMetric.Sample{Timestamp: entity.Time, Value: strconv.FormatFloat(*entity.Value, 'f', -1, 64)}
					nodeMetric.AddSample(DatahubMetric.TypeNodeCPUUsageSecondsPercentage, sample)
				}
			}
			nodeMetricList = append(nodeMetricList, nodeMetric)
		}
	}

	return nodeMetricList, nil
}
