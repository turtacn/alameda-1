package datahub

import (
	"strconv"

	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metrics"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes"
)

type MetricsQueryResponse struct {
	metrics.QueryResponse
}

func (resp MetricsQueryResponse) ListContainerMetricsResponse() (datahub_v1alpha1.ListContainerMetricsResponse, error) {

	var (
		convertedResp = datahub_v1alpha1.ListContainerMetricsResponse{}
	)

	convertedResp.Metrics = make([]*datahub_v1alpha1.MetricResult, 0)
	for _, result := range resp.Results {
		series := &datahub_v1alpha1.MetricResult{}

		series.Labels = result.Labels
		for _, sample := range result.Samples {
			s := &datahub_v1alpha1.Sample{}

			timestampProto, err := ptypes.TimestampProto(sample.Time)
			if err != nil {
				scope.Error("convert time.Time to google.protobuf.Timestamp failed")
				return convertedResp, err
			}
			s.Time = timestampProto
			s.NumValue = strconv.FormatFloat(sample.Value, 'f', -1, 64)
			series.Samples = append(series.Samples, s)
		}
		convertedResp.Metrics = append(convertedResp.Metrics, series)
	}

	return convertedResp, nil
}

func (resp MetricsQueryResponse) ListNodeMetricsResponse() (datahub_v1alpha1.ListNodeMetricsResponse, error) {

	var (
		convertedResp = datahub_v1alpha1.ListNodeMetricsResponse{}
	)

	convertedResp.Metrics = make([]*datahub_v1alpha1.MetricResult, 0)
	for _, result := range resp.Results {
		series := &datahub_v1alpha1.MetricResult{}

		series.Labels = result.Labels
		for _, sample := range result.Samples {
			s := &datahub_v1alpha1.Sample{}

			timestampProto, err := ptypes.TimestampProto(sample.Time)
			if err != nil {
				scope.Error("convert time.Time to google.protobuf.Timestamp failed")
				return convertedResp, err
			}
			s.Time = timestampProto
			s.NumValue = strconv.FormatFloat(sample.Value, 'f', -1, 64)
			series.Samples = append(series.Samples, s)
		}
		convertedResp.Metrics = append(convertedResp.Metrics, series)
	}

	return convertedResp, nil
}
