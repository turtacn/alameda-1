package stats

import (
	"fmt"
	"math"
	"strconv"

	stats_errors "github.com/containers-ai/alameda/ai-dispatcher/pkg/stats/errors"
	"github.com/containers-ai/alameda/pkg/utils"
	"github.com/containers-ai/alameda/pkg/utils/log"
	datahub_common "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	datahub_predictions "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/predictions"
	"github.com/spf13/viper"
)

type MeasurementData struct {
	predictData data
	metricData  data
}

type data struct {
	time  int64
	value float64
}

var scope = log.RegisterScope("statistic", "ai-dispatcher statistic", 0)

func NewMeasurementDataSet(metricSamples []*datahub_common.Sample,
	predictSamples []*datahub_predictions.Sample, granularity int64) map[int64]*MeasurementData {
	scope.Debugf("NewMeasurementDataSet metric samples: %s", utils.InterfaceToString(metricSamples))
	scope.Debugf("NewMeasurementDataSet predict samples: %s", utils.InterfaceToString(predictSamples))
	measurementDataSet := map[int64]*MeasurementData{}
	for _, metricSample := range metricSamples {
		//for time := range measurementData {
		for _, predictSample := range predictSamples {
			predictValue, err := strconv.ParseFloat(predictSample.GetNumValue(), 64)
			if err != nil {
				scope.Errorf("parse predict value failed: %s", err.Error())
				continue
			}
			//scope.Debugf("%v", predictSample.GetTime().GetSeconds()-metricSample.GetTime().GetSeconds())
			if math.Abs(float64(predictSample.GetTime().GetSeconds()-metricSample.GetTime().GetSeconds())) <
				float64(granularity) {
				metricValue, err := strconv.ParseFloat(metricSample.GetNumValue(), 64)
				if err != nil {
					scope.Errorf("parse metric value failed: %s", err.Error())
					continue
				}
				measurementDataSet[predictSample.GetTime().GetSeconds()] = &MeasurementData{
					predictData: data{
						time:  predictSample.GetTime().GetSeconds(),
						value: predictValue,
					},
					metricData: data{
						time:  metricSample.GetTime().GetSeconds(),
						value: metricValue,
					},
				}
				break
			}
		}
	}
	if len(measurementDataSet) == 0 {
		scope.Warnf("No measurementDataSet generated due to no data overlapped between metric and prediction")
	}
	return measurementDataSet
}

func MAPE(measurementDataSet map[int64]*MeasurementData,
	granularity int64) (float64, error) {
	nPts := 0.0
	result := 0.0
	scope.Debugf("Start MAPE calculation")
	for _, data := range measurementDataSet {
		metricValue := data.GetMetricData()
		if metricValue == 0 {
			scope.Warnf("Real value is 0 in MAPE measurement, skip this point")
			continue
		}
		predictValue := data.GetPredictData()
		nPts = nPts + 1
		deltaRatio := math.Abs(predictValue-metricValue) / metricValue
		result = result + deltaRatio
		scope.Debugf("(real value: %v, predict value: %v, delta ratio: %v)",
			metricValue, predictValue, deltaRatio)
	}
	if nPts == 0 {
		return 0, fmt.Errorf(stats_errors.ErrorNoDataPoints)
	}
	if granularity == 30 &&
		nPts < viper.GetFloat64("measurements.minimumDataPoints") {
		return 0, fmt.Errorf(stats_errors.ErrorDataPointsNotEnough)
	}
	resultPercentage := 100 * (result / nPts)
	scope.Debugf("MAPE calculation result: %v with sum %v and %v points",
		resultPercentage, result, nPts)
	return resultPercentage, nil
}

func RMSE(measurementDataSet map[int64]*MeasurementData,
	metricType datahub_common.MetricType, granularity int64) (float64, error) {
	nPts := 0.0
	result := 0.0
	normalize := 1.0
	if metricType == datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE {
		normalize = viper.GetFloat64("measurements.rmse.normalization.cpu")
	} else if metricType == datahub_common.MetricType_MEMORY_USAGE_BYTES {
		normalize = viper.GetFloat64("measurements.rmse.normalization.memory")
	} else if metricType == datahub_common.MetricType_DUTY_CYCLE {
		normalize = viper.GetFloat64("measurements.rmse.normalization.dutyCycle")
	}
	scope.Debugf("Start RMSE calculation")
	for _, data := range measurementDataSet {
		metricValue := data.GetMetricData()
		predictValue := data.GetPredictData()
		nPts = nPts + 1
		square := math.Pow(math.Abs((predictValue-metricValue)/normalize), 2)
		result = result + square
		scope.Debugf("(Use %v to normalize. normalized real value: %v, normalized predict value: %v, square %v)",
			normalize, metricValue/normalize, predictValue/normalize, square)
	}
	if nPts == 0 {
		return 0, fmt.Errorf(stats_errors.ErrorNoDataPoints)
	}
	if granularity == 30 &&
		nPts < viper.GetFloat64("measurements.minimumDataPoints") {
		return 0, fmt.Errorf(stats_errors.ErrorDataPointsNotEnough)
	}
	resultVal := math.Sqrt(result / nPts)
	scope.Debugf("RMSE calculation result: %v with sum square %v and %v points",
		resultVal, result, nPts)
	return resultVal, nil
}

func (mData MeasurementData) GetMetricData() float64 {
	return mData.metricData.value
}

func (mData MeasurementData) GetPredictData() float64 {
	return mData.predictData.value
}
