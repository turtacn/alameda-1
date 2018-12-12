package containerCPUUsageTotalRate

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/containers-ai/alameda/operator/pkg/kubernetes/metrics"
	"github.com/containers-ai/alameda/operator/pkg/kubernetes/metrics/prometheus/factory"
	"github.com/containers-ai/alameda/operator/pkg/kubernetes/metrics/prometheus/request"
)

var (
	metricName            = "namespace_pod_name_container_name:container_cpu_usage_seconds_total:sum_rate"
	defaultLabelSelectors = []metrics.LabelSelector{
		metrics.LabelSelector{
			Key:   "container_name",
			Op:    metrics.StringOperatorNotEqueal,
			Value: "POD",
		},
		metrics.LabelSelector{
			Key:   "container_name",
			Op:    metrics.StringOperatorNotEqueal,
			Value: "",
		},
	}
)

type QueryRequestFactory struct {
	Query      metrics.Query
	FactoryOpt factory.QueryRequestFactoryOpt
}

func NewQueryRequestFactory(q metrics.Query, opts ...factory.QueryRequestFactoryOpts) metrics.BackendQueryRequestFactory {

	var factoryOpt factory.QueryRequestFactoryOpt

	for _, o := range opts {
		o(&factoryOpt)
	}

	return &QueryRequestFactory{
		Query:      q,
		FactoryOpt: factoryOpt,
	}
}

func (f *QueryRequestFactory) BuildServiceRequest() (interface{}, error) {

	// Get query url
	u, err := f.queryUrl(f.Query)
	if err != nil {
		return metrics.QueryResponse{}, errors.New("Query: " + err.Error())
	}

	// Build http request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return metrics.QueryResponse{}, errors.New("Query: " + err.Error())
	}
	if token := f.FactoryOpt.PromAuth; token != "" {
		h := http.Header{
			"Authorization": []string{fmt.Sprintf(" Bearer %s", token)},
		}
		req.Header = h
	}

	return *req, nil
}

// getQueryUrl Return url.URL by the metrics.Query instance
func (f *QueryRequestFactory) queryUrl(q metrics.Query) (url.URL, error) {

	var (
		u *url.URL

		endpoint        string
		queryExpression string
		queryParameters url.Values
	)

	endpoint = request.GetQueryEndpointByTimeSelector(q.TimeSelector)
	queryExpression = getQueryExpression(q)
	queryParameters = request.GetQueryParametersByTimeSelector(q.TimeSelector)

	u, err := url.Parse(f.FactoryOpt.PromAddr + endpoint)
	if err != nil {
		return url.URL{}, errors.New("parse requset url failed: " + err.Error())
	}
	queryParameters.Set("query", queryExpression)
	u.RawQuery = queryParameters.Encode()

	return *u, nil
}

func getQueryExpression(q metrics.Query) string {

	var (
		expression string // query represent the query expression to prometheus api
		lss        = q.LabelSelectors
	)

	lss = append(lss, defaultLabelSelectors...)

	// build prometheus query expression
	labelSelectorString := ""
	for _, ls := range lss {

		k := ls.Key
		v := ls.Value
		op := request.StringOperatorLiteral[ls.Op]

		labelSelectorString += fmt.Sprintf("%s %s \"%s\",", k, op, v)
	}
	labelSelectorString = strings.TrimSuffix(labelSelectorString, ",")
	expression = fmt.Sprintf("%s{%s}", metricName, labelSelectorString)

	switch q.TimeSelector.(type) {
	case *metrics.Since:
		d := q.TimeSelector.(*metrics.Since)
		rangeDurationString := fmt.Sprintf("[%ss]", strconv.FormatFloat(d.Duration.Seconds(), 'f', 0, 64))
		expression = expression + rangeDurationString
	}

	return expression
}
