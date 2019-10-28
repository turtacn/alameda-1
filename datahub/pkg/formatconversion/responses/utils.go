package responses

import (
	FormatTypes "github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	ApiCommon "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	ApiPredictions "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/predictions"
	"github.com/golang/protobuf/ptypes"
)

func produceMetricDataFromSamples(metricType ApiCommon.MetricType, samples []FormatTypes.Sample, MetricDataChan chan<- ApiCommon.MetricData) {
	datahubMetricData := ApiCommon.MetricData{
		MetricType: metricType,
	}

	for _, sample := range samples {
		// TODO: Send error to caller
		googleTimestamp, err := ptypes.TimestampProto(sample.Timestamp)
		if err != nil {
			googleTimestamp = nil
		}

		datahubSample := ApiCommon.Sample{Time: googleTimestamp, NumValue: sample.Value}
		datahubMetricData.Data = append(datahubMetricData.Data, &datahubSample)
	}

	MetricDataChan <- datahubMetricData
}

func producePredictionMetricDataFromSamples(metricType ApiCommon.MetricType, granularity int64, samples []FormatTypes.PredictionSample, MetricDataChan chan<- ApiPredictions.MetricData) {
	datahubMetricData := ApiPredictions.MetricData{
		MetricType:  metricType,
		Granularity: granularity,
	}

	for _, sample := range samples {
		// TODO: Send error to caller
		googleTimestamp, err := ptypes.TimestampProto(sample.Timestamp)
		if err != nil {
			googleTimestamp = nil
		}

		datahubSample := ApiPredictions.Sample{Time: googleTimestamp, NumValue: sample.Value, ModelId: sample.ModelId, PredictionId: sample.PredictionId}
		datahubMetricData.Data = append(datahubMetricData.Data, &datahubSample)
	}

	MetricDataChan <- datahubMetricData
}
