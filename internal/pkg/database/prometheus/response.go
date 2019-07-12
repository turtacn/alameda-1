package prometheus

import (
	"encoding/json"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"time"
)

const (
	// Status string literal of prometheus read/write request
	StatusSuccess = "success"
	// Status string literal of prometheus read/write request
	StatusError = "error"
)

// Structure of prometheus http response
type Response struct {
	Status    string `json:"status"`
	Data      Data   `json:"data"`
	ErrorType string `json:"errorType"`
	Error     string `json:"error"`
}

// Represents the general value of prometheus http response field "data", the format of field "result" will depends on field "resultType"
type Data struct {
	ResultType ResultType    `json:"resultType"`
	Result     []interface{} `json:"result"`
}

// Structure to store metrics data from prometheus response
type Entity struct {
	Labels map[string]string
	Values []UnixTimeWithSampleValue
}

// Prometheus http resultType
type ResultType string

// Prometheus http matrix resultType
var MatrixResultType ResultType = "matrix"

// Prometheus http vector resultType
var VectorResultType ResultType = "vector"

// Prometheus http scalar resultType
var ScalarResultType ResultType = "scalar"

// Prometheus http string resultType
var StringResultType ResultType = "string"

type MatrixResult struct {
	Metric map[string]string `json:"metric"`
	Values []Value           `json:"values"`
}

type VectorResult struct {
	Metric map[string]string `json:"metric"`
	Value  Value             `json:"value"`
}

type ScalarResult UnixTimeWithScalarValue

type StringResult UnixTimeWithStringValue

type Value []interface{}

type UnixTimeWithSampleValue struct {
	UnixTime    time.Time
	SampleValue string
}

type UnixTimeWithScalarValue struct {
	UnixTime    time.Time
	ScalarValue string
}

type UnixTimeWithStringValue struct {
	UnixTime    time.Time
	StringValue string
}

// MatrixResponse MatrixResponse
type MatrixResponse struct {
	Status string
	Data   MatrixData
}

// MatrixData MatrixData
type MatrixData struct {
	ResultType ResultType
	Result     []struct {
		Metric map[string]string
		Values []UnixTimeWithSampleValue
	}
}

// VectorResponse VectorResponse
type VectorResponse struct {
	Status string
	Data   VectorData
}

// VectorData VectorData
type VectorData struct {
	ResultType ResultType
	Result     []struct {
		Metric map[string]string
		Value  UnixTimeWithSampleValue
	}
}

func (r Response) GetMatrixResponse() (MatrixResponse, error) {

	var (
		response = MatrixResponse{}
	)

	for _, r := range r.Data.Result {

		matrixResult := MatrixResult{}
		if _, ok := r.(map[string]interface{}); !ok {
			return response, errors.Errorf("error while building sample, cannot convert type %s to map[string]interface{}", reflect.TypeOf(r).String())
		}
		resultStr, err := json.Marshal(r.(map[string]interface{}))
		if err != nil {
			return response, err
		}
		err = json.Unmarshal(resultStr, &matrixResult)
		if err != nil {
			return response, err
		}

		var typeSpecifiedMatrixResult = struct {
			Metric map[string]string
			Values []UnixTimeWithSampleValue
		}{
			Metric: matrixResult.Metric,
		}

		for _, value := range matrixResult.Values {

			if _, ok := value[0].(float64); !ok {
				return response, errors.Errorf("error while building sample, cannot convert type %s to float64", reflect.TypeOf(value[0]))
			}
			unixTime := time.Unix(int64(value[0].(float64)), 0)

			if _, ok := value[1].(string); !ok {
				return response, errors.Errorf("error while building sample, cannot convert type %s to string", reflect.TypeOf(value[1]))
			}
			sampleValue := value[1].(string)

			unixTimeWithSampleValue := UnixTimeWithSampleValue{
				UnixTime:    unixTime,
				SampleValue: sampleValue,
			}
			typeSpecifiedMatrixResult.Values = append(typeSpecifiedMatrixResult.Values, unixTimeWithSampleValue)
		}

		response.Data.Result = append(response.Data.Result, typeSpecifiedMatrixResult)
	}

	return response, nil
}

func (r Response) GetEntities() ([]Entity, error) {
	if r.Status != StatusSuccess {
		return make([]Entity, 0), errors.Errorf("GetEntities failed: response status is not %s", StatusSuccess)
	}

	switch r.Data.ResultType {
	case MatrixResultType:
		return r.GetEntitiesMatrix()
	case VectorResultType:
		return r.GetEntitiesVector()
	default:
		return make([]Entity, 0), errors.Errorf("GetEntities failed: result type not supported %s", string(r.Data.ResultType))
	}
}

func (r Response) GetEntitiesMatrix() ([]Entity, error) {
	var (
		entities = make([]Entity, 0)
	)

	if r.Status != StatusSuccess {
		return entities, errors.Errorf("GetEntities failed: response status is not %s", StatusSuccess)
	}

	for _, r := range r.Data.Result {
		matrixResult := MatrixResult{}
		if _, ok := r.(map[string]interface{}); !ok {
			return entities, errors.Errorf("error while building sample, cannot convert type %s to map[string]interface{}", reflect.TypeOf(r).String())
		}

		resultStr, err := json.Marshal(r.(map[string]interface{}))
		if err != nil {
			return entities, err
		}

		err = json.Unmarshal(resultStr, &matrixResult)
		if err != nil {
			return entities, err
		}

		entity := Entity{
			Labels: matrixResult.Metric,
			Values: make([]UnixTimeWithSampleValue, 0),
		}
		for _, value := range matrixResult.Values {

			if _, ok := value[0].(float64); !ok {
				return entities, errors.Errorf("error while building sample, cannot convert type %s to float64", reflect.TypeOf(value[0]))
			}
			unixTime := time.Unix(int64(value[0].(float64)), 0)

			if _, ok := value[1].(string); !ok {
				return entities, errors.Errorf("error while building sample, cannot convert type %s to string", reflect.TypeOf(value[1]))
			}
			sampleValue := value[1].(string)

			unixTimeWithSampleValue := UnixTimeWithSampleValue{
				UnixTime:    unixTime,
				SampleValue: sampleValue,
			}
			entity.Values = append(entity.Values, unixTimeWithSampleValue)
		}

		entities = append(entities, entity)
	}

	return entities, nil
}

func (r Response) GetEntitiesVector() ([]Entity, error) {
	var (
		entities = make([]Entity, 0)
	)

	if r.Status != StatusSuccess {
		return entities, errors.Errorf("GetEntities failed: response status is not %s", StatusSuccess)
	}

	for _, r := range r.Data.Result {
		vectorResult := VectorResult{}
		if _, ok := r.(map[string]interface{}); !ok {
			return entities, errors.Errorf("error while building sample, cannot convert type %s to map[string]interface{}", reflect.TypeOf(r).String())
		}

		resultStr, err := json.Marshal(r.(map[string]interface{}))
		if err != nil {
			return entities, err
		}

		err = json.Unmarshal(resultStr, &vectorResult)
		if err != nil {
			return entities, err
		}

		entity := Entity{
			Labels: vectorResult.Metric,
			Values: make([]UnixTimeWithSampleValue, 0),
		}

		if _, ok := vectorResult.Value[0].(float64); !ok {
			return entities, errors.Errorf("error while building sample, cannot convert type %s to float64", reflect.TypeOf(vectorResult.Value[0]))
		}
		unixTime := time.Unix(int64(vectorResult.Value[0].(float64)), 0)

		if _, ok := vectorResult.Value[1].(string); !ok {
			return entities, errors.Errorf("error while building sample, cannot convert type %s to string", reflect.TypeOf(vectorResult.Value[1]))
		}
		sampleValue := vectorResult.Value[1].(string)

		unixTimeWithSampleValue := UnixTimeWithSampleValue{
			UnixTime:    unixTime,
			SampleValue: sampleValue,
		}
		entity.Values = append(entity.Values, unixTimeWithSampleValue)

		entities = append(entities, entity)
	}

	return entities, nil
}

func decodeHTTPResponse(httpResponse *http.Response, response *Response) error {

	var (
		err                    error
		httpResponseBody       []byte
		httpResponseBodyReader io.Reader
	)

	defer httpResponse.Body.Close()

	httpResponseBody, err = ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return errors.New("decode http response failed: read http response body failed: " + err.Error())
	}

	httpResponseBodyReader = strings.NewReader(string(httpResponseBody))
	err = json.NewDecoder(httpResponseBodyReader).Decode(&response)
	if err != nil {
		return errors.New("decode http response failed: " + err.Error() + " \n received response: " + string(httpResponseBody))
	}

	return nil
}
