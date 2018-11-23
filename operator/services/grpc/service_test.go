package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/containers-ai/alameda/operator/pkg/kubernetes/metrics"
	"github.com/containers-ai/alameda/operator/pkg/kubernetes/metrics/mocks"
	operator_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/operator"
	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

var (
	currentTime      = time.Unix(currentPtypeTime.GetSeconds(), int64(currentPtypeTime.GetNanos()))
	currentPtypeTime = ptypes.TimestampNow()
	oldTime          = currentTime.Add(time.Minute * (-1))
	oldPtypeTime, _  = ptypes.TimestampProto(oldTime)
	d, _             = time.ParseDuration("300s")
	dPtype           = ptypes.DurationProto(d)

	tests = []struct {
		have *operator_v1alpha1.ListMetricsRequest
		want *operator_v1alpha1.ListMetricsResponse
	}{
		{
			have: &operator_v1alpha1.ListMetricsRequest{
				MetricType: operator_v1alpha1.MetricType_CONTAINER_CPU_USAGE_TOTAL,
				Conditions: []*operator_v1alpha1.LabelSelector{
					&operator_v1alpha1.LabelSelector{
						Key:   "namespace",
						Op:    operator_v1alpha1.StrOp_Equal,
						Value: "test1",
					},
				},
			},
			want: &operator_v1alpha1.ListMetricsResponse{
				Metrics: []*operator_v1alpha1.MetricResult{
					&operator_v1alpha1.MetricResult{
						Labels: map[string]string{
							"namespace": "test1",
						},
						Samples: []*operator_v1alpha1.Sample{
							&operator_v1alpha1.Sample{
								Time:  currentPtypeTime,
								Value: 0.026015485,
							},
						},
					},
				},
				Status: &status.Status{
					Code: int32(code.Code_OK),
				},
			},
		},
		{
			have: &operator_v1alpha1.ListMetricsRequest{
				MetricType: operator_v1alpha1.MetricType_CONTAINER_CPU_USAGE_TOTAL,
				Conditions: []*operator_v1alpha1.LabelSelector{
					&operator_v1alpha1.LabelSelector{
						Key:   "namespace",
						Op:    operator_v1alpha1.StrOp_Equal,
						Value: "test2",
					},
				},
				TimeSelector: &operator_v1alpha1.ListMetricsRequest_Duration{
					Duration: dPtype,
				},
			},
			want: &operator_v1alpha1.ListMetricsResponse{
				Metrics: []*operator_v1alpha1.MetricResult{
					&operator_v1alpha1.MetricResult{
						Labels: map[string]string{
							"namespace": "test2",
						},
						Samples: []*operator_v1alpha1.Sample{
							&operator_v1alpha1.Sample{
								Time:  currentPtypeTime,
								Value: 0.026015485,
							},
							&operator_v1alpha1.Sample{
								Time:  oldPtypeTime,
								Value: 0.016015485,
							},
						},
					},
				},
				Status: &status.Status{
					Code: int32(code.Code_OK),
				},
			},
		},
		{
			have: &operator_v1alpha1.ListMetricsRequest{
				MetricType: operator_v1alpha1.MetricType_CONTAINER_CPU_USAGE_TOTAL,
				Conditions: []*operator_v1alpha1.LabelSelector{
					&operator_v1alpha1.LabelSelector{
						Key:   "namespace",
						Op:    operator_v1alpha1.StrOp_Equal,
						Value: "test3",
					},
				},
				TimeSelector: &operator_v1alpha1.ListMetricsRequest_Time{
					Time: currentPtypeTime,
				},
			},
			want: &operator_v1alpha1.ListMetricsResponse{
				Metrics: []*operator_v1alpha1.MetricResult{
					&operator_v1alpha1.MetricResult{
						Labels: map[string]string{
							"namespace": "test3",
						},
						Samples: []*operator_v1alpha1.Sample{
							&operator_v1alpha1.Sample{
								Time:  currentPtypeTime,
								Value: 0.026015485,
							},
						},
					},
				},
				Status: &status.Status{
					Code: int32(code.Code_OK),
				},
			},
		},
		{
			have: &operator_v1alpha1.ListMetricsRequest{
				MetricType: operator_v1alpha1.MetricType_CONTAINER_CPU_USAGE_TOTAL,
				Conditions: []*operator_v1alpha1.LabelSelector{
					&operator_v1alpha1.LabelSelector{
						Key:   "namespace",
						Op:    operator_v1alpha1.StrOp_Equal,
						Value: "test3",
					},
				},
				TimeSelector: &operator_v1alpha1.ListMetricsRequest_TimeRange{
					TimeRange: &operator_v1alpha1.TimeRange{
						StartTime: oldPtypeTime,
						EndTime:   currentPtypeTime,
						Step:      dPtype,
					},
				},
			},
			want: &operator_v1alpha1.ListMetricsResponse{
				Metrics: []*operator_v1alpha1.MetricResult{
					&operator_v1alpha1.MetricResult{
						Labels: map[string]string{
							"namespace": "test3",
						},
						Samples: []*operator_v1alpha1.Sample{
							&operator_v1alpha1.Sample{
								Time:  currentPtypeTime,
								Value: 0.026015485,
							},
							&operator_v1alpha1.Sample{
								Time:  oldPtypeTime,
								Value: 0.016015485,
							},
						},
					},
				},
				Status: &status.Status{
					Code: int32(code.Code_OK),
				},
			},
		},
		{
			have: &operator_v1alpha1.ListMetricsRequest{
				MetricType: operator_v1alpha1.MetricType_CONTAINER_CPU_USAGE_TOTAL,
				Conditions: []*operator_v1alpha1.LabelSelector{
					&operator_v1alpha1.LabelSelector{
						Key:   "namespace",
						Op:    operator_v1alpha1.StrOp_Equal,
						Value: "\"test\"",
					},
				},
			},
			want: &operator_v1alpha1.ListMetricsResponse{
				Status: &status.Status{
					Code:    int32(code.Code_INVALID_ARGUMENT),
					Message: "Validate: Condition Value \"\"test\"\" is invalid",
				},
			},
		},
	}

	mock = []struct {
		query    metrics.Query
		response metrics.QueryResponse
	}{
		{
			query: metrics.Query{
				Metric: metrics.MetricTypeContainerCPUUsageTotal,
				LabelSelectors: []metrics.LabelSelector{metrics.LabelSelector{
					Key:   "namespace",
					Op:    metrics.StringOperatorEqueal,
					Value: "test1",
				}},
			},
			response: metrics.QueryResponse{
				Results: []metrics.Data{
					metrics.Data{
						Labels: map[string]string{
							"namespace": "test1",
						},
						Samples: []metrics.Sample{
							metrics.Sample{
								Time:  currentTime,
								Value: 0.026015485,
							},
						},
					},
				},
			},
		},
		{
			query: metrics.Query{
				Metric: metrics.MetricTypeContainerCPUUsageTotal,
				LabelSelectors: []metrics.LabelSelector{metrics.LabelSelector{
					Key:   "namespace",
					Op:    metrics.StringOperatorEqueal,
					Value: "test2",
				}},
				TimeSelector: &metrics.Since{
					Duration: d,
				},
			},
			response: metrics.QueryResponse{
				Results: []metrics.Data{
					metrics.Data{
						Labels: map[string]string{
							"namespace": "test2",
						},
						Samples: []metrics.Sample{
							metrics.Sample{
								Time:  currentTime,
								Value: 0.026015485,
							},
							metrics.Sample{
								Time:  oldTime,
								Value: 0.016015485,
							},
						},
					},
				},
			},
		},
		{
			query: metrics.Query{
				Metric: metrics.MetricTypeContainerCPUUsageTotal,
				LabelSelectors: []metrics.LabelSelector{metrics.LabelSelector{
					Key:   "namespace",
					Op:    metrics.StringOperatorEqueal,
					Value: "test3",
				}},
				TimeSelector: &metrics.Timestamp{
					T: currentTime,
				},
			},
			response: metrics.QueryResponse{
				Results: []metrics.Data{
					metrics.Data{
						Labels: map[string]string{
							"namespace": "test3",
						},
						Samples: []metrics.Sample{
							metrics.Sample{
								Time:  currentTime,
								Value: 0.026015485,
							},
						},
					},
				},
			},
		},
		{
			query: metrics.Query{
				Metric: metrics.MetricTypeContainerCPUUsageTotal,
				LabelSelectors: []metrics.LabelSelector{metrics.LabelSelector{
					Key:   "namespace",
					Op:    metrics.StringOperatorEqueal,
					Value: "test3",
				}},
				TimeSelector: &metrics.TimeRange{
					StartTime: oldTime,
					EndTime:   currentTime,
					Step:      d,
				},
			},
			response: metrics.QueryResponse{
				Results: []metrics.Data{
					metrics.Data{
						Labels: map[string]string{
							"namespace": "test3",
						},
						Samples: []metrics.Sample{
							metrics.Sample{
								Time:  currentTime,
								Value: 0.026015485,
							},
							metrics.Sample{
								Time:  oldTime,
								Value: 0.016015485,
							},
						},
					},
				},
			},
		},
		{
			query: metrics.Query{
				Metric: metrics.MetricTypeContainerCPUUsageTotal,
				LabelSelectors: []metrics.LabelSelector{metrics.LabelSelector{
					Key:   "namespace",
					Op:    metrics.StringOperatorEqueal,
					Value: "\"test\"",
				}},
			},
			response: metrics.QueryResponse{
				Results: []metrics.Data{
					metrics.Data{
						Labels: map[string]string{
							"namespace": "\"test\"",
						},
						Samples: []metrics.Sample{
							metrics.Sample{
								Time:  currentTime,
								Value: 0.026015485,
							},
						},
					},
				},
			},
		},
	}
)

func TestListMetrics(t *testing.T) {

	ctl := gomock.NewController(t)
	defer ctl.Finish()
	mockDB := mocks.NewMockMetricsDB(ctl)
	s := Service{
		MetricsDB: mockDB,
	}

	assert := assert.New(t)
	for i, test := range tests {

		mockQuery := mock[i].query
		mockResponse := mock[i].response
		mockDB.EXPECT().Query(mockQuery).Return(mockResponse, nil).AnyTimes()

		resp, err := s.ListMetrics(context.Background(), test.have)
		require.NoError(t, err)
		assert.Equal(test.want, resp)
	}
}
