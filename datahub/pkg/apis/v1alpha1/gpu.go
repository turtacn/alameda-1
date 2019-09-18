package v1alpha1

import (
	DaoGpu "github.com/containers-ai/alameda/datahub/pkg/dao/gpu/nvidia/impl"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	DatahubV1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (s *ServiceV1alpha1) ListGpus(ctx context.Context, in *DatahubV1alpha1.ListGpusRequest) (*DatahubV1alpha1.ListGpusResponse, error) {
	scope.Debug("Request received from ListGpus grpc function: " + AlamedaUtils.InterfaceToString(in))

	gpuDAO := DaoGpu.NewGpuWithConfig(*s.Config.InfluxDB)
	metrics, err := gpuDAO.ListGpus(in.GetHost(), in.GetMinorNumber(), DBCommon.BuildQueryConditionV1(in.GetQueryCondition()))
	if err != nil {
		scope.Errorf("failed to ListGpus: %+v", err)
		return &DatahubV1alpha1.ListGpusResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	gpus := make([]*DatahubV1alpha1.Gpu, 0)
	for _, metric := range metrics {
		metadata := &DatahubV1alpha1.GpuMetadata{
			Host:        metric.Metadata.Host,
			Instance:    metric.Metadata.Instance,
			Job:         metric.Metadata.Job,
			MinorNumber: metric.Metadata.MinorNumber,
		}
		spec := &DatahubV1alpha1.GpuSpec{
			MemoryTotal: metric.Spec.MemoryTotal,
		}
		gpu := &DatahubV1alpha1.Gpu{
			Name:     metric.Name,
			Uuid:     metric.Uuid,
			Metadata: metadata,
			Spec:     spec,
		}
		gpus = append(gpus, gpu)
	}

	response := &DatahubV1alpha1.ListGpusResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		Gpus: gpus,
	}

	return response, nil
}

func (s *ServiceV1alpha1) ListGpuMetrics(ctx context.Context, in *DatahubV1alpha1.ListGpuMetricsRequest) (*DatahubV1alpha1.ListGpuMetricsResponse, error) {
	scope.Debug("Request received from ListGpuMetrics grpc function: " + AlamedaUtils.InterfaceToString(in))

	metricDAO := DaoGpu.NewMetricWithConfig(*s.Config.InfluxDB)
	metrics, err := metricDAO.ListMetrics(in.GetHost(), in.GetMinorNumber(), DBCommon.BuildQueryConditionV1(in.GetQueryCondition()))
	if err != nil {
		scope.Errorf("failed to ListGpuMetrics: %+v", err)
		return &DatahubV1alpha1.ListGpuMetricsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	gpuMetrics := make([]*DatahubV1alpha1.GpuMetric, 0)
	for _, metric := range metrics {
		gpuMetricExtended := daoGpuMetricExtended{metric}
		datahubGpuMetric := gpuMetricExtended.datahubGpuMetric()
		gpuMetrics = append(gpuMetrics, datahubGpuMetric)
	}

	response := &DatahubV1alpha1.ListGpuMetricsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		GpuMetrics: gpuMetrics,
	}

	return response, nil
}

func (s *ServiceV1alpha1) ListGpuPredictions(ctx context.Context, in *DatahubV1alpha1.ListGpuPredictionsRequest) (*DatahubV1alpha1.ListGpuPredictionsResponse, error) {
	return &DatahubV1alpha1.ListGpuPredictionsResponse{}, nil
}

func (s *ServiceV1alpha1) CreateGpuPredictions(ctx context.Context, in *DatahubV1alpha1.CreateGpuPredictionsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateGpuPredictions grpc function: " + AlamedaUtils.InterfaceToString(in))

	requestExtended := CreateGpuPredictionsRequestExtended{*in}
	if requestExtended.validate() != nil {
		return &status.Status{
			Code: int32(code.Code_INVALID_ARGUMENT),
		}, nil
	}

	predictionDAO := DaoGpu.NewPredictionWithConfig(*s.Config.InfluxDB)
	err := predictionDAO.CreatePredictions(requestExtended.GpuPredictions())
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
