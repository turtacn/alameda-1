package prometheus

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/containers-ai/alameda/pkg/utils/log"
)

const (
	apiPrefix    = "/api/v1"
	epQuery      = "/query"
	epQueryRange = "/query_range"

	// StatusSuccess Status string literal of prometheus api request
	StatusSuccess = "success"
	// StatusError Status string literal of prometheus api request
	StatusError = "error"

	defaultStepTimeString = "30"
)

var (
	scope = log.RegisterScope("prometheus", "metrics repository", 0)
)

type Entity struct {
	Labels map[string]string
	Values []UnixTimeWithSampleValue
}

// Prometheus Prometheus api client
type Prometheus struct {
	config Config
	client http.Client
}

// New New Prometheus api client with configuration
func New(config Config) (*Prometheus, error) {

	var (
		err error

		p                = &Prometheus{}
		requestTimeout   = 30 * time.Second
		handShakeTimeout = 5 * time.Second
	)

	if err = config.Validate(); err != nil {
		return p, errors.New("create prometheus instance failed: " + err.Error())
	}

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
			return p, errors.New("create prometheus instance failed: open bearer token file for prometheus failed: " + err.Error())
		}
		p.config.bearerToken = string(token)
	}

	return p, nil
}

// QueryRange Query data over a range of time from prometheus
func (p *Prometheus) QueryRange(query string, startTime, endTime time.Time) (Response, error) {

	var (
		err error

		endpoint        = apiPrefix + epQueryRange
		queryParameters = url.Values{}

		u            *url.URL
		httpRequest  *http.Request
		httpResponse *http.Response

		response Response
	)

	startTimeString := strconv.FormatInt(int64(startTime.Unix()), 10)
	endTimeString := strconv.FormatInt(int64(endTime.Unix()), 10)

	queryParameters.Set("query", query)
	queryParameters.Set("start", startTimeString)
	queryParameters.Set("end", endTimeString)
	queryParameters.Set("step", defaultStepTimeString)

	u, err = url.Parse(p.config.URL + endpoint)
	if err != nil {
		return Response{}, errors.New("QueryRange failed: url parse failed: " + err.Error())
	}
	u.RawQuery = queryParameters.Encode()

	httpRequest, err = http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return Response{}, errors.New("Query: " + err.Error())
	}
	if token := p.config.bearerToken; token != "" {
		h := http.Header{
			"Authorization": []string{fmt.Sprintf(" Bearer %s", token)},
		}
		httpRequest.Header = h
	}

	httpResponse, err = p.client.Do(httpRequest)
	if err != nil {
		return Response{}, errors.New("QueryRange failed: send http request failed" + err.Error())
	}
	err = decodeHTTPResponse(httpResponse, &response)
	if err != nil {
		return Response{}, errors.New("QueryRange failed: " + err.Error())
	}

	return response, nil
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
