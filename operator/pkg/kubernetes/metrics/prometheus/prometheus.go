package prometheus

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/containers-ai/alameda/operator/pkg/kubernetes/metrics"
	"github.com/containers-ai/alameda/operator/pkg/utils/log"
)

const (
	apiPrefix    = "/api/v1"
	epQuery      = "/query"
	epQueryRange = "/query_range"
)

var (
	// defaultLabelSelectors LabelSelectors must apply when query
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

type prometheus struct {
	config Config
	client http.Client
}

func New(config Config) (metrics.MetricsDB, error) {

	p := &prometheus{}

	p.config = config

	if p.config.TLSConfig != nil {
		p.client = http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: p.config.TLSConfig.InsecureSkipVerify},
			},
		}
	} else {
		p.client = http.Client{}
	}

	if p.config.BearerTokenFile != "" {
		token, err := ioutil.ReadFile(config.BearerTokenFile)
		if err != nil {
			log.GetLogger().Error(err, "open bearer token file for prometheus failed")
			return nil, errors.New("open bearer token file for prometheus failed")
		}
		p.config.bearerToken = string(token)
	}

	return p, nil
}

func (p *prometheus) Connect() error {
	return nil
}

func (p *prometheus) Close() error {

	return nil
}

func (p *prometheus) Query(q metrics.Query) (metrics.QueryResponse, error) {

	// Get query url
	u, err := p.queryUrl(q)
	if err != nil {
		return metrics.QueryResponse{}, err
	}

	// Build http request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return metrics.QueryResponse{}, err
	}
	if token := p.config.bearerToken; token != "" {
		h := http.Header{
			"Authorization": []string{fmt.Sprintf(" Bearer %s", token)},
		}
		req.Header = h
	}

	// Send request to prometheus
	resp, err := p.client.Do(req)
	if err != nil {
		log.GetLogger().Error(err, "send http request failed")
		return metrics.QueryResponse{}, err
	}

	// Convert http response to metrics response
	response, err := convertQueryResponse(resp)
	if err != nil {
		log.GetLogger().Error(err, err.Error())
		return metrics.QueryResponse{}, err
	}

	return response, nil
}

// getQueryUrl Return url.URL by the metrics.Query instance
func (p *prometheus) queryUrl(q metrics.Query) (url.URL, error) {

	var (
		u *url.URL
		v = &url.Values{}
	)

	// get query end point base on query time selector
	ep := p.queryEndpoint(&q)

	// set query expression in query parameters
	setQueryExpressionParameter(v, q)

	// set query time in query parameters
	setQueryTimeParameter(v, q)

	u, err := url.Parse(ep)
	if err != nil {
		log.GetLogger().Error(err, "parse requset url failed")
		return url.URL{}, err
	}
	u.RawQuery = v.Encode()

	return *u, nil
}

// queryEndpoint Return query endpoint base on prometheus protocol,address and type of query time
func (p *prometheus) queryEndpoint(q *metrics.Query) string {

	var ep string

	protocol := p.config.Protocol
	host := p.config.Host
	port := p.config.Port
	ep = fmt.Sprintf("%s://%s:%s", protocol, host, port)

	// append query endpoint into ep and set query parameter
	switch q.TimeSelector.(type) {
	case nil:
		ep += fmt.Sprintf("%s%s", apiPrefix, epQuery)

	case *metrics.Timestamp:
		ep += fmt.Sprintf("%s%s", apiPrefix, epQuery)

	case *metrics.TimeRange:
		ep += fmt.Sprintf("%s%s", apiPrefix, epQueryRange)

	case *metrics.Since:
		ep += fmt.Sprintf("%s%s", apiPrefix, epQuery)
	}

	return ep
}

func setQueryExpressionParameter(v *url.Values, q metrics.Query) {

	var (
		queryExpression string // query represent the query expression to prometheus api
		lss             = q.LabelSelectors
	)

	lss = append(lss, defaultLabelSelectors...)

	// build prometheus query expression
	labelSelectorString := ""
	for _, ls := range lss {

		k := ls.Key
		v := ls.Value
		op := StringOperatorLiteral[ls.Op]

		labelSelectorString += fmt.Sprintf("%s %s \"%s\",", k, op, v)
	}
	labelSelectorString = strings.TrimSuffix(labelSelectorString, ",")
	queryExpression = fmt.Sprintf("%s{%s}", MetricTypeName[q.Metric], labelSelectorString)

	switch q.TimeSelector.(type) {
	case *metrics.Since:
		d := q.TimeSelector.(*metrics.Since)
		rangeDurationString := fmt.Sprintf("[%ss]", strconv.FormatFloat(d.Duration.Seconds(), 'f', 0, 64))
		queryExpression = queryExpression + rangeDurationString
	}

	v.Set("query", queryExpression)
}

func setQueryTimeParameter(v *url.Values, q metrics.Query) {

	switch q.TimeSelector.(type) {
	case *metrics.Timestamp:
		t := q.TimeSelector.(*metrics.Timestamp)
		tStr := strconv.FormatInt(int64(t.T.Unix()), 10)
		v.Set("time", tStr)

	case *metrics.TimeRange:
		t := q.TimeSelector.(*metrics.TimeRange)
		startTime := t.StartTime
		endTime := t.EndTime
		step := t.Step
		startTimeString := strconv.FormatInt(int64(startTime.Unix()), 10)
		endTimeString := strconv.FormatInt(int64(endTime.Unix()), 10)
		stepString := strconv.FormatFloat(step.Seconds(), 'f', 0, 64)

		v.Set("start", startTimeString)
		v.Set("end", endTimeString)
		v.Set("step", stepString)
	}
}

// convertQueryResponse convert the response from prometheus query request to instance metrics.QueryResponse
// The format of prometheus response vary from resultType.
func convertQueryResponse(resp *http.Response) (metrics.QueryResponse, error) {

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return metrics.QueryResponse{}, err
	}

	// unmarshal json response to struct Response
	// if unmarshal error or receive error from prometheus, return error
	var r Response
	err = json.Unmarshal(body, &r)
	if err != nil {
		return metrics.QueryResponse{}, err
	} else if r.Error != "" || r.ErrorType != "" {
		return metrics.QueryResponse{}, errors.New(r.Error)
	}

	// build QueryResponse instance base on different type of prometheus resultType
	// For each case, get the associated type result first by converting the result of http response to []byte first and then unmarshal.
	// After getting the result, extract labels and value(s) from the result to serie instance.
	queryResponse := metrics.QueryResponse{}
	switch r.Data.ResultType {
	case "matrix":
		for _, r := range r.Data.Result {

			result := MatrixResult{}
			if _, ok := r.(map[string]interface{}); !ok {
				return metrics.QueryResponse{}, fmt.Errorf("cannot convert type %s to map[string]interface{}", reflect.TypeOf(r).String())
			}
			resultStr, err := json.Marshal(r.(map[string]interface{}))
			if err != nil {
				return metrics.QueryResponse{}, err
			}
			err = json.Unmarshal(resultStr, &result)
			if err != nil {
				return metrics.QueryResponse{}, err
			}

			serie := metrics.Data{}
			serie.Labels = result.Metric
			for _, value := range result.Values {

				if _, ok := value[0].(float64); !ok {
					return metrics.QueryResponse{}, fmt.Errorf("cannot convert type %s to float64", reflect.TypeOf(value[0]))
				}
				unixTime := time.Unix(int64(value[0].(float64)), 0)

				if _, ok := value[1].(string); !ok {
					return metrics.QueryResponse{}, fmt.Errorf("cannot convert type %s to float64", reflect.TypeOf(value[1]))
				}
				sampleValue, err := strconv.ParseFloat(value[1].(string), 64)
				if err != nil {
					return metrics.QueryResponse{}, err
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
				return metrics.QueryResponse{}, fmt.Errorf("cannot convert type %s to map[string]interface{}", reflect.TypeOf(r).String())
			}
			resultStr, err := json.Marshal(r.(map[string]interface{}))
			if err != nil {
				return metrics.QueryResponse{}, err
			}
			err = json.Unmarshal(resultStr, &result)
			if err != nil {
				return metrics.QueryResponse{}, err
			}

			serie := metrics.Data{}
			serie.Labels = result.Metric
			value := result.Value

			if _, ok := value[0].(float64); !ok {
				return metrics.QueryResponse{}, fmt.Errorf("cannot convert type %s to float64", reflect.TypeOf(value[0]))
			}
			unixTime := time.Unix(int64(value[0].(float64)), 0)

			if _, ok := value[1].(string); !ok {
				return metrics.QueryResponse{}, fmt.Errorf("cannot convert type %s to float64", reflect.TypeOf(value[1]))
			}
			sampleValue, err := strconv.ParseFloat(value[1].(string), 64)
			if err != nil {
				return metrics.QueryResponse{}, err
			}

			sample := metrics.Sample{
				Time:  unixTime,
				Value: sampleValue,
			}
			serie.Samples = append(serie.Samples, sample)

			queryResponse.Results = append(queryResponse.Results, serie)
		}
	case "scalar":
		return metrics.QueryResponse{}, errors.New(fmt.Sprintf("not implement for resultType \"%s\"", r.Data.ResultType))
	case "string":
		return metrics.QueryResponse{}, errors.New(fmt.Sprintf("not implement for resultType \"%s\"", r.Data.ResultType))
	default:
		return metrics.QueryResponse{}, errors.New(fmt.Sprintf("not implement for resultType \"%s\"", r.Data.ResultType))
	}

	return queryResponse, nil
}
