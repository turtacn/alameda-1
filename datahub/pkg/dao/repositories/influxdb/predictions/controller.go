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
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"github.com/pkg/errors"
	"strconv"
)

type ControllerRepository struct {
	influxDB *InternalInflux.InfluxClient
}

func NewControllerRepositoryWithConfig(influxDBCfg InternalInflux.Config) *ControllerRepository {
	return &ControllerRepository{
		influxDB: &InternalInflux.InfluxClient{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

func (r *ControllerRepository) CreatePredictions(predictions DaoPredictionTypes.ControllerPredictionMap) error {
	points := make([]*InfluxClient.Point, 0)

	for _, prediction := range predictions.MetricMap {
		r.appendPoints(FormatEnum.MetricKindRaw, prediction.ObjectMeta, prediction.Kind, prediction.PredictionRaw, &points)
		r.appendPoints(FormatEnum.MetricKindUpperBound, prediction.ObjectMeta, prediction.Kind, prediction.PredictionUpperBound, &points)
		r.appendPoints(FormatEnum.MetricKindLowerBound, prediction.ObjectMeta, prediction.Kind, prediction.PredictionLowerBound, &points)
	}

	// Batch write influxdb data points
	err := r.influxDB.WritePoints(points, InfluxClient.BatchPointsConfig{
		Database: string(RepoInflux.Prediction),
	})
	if err != nil {
		return errors.Wrap(err, "failed to batch write controller prediction in influxdb")
	}

	return nil
}

func (r *ControllerRepository) ListPredictions(request DaoPredictionTypes.ListControllerPredictionsRequest) ([]*DaoPredictionTypes.ControllerPrediction, error) {
	controllerPredictionList := make([]*DaoPredictionTypes.ControllerPrediction, 0)

	statement := InternalInflux.Statement{
		QueryCondition: &request.QueryCondition,
		Measurement:    Controller,
		GroupByTags:    []string{string(EntityInfluxPrediction.ControllerName)},
	}

	for _, objectMeta := range request.ObjectMeta {
		keyList := objectMeta.GenerateKeyList()
		keyList = append(keyList, string(EntityInfluxPrediction.ControllerKind))
		keyList = append(keyList, string(EntityInfluxPrediction.ControllerGranularity))
		keyList = append(keyList, string(EntityInfluxPrediction.ControllerModelId))
		keyList = append(keyList, string(EntityInfluxPrediction.ControllerPredictionId))

		valueList := objectMeta.GenerateValueList()
		valueList = append(valueList, request.Kind)
		valueList = append(valueList, strconv.FormatInt(request.Granularity, 10))
		valueList = append(valueList, request.ModelId)
		valueList = append(valueList, request.PredictionId)

		condition := statement.GenerateCondition(keyList, valueList, "AND")
		statement.AppendWhereClauseDirectly("OR", condition)
	}

	if len(request.ObjectMeta) == 0 {
		if request.Kind != "" && request.Kind != ApiResources.Kind_name[0] {
			statement.AppendWhereClause("AND", string(EntityInfluxPrediction.ControllerKind), "=", request.Kind)
		}
		statement.AppendWhereClause("AND", string(EntityInfluxPrediction.ControllerGranularity), "=", strconv.FormatInt(request.Granularity, 10))
		statement.AppendWhereClause("AND", string(EntityInfluxPrediction.ControllerModelId), "=", request.ModelId)
		statement.AppendWhereClause("AND", string(EntityInfluxPrediction.ControllerPredictionId), "=", request.PredictionId)
	}

	statement.AppendWhereClauseFromTimeCondition()
	statement.SetLimitClauseFromQueryCondition()
	statement.SetOrderClauseFromQueryCondition()
	cmd := statement.BuildQueryCmd()

	response, err := r.influxDB.QueryDB(cmd, string(RepoInflux.Prediction))
	if err != nil {
		return make([]*DaoPredictionTypes.ControllerPrediction, 0), errors.Wrap(err, "failed to list controller prediction")
	}

	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			row := group.GetRow(0)
			controllerPrediction := DaoPredictionTypes.NewControllerPrediction()
			controllerPrediction.ObjectMeta.Initialize(row)
			controllerPrediction.Kind = row[string(EntityInfluxPrediction.ControllerKind)]
			for j := 0; j < group.GetRowNum(); j++ {
				row := group.GetRow(j)
				if row["value"] != "" {
					entity := EntityInfluxPrediction.NewControllerEntity(group.GetRow(j))
					sample := FormatTypes.PredictionSample{Timestamp: entity.Time, Value: *entity.Value, ModelId: *entity.ModelId, PredictionId: *entity.PredictionId}
					granularity, _ := strconv.ParseInt(*entity.Granularity, 10, 64)
					switch *entity.MetricType {
					case FormatEnum.MetricKindRaw:
						controllerPrediction.AddRawSample(*entity.Metric, granularity, sample)
					case FormatEnum.MetricKindUpperBound:
						controllerPrediction.AddUpperBoundSample(*entity.Metric, granularity, sample)
					case FormatEnum.MetricKindLowerBound:
						controllerPrediction.AddLowerBoundSample(*entity.Metric, granularity, sample)
					}
				}
			}
			controllerPredictionList = append(controllerPredictionList, controllerPrediction)
		}
	}

	return controllerPredictionList, nil
}

func (r *ControllerRepository) appendPoints(kind FormatEnum.MetricKind, objectMeta Metadata.ObjectMeta, ctlKind string, predictions map[FormatEnum.MetricType]*FormatTypes.PredictionMetricData, points *[]*InfluxClient.Point) error {
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
				string(EntityInfluxPrediction.ControllerName):        objectMeta.Name,
				string(EntityInfluxPrediction.ControllerNamespace):   objectMeta.Namespace,
				string(EntityInfluxPrediction.ControllerClusterName): objectMeta.ClusterName,
				string(EntityInfluxPrediction.ControllerMetric):      metricType,
				string(EntityInfluxPrediction.ControllerMetricType):  kind,
				string(EntityInfluxPrediction.ControllerGranularity): strconv.FormatInt(granularity, 10),
				string(EntityInfluxPrediction.ControllerKind):        ctlKind,
			}

			// Pack influx fields
			fields := map[string]interface{}{
				string(EntityInfluxPrediction.ControllerModelId):      sample.ModelId,
				string(EntityInfluxPrediction.ControllerPredictionId): sample.PredictionId,
				string(EntityInfluxPrediction.ControllerValue):        valueInFloat64,
			}

			// Add to influx point list
			point, err := InfluxClient.NewPoint(string(Controller), tags, fields, sample.Timestamp)
			if err != nil {
				return errors.Wrap(err, "failed to instance influxdb data point")
			}
			*points = append(*points, point)
		}
	}

	return nil
}
