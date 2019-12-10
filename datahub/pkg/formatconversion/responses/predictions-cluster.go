package responses

import (
	DaoPredictionTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/predictions/types"
	FormatEnum "github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	ApiPredictions "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/predictions"
)

type ClusterPredictionExtended struct {
	*DaoPredictionTypes.ClusterPrediction
}

func (d *ClusterPredictionExtended) ProducePredictions() *ApiPredictions.ClusterPrediction {
	var (
		rawDataChan        = make(chan ApiPredictions.MetricData)
		upperBoundDataChan = make(chan ApiPredictions.MetricData)
		lowerBoundDataChan = make(chan ApiPredictions.MetricData)
		numOfGoroutine     = 0

		datahubClusterPrediction ApiPredictions.ClusterPrediction
	)

	datahubClusterPrediction = ApiPredictions.ClusterPrediction{
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
		datahubClusterPrediction.PredictedRawData = append(datahubClusterPrediction.PredictedRawData, &receivedPredictionData)
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
		datahubClusterPrediction.PredictedUpperboundData = append(datahubClusterPrediction.PredictedUpperboundData, &receivedPredictionData)
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
		datahubClusterPrediction.PredictedLowerboundData = append(datahubClusterPrediction.PredictedLowerboundData, &receivedPredictionData)
	}

	return &datahubClusterPrediction
}
