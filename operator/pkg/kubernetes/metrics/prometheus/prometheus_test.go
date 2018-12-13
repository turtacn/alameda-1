package prometheus

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/containers-ai/alameda/operator/pkg/kubernetes/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	mux    *http.ServeMux
	server *httptest.Server
	prom   *prometheus
)

func setup() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	cfg := Config{
		URL: server.URL,
	}

	p := New(cfg)
	prom = p.(*prometheus)
}

func TestQuery(t *testing.T) {

	setup()

	var (
		timestamp  = 1435781430
		timestamps = []int{1543286478, 1543286508}

		tests = []struct {
			have metrics.Query
			want metrics.QueryResponse
		}{
			{
				have: metrics.Query{
					Metric: metrics.MetricTypeContainerCPUUsageTotal,
					LabelSelectors: []metrics.LabelSelector{
						metrics.LabelSelector{Key: "namespace", Op: metrics.StringOperatorEqueal, Value: "default"},
						metrics.LabelSelector{Key: "pod_name", Op: metrics.StringOperatorEqueal, Value: "docker-registry-1-mbjnw"},
					},
				},
				want: metrics.QueryResponse{
					Metric: metrics.MetricTypeContainerCPUUsageTotal,
					Results: []metrics.Data{
						metrics.Data{
							Labels: map[string]string{
								"namespace": "default",
								"pod_name":  "docker-registry-1-mbjnw",
							},
							Samples: []metrics.Sample{
								metrics.Sample{
									Time:  time.Unix(int64(timestamp), int64(0)),
									Value: float64(101.1),
								},
							},
						},
					},
				},
			},
			{
				have: metrics.Query{
					Metric: metrics.MetricTypeContainerCPUUsageTotal,
					LabelSelectors: []metrics.LabelSelector{
						metrics.LabelSelector{Key: "namespace", Op: metrics.StringOperatorEqueal, Value: "default"},
						metrics.LabelSelector{Key: "pod_name", Op: metrics.StringOperatorEqueal, Value: "docker-registry-1-mbjnw"},
					},
					TimeSelector: &metrics.Since{Duration: 60 * time.Second},
				},
				want: metrics.QueryResponse{
					Metric: metrics.MetricTypeContainerCPUUsageTotal,
					Results: []metrics.Data{
						metrics.Data{
							Labels: map[string]string{
								"namespace": "default",
								"pod_name":  "docker-registry-1-mbjnw",
							},
							Samples: []metrics.Sample{
								metrics.Sample{
									Time:  time.Unix(int64(timestamps[0]), int64(0)),
									Value: float64(3121.990940488),
								},
								metrics.Sample{
									Time:  time.Unix(int64(timestamps[1]), int64(0)),
									Value: float64(3122.026482446),
								},
							},
						},
					},
				},
			},
			{
				have: metrics.Query{
					Metric: metrics.MetricTypeNodeCPUUsageSecondsAvg1M,
					LabelSelectors: []metrics.LabelSelector{
						metrics.LabelSelector{Key: "node_name", Op: metrics.StringOperatorEqueal, Value: "ip-172-31-17-113.us-west-2.compute.internal"},
					},
				},
				want: metrics.QueryResponse{
					Metric: metrics.MetricTypeNodeCPUUsageSecondsAvg1M,
					Results: []metrics.Data{
						metrics.Data{
							Labels: map[string]string{
								"node_name": "ip-172-31-17-113.us-west-2.compute.internal",
							},
							Samples: []metrics.Sample{
								metrics.Sample{
									Time:  time.Unix(int64(timestamp), int64(0)),
									Value: float64(0.023583333330073675),
								},
							},
						},
					},
				},
			},
			{
				have: metrics.Query{
					Metric: metrics.MetricTypeNodeMemoryUsageBytes,
					LabelSelectors: []metrics.LabelSelector{
						metrics.LabelSelector{Key: "node_name", Op: metrics.StringOperatorEqueal, Value: "ip-172-31-17-113.us-west-2.compute.internal"},
					},
				},
				want: metrics.QueryResponse{
					Metric: metrics.MetricTypeNodeMemoryUsageBytes,
					Results: []metrics.Data{
						metrics.Data{
							Labels: map[string]string{
								"node_name": "ip-172-31-17-113.us-west-2.compute.internal",
							},
							Samples: []metrics.Sample{
								metrics.Sample{
									Time:  time.Unix(int64(timestamp), int64(0)),
									Value: float64(6858416128),
								},
							},
						},
					},
				},
			},
		}

		mockPrometheusResponses = []Response{
			Response{
				Status: "success",
				Data: Data{
					ResultType: VectorResultType,
					Result: []interface{}{
						VectorResult{
							Metric: map[string]string{
								"namespace": "default",
								"pod_name":  "docker-registry-1-mbjnw",
							},
							Value: []interface{}{
								float64(timestamp),
								"101.1",
							},
						},
					},
				},
			},
			Response{
				Status: "success",
				Data: Data{
					ResultType: MatrixResultType,
					Result: []interface{}{
						MatrixResult{
							Metric: map[string]string{
								"namespace": "default",
								"pod_name":  "docker-registry-1-mbjnw",
							},
							Values: []Value{
								[]interface{}{float64(timestamps[0]), "3121.990940488"},
								[]interface{}{float64(timestamps[1]), "3122.026482446"},
							},
						},
					},
				},
			},
			Response{
				metricType: metrics.MetricTypeNodeCPUUsageSecondsAvg1M,
				Status:     "success",
				Data: Data{
					ResultType: VectorResultType,
					Result: []interface{}{
						VectorResult{
							Metric: map[string]string{
								"node": "ip-172-31-17-113.us-west-2.compute.internal",
							},
							Value: []interface{}{
								float64(timestamp),
								"0.023583333330073675",
							},
						},
					},
				},
			},
			Response{
				metricType: metrics.MetricTypeNodeMemoryUsageBytes,
				Status:     "success",
				Data: Data{
					ResultType: VectorResultType,
					Result: []interface{}{
						VectorResult{
							Metric: map[string]string{
								"node_name": "ip-172-31-17-113.us-west-2.compute.internal",
							},
							Value: []interface{}{
								float64(timestamp),
								"6858416128",
							},
						},
					},
				},
			},
		}

		testIndex = 0
	)

	mux.HandleFunc(apiPrefix+epQuery, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		mockResponse := mockPrometheusResponses[testIndex]
		json.NewEncoder(w).Encode(mockResponse)
	})

	for i, test := range tests {

		testIndex = i
		resp, err := prom.Query(test.have)

		assert := assert.New(t)
		require.Nil(t, err)
		assert.Equal(test.want, resp)
	}
	server.Close()
}
