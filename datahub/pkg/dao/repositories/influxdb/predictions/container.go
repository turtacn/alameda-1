package predictions

import (
	//"fmt"
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

// ContainerRepository Repository to access containers' prediction data
type ContainerRepository struct {
	influxDB *InternalInflux.InfluxClient
}

// NewContainerRepositoryWithConfig New container repository with influxDB configuration
func NewContainerRepositoryWithConfig(influxDBCfg InternalInflux.Config) *ContainerRepository {
	return &ContainerRepository{
		influxDB: &InternalInflux.InfluxClient{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

func (r *ContainerRepository) CreatePredictions(predictions []*DaoPredictionTypes.ContainerPredictionSample) error {
	points := make([]*InfluxClient.Point, 0)

	for _, predictionSample := range predictions {
		granularity := int64(30)
		if predictionSample.Predictions.Granularity != 0 {
			granularity = predictionSample.Predictions.Granularity
		}

		for _, sample := range predictionSample.Predictions.Data {
			// Parse float string to value
			valueInFloat64, err := DatahubUtils.StringToFloat64(sample.Value)
			if err != nil {
				return errors.Wrap(err, "new influxdb data point failed")
			}

			// Pack influx tags
			tags := map[string]string{
				string(EntityInfluxPrediction.ContainerNamespace):   predictionSample.Namespace,
				string(EntityInfluxPrediction.ContainerPodName):     predictionSample.PodName,
				string(EntityInfluxPrediction.ContainerName):        predictionSample.ContainerName,
				string(EntityInfluxPrediction.ContainerMetric):      predictionSample.MetricType,
				string(EntityInfluxPrediction.ContainerKind):        predictionSample.MetricKind,
				string(EntityInfluxPrediction.ContainerGranularity): strconv.FormatInt(granularity, 10),
			}

			// Pack influx fields
			fields := map[string]interface{}{
				string(EntityInfluxPrediction.ContainerModelId):      sample.ModelId,
				string(EntityInfluxPrediction.ContainerPredictionId): sample.PredictionId,
				string(EntityInfluxPrediction.ContainerValue):        valueInFloat64,
			}

			// Add to influx point list
			point, err := InfluxClient.NewPoint(string(Container), tags, fields, sample.Timestamp)
			if err != nil {
				return errors.Wrap(err, "failed to instance influxdb data point")
			}
			points = append(points, point)
		}
	}

	// Batch write influxdb data points
	err := r.influxDB.WritePoints(points, InfluxClient.BatchPointsConfig{
		Database: string(RepoInflux.Prediction),
	})
	if err != nil {
		return errors.Wrap(err, "failed to batch write influxdb data points")
	}

	return nil
}

// ListContainerPredictionsByRequest list containers' prediction from influxDB
func (r *ContainerRepository) ListPredictions(request DaoPredictionTypes.ListPodPredictionsRequest) ([]*DaoPredictionTypes.ContainerPrediction, error) {
	containerPredictionList := make([]*DaoPredictionTypes.ContainerPrediction, 0)

	statement := InternalInflux.Statement{
		QueryCondition: &request.QueryCondition,
		Measurement:    Container,
		GroupByTags:    []string{string(EntityInfluxPrediction.ContainerNamespace), string(EntityInfluxPrediction.ContainerPodName), string(EntityInfluxPrediction.ContainerName)},
	}

	granularity := request.Granularity
	if granularity == 0 {
		granularity = 30
	}

	for _, objMeta := range request.ObjectMeta {
		if objMeta.Namespace == "" && objMeta.Name == "" && request.ModelId == "" && request.PredictionId == "" {
			statement.WhereClause = ""
			break
		}

		keyList := []string{
			string(EntityInfluxPrediction.ContainerNamespace),
			string(EntityInfluxPrediction.ContainerPodName),
			string(EntityInfluxPrediction.ContainerModelId),
			string(EntityInfluxPrediction.ContainerPredictionId),
			string(EntityInfluxPrediction.ContainerGranularity),
		}
		valueList := []string{
			objMeta.Namespace,
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
		return make([]*DaoPredictionTypes.ContainerPrediction, 0), errors.Wrap(err, "failed to list container prediction")
	}

	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			containerPrediction := DaoPredictionTypes.NewContainerPrediction()
			containerPrediction.Namespace = group.Tags[string(EntityInfluxPrediction.ContainerNamespace)]
			containerPrediction.PodName = group.Tags[string(EntityInfluxPrediction.ContainerPodName)]
			containerPrediction.ContainerName = group.Tags[string(EntityInfluxPrediction.ContainerName)]
			for j := 0; j < group.GetRowNum(); j++ {
				row := group.GetRow(j)
				if row["value"] != "" {
					entity := EntityInfluxPrediction.NewContainerEntityFromMap(group.GetRow(j))
					sample := FormatTypes.PredictionSample{Timestamp: entity.Time, Value: *entity.Value, ModelId: *entity.ModelId, PredictionId: *entity.PredictionId}
					granularity, _ := strconv.ParseInt(*entity.Granularity, 10, 64)
					switch *entity.Kind {
					case FormatEnum.MetricKindRaw:
						containerPrediction.AddRawSample(*entity.Metric, granularity, sample)
					case FormatEnum.MetricKindUpperBound:
						containerPrediction.AddUpperBoundSample(*entity.Metric, granularity, sample)
					case FormatEnum.MetricKindLowerBound:
						containerPrediction.AddLowerBoundSample(*entity.Metric, granularity, sample)
					}
				}
			}
			containerPredictionList = append(containerPredictionList, containerPrediction)
		}
	}

	return containerPredictionList, nil
}
