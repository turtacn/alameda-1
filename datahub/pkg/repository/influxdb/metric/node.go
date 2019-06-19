package metric

import (
	"fmt"
	node_entity "github.com/containers-ai/alameda/datahub/pkg/entity/influxdb/metric/node"
	"github.com/containers-ai/alameda/datahub/pkg/metric"
	"github.com/containers-ai/alameda/datahub/pkg/repository/influxdb"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"
	"strings"
	"time"
)

// ContainerRepository is used to operate node measurement of cluster_status database
type NodeRepository struct {
	influxDB *influxdb.InfluxDBRepository
}

// NewContainerRepositoryWithConfig New container repository with influxDB configuration
func NewNodeRepositoryWithConfig(influxDBCfg influxdb.Config) *NodeRepository {
	return &NodeRepository{
		influxDB: &influxdb.InfluxDBRepository{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

// ListContainerPredictionsByRequest list containers' prediction from influxDB
func (r *NodeRepository) ListNodeMetrics(in *datahub_v1alpha1.ListNodeMetricsRequest) ([]*datahub_v1alpha1.NodeMetric, error) {
	nodeMetricList := make([]*datahub_v1alpha1.NodeMetric, 0)

	groupByTime := fmt.Sprintf("%s(%ds)", node_entity.NodeTime, in.GetQueryCondition().GetTimeRange().GetStep().GetSeconds())
	SelectedFields := fmt.Sprintf("sum(%s) as %s", node_entity.Value, node_entity.Value)

	whereClause := r.buildInfluxQLWhereClauseFromRequest(in)
	influxdbStatement := influxdb.Statement{
		Measurement:    influxdb.Measurement(node_entity.MetricMeasurementName),
		SelectedFields: []string{SelectedFields},
		WhereClause:    whereClause,
		GroupByTags:    []string{node_entity.Name, node_entity.MetricType, groupByTime},
	}

	queryCondition := influxdb.QueryCondition{
		//StartTime:      in.GetQueryCondition().GetTimeRange().GetStartTime(),
		//EndTime:        request.QueryCondition.EndTime,
		//StepTime:       request.QueryCondition.StepTime,
		TimestampOrder: influxdb.Order(in.GetQueryCondition().GetOrder()),
		Limit:          int(in.GetQueryCondition().GetLimit()),
	}
	//influxdbStatement.AppendTimeConditionIntoWhereClause(queryCondition)
	influxdbStatement.SetLimitClauseFromQueryCondition(queryCondition)
	influxdbStatement.SetOrderClauseFromQueryCondition(queryCondition)
	cmd := influxdbStatement.BuildQueryCmd()

	results, err := r.influxDB.QueryDB(cmd, node_entity.MetricDatabaseName)
	if err != nil {
		return nodeMetricList, errors.Wrap(err, "list container prediction failed")
	}

	rows := influxdb.PackMap(results)

	nodeMetricList = r.getNodeMetricsFromInfluxRows(rows)
	return nodeMetricList, nil
}

func (r *NodeRepository) getNodeMetricsFromInfluxRows(rows []*influxdb.InfluxDBRow) []*datahub_v1alpha1.NodeMetric {
	nodeMap := map[string]*datahub_v1alpha1.NodeMetric{}
	nodeMetricMap := map[string]*datahub_v1alpha1.MetricData{}
	nodeMetricSampleMap := map[string][]*datahub_v1alpha1.Sample{}

	for _, row := range rows {
		nodeName := row.Tags[node_entity.Name]
		metricType := row.Tags[node_entity.MetricType]

		metricValue := datahub_v1alpha1.MetricType(datahub_v1alpha1.MetricType_value[metricType])
		switch metricType {
		case metric.TypeContainerCPUUsageSecondsPercentage:
			metricValue = datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE
		case metric.TypeContainerMemoryUsageBytes:
			metricValue = datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES
		}

		nodeMap[nodeName] = &datahub_v1alpha1.NodeMetric{}
		nodeMap[nodeName].Name = nodeName

		metricKey := nodeName + "|" + metricType
		nodeMetricMap[metricKey] = &datahub_v1alpha1.MetricData{}
		nodeMetricMap[metricKey].MetricType = metricValue

		for _, data := range row.Data {
			t, _ := time.Parse(time.RFC3339, data[node_entity.NodeTime])
			value := data[node_entity.Value]

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

func (r *NodeRepository) buildInfluxQLWhereClauseFromRequest(in *datahub_v1alpha1.ListNodeMetricsRequest) string {
	whereClause := ""

	whereNames := ""
	nodeList := in.GetNodeNames()
	for _, value := range nodeList {
		whereNames += fmt.Sprintf("\"%s\"='%s' OR ", node_entity.Name, value)
	}

	whereNames = strings.TrimSuffix(whereNames, "OR ")
	whereNames = "(" + whereNames + ")"

	startTime := in.GetQueryCondition().GetTimeRange().GetStartTime().GetSeconds()
	endTime := in.GetQueryCondition().GetTimeRange().GetEndTime().GetSeconds()

	r.influxDB.AddWhereConditionDirect(&whereClause, whereNames)
	r.influxDB.AddTimeCondition(&whereClause, ">=", startTime)
	r.influxDB.AddTimeCondition(&whereClause, "<=", endTime)

	return whereClause
}
