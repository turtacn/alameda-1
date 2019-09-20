package prediction

import (
	DaoGpu "github.com/containers-ai/alameda/datahub/pkg/dao/gpu/nvidia"
	EntityInfluxGpuPrediction "github.com/containers-ai/alameda/datahub/pkg/entity/influxdb/gpu/nvidia/prediction"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb"
	DatahubUtils "github.com/containers-ai/alameda/datahub/pkg/utils"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InternalInfluxModels "github.com/containers-ai/alameda/internal/pkg/database/influxdb/models"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"github.com/pkg/errors"
	"strconv"
)

type MemoryUsedBytesLowerBoundRepository struct {
	influxDB *InternalInflux.InfluxClient
}

func NewMemoryUsedBytesLowerBoundRepositoryWithConfig(cfg InternalInflux.Config) *MemoryUsedBytesLowerBoundRepository {
	return &MemoryUsedBytesLowerBoundRepository{
		influxDB: InternalInflux.NewClient(&cfg),
	}
}

func (r *MemoryUsedBytesLowerBoundRepository) CreatePredictions(predictions []*DaoGpu.GpuPrediction) error {
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
				EntityInfluxGpuPrediction.MemoryUsedBytesMinorNumber: prediction.Metadata.MinorNumber,
				EntityInfluxGpuPrediction.MemoryUsedBytesValue:       valueInFloat64,
			}

			// Add to influx point list
			point, err := InfluxClient.NewPoint(string(MemoryUsedBytesLowerBound), tags, fields, metric.Timestamp)
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

func (r *MemoryUsedBytesLowerBoundRepository) ListPredictions(host, minorNumber, granularity string, condition *DBCommon.QueryCondition) ([]*EntityInfluxGpuPrediction.MemoryUsedBytesEntity, error) {
	entities := make([]*EntityInfluxGpuPrediction.MemoryUsedBytesEntity, 0)

	influxdbStatement := InternalInflux.Statement{
		QueryCondition: condition,
		Measurement:    MemoryUsedBytesLowerBound,
		GroupByTags:    []string{"host", "uuid"},
	}

	influxdbStatement.AppendWhereClauseFromTimeCondition()
	influxdbStatement.AppendWhereClause(EntityInfluxGpuPrediction.MemoryUsedBytesHost, "=", host)
	influxdbStatement.AppendWhereClause(EntityInfluxGpuPrediction.MemoryUsedBytesMinorNumber, "=", minorNumber)
	influxdbStatement.AppendWhereClause(EntityInfluxGpuPrediction.MemoryUsedBytesGranularity, "=", granularity)
	influxdbStatement.SetOrderClauseFromQueryCondition()
	influxdbStatement.SetLimitClauseFromQueryCondition()
	cmd := influxdbStatement.BuildQueryCmd()

	response, err := r.influxDB.QueryDB(cmd, string(RepoInflux.GpuPrediction))
	if err != nil {
		return entities, errors.Wrap(err, "failed to list nvidia gpu memory used bytes lower bound predictions")
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
