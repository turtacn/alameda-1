package responses

import (
	DaoPredictionTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/predictions/types"
	FormatEnum "github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	ApiPredictions "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/predictions"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

type ControllerPredictionExtended struct {
	*DaoPredictionTypes.ControllerPrediction
}

func (d *ControllerPredictionExtended) ProducePredictions() *ApiPredictions.ControllerPrediction {
	var (
		rawDataChan        = make(chan ApiPredictions.MetricData)
		upperBoundDataChan = make(chan ApiPredictions.MetricData)
		lowerBoundDataChan = make(chan ApiPredictions.MetricData)
		numOfGoroutine     = 0

		datahubControllerPrediction ApiPredictions.ControllerPrediction
	)

	var ctlKind ApiResources.Kind
	if value, ok := ApiResources.Kind_value[d.Kind]; ok {
		ctlKind = ApiResources.Kind(value)
	}

	datahubControllerPrediction = ApiPredictions.ControllerPrediction{
		ObjectMeta: NewObjectMeta(&d.ObjectMeta),
		Kind:       ctlKind,
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
		datahubControllerPrediction.PredictedRawData = append(datahubControllerPrediction.PredictedRawData, &receivedPredictionData)
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
		datahubControllerPrediction.PredictedUpperboundData = append(datahubControllerPrediction.PredictedUpperboundData, &receivedPredictionData)
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
		datahubControllerPrediction.PredictedLowerboundData = append(datahubControllerPrediction.PredictedLowerboundData, &receivedPredictionData)
	}

	return &datahubControllerPrediction
}
