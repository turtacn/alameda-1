package grpc

import (
	"errors"
	"testing"
	"time"

	operator_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/operator"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
)

func TestValidateListMetricsRequest(t *testing.T) {

	assert := assert.New(t)

	start := time.Now()
	end := start.AddDate(1, 0, 0)
	startPtype, _ := ptypes.TimestampProto(start)
	endPtype, _ := ptypes.TimestampProto(end)
	d, _ := time.ParseDuration("30s")
	step := ptypes.DurationProto(d)

	tests := []struct {
		have *operator_v1alpha1.ListMetricsRequest
		want error
	}{
		{
			have: &operator_v1alpha1.ListMetricsRequest{},
			want: nil,
		},
		{
			have: &operator_v1alpha1.ListMetricsRequest{
				Conditions: []*operator_v1alpha1.LabelSelector{
					&operator_v1alpha1.LabelSelector{
						Key:   "namespace",
						Op:    operator_v1alpha1.StrOp_EQUAL,
						Value: "test",
					},
				},
			},
			want: nil,
		},
		{
			have: &operator_v1alpha1.ListMetricsRequest{
				Conditions: []*operator_v1alpha1.LabelSelector{
					&operator_v1alpha1.LabelSelector{
						Key:   `\"node_name`,
						Op:    operator_v1alpha1.StrOp_EQUAL,
						Value: "test",
					},
				},
			},
			want: errors.New(""),
		},
		{
			have: &operator_v1alpha1.ListMetricsRequest{
				Conditions: []*operator_v1alpha1.LabelSelector{
					&operator_v1alpha1.LabelSelector{
						Key:   `\\\\"test`,
						Op:    operator_v1alpha1.StrOp_EQUAL,
						Value: `test`,
					},
				},
			},
			want: errors.New(""),
		},
		{
			have: &operator_v1alpha1.ListMetricsRequest{
				Conditions: []*operator_v1alpha1.LabelSelector{
					&operator_v1alpha1.LabelSelector{
						Key:   `test`,
						Op:    operator_v1alpha1.StrOp_EQUAL,
						Value: `\"test`,
					},
				},
			},
			want: errors.New(""),
		},
		{
			have: &operator_v1alpha1.ListMetricsRequest{
				Conditions: []*operator_v1alpha1.LabelSelector{
					&operator_v1alpha1.LabelSelector{
						Key:   `test`,
						Op:    operator_v1alpha1.StrOp_EQUAL,
						Value: `\\\\"test`,
					},
				},
			},
			want: errors.New(""),
		},
		{
			have: &operator_v1alpha1.ListMetricsRequest{
				TimeSelector: &operator_v1alpha1.ListMetricsRequest_Time{
					Time: ptypes.TimestampNow(),
				},
			},
			want: nil,
		},
		{
			have: &operator_v1alpha1.ListMetricsRequest{
				TimeSelector: &operator_v1alpha1.ListMetricsRequest_TimeRange{
					TimeRange: &operator_v1alpha1.TimeRange{},
				},
			},
			want: errors.New(""),
		},
		{
			have: &operator_v1alpha1.ListMetricsRequest{
				TimeSelector: &operator_v1alpha1.ListMetricsRequest_TimeRange{
					TimeRange: &operator_v1alpha1.TimeRange{
						StartTime: startPtype,
						EndTime:   endPtype,
						Step:      step,
					},
				},
			},
			want: nil,
		},
		{
			have: &operator_v1alpha1.ListMetricsRequest{
				TimeSelector: &operator_v1alpha1.ListMetricsRequest_TimeRange{
					TimeRange: &operator_v1alpha1.TimeRange{
						StartTime: endPtype,
						EndTime:   startPtype,
						Step:      step,
					},
				},
			},
			want: errors.New(""),
		},
	}

	for _, test := range tests {
		err := ValidateListMetricsRequest(*test.have)
		assert.Condition(func() bool {
			success := false
			if (test.want == nil && err == nil) || (test.want != nil && err != nil) {
				success = true
			}
			return success
		}, "test %s  ", test.have.String())
	}
}
