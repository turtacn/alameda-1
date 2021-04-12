package prediction

import (
	"fmt"
	DaoPrediction "github.com/containers-ai/alameda/datahub/pkg/dao/prediction"
	EntityInfluxPredictionNode "github.com/containers-ai/alameda/datahub/pkg/entity/influxdb/prediction/node"
	Metric "github.com/containers-ai/alameda/datahub/pkg/metric"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb"
	Utils "github.com/containers-ai/alameda/datahub/pkg/utils"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"time"
	"github.com/containers-ai/alameda/pkg/utils/log"
)


var (
	scope  = log.RegisterScope("prediction_db_measurements", "", 0)
)

type NodeRepository struct {
	influxDB *InternalInflux.InfluxClient
}

func NewNodeRepositoryWithConfig(influxDBCfg InternalInflux.Config) *NodeRepository {
	scope.Infof("influxdb-NewNodeRepositoryWithConfig input %v", influxDBCfg)
	return &NodeRepository{
		influxDB: &InternalInflux.InfluxClient{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

func (r *NodeRepository) CreateNodePrediction(in *datahub_v1alpha1.CreateNodePredictionsRequest) error {

	scope.Infof("influxdb-CreateNodePrediction input %v", in)
	points := make([]*InfluxClient.Point, 0)

	for _, nodePrediction := range in.GetNodePredictions() {
		nodeName := nodePrediction.GetName()
		isScheduled := nodePrediction.GetIsScheduled()

		r.appendMetricDataToPoints(Metric.NodeMetricKindRaw, nodePrediction.GetPredictedRawData(), &points, nodeName, isScheduled)
		r.appendMetricDataToPoints(Metric.NodeMetricKindUpperbound, nodePrediction.GetPredictedUpperboundData(), &points, nodeName, isScheduled)
		r.appendMetricDataToPoints(Metric.NodeMetricKindLowerbound, nodePrediction.GetPredictedLowerboundData(), &points, nodeName, isScheduled)
	}

	err := r.influxDB.WritePoints(points, InfluxClient.BatchPointsConfig{
		Database: string(RepoInflux.Prediction),
	})
	if err != nil {
		scope.Errorf("influxdb-CreateNodePrediction error %v", err)
		return errors.Wrap(err, "create node prediction failed")
	}

	return nil
}

func (r *NodeRepository) appendMetricDataToPoints(kind Metric.ContainerMetricKind, metricDataList []*datahub_v1alpha1.MetricData, points *[]*InfluxClient.Point, nodeName string, isScheduled bool) error {
	for _, metricData := range metricDataList {
		metricType := ""
		switch metricData.GetMetricType() {
		case datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
			metricType = Metric.TypeContainerCPUUsageSecondsPercentage
		case datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES:
			metricType = Metric.TypeContainerMemoryUsageBytes
		}

		if metricType == "" {
			return errors.New("No corresponding metricType")
		}

		granularity := metricData.GetGranularity()
		if granularity == 0 {
			granularity = 30
		}

		for _, data := range metricData.GetData() {
			tempTimeSeconds := data.GetTime().Seconds
			value := data.GetNumValue()
			valueInFloat64, err := Utils.StringToFloat64(value)
			if err != nil {
				return errors.Wrap(err, "new influxdb data point failed")
			}

			tags := map[string]string{
				EntityInfluxPredictionNode.Name:        nodeName,
				EntityInfluxPredictionNode.IsScheduled: strconv.FormatBool(isScheduled),
				EntityInfluxPredictionNode.Metric:      metricType,
				EntityInfluxPredictionNode.Kind:        kind,
				EntityInfluxPredictionNode.Granularity: strconv.FormatInt(granularity, 10),
			}
			fields := map[string]interface{}{
				EntityInfluxPredictionNode.Value: valueInFloat64,
			}
			point, err := InfluxClient.NewPoint(string(Node), tags, fields, time.Unix(tempTimeSeconds, 0))
			if err != nil {
				return errors.Wrap(err, "new influxdb data point failed")
			}
			*points = append(*points, point)
		}
	}

	return nil
}

func (r *NodeRepository) ListNodePredictionsByRequest(request DaoPrediction.ListNodePredictionsRequest) ([]*datahub_v1alpha1.NodePrediction, error) {

	scope.Infof("influxdb-ListNodePredictionsByRequest input %v", request)

	whereClause := r.buildInfluxQLWhereClauseFromRequest(request)

	queryCondition := DBCommon.QueryCondition{
		StartTime:      request.QueryCondition.StartTime,
		EndTime:        request.QueryCondition.EndTime,
		StepTime:       request.QueryCondition.StepTime,
		TimestampOrder: request.QueryCondition.TimestampOrder,
		Limit:          request.QueryCondition.Limit,
	}

	influxdbStatement := InternalInflux.Statement{
		QueryCondition: &queryCondition,
		Measurement:    Node,
		WhereClause:    whereClause,
		//GroupByTags: []string{node_entity.Name, node_entity.Metric, node_entity.IsScheduled, node_entity.Kind, node_entity.Granularity},
		GroupByTags: []string{EntityInfluxPredictionNode.Name, EntityInfluxPredictionNode.Metric, EntityInfluxPredictionNode.IsScheduled, EntityInfluxPredictionNode.Kind},
	}

	influxdbStatement.AppendWhereClauseFromTimeCondition()
	influxdbStatement.SetLimitClauseFromQueryCondition()
	influxdbStatement.SetOrderClauseFromQueryCondition()
	cmd := influxdbStatement.BuildQueryCmd()

	results, err := r.influxDB.QueryDB(cmd, string(RepoInflux.Prediction))
	if err != nil {
		scope.Errorf("influxdb-ListNodePredictionsByRequest error %v", err)
		return []*datahub_v1alpha1.NodePrediction{}, errors.Wrap(err, "list node prediction failed")
	}

	rows := InternalInflux.PackMap(results)
	nodePredictions := r.getNodePredictionsFromInfluxRows(rows)

	scope.Infof("influxdb-ListNodePredictionsByRequest return %d %v",  len(nodePredictions), nodePredictions)
	return nodePredictions, nil
}

func (r *NodeRepository) getNodePredictionsFromInfluxRows(rows []*InternalInflux.InfluxRow) []*datahub_v1alpha1.NodePrediction {
	nodeMap := map[string]*datahub_v1alpha1.NodePrediction{}
	nodeMetricKindMap := map[string]*datahub_v1alpha1.MetricData{}
	nodeMetricKindSampleMap := map[string][]*datahub_v1alpha1.Sample{}

	for _, row := range rows {
		name := row.Tags[EntityInfluxPredictionNode.Name]
		metricType := row.Tags[EntityInfluxPredictionNode.Metric]
		isScheduled := row.Tags[EntityInfluxPredictionNode.IsScheduled]

		metricValue := datahub_v1alpha1.MetricType(datahub_v1alpha1.MetricType_value[metricType])
		switch metricType {
		case Metric.TypeContainerCPUUsageSecondsPercentage:
			metricValue = datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE
		case Metric.TypeContainerMemoryUsageBytes:
			metricValue = datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES
		}

		kind := Metric.NodeMetricKindRaw
		if val, ok := row.Tags[EntityInfluxPredictionNode.Kind]; ok {
			if val != "" {
				kind = val
			}
		}

		granularity := int64(30)
		if val, ok := row.Tags[EntityInfluxPredictionNode.Granularity]; ok {
			if val != "" {
				granularity, _ = strconv.ParseInt(val, 10, 64)
			}
		}

		nodeKey := name + "|" + isScheduled
		nodeMap[nodeKey] = &datahub_v1alpha1.NodePrediction{}
		nodeMap[nodeKey].Name = name
		nodeMap[nodeKey].IsScheduled, _ = strconv.ParseBool(isScheduled)

		metricKey := nodeKey + "|" + kind + "|" + metricType
		nodeMetricKindMap[metricKey] = &datahub_v1alpha1.MetricData{}
		nodeMetricKindMap[metricKey].MetricType = metricValue
		nodeMetricKindMap[metricKey].Granularity = granularity

		for _, data := range row.Data {
			t, _ := time.Parse(time.RFC3339, data[EntityInfluxPredictionNode.Time])
			value := data[EntityInfluxPredictionNode.Value]

			googleTimestamp, _ := ptypes.TimestampProto(t)

			tempSample := &datahub_v1alpha1.Sample{
				Time:     googleTimestamp,
				NumValue: value,
			}
			nodeMetricKindSampleMap[metricKey] = append(nodeMetricKindSampleMap[metricKey], tempSample)
		}
	}

	for k := range nodeMetricKindSampleMap {
		name := strings.Split(k, "|")[0]
		isScheduled := strings.Split(k, "|")[1]
		kind := strings.Split(k, "|")[2]
		metricType := strings.Split(k, "|")[3]

		nodeKey := name + "|" + isScheduled
		metricKey := nodeKey + "|" + kind + "|" + metricType

		nodeMetricKindMap[metricKey].Data = nodeMetricKindSampleMap[metricKey]

		if kind == Metric.NodeMetricKindUpperbound {
			nodeMap[nodeKey].PredictedUpperboundData = append(nodeMap[nodeKey].PredictedUpperboundData, nodeMetricKindMap[metricKey])
		} else if kind == Metric.NodeMetricKindLowerbound {
			nodeMap[nodeKey].PredictedLowerboundData = append(nodeMap[nodeKey].PredictedLowerboundData, nodeMetricKindMap[metricKey])
		} else {
			nodeMap[nodeKey].PredictedRawData = append(nodeMap[nodeKey].PredictedRawData, nodeMetricKindMap[metricKey])
		}
	}

	nodeList := make([]*datahub_v1alpha1.NodePrediction, 0)
	for k := range nodeMap {
		nodeList = append(nodeList, nodeMap[k])
	}

	return nodeList
}

func (r *NodeRepository) buildInfluxQLWhereClauseFromRequest(request DaoPrediction.ListNodePredictionsRequest) string {

	var (
		whereClause string
		conditions  string
	)

	for _, nodeName := range request.NodeNames {
		conditions += fmt.Sprintf(`"%s" = '%s' or `, EntityInfluxPredictionNode.Name, nodeName)
	}

	conditions = strings.TrimSuffix(conditions, "or ")

	if conditions != "" {
		if request.Granularity == 30 {
			conditions += fmt.Sprintf(` AND ("%s"='' OR "%s"='%d')`, EntityInfluxPredictionNode.Granularity, EntityInfluxPredictionNode.Granularity, request.Granularity)
		} else {
			conditions += fmt.Sprintf(` AND "%s"='%d'`, EntityInfluxPredictionNode.Granularity, request.Granularity)
		}
	} else {
		if request.Granularity == 30 {
			conditions += fmt.Sprintf(`("%s"='' OR "%s"='%d')`, EntityInfluxPredictionNode.Granularity, EntityInfluxPredictionNode.Granularity, request.Granularity)
		} else {
			conditions += fmt.Sprintf(`"%s"='%d'`, EntityInfluxPredictionNode.Granularity, request.Granularity)
		}
	}

	if conditions != "" {
		whereClause = fmt.Sprintf("where %s", conditions)
	}

	return whereClause
}
