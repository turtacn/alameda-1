package requests

import (
	DaoGpu "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/gpu/influxdb"
	FormatEnum "github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	FormatTypes "github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	ApiCommon "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	ApiGpu "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/gpu"
	"github.com/golang/protobuf/ptypes"
)

type CreateGpuPredictionsRequestExtended struct {
	ApiGpu.CreateGpuPredictionsRequest
}

func (r *CreateGpuPredictionsRequestExtended) Validate() error {
	return nil
}

func (r *CreateGpuPredictionsRequestExtended) ProducePredictions() DaoGpu.GpuPredictionMap {
	gpuPredictionMap := make(map[FormatEnum.GpuMetricType][]*DaoGpu.GpuPrediction)

	rawDataTypeMap := map[ApiCommon.MetricType]FormatEnum.GpuMetricType{
		ApiCommon.MetricType_MEMORY_USAGE_BYTES:  FormatEnum.TypeGpuMemoryUsedBytes,
		ApiCommon.MetricType_POWER_USAGE_WATTS:   FormatEnum.TypeGpuPowerUsageMilliWatts,
		ApiCommon.MetricType_TEMPERATURE_CELSIUS: FormatEnum.TypeGpuTemperatureCelsius,
		ApiCommon.MetricType_DUTY_CYCLE:          FormatEnum.TypeGpuDutyCycle,
	}

	lowerBoundTypeMap := map[ApiCommon.MetricType]FormatEnum.GpuMetricType{
		ApiCommon.MetricType_MEMORY_USAGE_BYTES:  FormatEnum.TypeGpuMemoryUsedBytesLowerBound,
		ApiCommon.MetricType_POWER_USAGE_WATTS:   FormatEnum.TypeGpuPowerUsageMilliWattsLowerBound,
		ApiCommon.MetricType_TEMPERATURE_CELSIUS: FormatEnum.TypeGpuTemperatureCelsiusLowerBound,
		ApiCommon.MetricType_DUTY_CYCLE:          FormatEnum.TypeGpuDutyCycleLowerBound,
	}

	upperBoundTypeMap := map[ApiCommon.MetricType]FormatEnum.GpuMetricType{
		ApiCommon.MetricType_MEMORY_USAGE_BYTES:  FormatEnum.TypeGpuMemoryUsedBytesUpperBound,
		ApiCommon.MetricType_POWER_USAGE_WATTS:   FormatEnum.TypeGpuPowerUsageMilliWattsUpperBound,
		ApiCommon.MetricType_TEMPERATURE_CELSIUS: FormatEnum.TypeGpuTemperatureCelsiusUpperBound,
		ApiCommon.MetricType_DUTY_CYCLE:          FormatEnum.TypeGpuDutyCycleUpperBound,
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

			if _, exist := gpuPredictionMap[metricType]; !exist {
				gpuPredictionMap[metricType] = make([]*DaoGpu.GpuPrediction, 0)
			}

			for _, sample := range data.GetData() {
				timestamp, err := ptypes.Timestamp(sample.GetTime())
				if err != nil {
					scope.Error(" failed: " + err.Error())
				}
				sample := FormatTypes.PredictionSample{
					Timestamp:    timestamp,
					Value:        sample.GetNumValue(),
					ModelId:      sample.GetModelId(),
					PredictionId: sample.GetPredictionId(),
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

			if _, exist := gpuPredictionMap[metricType]; !exist {
				gpuPredictionMap[metricType] = make([]*DaoGpu.GpuPrediction, 0)
			}

			for _, sample := range data.GetData() {
				timestamp, err := ptypes.Timestamp(sample.GetTime())
				if err != nil {
					scope.Error(" failed: " + err.Error())
				}
				sample := FormatTypes.PredictionSample{
					Timestamp:    timestamp,
					Value:        sample.GetNumValue(),
					ModelId:      sample.GetModelId(),
					PredictionId: sample.GetPredictionId(),
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

			if _, exist := gpuPredictionMap[metricType]; !exist {
				gpuPredictionMap[metricType] = make([]*DaoGpu.GpuPrediction, 0)
			}

			for _, sample := range data.GetData() {
				timestamp, err := ptypes.Timestamp(sample.GetTime())
				if err != nil {
					scope.Error(" failed: " + err.Error())
				}
				sample := FormatTypes.PredictionSample{
					Timestamp:    timestamp,
					Value:        sample.GetNumValue(),
					ModelId:      sample.GetModelId(),
					PredictionId: sample.GetPredictionId(),
				}
				gpuPrediction.Metrics = append(gpuPrediction.Metrics, sample)
			}

			gpuPredictionMap[metricType] = append(gpuPredictionMap[metricType], &gpuPrediction)
		}
	}

	return gpuPredictionMap
}
