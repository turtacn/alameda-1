package grpc

import (
	"errors"
	"fmt"
	"regexp"

	v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/operator"
	"github.com/golang/protobuf/ptypes"
)

func ValidateListMetricsRequest(req *v1alpha1.ListMetricsRequest) error {

	// Validate Conditions Key and Value
	// Forbit value of key and value contain charater "
	reg, _ := regexp.Compile(`(\\)*\"`)
	for _, ls := range req.GetConditions() {
		k := ls.GetKey()
		v := ls.GetValue()
		if reg.MatchString(k) {
			return errors.New(fmt.Sprintf("Validate: Condition Key \"%s\" is invalid", k))
		}
		if reg.MatchString(v) {
			return errors.New(fmt.Sprintf("Validate: Condition Value \"%s\" is invalid", v))
		}
	}

	// Validate Conditions Key and Value
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
			return errors.New("Validate: must provide both \"start_time\",\"end_time\" and \"step\"")
		}
		_, err := ptypes.Duration(req.GetTimeRange().GetStep())
		if err != nil {
			return errors.New("Validate: " + err.Error())
		}
	}

	return nil
}
