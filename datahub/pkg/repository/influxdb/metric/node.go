package metric

import (
	"fmt"
	EntityInfluxMetricNode "github.com/containers-ai/alameda/datahub/pkg/entity/influxdb/metric/node"
	Metric "github.com/containers-ai/alameda/datahub/pkg/metric"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"
	"strings"
	"time"
)

// ContainerRepository is used to operate node measurement of cluster_status database
type NodeRepository struct {
	influxDB *InternalInflux.InfluxClient
}

// NewContainerRepositoryWithConfig New container repository with influxDB configuration
func NewNodeRepositoryWithConfig(influxDBCfg InternalInflux.Config) *NodeRepository {
	return &NodeRepository{
		influxDB: &InternalInflux.InfluxClient{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

// ListContainerPredictionsByRequest list containers' prediction from influxDB
func (r *NodeRepository) ListNodeMetrics(in *datahub_v1alpha1.ListNodeMetricsRequest) ([]*datahub_v1alpha1.NodeMetric, error) {
	nodeMetricList := make([]*datahub_v1alpha1.NodeMetric, 0)

	groupByTime := fmt.Sprintf("%s(%ds)", EntityInfluxMetricNode.NodeTime, in.GetQueryCondition().GetTimeRange().GetStep().GetSeconds())
	SelectedFields := fmt.Sprintf("sum(%s) as %s", EntityInfluxMetricNode.Value, EntityInfluxMetricNode.Value)

	influxdbStatement := InternalInflux.Statement{
		Measurement:    InternalInflux.Measurement(EntityInfluxMetricNode.MetricMeasurementName),
		SelectedFields: []string{SelectedFields},
		GroupByTags:    []string{EntityInfluxMetricNode.Name, EntityInfluxMetricNode.MetricType, groupByTime},
	}

	nodeList := in.GetNodeNames()
	whereNames := ""
	for _, value := range nodeList {
		whereNames += fmt.Sprintf("\"%s\"='%s' OR ", EntityInfluxMetricNode.Name, value)
	}
	whereNames = strings.TrimSuffix(whereNames, "OR ")
	whereNames = "(" + whereNames + ")"

	influxdbStatement.AppendWhereClauseDirectly(whereNames)
	influxdbStatement.AppendWhereClauseWithTime(">=", in.GetQueryCondition().GetTimeRange().GetStartTime().GetSeconds())
	influxdbStatement.AppendWhereClauseWithTime("<=", in.GetQueryCondition().GetTimeRange().GetEndTime().GetSeconds())
	influxdbStatement.SetLimitClauseFromQueryCondition()
	influxdbStatement.SetOrderClauseFromQueryCondition()

	cmd := influxdbStatement.BuildQueryCmd()

	results, err := r.influxDB.QueryDB(cmd, EntityInfluxMetricNode.MetricDatabaseName)
	if err != nil {
		return nodeMetricList, errors.Wrap(err, "list container prediction failed")
	}

	rows := InternalInflux.PackMap(results)

	nodeMetricList = r.getNodeMetricsFromInfluxRows(rows)
	return nodeMetricList, nil
}

func (r *NodeRepository) getNodeMetricsFromInfluxRows(rows []*InternalInflux.InfluxRow) []*datahub_v1alpha1.NodeMetric {
	nodeMap := map[string]*datahub_v1alpha1.NodeMetric{}
	nodeMetricMap := map[string]*datahub_v1alpha1.MetricData{}
	nodeMetricSampleMap := map[string][]*datahub_v1alpha1.Sample{}

	for _, row := range rows {
		nodeName := row.Tags[EntityInfluxMetricNode.Name]
		metricType := row.Tags[EntityInfluxMetricNode.MetricType]

		metricValue := datahub_v1alpha1.MetricType(datahub_v1alpha1.MetricType_value[metricType])
		switch metricType {
		case Metric.TypeContainerCPUUsageSecondsPercentage:
			metricValue = datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE
		case Metric.TypeContainerMemoryUsageBytes:
			metricValue = datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES
		}

		nodeMap[nodeName] = &datahub_v1alpha1.NodeMetric{}
		nodeMap[nodeName].Name = nodeName

		metricKey := nodeName + "|" + metricType
		nodeMetricMap[metricKey] = &datahub_v1alpha1.MetricData{}
		nodeMetricMap[metricKey].MetricType = metricValue

		for _, data := range row.Data {
			t, _ := time.Parse(time.RFC3339, data[EntityInfluxMetricNode.NodeTime])
			value := data[EntityInfluxMetricNode.Value]

			googleTimestamp, _ := ptypes.TimestampProto(t)

			tempSample := &datahub_v1alpha1.Sample{
				Time:     googleTimestamp,
				NumValue: value,
			}
			nodeMetricSampleMap[metricKey] = append(nodeMetricSampleMap[metricKey], tempSample)
		}
	}

	for k := range nodeMetricMap {
		nodeName := strings.Split(k, "|")[0]
		metricType := strings.Split(k, "|")[1]

		metricKey := nodeName + "|" + metricType

		nodeMetricMap[metricKey].Data = nodeMetricSampleMap[metricKey]
		nodeMap[nodeName].MetricData = append(nodeMap[nodeName].MetricData, nodeMetricMap[metricKey])
	}

	nodeList := make([]*datahub_v1alpha1.NodeMetric, 0)
	for k := range nodeMap {
		nodeList = append(nodeList, nodeMap[k])
	}

	return nodeList
}
