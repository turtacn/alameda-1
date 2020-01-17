package dispatcher

import (
	"encoding/json"
	"time"

	"github.com/containers-ai/alameda/ai-dispatcher/consts"
	"github.com/containers-ai/alameda/ai-dispatcher/pkg/metrics"
	"github.com/containers-ai/alameda/ai-dispatcher/pkg/queue"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func ModelCompleteNotification(modelMapper *ModelMapper,
	datahubGrpcCn *grpc.ClientConn, metricExporter *metrics.Exporter) {

	reconnectInterval := viper.GetInt64("queue.consumer.reconnectInterval")
	queueConnRetryItvMS := viper.GetInt64("queue.retry.connectIntervalMs")

	modelCompleteQueue := "model_complete"
	queueURL := viper.GetString("queue.url")
	for {
		queueConn := queue.GetQueueConn(queueURL, queueConnRetryItvMS)
		queueConsumer := queue.NewRabbitMQConsumer(queueConn)
		msgCH, err := queueConsumer.ConsumeJsonString(modelCompleteQueue)
		if err != nil {
			if queueConn != nil {
				queueConn.Close()
			}
			scope.Warnf("Consume message from model complete queue error: %s", err.Error())
			time.Sleep(time.Duration(reconnectInterval) * time.Second)
			continue
		}
		for cmsg := range msgCH {
			msg := string(cmsg.Body)
			var msgMap map[string]interface{}
			msgByte := []byte(msg)
			if err := json.Unmarshal(msgByte, &msgMap); err != nil {
				scope.Errorf("decode model complete job from queue failed: %s", err.Error())
				break
			}

			unit := msgMap["unit"].(map[string]interface{})
			unitType := msgMap["unit_type"].(string)
			dataGranularity := msgMap["data_granularity"].(string)
			jobCreateTime := int64(msgMap["job_create_time"].(float64))
			if unitType == consts.UnitTypeNode {
				nodeName := unit["name"].(string)
				clusterID := msgMap["cluster_name"].(string)
				metricType := msgMap["metric_type_str"].(string)
				modelMapper.RemoveModelInfo(clusterID, unitType, dataGranularity, metricType, map[string]string{
					"name": nodeName,
				})

				mt := time.Now().Unix() - jobCreateTime
				scope.Infof("[NODE][%s][%s] Export metric %s model time value %v",
					dataGranularity, nodeName, metricType, mt)
				metricExporter.ExportNodeMetricModelTime(clusterID, nodeName, dataGranularity, metricType,
					time.Now().Unix(), float64(mt))
			} else if unitType == consts.UnitTypePod {
				podNamespacedName := unit["namespaced_name"].(map[string]interface{})
				podNS := podNamespacedName["namespace"].(string)
				podName := podNamespacedName["name"].(string)
				clusterID := msgMap["cluster_name"].(string)
				metricType := msgMap["metric_type_str"].(string)
				ctName := msgMap["container_name"].(string)
				modelMapper.RemoveModelInfo(clusterID, unitType, dataGranularity, metricType, map[string]string{
					"namespace":     podNS,
					"name":          podName,
					"containerName": ctName,
				})

				mt := time.Now().Unix() - jobCreateTime
				scope.Infof("[POD][%s][%s/%s/%s] Export metric %s model time with value %v",
					dataGranularity, podNS, podName, ctName, metricType, mt)
				metricExporter.ExportContainerMetricModelTime(clusterID, podNS, podName, ctName, dataGranularity, metricType,
					time.Now().Unix(), float64(mt))
			} else if unitType == consts.UnitTypeGPU {
				gpuHost := unit["host"].(string)
				gpuMinorNumber := unit["minor_number"].(string)
				clusterID := msgMap["cluster_name"].(string)
				metricType := msgMap["metric_type_str"].(string)
				modelMapper.RemoveModelInfo(clusterID, unitType, dataGranularity,
					metricType, map[string]string{
						"host":        gpuHost,
						"minorNumber": gpuMinorNumber,
					})
				mt := time.Now().Unix() - jobCreateTime
				scope.Infof("[GPU][%s][%s/%s] Export metric %s model time value %v",
					dataGranularity, gpuHost, gpuMinorNumber, metricType, mt)
				metricExporter.ExportGPUMetricModelTime(clusterID, gpuHost, gpuMinorNumber,
					dataGranularity, metricType, time.Now().Unix(), float64(mt))
			} else if unitType == consts.UnitTypeApplication {
				appNamespacedName := unit["namespaced_name"].(map[string]interface{})
				appNS := appNamespacedName["namespace"].(string)
				appName := appNamespacedName["name"].(string)
				clusterID := msgMap["cluster_name"].(string)
				metricType := msgMap["metric_type_str"].(string)
				modelMapper.RemoveModelInfo(clusterID, unitType, dataGranularity,
					metricType, map[string]string{
						"namespace": appNS,
						"name":      appName,
					})

				mt := time.Now().Unix() - jobCreateTime
				scope.Infof("[APPLICATION][%s][%s/%s] Export metric %s model time value %v",
					dataGranularity, appNS, appName, metricType, mt)
				metricExporter.ExportApplicationMetricModelTime(clusterID, appNS, appName,
					dataGranularity, metricType, time.Now().Unix(), float64(mt))
			} else if unitType == consts.UnitTypeNamespace {
				namespaceName := unit["name"].(string)
				clusterID := msgMap["cluster_name"].(string)
				metricType := msgMap["metric_type_str"].(string)
				modelMapper.RemoveModelInfo(clusterID, unitType, dataGranularity, metricType, map[string]string{
					"name": namespaceName,
				})

				mt := time.Now().Unix() - jobCreateTime
				scope.Infof("[NAMESPACE][%s][%s] Export metric %s model time value %v",
					dataGranularity, namespaceName, metricType, mt)
				metricExporter.ExportNamespaceMetricModelTime(clusterID, namespaceName, dataGranularity, metricType,
					time.Now().Unix(), float64(mt))
			} else if unitType == consts.UnitTypeCluster {
				clusterName := unit["name"].(string)
				clusterID := msgMap["cluster_name"].(string)
				metricType := msgMap["metric_type_str"].(string)
				modelMapper.RemoveModelInfo(clusterID, unitType, dataGranularity, metricType, map[string]string{
					"name": clusterName,
				})

				mt := time.Now().Unix() - jobCreateTime
				scope.Infof("[CLUSTER][%s][%s] Export metric %s model time value %v",
					dataGranularity, clusterName, metricType, mt)
				metricExporter.ExportClusterMetricModelTime(clusterName, dataGranularity, metricType,
					time.Now().Unix(), float64(mt))
			} else if unitType == consts.UnitTypeController {
				controllerNamespacedName := unit["namespaced_name"].(map[string]interface{})
				controllerNS := controllerNamespacedName["namespace"].(string)
				controllerName := controllerNamespacedName["name"].(string)
				kind := unit["kind"].(string)
				clusterID := msgMap["cluster_name"].(string)
				metricType := msgMap["metric_type_str"].(string)
				modelMapper.RemoveModelInfo(clusterID, unitType, dataGranularity,
					metricType, map[string]string{
						"namespace": controllerNS,
						"name":      controllerName,
						"kind":      kind,
					})

				mt := time.Now().Unix() - jobCreateTime
				scope.Infof("[CONTROLLER][%s][%s][%s/%s] Export metric %s model time value %v",
					kind, dataGranularity, controllerNS, controllerName, metricType, mt)
				metricExporter.ExportControllerMetricModelTime(clusterID, controllerNS, controllerName,
					kind, dataGranularity, metricType, time.Now().Unix(), float64(mt))
			}
		}
		scope.Warnf("Retry construct consume model complete channel")
		if queueConn != nil {
			queueConn.Close()
		}
		time.Sleep(time.Duration(reconnectInterval) * time.Second)
	}
}
