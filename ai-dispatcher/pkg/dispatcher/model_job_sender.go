package dispatcher

import (
	"context"
	"time"

	"github.com/containers-ai/alameda/ai-dispatcher/pkg/metrics"
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
	metricExporter *metrics.Exporter
}

func NewModelJobSender(datahubGrpcCn *grpc.ClientConn, modelMapper *ModelMapper,
	metricExporter *metrics.Exporter) *modelJobSender {
	return &modelJobSender{
		datahubGrpcCn:  datahubGrpcCn,
		modelThreshold: viper.GetFloat64("model.threshold"),
		modelMapper:    modelMapper,
		metricExporter: metricExporter,
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
				NodeNames:      []string{nodeName},
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
			nodeInfo := new(modelInfo)
			nodeInfo.Name = nodeName
			nodeInfo.ModelMetrics = []datahub_v1alpha1.MetricType{
				datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
				datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES,
			}
			nodeInfo.SetTimeStamp(time.Now().Unix())

			nodeStr, err := marshaler.MarshalToString(node)
			//nodeStr, err := marshaler.MarshalToString(nodeInfo)
			if err != nil {
				scope.Errorf("Encode pb message failed for node %s with granularity seconds %v. %s",
					nodeName, granularity, err.Error())
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

				scope.Infof("export node %s drift counter with granularity %s",
					nodeName, dataGranularity)
				dispatcher.metricExporter.AddNodeMetricDrift(nodeName, queue.GetGranularityStr(granularity), 1.0)

				err = queueSender.SendJsonString(modelQueueName, jobJSONStr)
				if err != nil {
					scope.Errorf("Send model job payload failed for node %s with granularity seconds %v. %s",
						nodeName, granularity, err.Error())
				}
			}
			continue
		}

		for _, nodePrediction := range nodePredictions {
			nodeInfo := new(modelInfo)
			nodeInfo.Name = nodeName
			nodeInfo.ModelMetrics = []datahub_v1alpha1.MetricType{}
			nodeInfo.SetTimeStamp(time.Now().Unix())

			nodeMetricsRes, err := datahubServiceClnt.ListNodeMetrics(context.Background(),
				&datahub_v1alpha1.ListNodeMetricsRequest{
					QueryCondition: queryCondition,
					NodeNames:      []string{nodePrediction.GetName()},
				})
			if err != nil {
				scope.Errorf("List nodes %s metric with granularity %v for sending model job failed: %s",
					nodeName, granularity, err.Error())
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
							mape, mapeErr := stats.MAPE(measurementDataSet)
							if mapeErr == nil {
								scope.Infof("export MAPE value %v for node %s metric %v with granularity %v", mape,
									nodeName, metricType, granularity)
								dispatcher.metricExporter.SetNodeMetricMAPE(nodeName,
									queue.GetMetricLabel(metricDatum.GetMetricType()), queue.GetGranularityStr(granularity), mape)
							}
							if mapeErr != nil {
								nodeInfo.ModelMetrics = append(nodeInfo.ModelMetrics, metricType)
								scope.Infof(
									"model job for node %s metric %v with granularity %v should be sent due to MAPE calculation failed: %s",
									nodeName, metricType, granularity, mapeErr.Error())
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

					scope.Infof("export node %s drift counter with granularity %s",
						nodeName, dataGranularity)
					dispatcher.metricExporter.AddNodeMetricDrift(nodeName, queue.GetGranularityStr(granularity), 1.0)

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
			podInfo := new(modelInfo)
			podInfo.NamespacedName = &namespacedName{
				Namespace: podNS,
				Name:      podName,
			}
			podInfo.Containers = containers
			podInfo.SetTimeStamp(time.Now().Unix())

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

				scope.Infof("export pod %s/%s drift counter with granularity %s",
					podNS, podName, dataGranularity)
				dispatcher.metricExporter.AddPodMetricDrift(podNS, podName,
					queue.GetGranularityStr(granularity), 1.0)

				err = queueSender.SendJsonString(modelQueueName, jobJSONStr)
				if err != nil {
					scope.Errorf("Send model job payload for pod (%s/%s) with granularity %v failed: %s",
						podNS, podName, granularity, err.Error())
				}
			}
			continue
		}

		for _, podPrediction := range podPredictions {
			podInfo := new(modelInfo)
			podInfo.NamespacedName = &namespacedName{
				Namespace: podNS,
				Name:      podName,
			}
			podInfo.Containers = []*container{}
			podInfo.SetTimeStamp(time.Now().Unix())

			podMetricsRes, err := datahubServiceClnt.ListPodMetrics(context.Background(),
				&datahub_v1alpha1.ListPodMetricsRequest{
					QueryCondition: queryCondition,
					NamespacedName: podPrediction.GetNamespacedName(),
				})
			if err != nil {
				scope.Errorf("List pods (%s/%s) metric with granularity %v for sending model job failed: %s",
					podNS, podName, granularity, err.Error())
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
									mape, mapeErr := stats.MAPE(measurementDataSet)
									if mapeErr == nil {
										scope.Infof("export MAPE value %v for pod %s/%s container %s metric %v with granularity %v", mape,
											podNS, podName, containerName, metricType, granularity)
										dispatcher.metricExporter.SetContainerMetricMAPE(podNS, podName, containerName,
											queue.GetMetricLabel(metricDatum.GetMetricType()), queue.GetGranularityStr(granularity), mape)
									}

									if mapeErr != nil {
										modelMetrics = append(modelMetrics, metricType)
										scope.Infof(
											"model job for pod %s/%s container %s metric %v with granularity %v should be sent due to MAPE calculation failed: %s",
											podNS, podName, containerName, metricType, granularity, mapeErr.Error())
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

					scope.Infof("export pod %s/%s drift counter with granularity %s",
						podNS, podName, dataGranularity)
					dispatcher.metricExporter.AddPodMetricDrift(podNS, podName,
						queue.GetGranularityStr(granularity), 1.0)

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

func (dispatcher *modelJobSender) sendGPUModelJobs(gpus []*datahub_v1alpha1.Gpu,
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
	for _, gpu := range gpus {
		gpuHost := gpu.GetMetadata().GetHost()
		gpuMinorNumber := gpu.GetMetadata().GetMinorNumber()
		gpuPredictRes, err := datahubServiceClnt.ListGpuPredictions(context.Background(),
			&datahub_v1alpha1.ListGpuPredictionsRequest{
				Host:           gpuHost,
				MinorNumber:    gpuMinorNumber,
				Granularity:    granularity,
				QueryCondition: queryCondition,
			})
		if err != nil {
			scope.Errorf("Get (gpu host: %s minor number: %s) Prediction with granularity %v for sending model job failed: %s",
				gpuHost, gpuMinorNumber, granularity, err.Error())
			continue
		}
		gpuPredictions := gpuPredictRes.GetGpuPredictions()
		if len(gpuPredictions) == 0 {
			scope.Infof("No predict found for (gpu host: %s minor number: %s) with granularity %v, send model job to queue.",
				gpuHost, gpuMinorNumber, granularity)
			gpuInfo := new(modelInfo)
			gpuInfo.Host = gpuHost
			gpuInfo.MinorNumber = gpuMinorNumber
			gpuInfo.ModelMetrics = []datahub_v1alpha1.MetricType{
				datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES,
				datahub_v1alpha1.MetricType_DUTY_CYCLE,
			}
			gpuInfo.SetTimeStamp(time.Now().Unix())

			gpuStr, err := marshaler.MarshalToString(gpu)
			//gpuStr, err := marshaler.MarshalToString(gpuInfo)
			if err != nil {
				scope.Errorf("Encode pb message failed for (gpu host: %s minor number: %s) with granularity seconds %v. %s",
					gpuHost, gpuMinorNumber, granularity, err.Error())
				continue
			}
			if len(gpuInfo.ModelMetrics) > 0 && gpuStr != "" {
				jb := queue.NewJobBuilder(pdUnit, granularity, gpuStr)
				jobJSONStr, err := jb.GetJobJSONString()
				if err != nil {
					scope.Errorf("Prepare model job payload failed for (gpu host: %s minor number: %s) with granularity seconds %v. %s",
						gpuHost, gpuMinorNumber, granularity, err.Error())
					continue
				}

				scope.Infof("export (gpu host: %s minor number: %s) drift counter with granularity %s",
					gpuHost, gpuMinorNumber, dataGranularity)
				dispatcher.metricExporter.AddGPUMetricDrift(gpuHost, gpuMinorNumber,
					queue.GetGranularityStr(granularity), 1.0)

				err = queueSender.SendJsonString(modelQueueName, jobJSONStr)
				if err != nil {
					scope.Errorf("Send model job payload failed for (gpu host: %s minor number: %s) with granularity seconds %v. %s",
						gpuHost, gpuMinorNumber, granularity, err.Error())
				}
			}
			continue
		}

		for _, gpuPrediction := range gpuPredictions {
			gpuInfo := new(modelInfo)
			gpuInfo.Name = gpuHost
			gpuInfo.MinorNumber = gpuMinorNumber
			gpuInfo.ModelMetrics = []datahub_v1alpha1.MetricType{}
			gpuInfo.SetTimeStamp(time.Now().Unix())

			gpuMetricsRes, err := datahubServiceClnt.ListGpuMetrics(context.Background(),
				&datahub_v1alpha1.ListGpuMetricsRequest{
					QueryCondition: queryCondition,
					Host:           gpuHost,
					MinorNumber:    gpuMinorNumber,
				})
			if err != nil {
				scope.Errorf("List gpu (gpu host: %s minor number: %s) metric with granularity %v for sending model job failed: %s",
					gpuHost, gpuMinorNumber, granularity, err.Error())
				continue
			}

			gpuMetrics := gpuMetricsRes.GetGpuMetrics()
			predictRawData := gpuPrediction.GetPredictedRawData()

			for _, predictRawDatum := range predictRawData {
				pData := predictRawDatum.GetData()
				for _, gpuMetric := range gpuMetrics {
					metricData := gpuMetric.GetMetricData()
					for _, metricDatum := range metricData {
						mData := metricDatum.GetData()
						if metricDatum.GetMetricType() == predictRawDatum.GetMetricType() {
							metricType := predictRawDatum.GetMetricType()
							scope.Infof("start MAPE calculation for (gpu host: %s minor number: %s) metric %v with granularity %v",
								gpuHost, gpuMinorNumber, metricType, granularity)
							measurementDataSet := stats.NewMeasurementDataSet(mData, pData, granularity)
							mape, mapeErr := stats.MAPE(measurementDataSet)
							if mapeErr == nil {
								scope.Infof("export MAPE value %v for (gpu host: %s minor number: %s) metric %v with granularity %v", mape,
									gpuHost, gpuMinorNumber, metricType, granularity)
								dispatcher.metricExporter.SetGPUMetricMAPE(gpuHost, gpuMinorNumber,
									queue.GetMetricLabel(metricDatum.GetMetricType()), queue.GetGranularityStr(granularity), mape)
							}
							if err != nil {
								gpuInfo.ModelMetrics = append(gpuInfo.ModelMetrics, metricType)
								scope.Infof(
									"model job for (gpu host: %s minor number: %s) metric %v with granularity %v should be sent due to MAPE calculation failed: %s",
									gpuHost, gpuMinorNumber, metricType, granularity, mapeErr.Error())
							} else if mape > dispatcher.modelThreshold {
								gpuInfo.ModelMetrics = append(gpuInfo.ModelMetrics, metricType)
								scope.Infof("model job (gpu host: %s minor number: %s) metric %v with granularity %v should be sent due to MAPE %v > %v",
									gpuHost, gpuMinorNumber, metricType, granularity, mape, dispatcher.modelThreshold)
							} else {
								scope.Infof("(gpu host: %s minor number: %s) metric %v with granularity %v MAPE %v <= %v, skip sending this model metric",
									gpuHost, gpuMinorNumber, metricType, granularity, mape, dispatcher.modelThreshold)
							}
						}
					}
				}
			}
			isModeling := dispatcher.modelMapper.IsModeling(pdUnit, dataGranularity, gpuInfo)
			if !isModeling || (isModeling && dispatcher.modelMapper.IsModelTimeout(
				pdUnit, dataGranularity, gpuInfo)) {
				gpuStr, err := marshaler.MarshalToString(gpu)
				//gpuStr, err := marshaler.MarshalToString(gpuInfo)
				if err != nil {
					scope.Errorf("Encode pb message failed for (gpu host: %s minor number: %s) with granularity seconds %v. %s",
						gpuHost, gpuMinorNumber, granularity, err.Error())
					continue
				}
				if len(gpuInfo.ModelMetrics) > 0 && gpuStr != "" {
					jb := queue.NewJobBuilder(pdUnit, granularity, gpuStr)
					jobJSONStr, err := jb.GetJobJSONString()
					if err != nil {
						scope.Errorf(
							"Prepare model job payload failed for (gpu host: %s, minor number: %s) with granularity seconds %v. %s",
							gpuHost, gpuMinorNumber, granularity, err.Error())
						continue
					}

					scope.Infof("export (gpu host: %s minor number: %s) drift counter with granularity %s",
						gpuHost, gpuMinorNumber, dataGranularity)
					dispatcher.metricExporter.AddGPUMetricDrift(gpuHost, gpuMinorNumber,
						queue.GetGranularityStr(granularity), 1.0)

					err = queueSender.SendJsonString(modelQueueName, jobJSONStr)
					if err == nil {
						dispatcher.modelMapper.AddModelInfo(pdUnit, dataGranularity, gpuInfo)
					} else {
						scope.Errorf(
							"Send model job payload failed for (gpu host: %s minor number: %s) with granularity seconds %v. %s",
							gpuHost, gpuMinorNumber, granularity, err.Error())
					}
				}
			}
		}
	}
}
