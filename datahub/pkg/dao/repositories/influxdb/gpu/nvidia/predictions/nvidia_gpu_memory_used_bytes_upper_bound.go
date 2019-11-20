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

type MemoryUsedBytesUpperBoundRepository struct {
	influxDB *InternalInflux.InfluxClient
}

func NewMemoryUsedBytesUpperBoundRepositoryWithConfig(cfg InternalInflux.Config) *MemoryUsedBytesUpperBoundRepository {
	return &MemoryUsedBytesUpperBoundRepository{
		influxDB: InternalInflux.NewClient(&cfg),
	}
}

func (r *MemoryUsedBytesUpperBoundRepository) CreatePredictions(predictions []*DaoGpu.GpuPrediction) error {
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
				EntityInfluxGpuPrediction.MemoryUsedBytesHost:        prediction.Metadata.Host,
				EntityInfluxGpuPrediction.MemoryUsedBytesName:        prediction.Name,
				EntityInfluxGpuPrediction.MemoryUsedBytesUuid:        prediction.Uuid,
				EntityInfluxGpuPrediction.MemoryUsedBytesGranularity: strconv.FormatInt(granularity, 10),
			}

			// Pack influx fields
			fields := map[string]interface{}{
				EntityInfluxGpuPrediction.MemoryUsedBytesModelId:      metric.ModelId,
				EntityInfluxGpuPrediction.MemoryUsedBytesPredictionId: metric.PredictionId,
				EntityInfluxGpuPrediction.MemoryUsedBytesMinorNumber:  prediction.Metadata.MinorNumber,
				EntityInfluxGpuPrediction.MemoryUsedBytesValue:        valueInFloat64,
			}

			// Add to influx point list
			point, err := InfluxClient.NewPoint(string(MemoryUsedBytesUpperBound), tags, fields, metric.Timestamp)
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

func (r *MemoryUsedBytesUpperBoundRepository) ListPredictions(host, minorNumber, modelId, predictionId, granularity string, condition *DBCommon.QueryCondition) ([]*EntityInfluxGpuPrediction.MemoryUsedBytesEntity, error) {
	entities := make([]*EntityInfluxGpuPrediction.MemoryUsedBytesEntity, 0)

	influxdbStatement := InternalInflux.Statement{
		QueryCondition: condition,
		Measurement:    MemoryUsedBytesUpperBound,
		GroupByTags:    []string{"host", "uuid"},
	}

	influxdbStatement.AppendWhereClause("AND", EntityInfluxGpuPrediction.MemoryUsedBytesHost, "=", host)
	influxdbStatement.AppendWhereClause("AND", EntityInfluxGpuPrediction.MemoryUsedBytesMinorNumber, "=", minorNumber)
	influxdbStatement.AppendWhereClause("AND", EntityInfluxGpuPrediction.MemoryUsedBytesModelId, "=", modelId)
	influxdbStatement.AppendWhereClause("AND", EntityInfluxGpuPrediction.MemoryUsedBytesPredictionId, "=", predictionId)
	influxdbStatement.AppendWhereClause("AND", EntityInfluxGpuPrediction.MemoryUsedBytesGranularity, "=", granularity)
	influxdbStatement.AppendWhereClauseFromTimeCondition()
	influxdbStatement.SetOrderClauseFromQueryCondition()
	influxdbStatement.SetLimitClauseFromQueryCondition()
	cmd := influxdbStatement.BuildQueryCmd()

	response, err := r.influxDB.QueryDB(cmd, string(RepoInflux.GpuPrediction))
	if err != nil {
		return entities, errors.Wrap(err, "failed to list nvidia gpu memory used bytes upper bound predictions")
	}

	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			for j := 0; j < group.GetRowNum(); j++ {
				entity := EntityInfluxGpuPrediction.NewMemoryUsedBytesEntityFromMap(group.GetRow(j))
				entities = append(entities, &entity)
			}
		}
	}

	return entities, nil
}
