package predictions

import (
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

	statement := InternalInflux.Statement{
		QueryCondition: &request.QueryCondition,
		Measurement:    Node,
		GroupByTags:    []string{string(EntityInfluxPrediction.NodeName), string(EntityInfluxPrediction.NodeIsScheduled)},
	}

	granularity := request.Granularity
	if granularity == 0 {
		granularity = 30
	}

	for _, objMeta := range request.ObjectMeta {
		if objMeta.Name == "" && request.ModelId == "" && request.PredictionId == "" {
			statement.WhereClause = ""
			break
		}

		keyList := []string{
			string(EntityInfluxPrediction.NodeName),
			string(EntityInfluxPrediction.NodeModelId),
			string(EntityInfluxPrediction.NodePredictionId),
			string(EntityInfluxPrediction.NodeGranularity),
		}
		valueList := []string{
			objMeta.Name,
			request.ModelId,
			request.PredictionId,
			strconv.FormatInt(granularity, 10),
		}

		tempCondition := statement.GenerateCondition(keyList, valueList, "AND")
		statement.AppendWhereClauseDirectly("OR", tempCondition)
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
