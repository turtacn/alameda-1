package prometheus

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/containers-ai/alameda/operator/pkg/kubernetes/metrics"
	"github.com/containers-ai/alameda/operator/pkg/kubernetes/metrics/prometheus/containerCPUUsageTotal"
	"github.com/containers-ai/alameda/operator/pkg/kubernetes/metrics/prometheus/containerCPUUsageTotalRate"
	"github.com/containers-ai/alameda/operator/pkg/kubernetes/metrics/prometheus/containerMemoryUsage"
	"github.com/containers-ai/alameda/operator/pkg/kubernetes/metrics/prometheus/nodeCPUUsageSecondsAVG1M"
	"github.com/containers-ai/alameda/operator/pkg/kubernetes/metrics/prometheus/nodeMemoryUsageBytes"
	"github.com/containers-ai/alameda/operator/pkg/utils"
)

const (
	statusError = "error"
)

type Response struct {
	Status    string `json:"status"`
	Data      Data   `json:"data"`
	ErrorType string `json:"errorType"`
	Error     string `json:"error"`

	metricType metrics.MetricType
}

func (r *Response) setMetricType(t metrics.MetricType) {
	r.metricType = t
}

func (r Response) transformToMetricsQueryResponse() (metrics.QueryResponse, error) {

	var (
		err           error
		queryResponse metrics.QueryResponse
	)

	queryResponse, err = r.transformResponseByResultType(r.Data.ResultType)
	if err != nil {
		return queryResponse, errors.New("transform prometheus.Response to metircs.QueryResponse failed: " + err.Error())
	}

	queryResponse, err = transformResultsPrometheusLabelIntoLabelSelectorKeyByMetricType(queryResponse)
	if err != nil {
		return queryResponse, errors.New("transform prometheus.Response to metircs.QueryResponse failed: " + err.Error())
	}

	return queryResponse, nil
}

func (r Response) transformResponseByResultType(resultType ResultType) (metrics.QueryResponse, error) {

	var (
		queryResponse = initMetircQueryResponseForTransform(r.metricType)
	)

	switch resultType {
	case "matrix":
		for _, r := range r.Data.Result {

			result := MatrixResult{}
			if _, ok := r.(map[string]interface{}); !ok {
				return queryResponse, fmt.Errorf("error while building sample, cannot convert type %s to map[string]interface{}", reflect.TypeOf(r).String())
			}
			resultStr, err := json.Marshal(r.(map[string]interface{}))
			if err != nil {
				return queryResponse, err
			}
			err = json.Unmarshal(resultStr, &result)
			if err != nil {
				return queryResponse, err
			}

			serie := metrics.Data{}
			serie.Labels = result.Metric
			for _, value := range result.Values {

				if _, ok := value[0].(float64); !ok {
					return queryResponse, fmt.Errorf("error while building sample, cannot convert type %s to float64", reflect.TypeOf(value[0]))
				}
				unixTime := time.Unix(int64(value[0].(float64)), 0)

				if _, ok := value[1].(string); !ok {
					return queryResponse, fmt.Errorf("error while building sample, cannot convert type %s to string", reflect.TypeOf(value[1]))
				}
				sampleValue, err := strconv.ParseFloat(value[1].(string), 64)
				if err != nil {
					return queryResponse, err
				}

				sample := metrics.Sample{
					Time:  unixTime,
					Value: sampleValue,
				}
				serie.Samples = append(serie.Samples, sample)
			}
			queryResponse.Results = append(queryResponse.Results, serie)
		}
	case "vector":
		for _, r := range r.Data.Result {

			result := VectorResult{}

			if _, ok := r.(map[string]interface{}); !ok {
				return queryResponse, fmt.Errorf("error while building sample, cannot convert type %s to map[string]interface{}", reflect.TypeOf(r).String())
			}
			resultStr, err := json.Marshal(r.(map[string]interface{}))
			if err != nil {
				return queryResponse, err
			}
			err = json.Unmarshal(resultStr, &result)
			if err != nil {
				return queryResponse, err
			}

			serie := metrics.Data{}
			serie.Labels = result.Metric
			value := result.Value

			if _, ok := value[0].(float64); !ok {
				return queryResponse, fmt.Errorf("error while building sample, cannot convert type %s to float64", reflect.TypeOf(value[0]))
			}
			unixTime := time.Unix(int64(value[0].(float64)), 0)

			if _, ok := value[1].(string); !ok {
				return queryResponse, fmt.Errorf("error while building sample, cannot convert %+v(type %s) to string", value[1], reflect.TypeOf(value[1]))
			}
			sampleValue, err := strconv.ParseFloat(value[1].(string), 64)
			if err != nil {
				return queryResponse, err
			}

			sample := metrics.Sample{
				Time:  unixTime,
				Value: sampleValue,
			}
			serie.Samples = append(serie.Samples, sample)

			queryResponse.Results = append(queryResponse.Results, serie)
		}
	case "scalar":
		return queryResponse, errors.New(fmt.Sprintf("not implement for resultType \"%s\"", resultType))
	case "string":
		return queryResponse, errors.New(fmt.Sprintf("not implement for resultType \"%s\"", resultType))
	default:
		return queryResponse, errors.New(fmt.Sprintf("not implement for resultType \"%s\"", resultType))
	}

	return queryResponse, nil
}

func initMetircQueryResponseForTransform(t metrics.MetricType) metrics.QueryResponse {
	return metrics.QueryResponse{
		Metric:  t,
		Results: make([]metrics.Data, 0),
	}
}

func transformResultsPrometheusLabelIntoLabelSelectorKeyByMetricType(q metrics.QueryResponse) (metrics.QueryResponse, error) {

	var (
		result = q

		labelsInPrometheus = make([]string, 0)
		labelSelectorKeys  = make([]string, 0)
	)

	switch q.Metric {
	case metrics.MetricTypeContainerCPUUsageTotal:
		labelsInPrometheus = containerCPUUsageTotal.LabelMapper.PrometheusLabels
		labelSelectorKeys = containerCPUUsageTotal.LabelMapper.LabelSelectorKeys
	case metrics.MetricTypeContainerCPUUsageTotalRate:
		labelsInPrometheus = containerCPUUsageTotalRate.LabelMapper.PrometheusLabels
		labelSelectorKeys = containerCPUUsageTotalRate.LabelMapper.LabelSelectorKeys
	case metrics.MetricTypeContainerMemoryUsage:
		labelsInPrometheus = containerMemoryUsage.LabelMapper.PrometheusLabels
		labelSelectorKeys = containerMemoryUsage.LabelMapper.LabelSelectorKeys
	case metrics.MetricTypeNodeCPUUsageSecondsAvg1M:
		labelsInPrometheus = nodeCPUUsageSecondsAVG1M.LabelMapper.PrometheusLabels
		labelSelectorKeys = nodeCPUUsageSecondsAVG1M.LabelMapper.LabelSelectorKeys
	case metrics.MetricTypeNodeMemoryUsageBytes:
		labelsInPrometheus = nodeMemoryUsageBytes.LabelMapper.PrometheusLabels
		labelSelectorKeys = nodeMemoryUsageBytes.LabelMapper.LabelSelectorKeys
	default:
		return result, errors.New(fmt.Sprintf("transform prometheus label to metrics.LabelSelectorKey failed: no exist LabelSelectorKey mapper for metric type %d", q.Metric))
	}

	for _, data := range q.Results {
		labels := utils.StringStringMap(data.Labels)
		data.Labels = labels.ReplaceKeys(labelsInPrometheus, labelSelectorKeys)
	}

	return result, nil
}

type Data struct {
	ResultType ResultType    `json:"resultType"`
	Result     []interface{} `json:"result"`
}

type ResultType string

var MatrixResultType ResultType = "matrix"
var VectorResultType ResultType = "vector"
var ScalarResultType ResultType = "scalar"
var StringResultType ResultType = "string"

type MatrixResult struct {
	Metric map[string]string `json:"metric"`
	Values []Value           `json:"values"`
}

type VectorResult struct {
	Metric map[string]string `json:"metric"`
	Value  Value             `json:"value"`
}

type ScalarResult Value

type StringResult Value

type Value []interface{}
