package dispatcher

import (
	"strings"
	"time"

	"github.com/containers-ai/alameda/ai-dispatcher/consts"
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
		scope.Infof("%s metric %s drift with MAPE: %t", unitMeta["targetDisplayName"], metricType, mapeDrift)
		return mapeMetrics, mapeDrift
	} else if strings.ToLower(strings.TrimSpace(currentMeasure)) == "rmse" {
		scope.Infof("%s metric %s drift with MAPE: %t", unitMeta["targetDisplayName"], metricType, rmseDrift)
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
	scope.Debugf("%s Metric data: %v", targetDisplayName, mData)
	scope.Debugf("%s Predict data: %v", targetDisplayName, pData)

	measurementDataSet := stats.NewMeasurementDataSet(mData, pData, granularity)
	mape, mapeErr := stats.MAPE(measurementDataSet, granularity)
	if mapeErr == nil {
		scope.Infof("%s Export MAPE value %v metric %v",
			targetDisplayName, mape, metricType)
		if unitType == consts.UnitTypeNode {
			metricExporter.SetNodeMetricMAPE(unitMeta["clusterID"], unitMeta["nodeName"],
				queue.GetGranularityStr(granularity), metricType.String(), time.Now().Unix(), mape)
		} else if unitType == consts.UnitTypePod {
			metricExporter.SetContainerMetricMAPE(unitMeta["clusterID"], unitMeta["podNS"], unitMeta["podName"], unitMeta["containerName"],
				queue.GetGranularityStr(granularity), metricType.String(), time.Now().Unix(), mape)
		} else if unitType == consts.UnitTypeGPU {
			metricExporter.SetGPUMetricMAPE(unitMeta["clusterID"], unitMeta["gpuHost"], unitMeta["gpuMinorNumber"],
				queue.GetGranularityStr(granularity), metricType.String(), time.Now().Unix(), mape)
		} else if unitType == consts.UnitTypeApplication {
			metricExporter.SetApplicationMetricMAPE(unitMeta["clusterID"], unitMeta["applicationNS"], unitMeta["applicationName"],
				queue.GetGranularityStr(granularity), metricType.String(), time.Now().Unix(), mape)
		} else if unitType == consts.UnitTypeNamespace {
			metricExporter.SetNamespaceMetricMAPE(unitMeta["clusterID"], unitMeta["namespaceName"],
				queue.GetGranularityStr(granularity), metricType.String(), time.Now().Unix(), mape)
		} else if unitType == consts.UnitTypeController {
			metricExporter.SetControllerMetricMAPE(unitMeta["clusterID"], unitMeta["controllerNS"], unitMeta["controllerName"],
				unitMeta["controllerKind"],
				queue.GetGranularityStr(granularity), metricType.String(), time.Now().Unix(), mape)
		} else if unitType == consts.UnitTypeCluster {
			metricExporter.SetClusterMetricMAPE(unitMeta["clusterName"],
				queue.GetGranularityStr(granularity), metricType.String(), time.Now().Unix(), mape)
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
		if unitType == consts.UnitTypeNode {
			metricExporter.SetNodeMetricRMSE(unitMeta["clusterID"], unitMeta["nodeName"],
				queue.GetGranularityStr(granularity), metricType.String(), time.Now().Unix(), rmse)
		} else if unitType == consts.UnitTypePod {
			metricExporter.SetContainerMetricRMSE(unitMeta["clusterID"], unitMeta["podNS"], unitMeta["podName"], unitMeta["containerName"],
				queue.GetGranularityStr(granularity), metricType.String(), time.Now().Unix(), rmse)
		} else if unitType == consts.UnitTypeGPU {
			metricExporter.SetGPUMetricRMSE(unitMeta["clusterID"], unitMeta["gpuHost"], unitMeta["gpuMinorNumber"],
				queue.GetGranularityStr(granularity), metricType.String(), time.Now().Unix(), rmse)
		} else if unitType == consts.UnitTypeApplication {
			metricExporter.SetApplicationMetricRMSE(unitMeta["clusterID"], unitMeta["applicationNS"], unitMeta["applicationName"],
				queue.GetGranularityStr(granularity), metricType.String(), time.Now().Unix(), rmse)
		} else if unitType == consts.UnitTypeNamespace {
			metricExporter.SetNamespaceMetricRMSE(unitMeta["clusterID"], unitMeta["namespaceName"],
				queue.GetGranularityStr(granularity), metricType.String(), time.Now().Unix(), rmse)
		} else if unitType == consts.UnitTypeController {
			metricExporter.SetControllerMetricRMSE(unitMeta["clusterID"], unitMeta["controllerNS"], unitMeta["controllerName"],
				unitMeta["controllerKind"],
				queue.GetGranularityStr(granularity), metricType.String(), time.Now().Unix(), rmse)
		} else if unitType == consts.UnitTypeCluster {
			metricExporter.SetClusterMetricRMSE(unitMeta["clusterName"], queue.GetGranularityStr(granularity),
				metricType.String(), time.Now().Unix(), rmse)
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
