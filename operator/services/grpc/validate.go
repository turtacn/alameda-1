package grpc

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/containers-ai/alameda/operator/pkg/kubernetes/metrics"
	v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/operator"
	"github.com/golang/protobuf/ptypes"
)

func ValidateListMetricsRequest(req v1alpha1.ListMetricsRequest) error {

	var err error

	err = validateListMetricsRequestConditions(req)
	if err != nil {
		return err
	}

	err = validateListMetricsRequestTimeSelector(req)
	if err != nil {
		return err
	}

	return nil
}

func validateListMetricsRequestConditions(req v1alpha1.ListMetricsRequest) error {

	var err error

	err = validatelistMetricsRequestConditionKeyValue(req)
	if err != nil {
		return err
	}

	return nil
}

func validatelistMetricsRequestConditionKeyValue(req v1alpha1.ListMetricsRequest) error {

	reg, _ := regexp.Compile(`(\\)*\"`)
	for _, ls := range req.Conditions {
		k := ls.GetKey()
		v := ls.GetValue()
		if reg.MatchString(k) {
			return errors.New(fmt.Sprintf("Validate: Condition Key \"%s\" is invalid", k))
		}
		if !islistMetricsRequestConditionKeyAvailable(req.MetricType, k) {
			return errors.New(fmt.Sprintf("Validate: Condition Key \"%s\" is not available in this metric", k))
		}
		if reg.MatchString(v) {
			return errors.New(fmt.Sprintf("Validate: Condition Value \"%s\" is invalid", v))
		}
	}

	return nil
}

func islistMetricsRequestConditionKeyAvailable(metirc v1alpha1.MetricType, conditionKey string) bool {

	var isAvailable = false

	availableKeys := metrics.LabelSelectorKeysAvailableForMetrics[metrics.MetricType(int(metirc))]
	for _, avaliableKey := range availableKeys {
		if conditionKey == string(avaliableKey) {
			return true
		}
	}

	return isAvailable
}

func validateListMetricsRequestTimeSelector(req v1alpha1.ListMetricsRequest) error {

	switch req.TimeSelector.(type) {
	case nil:
	case *v1alpha1.ListMetricsRequest_Time:
	case *v1alpha1.ListMetricsRequest_Duration:
		_, err := ptypes.Duration(req.GetDuration())
		if err != nil {
			return err
		}
	case *v1alpha1.ListMetricsRequest_TimeRange:
		start := req.GetTimeRange().GetStartTime()
		end := req.GetTimeRange().GetEndTime()
		if start == nil || end == nil || req.GetTimeRange().GetStep() == nil {
			return errors.New("Validate: must provide both \"start_time\",\"end_time\" and \"step\"")
		}
		if start.Seconds >= end.Seconds && start.Nanos >= end.Nanos {
			return errors.New("Validate: \"start_time\" cannot greater than \"end_time\"")
		}
		_, err := ptypes.Duration(req.GetTimeRange().GetStep())
		if err != nil {
			return errors.New("Validate: " + err.Error())
		}
	}

	return nil
}
