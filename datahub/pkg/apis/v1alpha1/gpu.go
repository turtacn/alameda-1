package v1alpha1

import (
	DaoGpu "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/gpu/influxdb/nvidia"
	FormatEnum "github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	FormatRequest "github.com/containers-ai/alameda/datahub/pkg/formatconversion/requests"
	FormatResponse "github.com/containers-ai/alameda/datahub/pkg/formatconversion/responses"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiGpu "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/gpu"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
	"strconv"
)

func (s *ServiceV1alpha1) ListGpus(ctx context.Context, in *ApiGpu.ListGpusRequest) (*ApiGpu.ListGpusResponse, error) {
	scope.Debug("Request received from ListGpus grpc function: " + AlamedaUtils.InterfaceToString(in))

	queryCondition := &DBCommon.QueryCondition{}
	if in.GetQueryCondition() == nil {
		queryCondition = DBCommon.NewQueryCondition(1, 0, 0, 30)
	} else {
		queryCondition = DBCommon.BuildQueryConditionV1(in.GetQueryCondition())
	}

	gpuDAO := DaoGpu.NewGpuWithConfig(*s.Config.InfluxDB)
	metrics, err := gpuDAO.ListGpus(in.GetHost(), in.GetMinorNumber(), queryCondition)
	if err != nil {
		scope.Errorf("failed to ListGpus: %+v", err)
		return &ApiGpu.ListGpusResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	gpus := make([]*ApiGpu.Gpu, 0)
	for _, metric := range metrics {
		metadata := &ApiGpu.GpuMetadata{
			Host:        metric.Metadata.Host,
			Instance:    metric.Metadata.Instance,
			Job:         metric.Metadata.Job,
			MinorNumber: metric.Metadata.MinorNumber,
		}
		spec := &ApiGpu.GpuSpec{
			MemoryTotal: metric.Spec.MemoryTotal,
		}
		gpu := &ApiGpu.Gpu{
			Name:     metric.Name,
			Uuid:     metric.Uuid,
			Metadata: metadata,
			Spec:     spec,
		}
		gpus = append(gpus, gpu)
	}

	response := &ApiGpu.ListGpusResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		Gpus: gpus,
	}

	return response, nil
}

func (s *ServiceV1alpha1) ListGpuMetrics(ctx context.Context, in *ApiGpu.ListGpuMetricsRequest) (*ApiGpu.ListGpuMetricsResponse, error) {
	scope.Debug("Request received from ListGpuMetrics grpc function: " + AlamedaUtils.InterfaceToString(in))

	metricDAO := DaoGpu.NewMetricWithConfig(*s.Config.InfluxDB)
	metrics, err := metricDAO.ListMetrics(in.GetHost(), in.GetMinorNumber(), DBCommon.BuildQueryConditionV1(in.GetQueryCondition()))
	if err != nil {
		scope.Errorf("failed to ListGpuMetrics: %+v", err)
		return &ApiGpu.ListGpuMetricsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	gpuMetrics := make([]*ApiGpu.GpuMetric, 0)
	for _, metric := range metrics {
		gpuMetricExtended := FormatResponse.GpuMetricExtended{GpuMetric: metric}
		datahubGpuMetric := gpuMetricExtended.ProduceMetrics()
		gpuMetrics = append(gpuMetrics, datahubGpuMetric)
	}

	response := &ApiGpu.ListGpuMetricsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		GpuMetrics: gpuMetrics,
	}

	return response, nil
}

func (s *ServiceV1alpha1) ListGpuPredictions(ctx context.Context, in *ApiGpu.ListGpuPredictionsRequest) (*ApiGpu.ListGpuPredictionsResponse, error) {
	scope.Debug("Request received from ListGpuPredictions grpc function: " + AlamedaUtils.InterfaceToString(in))

	granularity := "30"
	if in.GetGranularity() != 0 {
		granularity = strconv.FormatInt(in.GetGranularity(), 10)
	}

	predictionDAO := DaoGpu.NewPredictionWithConfig(*s.Config.InfluxDB)
	predictionsMap, err := predictionDAO.ListPredictions(in.GetHost(), in.GetMinorNumber(), in.GetModelId(), in.GetPredictionId(), granularity, DBCommon.BuildQueryConditionV1(in.GetQueryCondition()))
	if err != nil {
		scope.Errorf("failed to ListGpuPredictions: %+v", err)
		return &ApiGpu.ListGpuPredictionsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	gpuPredictions := make([]*ApiGpu.GpuPrediction, 0)
	for metricType, predictions := range predictionsMap {
		for _, prediction := range predictions {
			gpu := &ApiGpu.GpuPrediction{}
			gpuPredictionExtended := FormatResponse.GpuPredictionExtended{GpuPrediction: prediction}
			gpuPrediction := gpuPredictionExtended.ProducePredictions(metricType)
			found := false

			// Look up if gpu is found
			for _, gpu = range gpuPredictions {
				if gpu.Uuid == gpuPrediction.Uuid {
					found = true
					break
				}
			}

			if found == false {
				gpuPredictions = append(gpuPredictions, gpuPrediction)
			} else {
				switch metricType {
				case FormatEnum.TypeGpuDutyCycle:
					gpu.PredictedRawData = append(gpu.PredictedRawData, gpuPrediction.PredictedRawData[0])
					break
				case FormatEnum.TypeGpuDutyCycleLowerBound:
					gpu.PredictedLowerboundData = append(gpu.PredictedLowerboundData, gpuPrediction.PredictedLowerboundData[0])
					break
				case FormatEnum.TypeGpuDutyCycleUpperBound:
					gpu.PredictedUpperboundData = append(gpu.PredictedUpperboundData, gpuPrediction.PredictedUpperboundData[0])
					break
				case FormatEnum.TypeGpuMemoryUsedBytes:
					gpu.PredictedRawData = append(gpu.PredictedRawData, gpuPrediction.PredictedRawData[0])
					break
				case FormatEnum.TypeGpuMemoryUsedBytesLowerBound:
					gpu.PredictedLowerboundData = append(gpu.PredictedLowerboundData, gpuPrediction.PredictedLowerboundData[0])
					break
				case FormatEnum.TypeGpuMemoryUsedBytesUpperBound:
					gpu.PredictedUpperboundData = append(gpu.PredictedUpperboundData, gpuPrediction.PredictedUpperboundData[0])
					break
				case FormatEnum.TypeGpuPowerUsageMilliWatts:
					gpu.PredictedRawData = append(gpu.PredictedRawData, gpuPrediction.PredictedRawData[0])
					break
				case FormatEnum.TypeGpuPowerUsageMilliWattsLowerBound:
					gpu.PredictedLowerboundData = append(gpu.PredictedLowerboundData, gpuPrediction.PredictedLowerboundData[0])
					break
				case FormatEnum.TypeGpuPowerUsageMilliWattsUpperBound:
					gpu.PredictedUpperboundData = append(gpu.PredictedUpperboundData, gpuPrediction.PredictedUpperboundData[0])
					break
				case FormatEnum.TypeGpuTemperatureCelsius:
					gpu.PredictedRawData = append(gpu.PredictedRawData, gpuPrediction.PredictedRawData[0])
					break
				case FormatEnum.TypeGpuTemperatureCelsiusLowerBound:
					gpu.PredictedLowerboundData = append(gpu.PredictedLowerboundData, gpuPrediction.PredictedLowerboundData[0])
					break
				case FormatEnum.TypeGpuTemperatureCelsiusUpperBound:
					gpu.PredictedUpperboundData = append(gpu.PredictedUpperboundData, gpuPrediction.PredictedUpperboundData[0])
					break
				}
			}
		}
	}

	response := &ApiGpu.ListGpuPredictionsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		GpuPredictions: gpuPredictions,
	}

	return response, nil
}

func (s *ServiceV1alpha1) CreateGpuPredictions(ctx context.Context, in *ApiGpu.CreateGpuPredictionsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateGpuPredictions grpc function: " + AlamedaUtils.InterfaceToString(in))

	requestExtended := FormatRequest.CreateGpuPredictionsRequestExtended{CreateGpuPredictionsRequest: *in}
	if requestExtended.Validate() != nil {
		return &status.Status{
			Code: int32(code.Code_INVALID_ARGUMENT),
		}, nil
	}

	predictionDAO := DaoGpu.NewPredictionWithConfig(*s.Config.InfluxDB)
	err := predictionDAO.CreatePredictions(requestExtended.ProducePredictions())
	if err != nil {
		scope.Errorf("failed to create gpu predictions: %+v", err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}
