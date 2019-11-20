package predictions

import (
	EntityInfluxGpuPrediction "github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/gpu/nvidia/predictions"
	DaoGpu "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/gpu/influxdb"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb"
	DatahubUtils "github.com/containers-ai/alameda/datahub/pkg/utils"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InternalInfluxModels "github.com/containers-ai/alameda/internal/pkg/database/influxdb/models"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"github.com/pkg/errors"
	"strconv"
)

type DutyCycleUpperBoundRepository struct {
	influxDB *InternalInflux.InfluxClient
}

func NewDutyCycleUpperBoundRepositoryWithConfig(cfg InternalInflux.Config) *DutyCycleUpperBoundRepository {
	return &DutyCycleUpperBoundRepository{
		influxDB: InternalInflux.NewClient(&cfg),
	}
}

func (r *DutyCycleUpperBoundRepository) CreatePredictions(predictions []*DaoGpu.GpuPrediction) error {
	points := make([]*InfluxClient.Point, 0)

	for _, prediction := range predictions {
		granularity := int64(30)
		if prediction.Granularity != 0 {
			granularity = prediction.Granularity
		}

		for _, metric := range prediction.Metrics {
			// Parse float string to value
			valueInFloat64, err := DatahubUtils.StringToFloat64(metric.Value)
			if err != nil {
				return errors.Wrap(err, "failed to parse string to float64")
			}

			// Pack influx tags
			tags := map[string]string{
				EntityInfluxGpuPrediction.DutyCycleHost:        prediction.Metadata.Host,
				EntityInfluxGpuPrediction.DutyCycleName:        prediction.Name,
				EntityInfluxGpuPrediction.DutyCycleUuid:        prediction.Uuid,
				EntityInfluxGpuPrediction.DutyCycleGranularity: strconv.FormatInt(granularity, 10),
			}

			// Pack influx fields
			fields := map[string]interface{}{
				EntityInfluxGpuPrediction.DutyCycleModelId:      metric.ModelId,
				EntityInfluxGpuPrediction.DutyCyclePredictionId: metric.PredictionId,
				EntityInfluxGpuPrediction.DutyCycleMinorNumber:  prediction.Metadata.MinorNumber,
				EntityInfluxGpuPrediction.DutyCycleValue:        valueInFloat64,
			}

			// Add to influx point list
			point, err := InfluxClient.NewPoint(string(DutyCycleUpperBound), tags, fields, metric.Timestamp)
			if err != nil {
				return errors.Wrap(err, "failed to instance influxdb data point")
			}
			points = append(points, point)
		}
	}

	// Batch write influxdb data points
	err := r.influxDB.WritePoints(points, InfluxClient.BatchPointsConfig{
		Database: string(RepoInflux.GpuPrediction),
	})
	if err != nil {
		return errors.Wrap(err, "failed to batch write influxdb data points")
	}

	return nil
}

func (r *DutyCycleUpperBoundRepository) ListPredictions(host, minorNumber, modelId, predictionId, granularity string, condition *DBCommon.QueryCondition) ([]*EntityInfluxGpuPrediction.DutyCycleEntity, error) {
	entities := make([]*EntityInfluxGpuPrediction.DutyCycleEntity, 0)

	influxdbStatement := InternalInflux.Statement{
		QueryCondition: condition,
		Measurement:    DutyCycleUpperBound,
		GroupByTags:    []string{"host", "uuid"},
	}

	influxdbStatement.AppendWhereClause("AND", EntityInfluxGpuPrediction.DutyCycleHost, "=", host)
	influxdbStatement.AppendWhereClause("AND", EntityInfluxGpuPrediction.DutyCycleMinorNumber, "=", minorNumber)
	influxdbStatement.AppendWhereClause("AND", EntityInfluxGpuPrediction.DutyCycleModelId, "=", modelId)
	influxdbStatement.AppendWhereClause("AND", EntityInfluxGpuPrediction.DutyCyclePredictionId, "=", predictionId)
	influxdbStatement.AppendWhereClause("AND", EntityInfluxGpuPrediction.DutyCycleGranularity, "=", granularity)
	influxdbStatement.AppendWhereClauseFromTimeCondition()
	influxdbStatement.SetOrderClauseFromQueryCondition()
	influxdbStatement.SetLimitClauseFromQueryCondition()
	cmd := influxdbStatement.BuildQueryCmd()

	response, err := r.influxDB.QueryDB(cmd, string(RepoInflux.GpuPrediction))
	if err != nil {
		return entities, errors.Wrap(err, "failed to list nvidia gpu duty cycle lower bound predictions")
	}

	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			for j := 0; j < group.GetRowNum(); j++ {
				entity := EntityInfluxGpuPrediction.NewDutyCycleEntityFromMap(group.GetRow(j))
				entities = append(entities, &entity)
			}
		}
	}

	return entities, nil
}
