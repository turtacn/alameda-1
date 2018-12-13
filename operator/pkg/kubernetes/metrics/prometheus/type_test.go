package prometheus

import (
	"testing"
	"time"

	"github.com/containers-ai/alameda/operator/pkg/kubernetes/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransformToMetricsQueryResponse(t *testing.T) {

	var (
		assert = assert.New(t)

		timestamp = 1435781430

		tests = []struct {
			have Response
			want metrics.QueryResponse
		}{
			{
				have: Response{
					metricType: metrics.MetricTypeContainerCPUUsageTotal,
					Status:     "success",
					Data: Data{
						ResultType: VectorResultType,
						Result: []interface{}{
							map[string]interface{}{
								"metric": map[string]string{
									"container_name": "prometheus",
									"cpu":            "total",
									"namespace":      "openshift-monitoring",
									"pod_name":       "prometheus-k8s-0",
								},
								"value": Value{
									float64(1435781430),
									"150556.78898869",
								},
							},
						},
					},
				},
				want: metrics.QueryResponse{
					Metric: metrics.MetricTypeContainerCPUUsageTotal,
					Results: []metrics.Data{
						metrics.Data{
							Labels: map[string]string{
								"container_name": "prometheus",
								"cpu":            "total",
								"namespace":      "openshift-monitoring",
								"pod_name":       "prometheus-k8s-0",
							},
							Samples: []metrics.Sample{
								metrics.Sample{
									Time:  time.Unix(int64(timestamp), int64(0)),
									Value: float64(150556.78898869),
								},
							},
						},
					},
				},
			},
			{
				have: Response{
					metricType: metrics.MetricTypeNodeMemoryUsageBytes,
					Status:     "success",
					Data: Data{
						ResultType: VectorResultType,
						Result: []interface{}{
							map[string]interface{}{
								"metric": map[string]string{
									"node": "localhost",
								},
								"value": Value{
									float64(1435781430),
									"150556",
								},
							},
						},
					},
				},
				want: metrics.QueryResponse{
					Metric: metrics.MetricTypeNodeMemoryUsageBytes,
					Results: []metrics.Data{
						metrics.Data{
							Labels: map[string]string{
								"node_name": "localhost",
							},
							Samples: []metrics.Sample{
								metrics.Sample{
									Time:  time.Unix(int64(timestamp), int64(0)),
									Value: float64(150556),
								},
							},
						},
					},
				},
			},
		}
	)

	for _, tt := range tests {
		get, err := tt.have.transformToMetricsQueryResponse()
		require.NoError(t, err)
		assert.Equal(tt.want, get)
	}
}
