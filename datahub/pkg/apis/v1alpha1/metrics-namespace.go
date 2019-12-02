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

func (s *ServiceV1alpha1) CreateNamespaceMetrics(ctx context.Context, in *ApiMetrics.CreateNamespaceMetricsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateNamespaceMetrics grpc function: " + AlamedaUtils.InterfaceToString(in))

	requestExtended := FormatRequest.CreateNamespaceMetricsRequestExtended{CreateNamespaceMetricsRequest: *in}
	if err := requestExtended.Validate(); err != nil {
		return &status.Status{
			Code:    int32(code.Code_INVALID_ARGUMENT),
			Message: err.Error(),
		}, nil
	}

	metricDAO := DaoMetrics.NewNamespaceMetricsWriterDAO(*s.Config)
	err := metricDAO.CreateMetrics(ctx, requestExtended.ProduceMetrics())
	if err != nil {
		scope.Errorf("failed to create namespace metrics: %+v", err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *ServiceV1alpha1) ListNamespaceMetrics(ctx context.Context, in *ApiMetrics.ListNamespaceMetricsRequest) (*ApiMetrics.ListNamespaceMetricsResponse, error) {
	scope.Debug("Request received from ListNamespaceMetrics grpc function: " + AlamedaUtils.InterfaceToString(in))

	requestExtended := FormatRequest.ListNamespaceMetricsRequestExtended{Request: in}
	if err := requestExtended.Validate(); err != nil {
		return &ApiMetrics.ListNamespaceMetricsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INVALID_ARGUMENT),
				Message: err.Error(),
			},
		}, nil
	}
	requestExtended.SetDefaultWithMetricsDBType(s.Config.Apis.Metrics.Source)

	metricsDao := DaoMetrics.NewNamespaceMetricsReaderDAO(*s.Config)
	metricMap, err := metricsDao.ListMetrics(ctx, requestExtended.ProduceRequest())
	if err != nil {
		return &ApiMetrics.ListNamespaceMetricsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}
	i := 0
	datahubNamespaceMetrics := make([]*ApiMetrics.NamespaceMetric, len(metricMap.MetricMap))
	for _, metric := range metricMap.MetricMap {
		m := FormatResponse.NamespaceMetricExtended{NamespaceMetric: *metric}.ProduceMetrics()
		datahubNamespaceMetrics[i] = &m
		i++
	}

	return &ApiMetrics.ListNamespaceMetricsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		NamespaceMetrics: datahubNamespaceMetrics,
	}, nil
}
