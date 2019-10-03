package types

import (
	Metric "github.com/containers-ai/alameda/datahub/pkg/metric"
	DatahubV1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes"
)

func produceDatahubMetricDataFromSamples(metricType DatahubV1alpha1.MetricType, samples []Metric.Sample, MetricDataChan chan<- DatahubV1alpha1.MetricData) {
	var (
		datahubMetricData DatahubV1alpha1.MetricData
	)

	datahubMetricData = DatahubV1alpha1.MetricData{
		MetricType: metricType,
	}

	for _, sample := range samples {

		// TODO: Send error to caller
		googleTimestamp, err := ptypes.TimestampProto(sample.Timestamp)
		if err != nil {
			googleTimestamp = nil
		}

		datahubSample := DatahubV1alpha1.Sample{Time: googleTimestamp, NumValue: sample.Value}
		datahubMetricData.Data = append(datahubMetricData.Data, &datahubSample)
	}

	MetricDataChan <- datahubMetricData
}
