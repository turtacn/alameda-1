package prometheus

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
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

	var (
		requestFactory metrics.BackendQueryRequestFactory

		httpRequest   http.Request
		httpResponse  *http.Response
		queryResponse metrics.QueryResponse
	)

	requestFactory = p.queryRequestFactory(q)
	request, err := requestFactory.BuildServiceRequest()
	if err != nil {
		scope.Error("build service request failed: " + err.Error())
		return metrics.QueryResponse{}, errors.New("Query: " + err.Error())
	}
	httpRequest, ok := request.(http.Request)
	if !ok {
		scope.Error("type assert to http.Request failed: " + err.Error())
		return metrics.QueryResponse{}, errors.New("Query: " + err.Error())
	}

	// Send request to prometheus
	httpResponse, err = p.client.Do(&httpRequest)
	if err != nil {
		scope.Error("send http request to prometheus failed: " + err.Error())
		return metrics.QueryResponse{}, err
	}

	var response Response
	err = decodeHTTPResponse(httpResponse, &response)
	if err != nil {
		scope.Error("decode http response failed: " + err.Error())
		return metrics.QueryResponse{}, errors.New("Query: %s" + err.Error())
	}
	if response.Status == statusError {
		scope.Error("get error response from prometheus" + response.Error)
		return metrics.QueryResponse{}, errors.New("Query: %s" + response.Error)
	}
	response.setMetricType(q.Metric)
	queryResponse, err = response.transformToMetricsQueryResponse()
	if err != nil {
		scope.Error("transform Response to QueryResponse failed: " + err.Error())
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

func decodeHTTPResponse(httpResponse *http.Response, response *Response) error {

	var err error

	defer httpResponse.Body.Close()
	err = json.NewDecoder(httpResponse.Body).Decode(&response)
	if err != nil {
		return errors.New("decode http response failed: %s" + err.Error())
	}

	return nil
}
