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
	datahub_gpu "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/gpu"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
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
			jobCreateTime := int64(msgMap["job_create_time"].(float64))
			if unitType == UnitTypeNode {
				nodeName := unit["name"].(string)
				modelMapper.RemoveModelInfo(unitType, dataGranularity, nodeName)

				mt := time.Now().Unix() - jobCreateTime
				scope.Infof("export node %s model time metric with granularity %s value %v",
					nodeName, dataGranularity, mt)
				metricExporter.ExportNodeMetricModelTime(nodeName, dataGranularity, float64(mt))

				res, err := datahubServiceClnt.ListNodes(context.Background(),
					&datahub_resources.ListNodesRequest{
						ObjectMeta: []*datahub_resources.ObjectMeta{
							&datahub_resources.ObjectMeta{
								Name: nodeName,
							},
						},
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
				modelMapper.RemoveModelInfo(unitType, dataGranularity,
					fmt.Sprintf("%s/%s", podNS, podName))

				mt := time.Now().Unix() - jobCreateTime
				scope.Infof("export pod %s model time metric with granularity %s value %v",
					fmt.Sprintf("%s/%s", podNS, podName), dataGranularity, mt)
				metricExporter.ExportPodMetricModelTime(podNS, podName, dataGranularity, float64(mt))

				res, err := datahubServiceClnt.ListPods(context.Background(),
					&datahub_resources.ListPodsRequest{
						ObjectMeta: []*datahub_resources.ObjectMeta{
							&datahub_resources.ObjectMeta{
								Namespace: podNS,
								Name:      podName,
							},
						},
						ScalingTool: datahub_resources.ScalingTool_VPA,
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
				modelMapper.RemoveModelInfo(unitType, dataGranularity,
					fmt.Sprintf("%s/%s", gpuHost, gpuMinorNumber))

				mt := time.Now().Unix() - jobCreateTime
				scope.Infof("export gpu %s model time metric with granularity %s value %v",
					fmt.Sprintf("%s/%s", gpuHost, gpuMinorNumber), dataGranularity, mt)
				metricExporter.ExportGPUMetricModelTime(gpuHost, gpuMinorNumber,
					dataGranularity, float64(mt))

				res, err := datahubServiceClnt.ListGpus(context.Background(),
					&datahub_gpu.ListGpusRequest{
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
					dataGranularity, float64(mt))

				res, err := datahubServiceClnt.ListApplications(context.Background(),
					&datahub_resources.ListApplicationsRequest{
						ObjectMeta: []*datahub_resources.ObjectMeta{
							&datahub_resources.ObjectMeta{
								Namespace: appNS,
								Name:      appName,
							},
						},
					})
				if err == nil {
					applications := res.GetApplications()
					if len(applications) > 0 {
						scope.Infof("application %s/%s model job completed and send predict job for granularity %s",
							appNS, appName, dataGranularity)
						predictJobSender.SendApplicationPredictJobs(applications, queueSender,
							unitType, queue.GetGranularitySec(strings.Trim(dataGranularity, " ")))
					}
				}
			} else if unitType == UnitTypeNamespace {
				namespaceName := unit["name"].(string)
				modelMapper.RemoveModelInfo(unitType, dataGranularity, namespaceName)

				mt := time.Now().Unix() - jobCreateTime
				scope.Infof("export namespace %s model time metric with granularity %s value %v",
					namespaceName, dataGranularity, mt)
				metricExporter.ExportNamespaceMetricModelTime(namespaceName, dataGranularity, float64(mt))

				res, err := datahubServiceClnt.ListNamespaces(context.Background(),
					&datahub_resources.ListNamespacesRequest{
						ObjectMeta: []*datahub_resources.ObjectMeta{
							&datahub_resources.ObjectMeta{
								Name: namespaceName,
							},
						},
					})
				if err == nil {
					namespaces := res.GetNamespaces()
					if len(namespaces) > 0 {
						scope.Infof("namespace %s model job completed and send predict job for granularity %s",
							namespaceName, dataGranularity)
						predictJobSender.SendNamespacePredictJobs(namespaces, queueSender,
							unitType, queue.GetGranularitySec(strings.Trim(dataGranularity, " ")))
					}
				}
			} else if unitType == UnitTypeCluster {
				clusterName := unit["name"].(string)
				modelMapper.RemoveModelInfo(unitType, dataGranularity, clusterName)

				mt := time.Now().Unix() - jobCreateTime
				scope.Infof("export cluster %s model time metric with granularity %s value %v",
					clusterName, dataGranularity, mt)
				metricExporter.ExportClusterMetricModelTime(clusterName, dataGranularity, float64(mt))

				res, err := datahubServiceClnt.ListClusters(context.Background(),
					&datahub_resources.ListClustersRequest{
						ObjectMeta: []*datahub_resources.ObjectMeta{
							&datahub_resources.ObjectMeta{
								Name: clusterName,
							},
						},
					})
				if err == nil {
					clusters := res.GetClusters()
					if len(clusters) > 0 {
						scope.Infof("cluster %s model job completed and send predict job for granularity %s",
							clusterName, dataGranularity)
						predictJobSender.SendClusterPredictJobs(clusters, queueSender,
							unitType, queue.GetGranularitySec(strings.Trim(dataGranularity, " ")))
					}
				}
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
					kind, dataGranularity, float64(mt))

				res, err := datahubServiceClnt.ListControllers(context.Background(),
					&datahub_resources.ListControllersRequest{
						ObjectMeta: []*datahub_resources.ObjectMeta{
							&datahub_resources.ObjectMeta{
								Namespace: appNS,
								Name:      appName,
							},
						},
						Kind: datahub_resources.Kind(datahub_resources.Kind_value[kind]),
					})
				if err == nil {
					controllers := res.GetControllers()
					if len(controllers) > 0 {
						scope.Infof("controller %s/%s model job completed and send predict job for granularity %s",
							appNS, appName, dataGranularity)
						predictJobSender.SendControllerPredictJobs(controllers, queueSender,
							unitType, queue.GetGranularitySec(strings.Trim(dataGranularity, " ")))
					}
				}
			}
		}
		queueConn.Close()
	}
}
