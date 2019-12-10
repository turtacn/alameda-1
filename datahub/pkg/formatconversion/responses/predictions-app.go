package responses

import (
	DaoPredictionTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/predictions/types"
	FormatEnum "github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	ApiPredictions "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/predictions"
)

type ApplicationPredictionExtended struct {
	*DaoPredictionTypes.ApplicationPrediction
}

func (d *ApplicationPredictionExtended) ProducePredictions() *ApiPredictions.ApplicationPrediction {
	var (
		rawDataChan        = make(chan ApiPredictions.MetricData)
		upperBoundDataChan = make(chan ApiPredictions.MetricData)
		lowerBoundDataChan = make(chan ApiPredictions.MetricData)
		numOfGoroutine     = 0

		datahubApplicationPrediction ApiPredictions.ApplicationPrediction
	)

	datahubApplicationPrediction = ApiPredictions.ApplicationPrediction{
		ObjectMeta: NewObjectMeta(&d.ObjectMeta),
	}

	// Handle prediction raw data
	numOfGoroutine = 0
	for metricType, samples := range d.PredictionRaw {
		if datahubMetricType, exist := FormatEnum.TypeToDatahubMetricType[metricType]; exist {
			numOfGoroutine++
			go producePredictionMetricDataFromSamples(datahubMetricType, samples.Granularity, samples.Data, rawDataChan)
		}
	}
	for i := 0; i < numOfGoroutine; i++ {
		receivedPredictionData := <-rawDataChan
		datahubApplicationPrediction.PredictedRawData = append(datahubApplicationPrediction.PredictedRawData, &receivedPredictionData)
	}

	// Handle prediction upper bound data
	numOfGoroutine = 0
	for metricType, samples := range d.PredictionUpperBound {
		if datahubMetricType, exist := FormatEnum.TypeToDatahubMetricType[metricType]; exist {
			numOfGoroutine++
			go producePredictionMetricDataFromSamples(datahubMetricType, samples.Granularity, samples.Data, upperBoundDataChan)
		}
	}
	for i := 0; i < numOfGoroutine; i++ {
		receivedPredictionData := <-upperBoundDataChan
		datahubApplicationPrediction.PredictedUpperboundData = append(datahubApplicationPrediction.PredictedUpperboundData, &receivedPredictionData)
	}

	// Handle prediction lower bound data
	numOfGoroutine = 0
	for metricType, samples := range d.PredictionLowerBound {
		if datahubMetricType, exist := FormatEnum.TypeToDatahubMetricType[metricType]; exist {
			numOfGoroutine++
			go producePredictionMetricDataFromSamples(datahubMetricType, samples.Granularity, samples.Data, lowerBoundDataChan)
		}
	}
	for i := 0; i < numOfGoroutine; i++ {
		receivedPredictionData := <-lowerBoundDataChan
		datahubApplicationPrediction.PredictedLowerboundData = append(datahubApplicationPrediction.PredictedLowerboundData, &receivedPredictionData)
	}

	return &datahubApplicationPrediction
}
