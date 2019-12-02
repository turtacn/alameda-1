package v1alpha1

import (
	DaoMetrics "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics"
	FormatRequest "github.com/containers-ai/alameda/datahub/pkg/formatconversion/requests"
	FormatResponse "github.com/containers-ai/alameda/datahub/pkg/formatconversion/responses"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiMetrics "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/metrics"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (s *ServiceV1alpha1) CreateClusterMetrics(ctx context.Context, in *ApiMetrics.CreateClusterMetricsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateClusterMetrics grpc function: " + AlamedaUtils.InterfaceToString(in))

	requestExtended := FormatRequest.CreateClusterMetricsRequestExtended{CreateClusterMetricsRequest: *in}
	if err := requestExtended.Validate(); err != nil {
		return &status.Status{
			Code:    int32(code.Code_INVALID_ARGUMENT),
			Message: err.Error(),
		}, nil
	}

	metricDAO := DaoMetrics.NewClusterMetricsWriterDAO(*s.Config)
	err := metricDAO.CreateMetrics(ctx, requestExtended.ProduceMetrics())
	if err != nil {
		scope.Errorf("failed to create cluster metrics: %+v", err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *ServiceV1alpha1) ListClusterMetrics(ctx context.Context, in *ApiMetrics.ListClusterMetricsRequest) (*ApiMetrics.ListClusterMetricsResponse, error) {
	scope.Debug("Request received from ListClusterMetrics grpc function: " + AlamedaUtils.InterfaceToString(in))

	requestExtended := FormatRequest.ListClusterMetricsRequestExtended{Request: in}
	if err := requestExtended.Validate(); err != nil {
		return &ApiMetrics.ListClusterMetricsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INVALID_ARGUMENT),
				Message: err.Error(),
			},
		}, nil
	}
	requestExtended.SetDefaultWithMetricsDBType(s.Config.Apis.Metrics.Source)

	metricsDao := DaoMetrics.NewClusterMetricsReaderDAO(*s.Config)
	metricMap, err := metricsDao.ListMetrics(ctx, requestExtended.ProduceRequest())
	if err != nil {
		return &ApiMetrics.ListClusterMetricsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}
	i := 0
	datahubClusterMetrics := make([]*ApiMetrics.ClusterMetric, len(metricMap.MetricMap))
	for _, metric := range metricMap.MetricMap {
		m := FormatResponse.ClusterMetricExtended{ClusterMetric: *metric}.ProduceMetrics()
		datahubClusterMetrics[i] = &m
		i++
	}

	return &ApiMetrics.ListClusterMetricsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		ClusterMetrics: datahubClusterMetrics,
	}, nil
}
