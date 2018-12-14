package prometheus

import (
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metrics"
)

var (
	MetricTypeName = map[metrics.MetricType]string{
		metrics.MetricTypeContainerCPUUsageTotal:     "container_cpu_usage_seconds_total",
		metrics.MetricTypeContainerCPUUsageTotalRate: "namespace_pod_name_container_name:container_cpu_usage_seconds_total:sum_rate",
		metrics.MetricTypeContainerMemoryUsage:       "container_memory_usage_bytes",
	}

	StringOperatorLiteral = map[metrics.StringOperator]string{
		metrics.StringOperatorEqueal:    "=",
		metrics.StringOperatorNotEqueal: "!=",
	}
)

type Response struct {
	StatusCode int
	Status     string `json:"status"`
	Data       Data   `json:"data"`
	ErrorType  string `json:"errorType"`
	Error      string `json:"error"`
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
