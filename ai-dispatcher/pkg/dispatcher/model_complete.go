package dispatcher

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/containers-ai/alameda/ai-dispatcher/pkg/metrics"
	"github.com/containers-ai/alameda/ai-dispatcher/pkg/queue"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func ModelCompleteNotification(modelMapper *ModelMapper,
	datahubGrpcCn *grpc.ClientConn, metricExporter *metrics.Exporter) {

	datahubServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(datahubGrpcCn)
	predictJobSender := NewPredictJobSender(datahubGrpcCn)
	reconnectInterval := viper.GetInt64("queue.consumer.reconnectInterval")
	queueConnRetryItvMS := viper.GetInt64("queue.retry.connectIntervalMs")

	modelCompleteQueue := "model_complete"
	queueURL := viper.GetString("queue.url")
	for {
		queueConn := queue.GetQueueConn(queueURL, queueConnRetryItvMS)
		queueConsumer := queue.NewRabbitMQConsumer(queueConn)
		queueSender := queue.NewRabbitMQSender(queueConn)
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
			if unitType == UnitTypeNode {
				nodeName := unit["name"].(string)
				modelInfo := modelMapper.GetModelInfo(unitType, dataGranularity, nodeName)
				if modelInfo != nil {
					mt := time.Now().Unix() - modelInfo.Timestamp
					scope.Infof("export node %s model time metric with granularity %s value %v",
						nodeName, dataGranularity, mt)
					metricExporter.SetNodeMetricModelTime(nodeName, dataGranularity, float64(mt))
					modelMapper.RemoveModelInfo(unitType, dataGranularity, nodeName)
				}

				res, err := datahubServiceClnt.ListNodes(context.Background(),
					&datahub_v1alpha1.ListNodesRequest{
						NodeNames: []string{nodeName},
					})
				if err == nil {
					nodes := res.GetNodes()
					if len(nodes) > 0 {
						scope.Infof("node %s model job completed and send predict job for granularity %s",
							nodeName, dataGranularity)
						predictJobSender.SendNodePredictJobs(nodes, queueSender,
							unitType, queue.GetGranularitySec(strings.Trim(dataGranularity, " ")))
					}
				}
			} else if unitType == UnitTypePod {
				podNamespacedName := unit["namespaced_name"].(map[string]interface{})
				podNS := podNamespacedName["namespace"].(string)
				podName := podNamespacedName["name"].(string)
				modelInfo := modelMapper.GetModelInfo(unitType, dataGranularity, fmt.Sprintf("%s/%s", podNS, podName))
				if modelInfo != nil {
					mt := time.Now().Unix() - modelInfo.Timestamp
					scope.Infof("export pod %s model time metric with granularity %s value %v",
						fmt.Sprintf("%s/%s", podNS, podName), dataGranularity, mt)
					metricExporter.SetPodMetricModelTime(podNS, podName, dataGranularity, float64(mt))
					modelMapper.RemoveModelInfo(unitType, dataGranularity,
						fmt.Sprintf("%s/%s", podNS, podName))
				}

				res, err := datahubServiceClnt.ListAlamedaPods(context.Background(),
					&datahub_v1alpha1.ListAlamedaPodsRequest{
						NamespacedName: &datahub_v1alpha1.NamespacedName{
							Namespace: podNS,
							Name:      podName,
						},
					})
				if err == nil {
					pods := res.GetPods()
					if len(pods) > 0 {
						scope.Infof("pod %s/%s model job completed and send predict job for granularity %s",
							podNS, podName, dataGranularity)
						predictJobSender.SendPodPredictJobs(pods, queueSender,
							unitType, queue.GetGranularitySec(strings.Trim(dataGranularity, " ")))
					}
				}
			} else if unitType == UnitTypeGPU {
				gpuHost := unit["host"].(string)
				gpuMinorNumber := unit["minor_number"].(string)
				modelInfo := modelMapper.GetModelInfo(unitType, dataGranularity, fmt.Sprintf("%s/%s", gpuHost, gpuMinorNumber))
				if modelInfo != nil {
					mt := time.Now().Unix() - modelInfo.Timestamp
					scope.Infof("export gpu %s model time metric with granularity %s value %v",
						fmt.Sprintf("%s/%s", gpuHost, gpuMinorNumber), dataGranularity, mt)
					metricExporter.SetGPUMetricModelTime(gpuHost, gpuMinorNumber,
						dataGranularity, float64(mt))
					modelMapper.RemoveModelInfo(unitType, dataGranularity,
						fmt.Sprintf("%s/%s", gpuHost, gpuMinorNumber))
				}

				res, err := datahubServiceClnt.ListGpus(context.Background(),
					&datahub_v1alpha1.ListGpusRequest{
						Host:        gpuHost,
						MinorNumber: gpuMinorNumber,
					})
				if err == nil {
					gpus := res.GetGpus()
					if len(gpus) > 0 {
						scope.Infof("gpu (host: %s, minor number: %s) model job completed and send predict job for granularity %s",
							gpuHost, gpuMinorNumber, dataGranularity)
						predictJobSender.SendGPUPredictJobs(gpus, queueSender,
							unitType, queue.GetGranularitySec(strings.Trim(dataGranularity, " ")))
					}
				}
			}
		}
		queueConn.Close()
	}
}
