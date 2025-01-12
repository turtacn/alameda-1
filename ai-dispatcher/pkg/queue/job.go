package queue

import (
	"encoding/json"
	"time"

	"github.com/streadway/amqp"
)

type job struct {
	UnitType          string `json:"unitType"`
	Granularity       string `json:"granularity"`
	GranularitySec    int64  `json:"granularitySec"`
	PayloadJSONString string `json:"payloadJSONString"`
}

type jobBuilder struct {
	job *job
}

func NewJobBuilder(unitType string, granularitySec int64, payloadJSONString string) *jobBuilder {
	granularity := GetGranularityStr(granularitySec)
	job := &job{
		UnitType:          unitType,
		GranularitySec:    granularitySec,
		Granularity:       granularity,
		PayloadJSONString: payloadJSONString,
	}
	return &jobBuilder{job: job}
}

func (jobBuilder *jobBuilder) GetJobJSONString() (string, error) {
	jobJSONBin, err := json.Marshal(jobBuilder.job)
	if err != nil {
		return "", err
	}
	return string(jobJSONBin), err
}

func GetGranularityStr(granularitySec int64) string {
	if granularitySec == 30 {
		return "30s"
	} else if granularitySec == 3600 {
		return "1h"
	} else if granularitySec == 21600 {
		return "6h"
	} else if granularitySec == 86400 {
		return "24h"
	}
	return "30s"
}

func GetQueueConn(queueURL string, retryItvMS int64) *amqp.Connection {
	for {
		queueConn, err := amqp.Dial(queueURL)
		if err != nil {
			scope.Errorf("Queue connection constructs failed and will retry after %v milliseconds. %s", retryItvMS, err.Error())
			time.Sleep(time.Duration(retryItvMS) * time.Millisecond)
			continue
		}
		return queueConn
	}
}
