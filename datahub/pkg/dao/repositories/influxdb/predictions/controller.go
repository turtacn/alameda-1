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
		controllerName := prediction.ObjectMeta.Name
		ctlKind := prediction.CtlKind
		r.appendPoints(FormatEnum.MetricKindRaw, controllerName, ctlKind, prediction.PredictionRaw, &points)
		r.appendPoints(FormatEnum.MetricKindUpperBound, controllerName, ctlKind, prediction.PredictionUpperBound, &points)
		r.appendPoints(FormatEnum.MetricKindLowerBound, controllerName, ctlKind, prediction.PredictionLowerBound, &points)
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
			string(EntityInfluxPrediction.ControllerName),
			string(EntityInfluxPrediction.ControllerModelId),
			string(EntityInfluxPrediction.ControllerPredictionId),
			string(EntityInfluxPrediction.ControllerGranularity),
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
		return make([]*DaoPredictionTypes.ControllerPrediction, 0), errors.Wrap(err, "failed to list controller prediction")
	}

	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			controllerPrediction := DaoPredictionTypes.NewControllerPrediction()
			controllerPrediction.ObjectMeta.Name = group.Tags[string(EntityInfluxPrediction.ControllerName)]
			controllerPrediction.CtlKind = group.Tags[string(EntityInfluxPrediction.ControllerCtlKind)]
			for j := 0; j < group.GetRowNum(); j++ {
				row := group.GetRow(j)
				if row["value"] != "" {
					entity := EntityInfluxPrediction.NewControllerEntityFromMap(group.GetRow(j))
					sample := FormatTypes.PredictionSample{Timestamp: entity.Time, Value: *entity.Value, ModelId: *entity.ModelId, PredictionId: *entity.PredictionId}
					granularity, _ := strconv.ParseInt(*entity.Granularity, 10, 64)
					switch *entity.Kind {
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

func (r *ControllerRepository) appendPoints(kind FormatEnum.MetricKind, controllerName string, ctlKind string, predictions map[FormatEnum.MetricType]*FormatTypes.PredictionMetricData, points *[]*InfluxClient.Point) error {
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
				string(EntityInfluxPrediction.ControllerName):        controllerName,
				string(EntityInfluxPrediction.ControllerMetric):      metricType,
				string(EntityInfluxPrediction.ControllerKind):        kind,
				string(EntityInfluxPrediction.ControllerCtlKind):     ctlKind,
				string(EntityInfluxPrediction.ControllerGranularity): strconv.FormatInt(granularity, 10),
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
