package requests

import (
	DaoGpu "github.com/containers-ai/alameda/datahub/pkg/dao/gpu/nvidia"
	Metric "github.com/containers-ai/alameda/datahub/pkg/metric"
	DatahubV1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes"
)

type CreateGpuPredictionsRequestExtended struct {
	DatahubV1alpha1.CreateGpuPredictionsRequest
}

func (r *CreateGpuPredictionsRequestExtended) Validate() error {
	return nil
}

func (r *CreateGpuPredictionsRequestExtended) ProducePredictions() DaoGpu.GpuPredictionMap {
	gpuPredictionMap := make(map[Metric.GpuMetricType][]*DaoGpu.GpuPrediction)

	rawDataTypeMap := map[DatahubV1alpha1.MetricType]Metric.GpuMetricType{
		DatahubV1alpha1.MetricType_MEMORY_USAGE_BYTES:  Metric.TypeGpuMemoryUsedBytes,
		DatahubV1alpha1.MetricType_POWER_USAGE_WATTS:   Metric.TypeGpuPowerUsageMilliWatts,
		DatahubV1alpha1.MetricType_TEMPERATURE_CELSIUS: Metric.TypeGpuTemperatureCelsius,
		DatahubV1alpha1.MetricType_DUTY_CYCLE:          Metric.TypeGpuDutyCycle,
	}

	lowerBoundTypeMap := map[DatahubV1alpha1.MetricType]Metric.GpuMetricType{
		DatahubV1alpha1.MetricType_MEMORY_USAGE_BYTES:  Metric.TypeGpuMemoryUsedBytesLowerBound,
		DatahubV1alpha1.MetricType_POWER_USAGE_WATTS:   Metric.TypeGpuPowerUsageMilliWattsLowerBound,
		DatahubV1alpha1.MetricType_TEMPERATURE_CELSIUS: Metric.TypeGpuTemperatureCelsiusLowerBound,
		DatahubV1alpha1.MetricType_DUTY_CYCLE:          Metric.TypeGpuDutyCycleLowerBound,
	}

	upperBoundTypeMap := map[DatahubV1alpha1.MetricType]Metric.GpuMetricType{
		DatahubV1alpha1.MetricType_MEMORY_USAGE_BYTES:  Metric.TypeGpuMemoryUsedBytesUpperBound,
		DatahubV1alpha1.MetricType_POWER_USAGE_WATTS:   Metric.TypeGpuPowerUsageMilliWattsUpperBound,
		DatahubV1alpha1.MetricType_TEMPERATURE_CELSIUS: Metric.TypeGpuTemperatureCelsiusUpperBound,
		DatahubV1alpha1.MetricType_DUTY_CYCLE:          Metric.TypeGpuDutyCycleUpperBound,
	}

	for _, predictions := range r.GetGpuPredictions() {
		gpu := DaoGpu.Gpu{}
		gpu.Name = predictions.GetName()
		gpu.Uuid = predictions.GetUuid()
		gpu.Metadata.Host = predictions.GetMetadata().GetHost()
		gpu.Metadata.Instance = predictions.GetMetadata().GetInstance()
		gpu.Metadata.Job = predictions.GetMetadata().GetJob()
		gpu.Metadata.MinorNumber = predictions.GetMetadata().GetMinorNumber()

		// Prepare predicted raw data
		for _, data := range predictions.GetPredictedRawData() {
			metricType := rawDataTypeMap[data.GetMetricType()]

			gpuPrediction := DaoGpu.GpuPrediction{}
			gpuPrediction.Gpu = gpu
			gpuPrediction.Granularity = data.GetGranularity()
			gpuPrediction.ModelId = predictions.GetModelId()
			gpuPrediction.PredictionId = predictions.GetPredictionId()

			if _, exist := gpuPredictionMap[metricType]; !exist {
				gpuPredictionMap[metricType] = make([]*DaoGpu.GpuPrediction, 0)
			}

			for _, sample := range data.GetData() {
				timestamp, err := ptypes.Timestamp(sample.GetTime())
				if err != nil {
					scope.Error(" failed: " + err.Error())
				}
				sample := Metric.Sample{
					Timestamp: timestamp,
					Value:     sample.GetNumValue(),
				}
				gpuPrediction.Metrics = append(gpuPrediction.Metrics, sample)
			}

			gpuPredictionMap[metricType] = append(gpuPredictionMap[metricType], &gpuPrediction)
		}

		// Prepare predicted lower bound data
		for _, data := range predictions.GetPredictedLowerboundData() {
			metricType := lowerBoundTypeMap[data.GetMetricType()]

			gpuPrediction := DaoGpu.GpuPrediction{}
			gpuPrediction.Gpu = gpu
			gpuPrediction.Granularity = data.GetGranularity()
			gpuPrediction.ModelId = predictions.GetModelId()
			gpuPrediction.PredictionId = predictions.GetPredictionId()

			if _, exist := gpuPredictionMap[metricType]; !exist {
				gpuPredictionMap[metricType] = make([]*DaoGpu.GpuPrediction, 0)
			}

			for _, sample := range data.GetData() {
				timestamp, err := ptypes.Timestamp(sample.GetTime())
				if err != nil {
					scope.Error(" failed: " + err.Error())
				}
				sample := Metric.Sample{
					Timestamp: timestamp,
					Value:     sample.GetNumValue(),
				}
				gpuPrediction.Metrics = append(gpuPrediction.Metrics, sample)
			}

			gpuPredictionMap[metricType] = append(gpuPredictionMap[metricType], &gpuPrediction)
		}

		// Prepare predicted upper bound data
		for _, data := range predictions.GetPredictedUpperboundData() {
			metricType := upperBoundTypeMap[data.GetMetricType()]

			gpuPrediction := DaoGpu.GpuPrediction{}
			gpuPrediction.Gpu = gpu
			gpuPrediction.Granularity = data.GetGranularity()
			gpuPrediction.ModelId = predictions.GetModelId()
			gpuPrediction.PredictionId = predictions.GetPredictionId()

			if _, exist := gpuPredictionMap[metricType]; !exist {
				gpuPredictionMap[metricType] = make([]*DaoGpu.GpuPrediction, 0)
			}

			for _, sample := range data.GetData() {
				timestamp, err := ptypes.Timestamp(sample.GetTime())
				if err != nil {
					scope.Error(" failed: " + err.Error())
				}
				sample := Metric.Sample{
					Timestamp: timestamp,
					Value:     sample.GetNumValue(),
				}
				gpuPrediction.Metrics = append(gpuPrediction.Metrics, sample)
			}

			gpuPredictionMap[metricType] = append(gpuPredictionMap[metricType], &gpuPrediction)
		}
	}

	return gpuPredictionMap
}
