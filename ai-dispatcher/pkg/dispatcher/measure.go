package dispatcher

import (
	"strings"
	"time"

	"github.com/containers-ai/alameda/ai-dispatcher/pkg/metrics"
	"github.com/containers-ai/alameda/ai-dispatcher/pkg/queue"
	"github.com/containers-ai/alameda/ai-dispatcher/pkg/stats"
	stats_errors "github.com/containers-ai/alameda/ai-dispatcher/pkg/stats/errors"
	datahub_common "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	datahub_predictions "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/predictions"
	"github.com/spf13/viper"
)

func DriftEvaluation(unitType string, metricType datahub_common.MetricType, granularity int64,
	mData []*datahub_common.Sample, pData []*datahub_predictions.Sample,
	unitMeta map[string]string, metricExporter *metrics.Exporter) ([]datahub_common.MetricType, bool) {
	currentMeasure := viper.GetString("measurements.current")

	mapeMetrics, mapeDrift := mapeDriftEvaluation(unitType, metricType, granularity, mData, pData, unitMeta, metricExporter)
	rmseMetrics, rmseDrift := rmseDriftEvaluation(unitType, metricType, granularity, mData, pData, unitMeta, metricExporter)

	if strings.ToLower(strings.TrimSpace(currentMeasure)) == "mape" {
		scope.Infof("%s drift with MAPE: %t", unitMeta["targetDisplayName"], mapeDrift)
		return mapeMetrics, mapeDrift
	} else if strings.ToLower(strings.TrimSpace(currentMeasure)) == "rmse" {
		scope.Infof("%s drift with MAPE: %t", unitMeta["targetDisplayName"], rmseDrift)
		return rmseMetrics, rmseDrift
	}

	return []datahub_common.MetricType{}, false
}

func mapeDriftEvaluation(unitType string, metricType datahub_common.MetricType, granularity int64,
	mData []*datahub_common.Sample, pData []*datahub_predictions.Sample,
	unitMeta map[string]string, metricExporter *metrics.Exporter) ([]datahub_common.MetricType, bool) {
	shouldDrift := false
	modelThreshold := viper.GetFloat64("measurements.mape.threshold")
	metricsNeedToModel := []datahub_common.MetricType{}
	targetDisplayName := unitMeta["targetDisplayName"]
	scope.Infof("%s Start MAPE calculation for metric %v",
		targetDisplayName, metricType)
	scope.Infof("%s Metric data: %v",targetDisplayName, mData)
	scope.Infof("%s Predict data: %v",targetDisplayName, pData)

	measurementDataSet := stats.NewMeasurementDataSet(mData, pData, granularity)
	mape, mapeErr := stats.MAPE(measurementDataSet, granularity)
	if mapeErr == nil {
		scope.Infof("%s Export MAPE value %v metric %v",
			targetDisplayName, mape, metricType)
		if unitType == UnitTypeNode {
			metricExporter.SetNodeMetricMAPE(unitMeta["nodeName"],
				queue.GetMetricLabel(metricType), queue.GetGranularityStr(granularity), time.Now().Unix(), mape)
		} else if unitType == UnitTypePod {
			metricExporter.SetContainerMetricMAPE(unitMeta["podNS"], unitMeta["podName"], unitMeta["containerName"],
				queue.GetMetricLabel(metricType), queue.GetGranularityStr(granularity), time.Now().Unix(), mape)
		} else if unitType == UnitTypeGPU {
			metricExporter.SetGPUMetricMAPE(unitMeta["gpuHost"], unitMeta["gpuMinorNumber"],
				queue.GetMetricLabel(metricType), queue.GetGranularityStr(granularity), time.Now().Unix(), mape)
		} else if unitType == UnitTypeApplication {
			metricExporter.SetApplicationMetricMAPE(unitMeta["appNS"], unitMeta["appName"],
				queue.GetMetricLabel(metricType), queue.GetGranularityStr(granularity), time.Now().Unix(), mape)
		} else if unitType == UnitTypeNamespace {
			metricExporter.SetNamespaceMetricMAPE(unitMeta["nsName"],
				queue.GetMetricLabel(metricType), queue.GetGranularityStr(granularity), time.Now().Unix(), mape)
		} else if unitType == UnitTypeController {
			metricExporter.SetControllerMetricMAPE(unitMeta["controllerNS"], unitMeta["controllerName"],
				unitMeta["controllerKind"], queue.GetMetricLabel(metricType),
				queue.GetGranularityStr(granularity), time.Now().Unix(), mape)
		} else if unitType == UnitTypeCluster {
			metricExporter.SetClusterMetricMAPE(unitMeta["clusterName"], queue.GetMetricLabel(metricType),
				queue.GetGranularityStr(granularity), time.Now().Unix(), mape)
		}
	}

	if mapeErr != nil && stats_errors.DataPointsNotEnough(mapeErr) {
		scope.Infof("%s metric %v skip modeling due to not enough data points to calculate MAPE",
			targetDisplayName, metricType)
	} else if mapeErr != nil {
		metricsNeedToModel = append(metricsNeedToModel, metricType)
		scope.Infof(
			"%s MAPE calculation of metric %v failed due to: %s",
			targetDisplayName, metricType, mapeErr.Error())
	} else if mape > modelThreshold {
		metricsNeedToModel = append(metricsNeedToModel, metricType)
		shouldDrift = true
		scope.Infof("%s MAPE of metric %v  %v > %v",
			targetDisplayName, metricType, mape, modelThreshold)
	} else {
		scope.Infof("%s MAPE of metric %v %v <= %v",
			targetDisplayName, metricType, mape, modelThreshold)
	}
	return metricsNeedToModel, shouldDrift
}

func rmseDriftEvaluation(unitType string, metricType datahub_common.MetricType, granularity int64,
	mData []*datahub_common.Sample, pData []*datahub_predictions.Sample,
	unitMeta map[string]string, metricExporter *metrics.Exporter) ([]datahub_common.MetricType, bool) {
	shouldDrift := false
	modelThreshold := viper.GetFloat64("measurements.rmse.threshold")
	metricsNeedToModel := []datahub_common.MetricType{}
	targetDisplayName := unitMeta["targetDisplayName"]
	scope.Infof("%s Start RMSE calculation for metric %v",
		targetDisplayName, metricType)
	measurementDataSet := stats.NewMeasurementDataSet(mData, pData, granularity)
	rmse, rmseErr := stats.RMSE(measurementDataSet, metricType, granularity)
	if rmseErr == nil {
		scope.Infof("%s Export RMSE value %v for metric %v",
			targetDisplayName, rmse, metricType)
		if unitType == UnitTypeNode {
			metricExporter.SetNodeMetricRMSE(unitMeta["nodeName"],
				queue.GetMetricLabel(metricType), queue.GetGranularityStr(granularity), time.Now().Unix(), rmse)
		} else if unitType == UnitTypePod {
			metricExporter.SetContainerMetricRMSE(unitMeta["podNS"], unitMeta["podName"], unitMeta["containerName"],
				queue.GetMetricLabel(metricType), queue.GetGranularityStr(granularity), time.Now().Unix(), rmse)
		} else if unitType == UnitTypeGPU {
			metricExporter.SetGPUMetricRMSE(unitMeta["gpuHost"], unitMeta["gpuMinorNumber"],
				queue.GetMetricLabel(metricType), queue.GetGranularityStr(granularity), time.Now().Unix(), rmse)
		} else if unitType == UnitTypeApplication {
			metricExporter.SetApplicationMetricRMSE(unitMeta["appNS"], unitMeta["appName"],
				queue.GetMetricLabel(metricType), queue.GetGranularityStr(granularity), time.Now().Unix(), rmse)
		} else if unitType == UnitTypeNamespace {
			metricExporter.SetNamespaceMetricRMSE(unitMeta["nsName"],
				queue.GetMetricLabel(metricType), queue.GetGranularityStr(granularity), time.Now().Unix(), rmse)
		} else if unitType == UnitTypeController {
			metricExporter.SetControllerMetricRMSE(unitMeta["controllerNS"], unitMeta["controllerName"],
				unitMeta["controllerKind"], queue.GetMetricLabel(metricType),
				queue.GetGranularityStr(granularity), time.Now().Unix(), rmse)
		} else if unitType == UnitTypeCluster {
			metricExporter.SetClusterMetricRMSE(unitMeta["clusterName"],
				queue.GetMetricLabel(metricType), queue.GetGranularityStr(granularity), time.Now().Unix(), rmse)
		}
	}

	if rmseErr != nil && stats_errors.DataPointsNotEnough(rmseErr) {
		scope.Infof("%s metric %v skip modeling due to not enough data points to calculate RMSE",
			targetDisplayName, metricType)
	} else if rmseErr != nil {
		metricsNeedToModel = append(metricsNeedToModel, metricType)
		scope.Infof(
			"%s RMSE calculation of metric %v failed due to : %s",
			targetDisplayName, metricType, rmseErr.Error())
	} else if rmse > modelThreshold {
		metricsNeedToModel = append(metricsNeedToModel, metricType)
		shouldDrift = true
		scope.Infof("%s RMSE of metric %v %v > %v",
			targetDisplayName, metricType, rmse, modelThreshold)
	} else {
		scope.Infof("%s RMSE of metric %v %v <= %v",
			targetDisplayName, metricType, rmse, modelThreshold)
	}
	return metricsNeedToModel, shouldDrift
}
