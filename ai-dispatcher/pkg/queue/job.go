package queue

import (
	"encoding/json"
	"time"

	"github.com/containers-ai/alameda/ai-dispatcher/pkg/metrics"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/streadway/amqp"
)

type job struct {
	UnitType          string `json:"unitType"`
	Granularity       string `json:"granularity"`
	GranularitySec    int64  `json:"granularitySec"`
	PayloadJSONString string `json:"payloadJSONString"`
	CreateTimeStamp   int64  `json:"createTimestamp"`
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
		CreateTimeStamp:   time.Now().Unix(),
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

func GetGranularitySec(granularityStr string) int64 {
	if granularityStr == "30s" {
		return 30
	} else if granularityStr == "1h" {
		return 3600
	} else if granularityStr == "6h" {
		return 21600
	} else if granularityStr == "24h" {
		return 86400
	}
	return 30
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

func GetMetricLabel(mt datahub_v1alpha1.MetricType) string {
	if mt == datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE {
		return metrics.MetricTypeLabelCPUUsageSecondsPercentage
	} else if mt == datahub_v1alpha1.MetricType_DUTY_CYCLE {
		return metrics.MetricTypeLabelDutyCycle
	} else if mt == datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES {
		return metrics.MetricTypeLabelMemoryUsageBytes
	} else if mt == datahub_v1alpha1.MetricType_POWER_USAGE_WATTS {
		return metrics.MetricTypeLabelPowerUsageWatts
	} else if mt == datahub_v1alpha1.MetricType_TEMPERATURE_CELSIUS {
		return metrics.MetricTypeLabelTemperatureCelsius
	}
	return ""
}
