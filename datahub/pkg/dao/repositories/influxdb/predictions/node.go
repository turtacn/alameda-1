package predictions

import (
	EntityInfluxPrediction "github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/predictions"
	DaoPredictionTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/predictions/types"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb"
	FormatEnum "github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	FormatTypes "github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	Metadata "github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	DatahubUtils "github.com/containers-ai/alameda/datahub/pkg/utils"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InternalInfluxModels "github.com/containers-ai/alameda/internal/pkg/database/influxdb/models"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"github.com/pkg/errors"
	"strconv"
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

	for _, prediction := range predictions.MetricMap {
		r.appendPoints(FormatEnum.MetricKindRaw, prediction.ObjectMeta, prediction.IsScheduled, prediction.PredictionRaw, &points)
		r.appendPoints(FormatEnum.MetricKindUpperBound, prediction.ObjectMeta, prediction.IsScheduled, prediction.PredictionUpperBound, &points)
		r.appendPoints(FormatEnum.MetricKindLowerBound, prediction.ObjectMeta, prediction.IsScheduled, prediction.PredictionLowerBound, &points)
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

	statement := InternalInflux.Statement{
		QueryCondition: &request.QueryCondition,
		Measurement:    Node,
		GroupByTags:    []string{string(EntityInfluxPrediction.NodeName), string(EntityInfluxPrediction.NodeIsScheduled)},
	}

	for _, objectMeta := range request.ObjectMeta {
		// TODO: Add IsScheduled parameter
		keyList := objectMeta.GenerateKeyList()
		keyList = append(keyList, string(EntityInfluxPrediction.NodeGranularity))
		keyList = append(keyList, string(EntityInfluxPrediction.NodeModelId))
		keyList = append(keyList, string(EntityInfluxPrediction.NodePredictionId))

		valueList := objectMeta.GenerateValueList()
		valueList = append(valueList, strconv.FormatInt(request.Granularity, 10))
		valueList = append(valueList, request.ModelId)
		valueList = append(valueList, request.PredictionId)

		condition := statement.GenerateCondition(keyList, valueList, "AND")
		statement.AppendWhereClauseDirectly("OR", condition)
	}

	if len(request.ObjectMeta) == 0 {
		statement.AppendWhereClause("AND", string(EntityInfluxPrediction.NodeGranularity), "=", strconv.FormatInt(request.Granularity, 10))
		statement.AppendWhereClause("AND", string(EntityInfluxPrediction.NodeModelId), "=", request.ModelId)
		statement.AppendWhereClause("AND", string(EntityInfluxPrediction.NodePredictionId), "=", request.PredictionId)
	}

	statement.AppendWhereClauseFromTimeCondition()
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
			nodePrediction.ObjectMeta.Initialize(group.GetRow(0))
			nodePrediction.IsScheduled, _ = strconv.ParseBool(group.Tags[string(EntityInfluxPrediction.NodeIsScheduled)])
			for j := 0; j < group.GetRowNum(); j++ {
				row := group.GetRow(j)
				if row["value"] != "" {
					entity := EntityInfluxPrediction.NewNodeEntity(group.GetRow(j))
					sample := FormatTypes.PredictionSample{Timestamp: entity.Time, Value: *entity.Value, ModelId: *entity.ModelId, PredictionId: *entity.PredictionId}
					granularity, _ := strconv.ParseInt(*entity.Granularity, 10, 64)
					switch *entity.MetricType {
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

func (r *NodeRepository) appendPoints(kind FormatEnum.MetricKind, objectMeta Metadata.ObjectMeta, isScheduled bool, predictions map[FormatEnum.MetricType]*FormatTypes.PredictionMetricData, points *[]*InfluxClient.Point) error {
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
				string(EntityInfluxPrediction.NodeName):        objectMeta.Name,
				string(EntityInfluxPrediction.NodeClusterName): objectMeta.ClusterName,
				string(EntityInfluxPrediction.NodeMetric):      metricType,
				string(EntityInfluxPrediction.NodeMetricType):  kind,
				string(EntityInfluxPrediction.NodeGranularity): strconv.FormatInt(granularity, 10),
				string(EntityInfluxPrediction.NodeIsScheduled): strconv.FormatBool(isScheduled),
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
