package predictions

import (
	"fmt"
	EntityInfluxPrediction "github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/predictions"
	DaoPredictionTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/predictions/types"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb"
	FormatEnum "github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	FormatTypes "github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	DatahubUtils "github.com/containers-ai/alameda/datahub/pkg/utils"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InternalInfluxModels "github.com/containers-ai/alameda/internal/pkg/database/influxdb/models"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

type NodeRepository struct {
	influxDB *InternalInflux.InfluxClient
}

func NewNodeRepositoryWithConfig(influxDBCfg InternalInflux.Config) *NodeRepository {
	return &NodeRepository{
		influxDB: &InternalInflux.InfluxClient{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

func (r *NodeRepository) CreatePredictions(predictions DaoPredictionTypes.NodePredictionMap) error {
	points := make([]*InfluxClient.Point, 0)

	for _, nodePrediction := range predictions.MetricMap {
		nodeName := nodePrediction.ObjectMeta.Name
		isScheduled := nodePrediction.IsScheduled
		r.appendPoints(FormatEnum.MetricKindRaw, nodeName, isScheduled, nodePrediction.PredictionRaw, &points)
		r.appendPoints(FormatEnum.MetricKindUpperBound, nodeName, isScheduled, nodePrediction.PredictionUpperBound, &points)
		r.appendPoints(FormatEnum.MetricKindLowerBound, nodeName, isScheduled, nodePrediction.PredictionLowerBound, &points)
	}

	// Batch write influxdb data points
	err := r.influxDB.WritePoints(points, InfluxClient.BatchPointsConfig{
		Database: string(RepoInflux.Prediction),
	})
	if err != nil {
		return errors.Wrap(err, "failed to batch write node prediction in influxdb")
	}

	return nil
}

func (r *NodeRepository) ListPredictions(request DaoPredictionTypes.ListNodePredictionsRequest) ([]*DaoPredictionTypes.NodePrediction, error) {
	nodePredictionList := make([]*DaoPredictionTypes.NodePrediction, 0)

	whereClause := r.buildWhereClause(request)

	/*queryCondition := DBCommon.QueryCondition{
		StartTime:      request.QueryCondition.StartTime,
		EndTime:        request.QueryCondition.EndTime,
		StepTime:       request.QueryCondition.StepTime,
		TimestampOrder: request.QueryCondition.TimestampOrder,
		Limit:          request.QueryCondition.Limit,
	}*/

	statement := InternalInflux.Statement{
		QueryCondition: &request.QueryCondition,
		Measurement:    Node,
		WhereClause:    whereClause,
		//GroupByTags: []string{node_entity.Name, node_entity.Metric, node_entity.IsScheduled, node_entity.Kind, node_entity.Granularity},
		//GroupByTags: []string{string(EntityInfluxPrediction.NodeName), string(EntityInfluxPrediction.NodeMetric), string(EntityInfluxPrediction.NodeIsScheduled), string(EntityInfluxPrediction.NodeKind)},
		GroupByTags: []string{string(EntityInfluxPrediction.NodeName), string(EntityInfluxPrediction.NodeIsScheduled)},
	}

	statement.AppendWhereClauseFromTimeCondition()
	statement.AppendWhereClause("AND", string(EntityInfluxPrediction.NodeModelId), "=", request.ModelId)
	statement.AppendWhereClause("AND", string(EntityInfluxPrediction.NodePredictionId), "=", request.PredictionId)
	statement.SetLimitClauseFromQueryCondition()
	statement.SetOrderClauseFromQueryCondition()
	cmd := statement.BuildQueryCmd()

	response, err := r.influxDB.QueryDB(cmd, string(RepoInflux.Prediction))
	if err != nil {
		return make([]*DaoPredictionTypes.NodePrediction, 0), errors.Wrap(err, "failed to list node prediction")
	}

	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			nodePrediction := DaoPredictionTypes.NewNodePrediction()
			nodePrediction.ObjectMeta.Name = group.Tags[string(EntityInfluxPrediction.NodeName)]
			nodePrediction.IsScheduled, _ = strconv.ParseBool(group.Tags[string(EntityInfluxPrediction.NodeIsScheduled)])
			for j := 0; j < group.GetRowNum(); j++ {
				row := group.GetRow(j)
				if row["value"] != "" {
					entity := EntityInfluxPrediction.NewNodeEntityFromMap(group.GetRow(j))
					sample := FormatTypes.PredictionSample{Timestamp: entity.Time, Value: *entity.Value, ModelId: *entity.ModelId, PredictionId: *entity.PredictionId}
					granularity, _ := strconv.ParseInt(*entity.Granularity, 10, 64)
					switch *entity.Kind {
					case FormatEnum.MetricKindRaw:
						nodePrediction.AddRawSample(*entity.Metric, granularity, sample)
					case FormatEnum.MetricKindUpperBound:
						nodePrediction.AddUpperBoundSample(*entity.Metric, granularity, sample)
					case FormatEnum.MetricKindLowerBound:
						nodePrediction.AddLowerBoundSample(*entity.Metric, granularity, sample)
					}
				}
			}
			nodePredictionList = append(nodePredictionList, nodePrediction)
		}
	}

	return nodePredictionList, nil
}

func (r *NodeRepository) appendPoints(kind FormatEnum.MetricKind, nodeName string, isScheduled bool, predictions map[FormatEnum.MetricType]*FormatTypes.PredictionMetricData, points *[]*InfluxClient.Point) error {
	for metricType, metricData := range predictions {
		granularity := metricData.Granularity
		if granularity == 0 {
			granularity = 30
		}

		for _, sample := range metricData.Data {
			// Parse float string to value
			valueInFloat64, err := DatahubUtils.StringToFloat64(sample.Value)
			if err != nil {
				return errors.Wrap(err, "failed to parse string to float64")
			}

			// Pack influx tags
			tags := map[string]string{
				string(EntityInfluxPrediction.NodeName):        nodeName,
				string(EntityInfluxPrediction.NodeIsScheduled): strconv.FormatBool(isScheduled),
				string(EntityInfluxPrediction.NodeMetric):      metricType,
				string(EntityInfluxPrediction.NodeKind):        kind,
				string(EntityInfluxPrediction.NodeGranularity): strconv.FormatInt(granularity, 10),
			}

			// Pack influx fields
			fields := map[string]interface{}{
				string(EntityInfluxPrediction.NodeModelId):      sample.ModelId,
				string(EntityInfluxPrediction.NodePredictionId): sample.PredictionId,
				string(EntityInfluxPrediction.NodeValue):        valueInFloat64,
			}

			// Add to influx point list
			point, err := InfluxClient.NewPoint(string(Node), tags, fields, sample.Timestamp)
			if err != nil {
				return errors.Wrap(err, "failed to instance influxdb data point")
			}
			*points = append(*points, point)
		}
	}

	return nil
}

/*
func (r *NodeRepository) CreateNodePrediction(in *ApiPredictions.CreateNodePredictionsRequest) error {

	points := make([]*InfluxClient.Point, 0)

	for _, nodePrediction := range in.GetNodePredictions() {
		nodeName := nodePrediction.GetName()
		isScheduled := nodePrediction.GetIsScheduled()
		modelId := nodePrediction.GetModelId()
		predictionId := nodePrediction.GetPredictionId()

		r.appendMetricDataToPoints(Metric.NodeMetricKindRaw, nodePrediction.GetPredictedRawData(), &points, nodeName, isScheduled, modelId, predictionId)
		r.appendMetricDataToPoints(Metric.NodeMetricKindUpperbound, nodePrediction.GetPredictedUpperboundData(), &points, nodeName, isScheduled, modelId, predictionId)
		r.appendMetricDataToPoints(Metric.NodeMetricKindLowerbound, nodePrediction.GetPredictedLowerboundData(), &points, nodeName, isScheduled, modelId, predictionId)
	}

	err := r.influxDB.WritePoints(points, InfluxClient.BatchPointsConfig{
		Database: string(RepoInflux.Prediction),
	})
	if err != nil {
		return errors.Wrap(err, "create node prediction failed")
	}

	return nil
}

func (r *NodeRepository) appendMetricDataToPoints(kind FormatEnum.MetricKind, metricDataList []*ApiPredictions.MetricData, points *[]*InfluxClient.Point, nodeName string, isScheduled bool, modelId, predictionId string) error {
	for _, metricData := range metricDataList {
		metricType := ""
		switch metricData.GetMetricType() {
		case ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
			metricType = FormatEnum.MetricTypeCPUUsageSecondsPercentage
		case ApiCommon.MetricType_MEMORY_USAGE_BYTES:
			metricType = FormatEnum.MetricTypeMemoryUsageBytes
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
			valueInFloat64, err := DatahubUtils.StringToFloat64(value)
			if err != nil {
				return errors.Wrap(err, "new influxdb data point failed")
			}

			tags := map[string]string{
				string(EntityInfluxPrediction.NodeName):        nodeName,
				string(EntityInfluxPrediction.NodeIsScheduled): strconv.FormatBool(isScheduled),
				string(EntityInfluxPrediction.NodeMetric):      metricType,
				string(EntityInfluxPrediction.NodeKind):        kind,
				string(EntityInfluxPrediction.NodeGranularity): strconv.FormatInt(granularity, 10),
			}
			fields := map[string]interface{}{
				string(EntityInfluxPrediction.NodeModelId):      modelId,
				string(EntityInfluxPrediction.NodePredictionId): predictionId,
				string(EntityInfluxPrediction.NodeValue):        valueInFloat64,
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

func (r *NodeRepository) ListNodePredictionsByRequest(request DaoPredictionTypes.ListNodePredictionsRequest) ([]*ApiPredictions.NodePrediction, error) {
	whereClause := r.buildWhereClause(request)

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
		GroupByTags: []string{string(EntityInfluxPrediction.NodeName), string(EntityInfluxPrediction.NodeMetric), string(EntityInfluxPrediction.NodeIsScheduled), string(EntityInfluxPrediction.NodeKind)},
	}

	influxdbStatement.AppendWhereClauseFromTimeCondition()
	influxdbStatement.AppendWhereClause(string(EntityInfluxPrediction.NodeModelId), "=", request.ModelId)
	influxdbStatement.AppendWhereClause(string(EntityInfluxPrediction.NodePredictionId), "=", request.PredictionId)
	influxdbStatement.SetLimitClauseFromQueryCondition()
	influxdbStatement.SetOrderClauseFromQueryCondition()
	cmd := influxdbStatement.BuildQueryCmd()

	results, err := r.influxDB.QueryDB(cmd, string(RepoInflux.Prediction))
	if err != nil {
		return []*ApiPredictions.NodePrediction{}, errors.Wrap(err, "list node prediction failed")
	}

	rows := InternalInflux.PackMap(results)
	nodePredictions := r.getNodePredictionsFromInfluxRows(rows)

	return nodePredictions, nil
}

func (r *NodeRepository) getNodePredictionsFromInfluxRows(rows []*InternalInflux.InfluxRow) []*ApiPredictions.NodePrediction {
	nodeMap := map[string]*ApiPredictions.NodePrediction{}
	nodeMetricKindMap := map[string]*ApiCommon.MetricData{}
	nodeMetricKindSampleMap := map[string][]*ApiCommon.Sample{}

	for _, row := range rows {
		name := row.Tags[string(EntityInfluxPrediction.NodeName)]
		metricType := row.Tags[string(EntityInfluxPrediction.NodeMetric)]
		isScheduled := row.Tags[string(EntityInfluxPrediction.NodeIsScheduled)]

		metricValue := ApiCommon.MetricType(ApiCommon.MetricType_value[metricType])
		switch metricType {
		case FormatEnum.MetricTypeCPUUsageSecondsPercentage:
			metricValue = ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE
		case FormatEnum.MetricTypeMemoryUsageBytes:
			metricValue = ApiCommon.MetricType_MEMORY_USAGE_BYTES
		}

		kind := FormatEnum.MetricKindRaw
		if val, ok := row.Tags[string(EntityInfluxPrediction.NodeKind)]; ok {
			if val != "" {
				kind = val
			}
		}

		granularity := int64(30)
		if val, ok := row.Tags[string(EntityInfluxPrediction.NodeGranularity)]; ok {
			if val != "" {
				granularity, _ = strconv.ParseInt(val, 10, 64)
			}
		}

		for _, data := range row.Data {
			modelId := data[string(EntityInfluxPrediction.NodeModelId)]
			predictionId := data[string(EntityInfluxPrediction.NodePredictionId)]

			nodeKey := name + "|" + isScheduled + "|" + modelId + "|" + predictionId
			nodeMap[nodeKey] = &ApiPredictions.NodePrediction{}
			nodeMap[nodeKey].Name = name
			nodeMap[nodeKey].IsScheduled, _ = strconv.ParseBool(isScheduled)
			//nodeMap[nodeKey].ModelId = modelId
			//nodeMap[nodeKey].PredictionId = predictionId

			metricKey := nodeKey + "|" + kind + "|" + metricType
			nodeMetricKindMap[metricKey] = &ApiCommon.MetricData{}
			nodeMetricKindMap[metricKey].MetricType = metricValue
			nodeMetricKindMap[metricKey].Granularity = granularity

			t, _ := time.Parse(time.RFC3339, data[string(EntityInfluxPrediction.NodeTime)])
			value := data[string(EntityInfluxPrediction.NodeValue)]

			googleTimestamp, _ := ptypes.TimestampProto(t)

			tempSample := &ApiCommon.Sample{
				Time:     googleTimestamp,
				NumValue: value,
			}
			nodeMetricKindSampleMap[metricKey] = append(nodeMetricKindSampleMap[metricKey], tempSample)
		}
	}

	for k := range nodeMetricKindSampleMap {
		name := strings.Split(k, "|")[0]
		isScheduled := strings.Split(k, "|")[1]
		modelId := strings.Split(k, "|")[2]
		predictionId := strings.Split(k, "|")[3]
		kind := strings.Split(k, "|")[4]
		metricType := strings.Split(k, "|")[5]

		nodeKey := name + "|" + isScheduled + "|" + modelId + "|" + predictionId
		metricKey := nodeKey + "|" + kind + "|" + metricType

		nodeMetricKindMap[metricKey].Data = nodeMetricKindSampleMap[metricKey]

		//if kind == FormatTypes.NodeMetricKindUpperbound {
		//	nodeMap[nodeKey].PredictedUpperboundData = append(nodeMap[nodeKey].PredictedUpperboundData, nodeMetricKindMap[metricKey])
		//} else if kind == FormatTypes.NodeMetricKindLowerbound {
		//	nodeMap[nodeKey].PredictedLowerboundData = append(nodeMap[nodeKey].PredictedLowerboundData, nodeMetricKindMap[metricKey])
		//} else {
			nodeMap[nodeKey].PredictedRawData = append(nodeMap[nodeKey].PredictedRawData, nodeMetricKindMap[metricKey])
		//}
	}

	nodeList := make([]*ApiPredictions.NodePrediction, 0)
	for k := range nodeMap {
		nodeList = append(nodeList, nodeMap[k])
	}

	return nodeList
}
*/

func (r *NodeRepository) buildWhereClause(request DaoPredictionTypes.ListNodePredictionsRequest) string {
	whereClause := ""
	conditions := ""

	for _, objectMeta := range request.ObjectMeta {
		conditions += fmt.Sprintf(`"%s" = '%s' or `, EntityInfluxPrediction.NodeName, objectMeta.Name)
	}
	conditions = strings.TrimSuffix(conditions, "or ")

	if conditions != "" {
		if request.Granularity == 30 {
			conditions += fmt.Sprintf(` AND ("%s"='' OR "%s"='%d')`, EntityInfluxPrediction.NodeGranularity, EntityInfluxPrediction.NodeGranularity, request.Granularity)
		} else {
			conditions += fmt.Sprintf(` AND "%s"='%d'`, EntityInfluxPrediction.NodeGranularity, request.Granularity)
		}
	} else {
		if request.Granularity == 30 {
			conditions += fmt.Sprintf(`("%s"='' OR "%s"='%d')`, EntityInfluxPrediction.NodeGranularity, EntityInfluxPrediction.NodeGranularity, request.Granularity)
		} else {
			conditions += fmt.Sprintf(`"%s"='%d'`, EntityInfluxPrediction.NodeGranularity, request.Granularity)
		}
	}

	if conditions != "" {
		whereClause = fmt.Sprintf("where %s", conditions)
	}

	return whereClause
}
