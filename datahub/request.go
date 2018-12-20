package datahub

import (
	"errors"
	"fmt"
	"time"

	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metrics"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes"
)

type LabelSelector struct {
	datahub_v1alpha1.LabelSelector
}

type LabelSelectors []LabelSelector

func (lss LabelSelectors) MetricsLabelSelectors() []metrics.LabelSelector {

	var (
		convertedLSS = make([]metrics.LabelSelector, 0)
	)

	for _, labelSelector := range lss {

		k := labelSelector.GetKey()
		v := labelSelector.GetValue()
		var op metrics.StringOperator
		switch labelSelector.GetOp() {
		case datahub_v1alpha1.StrOp_Equal:
			op = metrics.StringOperatorEqueal
		case datahub_v1alpha1.StrOp_NotEqual:
			op = metrics.StringOperatorNotEqueal
		}

		convertedLSS = append(convertedLSS, metrics.LabelSelector{Key: k, Op: op, Value: v})
	}

	return convertedLSS
}

type ListContainerMetricsRequest struct {
	datahub_v1alpha1.ListContainerMetricsRequest
}

func (req ListContainerMetricsRequest) Validate() error {

	var (
		err error

		availableLabelSelectorKeys metrics.LabelSelectorKeys
	)

	if req.GetMetricType() == datahub_v1alpha1.ContainerMetricType_CONTAINER_METRICTYPE_UNDEFINED {
		return errors.New(fmt.Sprintf(`validate failed: metric type not support "%s"`, datahub_v1alpha1.ContainerMetricType_name[int32(req.GetMetricType())]))
	}

	availableLabelSelectorKeys = req.getAvailableLabelSelectorKeysByMetricType()
	err = req.validateConditionKeysAreAvailable(availableLabelSelectorKeys)
	if err != nil {
		return errors.New("vaildate failed: " + err.Error())
	}

	return nil
}

func (req ListContainerMetricsRequest) validateConditionKeysAreAvailable(availableLabelSelectorKeys metrics.LabelSelectorKeys) error {

	for _, condition := range req.GetConditions() {
		key := condition.GetKey()
		if !availableLabelSelectorKeys.Contains(metrics.LabelSelectorKey(key)) {
			return errors.New(fmt.Sprintf(`invalid condition key: "%s"`, key))
		}
	}
	return nil
}

func (req ListContainerMetricsRequest) getAvailableLabelSelectorKeysByMetricType() metrics.LabelSelectorKeys {

	var (
		availableLabelSelectorKeys metrics.LabelSelectorKeys
	)

	switch req.GetMetricType() {
	case datahub_v1alpha1.ContainerMetricType_CONTAINER_CPU_USAGE_SECONDS_PERCENTAGE:
		availableLabelSelectorKeys = metrics.LabelSelectorKeysAvailableForMetrics[metrics.MetricTypeContainerCPUUsageSecondsPercentage]
	case datahub_v1alpha1.ContainerMetricType_CONTAINER_MEMORY_USAGE_BYTES:
		availableLabelSelectorKeys = metrics.LabelSelectorKeysAvailableForMetrics[metrics.MetricTypeContainerMemoryUsageBytes]
	default:
		availableLabelSelectorKeys = make(metrics.LabelSelectorKeys, 0)
	}

	return availableLabelSelectorKeys
}

func (req ListContainerMetricsRequest) MetricsQuery() metrics.Query {

	var (
		q   = metrics.Query{}
		lss = make(LabelSelectors, 0)
	)

	switch req.MetricType {
	case datahub_v1alpha1.ContainerMetricType_CONTAINER_CPU_USAGE_SECONDS_PERCENTAGE:
		q.Metric = metrics.MetricTypeContainerCPUUsageSecondsPercentage
	case datahub_v1alpha1.ContainerMetricType_CONTAINER_MEMORY_USAGE_BYTES:
		q.Metric = metrics.MetricTypeContainerMemoryUsageBytes
	}

	for _, condition := range req.GetConditions() {
		lss = append(lss, LabelSelector{*condition})
	}
	q.LabelSelectors = lss.MetricsLabelSelectors()

	// assign difference type of time to query instance by type of gRPC request time
	switch req.TimeSelector.(type) {
	case nil:
		q.TimeSelector = nil
	case *datahub_v1alpha1.ListContainerMetricsRequest_Time:
		q.TimeSelector = &metrics.Timestamp{T: time.Unix(req.GetTime().GetSeconds(), int64(req.GetTime().GetNanos()))}
	case *datahub_v1alpha1.ListContainerMetricsRequest_Duration:
		d, _ := ptypes.Duration(req.GetDuration())
		q.TimeSelector = &metrics.Since{
			Duration: d,
		}
	case *datahub_v1alpha1.ListContainerMetricsRequest_TimeRange:
		startTime := req.GetTimeRange().GetStartTime()
		endTime := req.GetTimeRange().GetEndTime()
		step, _ := ptypes.Duration(req.GetTimeRange().GetStep())
		q.TimeSelector = &metrics.TimeRange{
			StartTime: time.Unix(startTime.GetSeconds(), int64(startTime.GetNanos())),
			EndTime:   time.Unix(endTime.GetSeconds(), int64(endTime.GetNanos())),
			Step:      step,
		}
	}

	return q
}

type ListNodeMetricsRequest struct {
	datahub_v1alpha1.ListNodeMetricsRequest
}

func (req ListNodeMetricsRequest) Validate() error {

	var (
		err error

		availableLabelSelectorKeys metrics.LabelSelectorKeys
	)

	if req.GetMetricType() == datahub_v1alpha1.NodeMetricType_NODE_METRICTYPE_UNDEFINED {
		return errors.New(fmt.Sprintf(`validate failed: metric type not support "%s"`, datahub_v1alpha1.NodeMetricType_name[int32(req.GetMetricType())]))
	}

	availableLabelSelectorKeys = req.getAvailableLabelSelectorKeysByMetricType()
	err = req.validateConditionKeysAreAvailable(availableLabelSelectorKeys)
	if err != nil {
		return errors.New("vaildate failed: " + err.Error())
	}

	return nil
}

func (req ListNodeMetricsRequest) validateConditionKeysAreAvailable(availableLabelSelectorKeys metrics.LabelSelectorKeys) error {

	for _, condition := range req.GetConditions() {
		key := condition.GetKey()
		if !availableLabelSelectorKeys.Contains(metrics.LabelSelectorKey(key)) {
			return errors.New(fmt.Sprintf(`invalid condition key: "%s"`, key))
		}
	}
	return nil
}

func (req ListNodeMetricsRequest) getAvailableLabelSelectorKeysByMetricType() metrics.LabelSelectorKeys {

	var (
		availableLabelSelectorKeys metrics.LabelSelectorKeys
	)

	switch req.GetMetricType() {
	case datahub_v1alpha1.NodeMetricType_NODE_CPU_USAGE_SECONDS_PERCENTAGE:
		availableLabelSelectorKeys = metrics.LabelSelectorKeysAvailableForMetrics[metrics.MetricTypeNodeCPUUsageSecondsPercentage]
	case datahub_v1alpha1.NodeMetricType_NODE_MEMORY_USAGE_BYTES:
		availableLabelSelectorKeys = metrics.LabelSelectorKeysAvailableForMetrics[metrics.MetricTypeNodeMemoryUsageBytes]
	default:
		availableLabelSelectorKeys = make(metrics.LabelSelectorKeys, 0)
	}

	return availableLabelSelectorKeys
}

func (req ListNodeMetricsRequest) MetricsQuery() metrics.Query {

	var (
		q   = metrics.Query{}
		lss = make(LabelSelectors, 0)
	)

	switch req.MetricType {
	case datahub_v1alpha1.NodeMetricType_NODE_CPU_USAGE_SECONDS_PERCENTAGE:
		q.Metric = metrics.MetricTypeNodeCPUUsageSecondsPercentage
	case datahub_v1alpha1.NodeMetricType_NODE_MEMORY_USAGE_BYTES:
		q.Metric = metrics.MetricTypeNodeMemoryUsageBytes
	}

	for _, condition := range req.GetConditions() {
		lss = append(lss, LabelSelector{*condition})
	}
	q.LabelSelectors = lss.MetricsLabelSelectors()

	// assign difference type of time to query instance by type of gRPC request time
	switch req.TimeSelector.(type) {
	case nil:
		q.TimeSelector = nil
	case *datahub_v1alpha1.ListNodeMetricsRequest_Time:
		q.TimeSelector = &metrics.Timestamp{T: time.Unix(req.GetTime().GetSeconds(), int64(req.GetTime().GetNanos()))}
	case *datahub_v1alpha1.ListNodeMetricsRequest_Duration:
		d, _ := ptypes.Duration(req.GetDuration())
		q.TimeSelector = &metrics.Since{
			Duration: d,
		}
	case *datahub_v1alpha1.ListNodeMetricsRequest_TimeRange:
		startTime := req.GetTimeRange().GetStartTime()
		endTime := req.GetTimeRange().GetEndTime()
		step, _ := ptypes.Duration(req.GetTimeRange().GetStep())
		q.TimeSelector = &metrics.TimeRange{
			StartTime: time.Unix(startTime.GetSeconds(), int64(startTime.GetNanos())),
			EndTime:   time.Unix(endTime.GetSeconds(), int64(endTime.GetNanos())),
			Step:      step,
		}
	}

	return q
}
