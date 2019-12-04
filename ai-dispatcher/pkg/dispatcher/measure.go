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
		scope.Infof("drift with MAPE of %s: %t", unitMeta["targetDisplayName"], mapeDrift)
		return mapeMetrics, mapeDrift
	} else if strings.ToLower(strings.TrimSpace(currentMeasure)) == "rmse" {
		scope.Infof("drift with RMSE of %s: %t", unitMeta["targetDisplayName"], rmseDrift)
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
	scope.Infof("start MAPE calculation for %s metric %v with granularity %v",
		targetDisplayName, metricType, granularity)
	measurementDataSet := stats.NewMeasurementDataSet(mData, pData, granularity)
	mape, mapeErr := stats.MAPE(measurementDataSet, granularity)
	if mapeErr == nil {
		scope.Infof("export MAPE value %v for %s metric %v with granularity %v", mape,
			targetDisplayName, metricType, granularity)
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
		scope.Infof("%s metric %v with granularity %v skip modeling due to not enough data points to calculate MAPE",
			targetDisplayName, metricType, granularity)
	} else if mapeErr != nil {
		metricsNeedToModel = append(metricsNeedToModel, metricType)
		scope.Infof(
			"MAPE calculation of %s metric %v with granularity %v failed due to: %s",
			targetDisplayName, metricType, granularity, mapeErr.Error())
	} else if mape > modelThreshold {
		metricsNeedToModel = append(metricsNeedToModel, metricType)
		shouldDrift = true
		scope.Infof("%s metric %v with granularity %v MAPE %v > %v",
			targetDisplayName, metricType, granularity, mape, modelThreshold)
	} else {
		scope.Infof("%s metric %v with granularity %v MAPE %v <= %v",
			targetDisplayName, metricType, granularity, mape, modelThreshold)
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
	scope.Infof("start RMSE calculation for %s metric %v with granularity %v",
		targetDisplayName, metricType, granularity)
	measurementDataSet := stats.NewMeasurementDataSet(mData, pData, granularity)
	rmse, rmseErr := stats.RMSE(measurementDataSet, metricType, granularity)
	if rmseErr == nil {
		scope.Infof("export RMSE value %v for %s metric %v with granularity %v", rmse,
			targetDisplayName, metricType, granularity)
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
		scope.Infof("%s metric %v with granularity %v skip modeling due to not enough data points to calculate RMSE",
			targetDisplayName, metricType, granularity)
	} else if rmseErr != nil {
		metricsNeedToModel = append(metricsNeedToModel, metricType)
		scope.Infof(
			"RMSE calculation of %s metric %v with granularity %v failed due to : %s",
			targetDisplayName, metricType, granularity, rmseErr.Error())
	} else if rmse > modelThreshold {
		metricsNeedToModel = append(metricsNeedToModel, metricType)
		shouldDrift = true
		scope.Infof("%s metric %v with granularity %v RMSE %v > %v",
			targetDisplayName, metricType, granularity, rmse, modelThreshold)
	} else {
		scope.Infof("%s metric %v with granularity %v RMSE %v <= %v",
			targetDisplayName, metricType, granularity, rmse, modelThreshold)
	}
	return metricsNeedToModel, shouldDrift
}
