package dispatcher

import (
	"context"
	"time"

	"github.com/containers-ai/alameda/ai-dispatcher/pkg/queue"
	"github.com/containers-ai/alameda/ai-dispatcher/pkg/stats"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

type modelJobSender struct {
	datahubGrpcCn  *grpc.ClientConn
	modelThreshold float64
	modelMapper    *ModelMapper
}

func NewModelJobSender(datahubGrpcCn *grpc.ClientConn, modelMapper *ModelMapper) *modelJobSender {
	return &modelJobSender{
		datahubGrpcCn:  datahubGrpcCn,
		modelThreshold: viper.GetFloat64("model.threshold"),
		modelMapper:    modelMapper,
	}
}

func (dispatcher *modelJobSender) sendNodeModelJobs(nodes []*datahub_v1alpha1.Node,
	queueSender queue.QueueSender, pdUnit string, granularity int64, predictionStep int64) {

	marshaler := jsonpb.Marshaler{}
	dataGranularity := queue.GetGranularityStr(granularity)
	datahubServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(dispatcher.datahubGrpcCn)
	queryCondition := &datahub_v1alpha1.QueryCondition{
		Order: datahub_v1alpha1.QueryCondition_DESC,
		TimeRange: &datahub_v1alpha1.TimeRange{
			StartTime: &timestamp.Timestamp{
				Seconds: time.Now().Unix() - predictionStep*granularity,
			},
			Step: &duration.Duration{
				Seconds: granularity,
			},
		},
	}
	for _, node := range nodes {
		nodeName := node.GetName()
		nodePredictRes, err := datahubServiceClnt.ListNodePredictions(context.Background(),
			&datahub_v1alpha1.ListNodePredictionsRequest{
				NodeNames:      []string{node.GetName()},
				Granularity:    granularity,
				QueryCondition: queryCondition,
			})
		if err != nil {
			scope.Errorf("Get node %s Prediction with granularity %v for sending model job failed: %s",
				nodeName, granularity, err.Error())
			continue
		}
		nodePredictions := nodePredictRes.GetNodePredictions()
		if len(nodePredictions) == 0 {
			scope.Infof("No predict found for node %s with granularity %v, send model job to queue.",
				nodeName, granularity)
			nodeInfo := &modelInfo{
				&podModel{},
				&nodeModel{
					Name: nodeName,
					ModelMetrics: []datahub_v1alpha1.MetricType{
						datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
						datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES,
					},
				},
				time.Now().Unix(),
			}

			nodeStr, err := marshaler.MarshalToString(node)
			//nodeStr, err := marshaler.MarshalToString(nodeInfo)
			if err != nil {
				scope.Errorf("Encode pb message failed for node %s with granularity seconds %v. %s",
					node.GetName(), granularity, err.Error())
				continue
			}
			if len(nodeInfo.ModelMetrics) > 0 && nodeStr != "" {
				jb := queue.NewJobBuilder(pdUnit, granularity, nodeStr)
				jobJSONStr, err := jb.GetJobJSONString()
				if err != nil {
					scope.Errorf("Prepare model job payload failed for node %s with granularity seconds %v. %s",
						nodeName, granularity, err.Error())
					continue
				}
				err = queueSender.SendJsonString(modelQueueName, jobJSONStr)
			}
			continue
		}

		for _, nodePrediction := range nodePredictions {
			nodeInfo := &modelInfo{
				&podModel{},
				&nodeModel{
					Name:         nodeName,
					ModelMetrics: []datahub_v1alpha1.MetricType{},
				},
				time.Now().Unix(),
			}
			nodeMetricsRes, err := datahubServiceClnt.ListNodeMetrics(context.Background(),
				&datahub_v1alpha1.ListNodeMetricsRequest{
					QueryCondition: queryCondition,
					NodeNames:      []string{nodePrediction.GetName()},
				})
			if err != nil {
				scope.Errorf("List nodes metric with granularity %v for sending model job failed: %s",
					granularity, err.Error())
				continue
			}

			nodeMetrics := nodeMetricsRes.GetNodeMetrics()
			predictRawData := nodePrediction.GetPredictedRawData()

			for _, predictRawDatum := range predictRawData {
				pData := predictRawDatum.GetData()
				for _, nodeMetric := range nodeMetrics {
					metricData := nodeMetric.GetMetricData()
					for _, metricDatum := range metricData {
						mData := metricDatum.GetData()
						if metricDatum.GetMetricType() == predictRawDatum.GetMetricType() {
							metricType := predictRawDatum.GetMetricType()
							scope.Infof("start MAPE calculation for node %s metric %v with granularity %v",
								nodeName, metricType, granularity)
							measurementDataSet := stats.NewMeasurementDataSet(mData, pData, granularity)
							mape, err := stats.MAPE(measurementDataSet)
							if err != nil {
								nodeInfo.ModelMetrics = append(nodeInfo.ModelMetrics, metricType)
								scope.Infof(
									"model job for node %s metric %v with granularity %v should be sent due to MAPE calculation failed: %s",
									nodeName, metricType, granularity, err.Error())
							} else if mape > dispatcher.modelThreshold {
								nodeInfo.ModelMetrics = append(nodeInfo.ModelMetrics, metricType)
								scope.Infof("model job node %s metric %v with granularity %v should be sent due to MAPE %v > %v",
									nodeName, metricType, granularity, mape, dispatcher.modelThreshold)
							} else {
								scope.Infof("node %s metric %v with granularity %v MAPE %v <= %v, skip sending this model metric",
									nodeName, metricType, granularity, mape, dispatcher.modelThreshold)
							}
						}
					}
				}
			}
			isModeling := dispatcher.modelMapper.IsModeling(pdUnit, dataGranularity, nodeInfo)
			if !isModeling || (isModeling && dispatcher.modelMapper.IsModelTimeout(
				pdUnit, dataGranularity, nodeInfo)) {
				nodeStr, err := marshaler.MarshalToString(node)
				//nodeStr, err := marshaler.MarshalToString(nodeInfo)
				if err != nil {
					scope.Errorf("Encode pb message failed for node %s with granularity seconds %v. %s",
						node.GetName(), granularity, err.Error())
					continue
				}
				if len(nodeInfo.ModelMetrics) > 0 && nodeStr != "" {
					jb := queue.NewJobBuilder(pdUnit, granularity, nodeStr)
					jobJSONStr, err := jb.GetJobJSONString()
					if err != nil {
						scope.Errorf(
							"Prepare model job payload failed for node %s with granularity seconds %v. %s",
							nodeName, granularity, err.Error())
						continue
					}
					err = queueSender.SendJsonString(modelQueueName, jobJSONStr)
					if err == nil {
						dispatcher.modelMapper.AddModelInfo(pdUnit, dataGranularity, nodeInfo)
					} else {
						scope.Errorf(
							"Send model job payload failed for node %s with granularity seconds %v. %s",
							nodeName, granularity, err.Error())
					}
				}
			}
		}
	}
}

func (dispatcher *modelJobSender) sendPodModelJobs(pods []*datahub_v1alpha1.Pod, queueSender queue.QueueSender,
	pdUnit string, granularity int64, predictionStep int64) {

	marshaler := jsonpb.Marshaler{}
	dataGranularity := queue.GetGranularityStr(granularity)
	datahubServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(dispatcher.datahubGrpcCn)
	queryCondition := &datahub_v1alpha1.QueryCondition{
		Order: datahub_v1alpha1.QueryCondition_DESC,
		TimeRange: &datahub_v1alpha1.TimeRange{
			StartTime: &timestamp.Timestamp{
				Seconds: time.Now().Unix() - predictionStep*granularity,
			},
			Step: &duration.Duration{
				Seconds: granularity,
			},
		},
	}
	for _, pod := range pods {
		podNS := pod.GetNamespacedName().GetNamespace()
		podName := pod.GetNamespacedName().GetName()
		// send model jobs
		podPredictRes, err := datahubServiceClnt.ListPodPredictions(context.Background(),
			&datahub_v1alpha1.ListPodPredictionsRequest{
				Granularity:    granularity,
				QueryCondition: queryCondition,
			})
		if err != nil {
			scope.Errorf("Get pod (%s/%s) Prediction with granularity %v for sending model job failed: %s",
				podNS, podName, granularity, err.Error())
			continue
		}
		podPredictions := podPredictRes.GetPodPredictions()
		if len(podPredictions) == 0 {
			scope.Infof("No predict found for pod (%s/%s), send model job to queue.",
				podNS, podName)
			containers := []*container{}
			for _, ct := range pod.GetContainers() {
				containers = append(containers, &container{
					Name: ct.GetName(),
					ModelMetrics: []datahub_v1alpha1.MetricType{
						datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
						datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES,
					},
				})
			}
			podInfo := &modelInfo{
				&podModel{
					NamespacedName: &namespacedName{
						Namespace: podNS,
						Name:      podName,
					},
					Containers: containers,
				},
				&nodeModel{},
				time.Now().Unix(),
			}
			podStr, err := marshaler.MarshalToString(pod)
			//podStr, err := marshaler.MarshalToString(podInfo)
			if err != nil {
				scope.Errorf("Encode pb message failed for pod %s/%s with granularity %v seconds. %s",
					podNS, podName, granularity, err.Error())
				continue
			}
			if len(podInfo.Containers) > 0 && podStr != "" {
				jb := queue.NewJobBuilder(pdUnit, granularity, podStr)
				jobJSONStr, err := jb.GetJobJSONString()
				if err != nil {
					scope.Errorf("Prepare model job payload for pod (%s/%s) with granularity %v failed: %s",
						podNS, podName, granularity, err.Error())
					continue
				}
				err = queueSender.SendJsonString(modelQueueName, jobJSONStr)
			}
			continue
		}

		for _, podPrediction := range podPredictions {
			podInfo := &modelInfo{
				&podModel{
					NamespacedName: &namespacedName{
						Namespace: podNS,
						Name:      podName,
					},
					Containers: []*container{},
				},
				&nodeModel{},
				time.Now().Unix(),
			}

			podMetricsRes, err := datahubServiceClnt.ListPodMetrics(context.Background(),
				&datahub_v1alpha1.ListPodMetricsRequest{
					QueryCondition: queryCondition,
					NamespacedName: podPrediction.GetNamespacedName(),
				})
			if err != nil {
				scope.Errorf("List nodes metric with granularity %v for sending model job failed: %s",
					granularity, err.Error())
				continue
			}
			podMetrics := podMetricsRes.GetPodMetrics()
			containerPredictions := podPrediction.GetContainerPredictions()
			for _, containerPrediction := range containerPredictions {
				predictRawData := containerPrediction.GetPredictedRawData()
				for _, predictRawDatum := range predictRawData {
					pData := predictRawDatum.GetData()
					for _, podMetric := range podMetrics {
						containerMetrics := podMetric.GetContainerMetrics()
						for _, containerMetric := range containerMetrics {
							containerName := containerMetric.GetName()
							metricData := containerMetric.GetMetricData()
							modelMetrics := []datahub_v1alpha1.MetricType{}
							for _, metricDatum := range metricData {
								mData := metricDatum.GetData()
								if metricDatum.GetMetricType() == predictRawDatum.GetMetricType() {
									metricType := predictRawDatum.GetMetricType()
									scope.Infof("start MAPE calculation for pod %s/%s container %s metric %v with granularity %v",
										podNS, podName, containerName, metricType, granularity)
									measurementDataSet := stats.NewMeasurementDataSet(mData, pData, granularity)
									mape, err := stats.MAPE(measurementDataSet)
									if err != nil {
										modelMetrics = append(modelMetrics, metricType)
										scope.Infof(
											"model job for pod %s/%s container %s metric %v with granularity %v should be sent due to MAPE calculation failed: %s",
											podNS, podName, containerName, metricType, granularity, err.Error())
									} else if mape > dispatcher.modelThreshold {
										modelMetrics = append(modelMetrics, metricType)
										scope.Infof("pod %s/%s container %s metric %v with granularity %v should be sent due to MAPE %v > %v",
											podNS, podName, containerName, metricType, granularity, mape, dispatcher.modelThreshold)
									} else {
										scope.Infof("pod %s/%s container %s metric %v with granularity %v MAPE %v <= %v, skip sending this model metric",
											podNS, podName, containerName, metricType, granularity, mape, dispatcher.modelThreshold)
									}
								}
							}
							if len(modelMetrics) > 0 {
								podInfo.Containers = append(podInfo.Containers, &container{
									Name:         containerName,
									ModelMetrics: modelMetrics,
								})
							}
						}
					}
				}
			}

			isModeling := dispatcher.modelMapper.IsModeling(pdUnit, dataGranularity, podInfo)
			if !isModeling || (isModeling && dispatcher.modelMapper.IsModelTimeout(pdUnit, dataGranularity, podInfo)) {
				podStr, err := marshaler.MarshalToString(pod)
				//podStr, err := marshaler.MarshalToString(podInfo)
				if err != nil {
					scope.Errorf("Encode pb message failed for pod %s/%s with granularity %v seconds. %s",
						podNS, podName, granularity, err.Error())
					continue
				}
				if len(podInfo.Containers) > 0 && podStr != "" {
					jb := queue.NewJobBuilder(pdUnit, granularity, podStr)
					jobJSONStr, err := jb.GetJobJSONString()
					if err != nil {
						scope.Errorf("Prepare model job payload failed for pod %s/%s with granularity %v seconds. %s",
							podNS, podName, granularity, err.Error())
					}
					err = queueSender.SendJsonString(modelQueueName, jobJSONStr)
					if err == nil {
						dispatcher.modelMapper.AddModelInfo(pdUnit, dataGranularity, podInfo)
					} else {
						scope.Errorf("Send model job payload failed for pod %s/%s with granularity %v seconds. %s",
							podNS, podName, granularity, err.Error())
					}
				}
			}
		}
	}
}
