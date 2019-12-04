package dispatcher

import (
	"encoding/json"
	"fmt"
	"time"

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
		for {
			msg, ok, err := queueConsumer.ReceiveJsonString(modelCompleteQueue)
			if err != nil {
				scope.Errorf("Get message from model complete queue error: %s", err.Error())
				break
			}
			if !ok {
				scope.Infof("No jobs found in queue %s, retry to get jobs next %v seconds",
					modelCompleteQueue, reconnectInterval)
				time.Sleep(time.Duration(reconnectInterval) * time.Second)
				break
			}

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
			if unitType == UnitTypeNode {
				nodeName := unit["name"].(string)
				modelMapper.RemoveModelInfo(unitType, dataGranularity, nodeName)

				mt := time.Now().Unix() - jobCreateTime
				scope.Infof("export node %s model time metric with granularity %s value %v",
					nodeName, dataGranularity, mt)
				metricExporter.ExportNodeMetricModelTime(nodeName, dataGranularity,
					time.Now().Unix(), float64(mt))
			} else if unitType == UnitTypePod {
				podNamespacedName := unit["namespaced_name"].(map[string]interface{})
				podNS := podNamespacedName["namespace"].(string)
				podName := podNamespacedName["name"].(string)
				modelMapper.RemoveModelInfo(unitType, dataGranularity,
					fmt.Sprintf("%s/%s", podNS, podName))

				mt := time.Now().Unix() - jobCreateTime
				scope.Infof("export pod %s model time metric with granularity %s value %v",
					fmt.Sprintf("%s/%s", podNS, podName), dataGranularity, mt)
				metricExporter.ExportPodMetricModelTime(podNS, podName, dataGranularity,
					time.Now().Unix(), float64(mt))
			} else if unitType == UnitTypeGPU {
				gpuHost := unit["host"].(string)
				gpuMinorNumber := unit["minor_number"].(string)
				modelMapper.RemoveModelInfo(unitType, dataGranularity,
					fmt.Sprintf("%s/%s", gpuHost, gpuMinorNumber))

				mt := time.Now().Unix() - jobCreateTime
				scope.Infof("export gpu %s model time metric with granularity %s value %v",
					fmt.Sprintf("%s/%s", gpuHost, gpuMinorNumber), dataGranularity, mt)
				metricExporter.ExportGPUMetricModelTime(gpuHost, gpuMinorNumber,
					dataGranularity, time.Now().Unix(), float64(mt))
			} else if unitType == UnitTypeApplication {
				appNamespacedName := unit["namespaced_name"].(map[string]interface{})
				appNS := appNamespacedName["namespace"].(string)
				appName := appNamespacedName["name"].(string)
				modelMapper.RemoveModelInfo(unitType, dataGranularity,
					fmt.Sprintf("%s/%s", appNS, appName))

				mt := time.Now().Unix() - jobCreateTime
				scope.Infof("export application %s model time metric with granularity %s value %v",
					fmt.Sprintf("%s/%s", appNS, appName), dataGranularity, mt)
				metricExporter.ExportApplicationMetricModelTime(appNS, appName,
					dataGranularity, time.Now().Unix(), float64(mt))
			} else if unitType == UnitTypeNamespace {
				namespaceName := unit["name"].(string)
				modelMapper.RemoveModelInfo(unitType, dataGranularity, namespaceName)

				mt := time.Now().Unix() - jobCreateTime
				scope.Infof("export namespace %s model time metric with granularity %s value %v",
					namespaceName, dataGranularity, mt)
				metricExporter.ExportNamespaceMetricModelTime(namespaceName, dataGranularity,
					time.Now().Unix(), float64(mt))
			} else if unitType == UnitTypeCluster {
				clusterName := unit["name"].(string)
				modelMapper.RemoveModelInfo(unitType, dataGranularity, clusterName)

				mt := time.Now().Unix() - jobCreateTime
				scope.Infof("export cluster %s model time metric with granularity %s value %v",
					clusterName, dataGranularity, mt)
				metricExporter.ExportClusterMetricModelTime(clusterName, dataGranularity,
					time.Now().Unix(), float64(mt))
			} else if unitType == UnitTypeController {
				appNamespacedName := unit["namespaced_name"].(map[string]interface{})
				appNS := appNamespacedName["namespace"].(string)
				appName := appNamespacedName["name"].(string)
				kind := unit["kind"].(string)
				modelMapper.RemoveModelInfo(unitType, dataGranularity,
					fmt.Sprintf("%s/%s", appNS, appName))

				mt := time.Now().Unix() - jobCreateTime
				scope.Infof("export controller %s model time metric with granularity %s value %v",
					fmt.Sprintf("%s/%s", appNS, appName), dataGranularity, mt)
				metricExporter.ExportControllerMetricModelTime(appNS, appName,
					kind, dataGranularity, time.Now().Unix(), float64(mt))
			}
		}
		queueConn.Close()
	}
}
