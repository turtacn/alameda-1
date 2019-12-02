package v1alpha1

import (
	DaoMetric "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics"
	FormatRequest "github.com/containers-ai/alameda/datahub/pkg/formatconversion/requests"
	FormatResponse "github.com/containers-ai/alameda/datahub/pkg/formatconversion/responses"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiMetrics "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/metrics"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (s *ServiceV1alpha1) CreateNodeMetrics(ctx context.Context, in *ApiMetrics.CreateNodeMetricsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateNodeMetrics grpc function: " + AlamedaUtils.InterfaceToString(in))

	requestExtended := FormatRequest.CreateNodeMetricsRequestExtended{CreateNodeMetricsRequest: *in}
	if requestExtended.Validate() != nil {
		return &status.Status{
			Code: int32(code.Code_INVALID_ARGUMENT),
		}, nil
	}

	metricDAO := DaoMetric.NewNodeMetricsWriterDAO(*s.Config)
	err := metricDAO.CreateMetrics(ctx, requestExtended.ProduceMetrics())
	if err != nil {
		scope.Errorf("failed to create node metrics: %+v", err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *ServiceV1alpha1) ListNodeMetrics(ctx context.Context, in *ApiMetrics.ListNodeMetricsRequest) (*ApiMetrics.ListNodeMetricsResponse, error) {
	scope.Debug("Request received from ListNodeMetrics grpc function: " + AlamedaUtils.InterfaceToString(in))

	requestExt := FormatRequest.ListNodeMetricsRequestExtended{Request: in}
	if err := requestExt.Validate(); err != nil {
		return &ApiMetrics.ListNodeMetricsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INVALID_ARGUMENT),
				Message: err.Error(),
			},
		}, nil
	}
	requestExt.SetDefaultWithMetricsDBType(s.Config.Apis.Metrics.Source)

	metricDAO := DaoMetric.NewNodeMetricsReaderDAO(*s.Config)
	nodesMetricMap, err := metricDAO.ListMetrics(ctx, requestExt.ProduceRequest())
	if err != nil {
		scope.Errorf("ListNodeMetrics failed: %+v", err)
		return &ApiMetrics.ListNodeMetricsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	datahubNodeMetrics := make([]*ApiMetrics.NodeMetric, 0)
	for _, nodeMetric := range nodesMetricMap.MetricMap {
		nodeMetricExtended := FormatResponse.NodeMetricExtended{NodeMetric: nodeMetric}
		datahubNodeMetric := nodeMetricExtended.ProduceMetrics()
		datahubNodeMetrics = append(datahubNodeMetrics, datahubNodeMetric)
	}

	return &ApiMetrics.ListNodeMetricsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		NodeMetrics: datahubNodeMetrics,
	}, nil
}
