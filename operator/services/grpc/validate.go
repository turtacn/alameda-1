package grpc

import (
	"errors"

	v1alpha1 "github.com/containers-ai/api/operator/v1alpha1"
	"github.com/golang/protobuf/ptypes"
)

func ValidateListMetricsRequest(req *v1alpha1.ListMetricsRequest) error {

	switch req.TimeSelector.(type) {
	case nil:
	case *v1alpha1.ListMetricsRequest_Time:
	case *v1alpha1.ListMetricsRequest_Duration:
		_, err := ptypes.Duration(req.GetDuration())
		if err != nil {
			return err
		}
	case *v1alpha1.ListMetricsRequest_TimeRange:
		if req.GetTimeRange().GetStartTime() == nil || req.GetTimeRange().GetEndTime() == nil || req.GetTimeRange().GetStep() == nil {
			return errors.New("must provide both \"start_time\",\"end_time\" and \"step\"")
		}
		_, err := ptypes.Duration(req.GetTimeRange().GetStep())
		if err != nil {
			return err
		}
	}

	return nil
}
