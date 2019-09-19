package prediction

import (
	DaoGpu "github.com/containers-ai/alameda/datahub/pkg/dao/gpu/nvidia"
	EntityInfluxGpuPrediction "github.com/containers-ai/alameda/datahub/pkg/entity/influxdb/gpu/nvidia/prediction"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb"
	DatahubUtils "github.com/containers-ai/alameda/datahub/pkg/utils"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
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
				EntityInfluxGpuPrediction.DutyCycleMinorNumber: prediction.Metadata.MinorNumber,
				EntityInfluxGpuPrediction.DutyCycleValue:       valueInFloat64,
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
