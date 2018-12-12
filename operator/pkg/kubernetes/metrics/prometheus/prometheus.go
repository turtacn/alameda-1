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
	"github.com/containers-ai/alameda/operator/pkg/kubernetes/metrics/prometheus/containerCPUUsageTotal"
	"github.com/containers-ai/alameda/operator/pkg/kubernetes/metrics/prometheus/containerCPUUsageTotalRate"
	"github.com/containers-ai/alameda/operator/pkg/kubernetes/metrics/prometheus/containerMemoryUsage"
	"github.com/containers-ai/alameda/operator/pkg/kubernetes/metrics/prometheus/factory"
	"github.com/containers-ai/alameda/operator/pkg/kubernetes/metrics/prometheus/nodeCPUUsageSecondsAVG1M"
	"github.com/containers-ai/alameda/operator/pkg/kubernetes/metrics/prometheus/nodeMemoryUsageBytes"
	"github.com/containers-ai/alameda/operator/pkg/utils/log"
)

const (
	apiPrefix    = "/api/v1"
	epQuery      = "/query"
	epQueryRange = "/query_range"
)

var (
	scope = log.RegisterScope("prometheus", "metrics repository", 0)

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

func (p prometheus) BaseURL() string {
	return p.config.URL
}

func (p prometheus) BearerToken() string {
	return p.config.bearerToken
}

func New(config Config) metrics.MetricsDB {

	var (
		p                = &prometheus{}
		requestTimeout   = 30 * time.Second
		handShakeTimeout = 5 * time.Second
	)

	p.config = config

	u, _ := url.Parse(config.URL)
	p.client = http.Client{
		Timeout: requestTimeout,
	}
	if strings.ToLower(u.Scheme) == "https" {
		p.client.Transport = &http.Transport{
			TLSHandshakeTimeout: handShakeTimeout,
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: p.config.TLSConfig.InsecureSkipVerify},
		}
	}

	if p.config.BearerTokenFile != "" {
		token, err := ioutil.ReadFile(config.BearerTokenFile)
		if err != nil {
			scope.Error("open bearer token file for prometheus failed: " + err.Error())
			panic("open bearer token file for prometheus failed")
		}
		p.config.bearerToken = string(token)
	}

	return p
}

func (p *prometheus) Connect() error {
	return nil
}

func (p *prometheus) Close() error {

	return nil
}

func (p *prometheus) Query(q metrics.Query) (metrics.QueryResponse, error) {

	var requestFactory metrics.BackendQueryRequestFactory
	var req http.Request

	requestFactory = p.queryRequestFactory(q)
	request, err := requestFactory.BuildServiceRequest()
	if err != nil {
		scope.Error("build service request failed: " + err.Error())
		return metrics.QueryResponse{}, errors.New("Query: " + err.Error())
	}
	req, ok := request.(http.Request)
	if !ok {
		scope.Error("type assert to http.Request failed: " + err.Error())
		return metrics.QueryResponse{}, errors.New("Query: " + err.Error())
	}

	// Send request to prometheus
	resp, err := p.client.Do(&req)
	if err != nil {
		scope.Error("send http request to prometheus failed: " + err.Error())
		return metrics.QueryResponse{}, err
	}

	// Convert http response to metrics response
	response, err := getResponse(resp)
	if err != nil {
		scope.Error("get prometheus response error" + err.Error())
		return metrics.QueryResponse{}, errors.New("Query: %s" + err.Error())
	} else if response.Status == "error" {
		scope.Error("get error response from prometheus" + response.Error)
		return metrics.QueryResponse{}, errors.New("Query: %s" + response.Error)
	}

	// Convert Response to QueryResponse
	queryResponse, err := convertQueryResponse(response)
	if err != nil {
		scope.Error("convert Response to QueryResponse failed: " + err.Error())
		return metrics.QueryResponse{}, errors.New("Query: %s" + err.Error())
	}

	return queryResponse, nil
}

func (p *prometheus) queryRequestFactory(q metrics.Query) metrics.BackendQueryRequestFactory {

	var (
		baseURL     = p.BaseURL()
		bearerToken = p.BearerToken()
		factoryOpts = []factory.QueryRequestFactoryOpts{
			factory.PromAddr(baseURL),
			factory.PromAuth(bearerToken),
		}

		bqrf metrics.BackendQueryRequestFactory
	)

	switch q.Metric {
	case metrics.MetricTypeContainerCPUUsageTotal:
		bqrf = containerCPUUsageTotal.NewQueryRequestFactory(q, factoryOpts...)
	case metrics.MetricTypeContainerCPUUsageTotalRate:
		bqrf = containerCPUUsageTotalRate.NewQueryRequestFactory(q, factoryOpts...)
	case metrics.MetricTypeContainerMemoryUsage:
		bqrf = containerMemoryUsage.NewQueryRequestFactory(q, factoryOpts...)
	case metrics.MetricTypeNodeCPUUsageSecondsAvg1M:
		bqrf = nodeCPUUsageSecondsAVG1M.NewQueryRequestFactory(q, factoryOpts...)
	case metrics.MetricTypeNodeMemoryUsageBytes:
		bqrf = nodeMemoryUsageBytes.NewQueryRequestFactory(q, factoryOpts...)
	}

	return bqrf
}

// getResponse Convert http response to Response struct
func getResponse(resp *http.Response) (Response, error) {

	var r Response

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Response{}, err
	}

	// unmarshal json response to struct Response
	// if unmarshal error or receive error from prometheus, return error
	err = json.Unmarshal(body, &r)
	if err != nil {
		return Response{}, err
	}

	r.StatusCode = resp.StatusCode

	return r, nil
}

// convertQueryResponse convert the response from prometheus to metrics.QueryResponse
// The format of prometheus response vary from resultType.
func convertQueryResponse(r Response) (metrics.QueryResponse, error) {

	// build QueryResponse instance base on different type of prometheus resultType
	// For each case, get the associated type result first by converting the result of http response to []byte first and then unmarshal.
	// After getting the result, extract labels and value(s) from the result to serie instance.
	queryResponse := metrics.QueryResponse{}
	switch r.Data.ResultType {
	case "matrix":
		for _, r := range r.Data.Result {

			result := MatrixResult{}
			if _, ok := r.(map[string]interface{}); !ok {
				return metrics.QueryResponse{}, fmt.Errorf("error while building sample, cannot convert type %s to map[string]interface{}", reflect.TypeOf(r).String())
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
					return metrics.QueryResponse{}, fmt.Errorf("error while building sample, cannot convert type %s to float64", reflect.TypeOf(value[0]))
				}
				unixTime := time.Unix(int64(value[0].(float64)), 0)

				if _, ok := value[1].(string); !ok {
					return metrics.QueryResponse{}, fmt.Errorf("error while building sample, cannot convert type %s to string", reflect.TypeOf(value[1]))
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
				return metrics.QueryResponse{}, fmt.Errorf("error while building sample, cannot convert type %s to map[string]interface{}", reflect.TypeOf(r).String())
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
				return metrics.QueryResponse{}, fmt.Errorf("error while building sample, cannot convert type %s to float64", reflect.TypeOf(value[0]))
			}
			unixTime := time.Unix(int64(value[0].(float64)), 0)

			if _, ok := value[1].(string); !ok {
				return metrics.QueryResponse{}, fmt.Errorf("error while building sample, cannot convert %+v(type %s) to string", value[1], reflect.TypeOf(value[1]))
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
