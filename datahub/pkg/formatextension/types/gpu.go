package types

import (
	DaoGpu "github.com/containers-ai/alameda/datahub/pkg/dao/gpu/nvidia"
	Metric "github.com/containers-ai/alameda/datahub/pkg/metric"
	DatahubV1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
)

type GpuMetricExtended struct {
	*DaoGpu.GpuMetric
}

func (n *GpuMetricExtended) ProduceMetrics() *DatahubV1alpha1.GpuMetric {
	var (
		metricDataChan  = make(chan DatahubV1alpha1.MetricData)
		numOfGoroutines = 0

		datahubGpuMetadata DatahubV1alpha1.GpuMetadata
		datahubGpuMetric   DatahubV1alpha1.GpuMetric
	)

	datahubGpuMetadata = DatahubV1alpha1.GpuMetadata{
		Host:        n.Metadata.Host,
		Instance:    n.Metadata.Instance,
		Job:         n.Metadata.Job,
		MinorNumber: n.Metadata.MinorNumber,
	}

	datahubGpuMetric = DatahubV1alpha1.GpuMetric{
		Name:     n.Name,
		Uuid:     n.Uuid,
		Metadata: &datahubGpuMetadata,
	}

	for metricType, samples := range n.Metrics {
		if datahubMetricType, exist := Metric.TypeToDatahubMetricType[metricType]; exist {
			numOfGoroutines++
			go produceDatahubMetricDataFromSamples(datahubMetricType, samples, metricDataChan)
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

func (n *GpuPredictionExtended) ProducePredictions(metricType Metric.GpuMetricType) *DatahubV1alpha1.GpuPrediction {
	var (
		metricDataChan = make(chan DatahubV1alpha1.MetricData)

		datahubGpuMetadata   DatahubV1alpha1.GpuMetadata
		datahubGpuPrediction DatahubV1alpha1.GpuPrediction
	)

	datahubGpuMetadata = DatahubV1alpha1.GpuMetadata{
		Host:        n.Metadata.Host,
		Instance:    n.Metadata.Instance,
		Job:         n.Metadata.Job,
		MinorNumber: n.Metadata.MinorNumber,
	}

	datahubGpuPrediction = DatahubV1alpha1.GpuPrediction{
		Name:         n.Name,
		Uuid:         n.Uuid,
		Metadata:     &datahubGpuMetadata,
		ModelId:      n.ModelId,
		PredictionId: n.PredictionId,
	}

	if datahubMetricType, exist := Metric.TypeToDatahubMetricType[metricType]; exist {
		go produceDatahubMetricDataFromSamples(datahubMetricType, n.Metrics, metricDataChan)
	}

	receivedMetricData := <-metricDataChan
	receivedMetricData.Granularity = n.Granularity
	switch metricType {
	case Metric.TypeGpuDutyCycle:

		datahubGpuPrediction.PredictedRawData = append(datahubGpuPrediction.PredictedRawData, &receivedMetricData)
		break
	case Metric.TypeGpuDutyCycleLowerBound:
		datahubGpuPrediction.PredictedLowerboundData = append(datahubGpuPrediction.PredictedLowerboundData, &receivedMetricData)
		break
	case Metric.TypeGpuDutyCycleUpperBound:
		datahubGpuPrediction.PredictedUpperboundData = append(datahubGpuPrediction.PredictedUpperboundData, &receivedMetricData)
		break
	case Metric.TypeGpuMemoryUsedBytes:
		datahubGpuPrediction.PredictedRawData = append(datahubGpuPrediction.PredictedRawData, &receivedMetricData)
		break
	case Metric.TypeGpuMemoryUsedBytesLowerBound:
		datahubGpuPrediction.PredictedLowerboundData = append(datahubGpuPrediction.PredictedLowerboundData, &receivedMetricData)
		break
	case Metric.TypeGpuMemoryUsedBytesUpperBound:
		datahubGpuPrediction.PredictedUpperboundData = append(datahubGpuPrediction.PredictedUpperboundData, &receivedMetricData)
		break
	case Metric.TypeGpuPowerUsageMilliWatts:
		datahubGpuPrediction.PredictedRawData = append(datahubGpuPrediction.PredictedRawData, &receivedMetricData)
		break
	case Metric.TypeGpuPowerUsageMilliWattsLowerBound:
		datahubGpuPrediction.PredictedLowerboundData = append(datahubGpuPrediction.PredictedLowerboundData, &receivedMetricData)
		break
	case Metric.TypeGpuPowerUsageMilliWattsUpperBound:
		datahubGpuPrediction.PredictedUpperboundData = append(datahubGpuPrediction.PredictedUpperboundData, &receivedMetricData)
		break
	case Metric.TypeGpuTemperatureCelsius:
		datahubGpuPrediction.PredictedRawData = append(datahubGpuPrediction.PredictedRawData, &receivedMetricData)
		break
	case Metric.TypeGpuTemperatureCelsiusLowerBound:
		datahubGpuPrediction.PredictedLowerboundData = append(datahubGpuPrediction.PredictedLowerboundData, &receivedMetricData)
		break
	case Metric.TypeGpuTemperatureCelsiusUpperBound:
		datahubGpuPrediction.PredictedUpperboundData = append(datahubGpuPrediction.PredictedUpperboundData, &receivedMetricData)
		break
	}

	return &datahubGpuPrediction
}
