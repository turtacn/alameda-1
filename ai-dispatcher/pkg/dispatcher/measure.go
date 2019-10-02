package dispatcher

import (
	"strings"

	"github.com/containers-ai/alameda/ai-dispatcher/pkg/metrics"
	"github.com/containers-ai/alameda/ai-dispatcher/pkg/queue"
	"github.com/containers-ai/alameda/ai-dispatcher/pkg/stats"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/spf13/viper"
)

func DriftEvaluation(unitType string, metricType datahub_v1alpha1.MetricType, granularity int64,
	mData []*datahub_v1alpha1.Sample, pData []*datahub_v1alpha1.Sample,
	unitMeta map[string]string, metricExporter *metrics.Exporter) ([]datahub_v1alpha1.MetricType, bool) {
	currentMeasure := viper.GetString("measurements.current")

	mapeMetrics, mapeDrift := mapeDriftEvaluation(unitType, metricType, granularity, mData, pData, unitMeta, metricExporter)
	rmseMetrics, rmseDrift := rmseDriftEvaluation(unitType, metricType, granularity, mData, pData, unitMeta, metricExporter)

	if strings.ToLower(strings.TrimSpace(currentMeasure)) == "mape" {
		scope.Infof("drift with MAPE of %s: %t", unitMeta["targetDisplayName"], mapeDrift)
		return mapeMetrics, mapeDrift
	} else if strings.ToLower(strings.TrimSpace(currentMeasure)) == "rmse" {
		scope.Infof("drift with RMSE of %s: %t", unitMeta["targetDisplayName"], rmseDrift)
		return rmseMetrics, rmseDrift
	}

	return []datahub_v1alpha1.MetricType{}, false
}

func mapeDriftEvaluation(unitType string, metricType datahub_v1alpha1.MetricType, granularity int64,
	mData []*datahub_v1alpha1.Sample, pData []*datahub_v1alpha1.Sample,
	unitMeta map[string]string, metricExporter *metrics.Exporter) ([]datahub_v1alpha1.MetricType, bool) {
	shouldDrift := false
	modelThreshold := viper.GetFloat64("measurements.mape.threshold")
	metricsNeedToModel := []datahub_v1alpha1.MetricType{}
	targetDisplayName := unitMeta["targetDisplayName"]
	scope.Infof("start MAPE calculation for %s metric %v with granularity %v",
		targetDisplayName, metricType, granularity)
	measurementDataSet := stats.NewMeasurementDataSet(mData, pData, granularity)
	mape, mapeErr := stats.MAPE(measurementDataSet)
	if mapeErr == nil {
		scope.Infof("export MAPE value %v for %s metric %v with granularity %v", mape,
			targetDisplayName, metricType, granularity)
		if unitType == UnitTypeNode {
			metricExporter.SetNodeMetricMAPE(unitMeta["nodeName"],
				queue.GetMetricLabel(metricType), queue.GetGranularityStr(granularity), mape)
		} else if unitType == UnitTypePod {
			metricExporter.SetContainerMetricMAPE(unitMeta["podNS"], unitMeta["podName"], unitMeta["containerName"],
				queue.GetMetricLabel(metricType), queue.GetGranularityStr(granularity), mape)
		} else if unitType == UnitTypeGPU {
			metricExporter.SetGPUMetricMAPE(unitMeta["gpuHost"], unitMeta["gpuMinorNumber"],
				queue.GetMetricLabel(metricType), queue.GetGranularityStr(granularity), mape)
		}
	}
	if mapeErr != nil {
		metricsNeedToModel = append(metricsNeedToModel, metricType)
		scope.Infof(
			"MAPE calculation of %s metric %v with granularity %v failed due to: %s",
			targetDisplayName, metricType, granularity, mapeErr.Error())
	} else if mape > modelThreshold {
		metricsNeedToModel = append(metricsNeedToModel, metricType)
		scope.Infof("%s metric %v with granularity %v MAPE %v > %v",
			targetDisplayName, metricType, granularity, mape, modelThreshold)
		shouldDrift = true
	} else {
		scope.Infof("%s metric %v with granularity %v MAPE %v <= %v",
			targetDisplayName, metricType, granularity, mape, modelThreshold)
	}
	return metricsNeedToModel, shouldDrift
}

func rmseDriftEvaluation(unitType string, metricType datahub_v1alpha1.MetricType, granularity int64,
	mData []*datahub_v1alpha1.Sample, pData []*datahub_v1alpha1.Sample,
	unitMeta map[string]string, metricExporter *metrics.Exporter) ([]datahub_v1alpha1.MetricType, bool) {
	shouldDrift := false
	modelThreshold := viper.GetFloat64("measurements.rmse.threshold")
	metricsNeedToModel := []datahub_v1alpha1.MetricType{}
	targetDisplayName := unitMeta["targetDisplayName"]
	scope.Infof("start RMSE calculation for %s metric %v with granularity %v",
		targetDisplayName, metricType, granularity)
	measurementDataSet := stats.NewMeasurementDataSet(mData, pData, granularity)
	rmse, rmseErr := stats.RMSE(measurementDataSet, metricType)
	if rmseErr == nil {
		scope.Infof("export RMSE value %v for %s metric %v with granularity %v", rmse,
			targetDisplayName, metricType, granularity)
		if unitType == UnitTypeNode {
			metricExporter.SetNodeMetricRMSE(unitMeta["nodeName"],
				queue.GetMetricLabel(metricType), queue.GetGranularityStr(granularity), rmse)
		} else if unitType == UnitTypePod {
			metricExporter.SetContainerMetricRMSE(unitMeta["podNS"], unitMeta["podName"], unitMeta["containerName"],
				queue.GetMetricLabel(metricType), queue.GetGranularityStr(granularity), rmse)
		} else if unitType == UnitTypeGPU {
			metricExporter.SetGPUMetricRMSE(unitMeta["gpuHost"], unitMeta["gpuMinorNumber"],
				queue.GetMetricLabel(metricType), queue.GetGranularityStr(granularity), rmse)
		}
	}
	if rmseErr != nil {
		metricsNeedToModel = append(metricsNeedToModel, metricType)
		scope.Infof(
			"RMSE calculation of %s metric %v with granularity %v failed due to : %s",
			targetDisplayName, metricType, granularity, rmseErr.Error())
	} else if rmse > modelThreshold {
		metricsNeedToModel = append(metricsNeedToModel, metricType)
		scope.Infof("%s metric %v with granularity %v RMSE %v > %v",
			targetDisplayName, metricType, granularity, rmse, modelThreshold)
		shouldDrift = true
	} else {
		scope.Infof("%s metric %v with granularity %v RMSE %v <= %v",
			targetDisplayName, metricType, granularity, rmse, modelThreshold)
	}
	return metricsNeedToModel, shouldDrift
}
