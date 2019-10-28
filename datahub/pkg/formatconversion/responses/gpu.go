package responses

import (
	DaoGpu "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/gpu/influxdb"
	FormatEnum "github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	ApiCommon "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	ApiGpu "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/gpu"
	ApiPredictions "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/predictions"
)

type GpuMetricExtended struct {
	*DaoGpu.GpuMetric
}

func (n *GpuMetricExtended) ProduceMetrics() *ApiGpu.GpuMetric {
	var (
		metricDataChan  = make(chan ApiCommon.MetricData)
		numOfGoroutines = 0

		datahubGpuMetadata ApiGpu.GpuMetadata
		datahubGpuMetric   ApiGpu.GpuMetric
	)

	datahubGpuMetadata = ApiGpu.GpuMetadata{
		Host:        n.Metadata.Host,
		Instance:    n.Metadata.Instance,
		Job:         n.Metadata.Job,
		MinorNumber: n.Metadata.MinorNumber,
	}

	datahubGpuMetric = ApiGpu.GpuMetric{
		Name:     n.Name,
		Uuid:     n.Uuid,
		Metadata: &datahubGpuMetadata,
	}

	for metricType, samples := range n.Metrics {
		if datahubMetricType, exist := FormatEnum.TypeToDatahubMetricType[metricType]; exist {
			numOfGoroutines++
			go produceMetricDataFromSamples(datahubMetricType, samples, metricDataChan)
		}
	}

	for i := 0; i < numOfGoroutines; i++ {
		receivedMetricData := <-metricDataChan
		datahubGpuMetric.MetricData = append(datahubGpuMetric.MetricData, &receivedMetricData)
	}

	return &datahubGpuMetric
}

type GpuPredictionExtended struct {
	*DaoGpu.GpuPrediction
}

func (n *GpuPredictionExtended) ProducePredictions(metricType FormatEnum.GpuMetricType) *ApiGpu.GpuPrediction {
	var (
		metricDataChan = make(chan ApiPredictions.MetricData)

		datahubGpuMetadata   ApiGpu.GpuMetadata
		datahubGpuPrediction ApiGpu.GpuPrediction
	)

	datahubGpuMetadata = ApiGpu.GpuMetadata{
		Host:        n.Metadata.Host,
		Instance:    n.Metadata.Instance,
		Job:         n.Metadata.Job,
		MinorNumber: n.Metadata.MinorNumber,
	}

	datahubGpuPrediction = ApiGpu.GpuPrediction{
		Name:     n.Name,
		Uuid:     n.Uuid,
		Metadata: &datahubGpuMetadata,
	}

	if datahubMetricType, exist := FormatEnum.TypeToDatahubMetricType[metricType]; exist {
		go producePredictionMetricDataFromSamples(datahubMetricType, n.Granularity, n.Metrics, metricDataChan)
	}

	receivedMetricData := <-metricDataChan
	switch metricType {
	case FormatEnum.TypeGpuDutyCycle:
		datahubGpuPrediction.PredictedRawData = append(datahubGpuPrediction.PredictedRawData, &receivedMetricData)
		break
	case FormatEnum.TypeGpuDutyCycleLowerBound:
		datahubGpuPrediction.PredictedLowerboundData = append(datahubGpuPrediction.PredictedLowerboundData, &receivedMetricData)
		break
	case FormatEnum.TypeGpuDutyCycleUpperBound:
		datahubGpuPrediction.PredictedUpperboundData = append(datahubGpuPrediction.PredictedUpperboundData, &receivedMetricData)
		break
	case FormatEnum.TypeGpuMemoryUsedBytes:
		datahubGpuPrediction.PredictedRawData = append(datahubGpuPrediction.PredictedRawData, &receivedMetricData)
		break
	case FormatEnum.TypeGpuMemoryUsedBytesLowerBound:
		datahubGpuPrediction.PredictedLowerboundData = append(datahubGpuPrediction.PredictedLowerboundData, &receivedMetricData)
		break
	case FormatEnum.TypeGpuMemoryUsedBytesUpperBound:
		datahubGpuPrediction.PredictedUpperboundData = append(datahubGpuPrediction.PredictedUpperboundData, &receivedMetricData)
		break
	case FormatEnum.TypeGpuPowerUsageMilliWatts:
		datahubGpuPrediction.PredictedRawData = append(datahubGpuPrediction.PredictedRawData, &receivedMetricData)
		break
	case FormatEnum.TypeGpuPowerUsageMilliWattsLowerBound:
		datahubGpuPrediction.PredictedLowerboundData = append(datahubGpuPrediction.PredictedLowerboundData, &receivedMetricData)
		break
	case FormatEnum.TypeGpuPowerUsageMilliWattsUpperBound:
		datahubGpuPrediction.PredictedUpperboundData = append(datahubGpuPrediction.PredictedUpperboundData, &receivedMetricData)
		break
	case FormatEnum.TypeGpuTemperatureCelsius:
		datahubGpuPrediction.PredictedRawData = append(datahubGpuPrediction.PredictedRawData, &receivedMetricData)
		break
	case FormatEnum.TypeGpuTemperatureCelsiusLowerBound:
		datahubGpuPrediction.PredictedLowerboundData = append(datahubGpuPrediction.PredictedLowerboundData, &receivedMetricData)
		break
	case FormatEnum.TypeGpuTemperatureCelsiusUpperBound:
		datahubGpuPrediction.PredictedUpperboundData = append(datahubGpuPrediction.PredictedUpperboundData, &receivedMetricData)
		break
	}

	return &datahubGpuPrediction
}
