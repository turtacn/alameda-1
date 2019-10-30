package dispatcher

import (
	"context"
	"fmt"
	"time"

	"github.com/containers-ai/alameda/ai-dispatcher/pkg/metrics"
	"github.com/containers-ai/alameda/ai-dispatcher/pkg/queue"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	datahub_common "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	datahub_metrics "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/metrics"
	datahub_predictions "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/predictions"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/grpc"
)

type podModelJobSender struct {
	datahubGrpcCn  *grpc.ClientConn
	modelMapper    *ModelMapper
	metricExporter *metrics.Exporter
}

func NewPodModelJobSender(datahubGrpcCn *grpc.ClientConn, modelMapper *ModelMapper,
	metricExporter *metrics.Exporter) *podModelJobSender {
	return &podModelJobSender{
		datahubGrpcCn:  datahubGrpcCn,
		modelMapper:    modelMapper,
		metricExporter: metricExporter,
	}
}

func (sender *podModelJobSender) sendModelJobs(pods []*datahub_resources.Pod, queueSender queue.QueueSender,
	pdUnit string, granularity int64, predictionStep int64) {
	datahubServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(sender.datahubGrpcCn)

	for _, pod := range pods {
		podNS := pod.GetNamespacedName().GetNamespace()
		podName := pod.GetNamespacedName().GetName()
		lastPredictionContainers, err := sender.getLastPrediction(datahubServiceClnt, pod, granularity)
		if err != nil {
			scope.Errorf("Get pod (%s/%s) last Prediction failed: %s",
				podNS, podName, err.Error())
			continue
		}
		if lastPredictionContainers == nil && err == nil {
			scope.Infof("No prediction found of pod (%s/%s)", podNS, podName)
		}
		sender.sendJobByMetrics(pod, queueSender, pdUnit, granularity, predictionStep,
			datahubServiceClnt, lastPredictionContainers)
	}
}

func (sender *podModelJobSender) sendJob(pod *datahub_resources.Pod, queueSender queue.QueueSender, pdUnit string,
	granularity int64, podInfo *modelInfo) {
	marshaler := jsonpb.Marshaler{}
	podNS := pod.GetNamespacedName().GetNamespace()
	podName := pod.GetNamespacedName().GetName()
	dataGranularity := queue.GetGranularityStr(granularity)
	podStr, err := marshaler.MarshalToString(pod)
	if err != nil {
		scope.Errorf("Encode pb message failed for pod %s/%s with granularity %v seconds. %s",
			podNS, podName, granularity, err.Error())
		return
	}
	if len(podInfo.Containers) > 0 && podStr != "" {
		jb := queue.NewJobBuilder(pdUnit, granularity, podStr)
		jobJSONStr, err := jb.GetJobJSONString()
		if err != nil {
			scope.Errorf("Prepare model job payload failed for pod %s/%s with granularity %v seconds. %s",
				podNS, podName, granularity, err.Error())
			return
		}

		err = queueSender.SendJsonString(modelQueueName, jobJSONStr,
			fmt.Sprintf("%s/%s/%v", podNS, podName, granularity))
		if err == nil {
			sender.modelMapper.AddModelInfo(pdUnit, dataGranularity, podInfo)
		} else {
			scope.Errorf("Send model job payload failed for pod %s/%s with granularity %v seconds. %s",
				podNS, podName, granularity, err.Error())
		}
	}
}

func (sender *podModelJobSender) genPodInfo(podNS,
	podName string) *modelInfo {
	podInfo := new(modelInfo)
	podInfo.NamespacedName = &namespacedName{
		Namespace: podNS,
		Name:      podName,
	}
	podInfo.Containers = []*container{}
	podInfo.SetTimeStamp(time.Now().Unix())
	return podInfo
}

func (sender *podModelJobSender) genPodInfoWithAllMetrics(podNS,
	podName string, pod *datahub_resources.Pod) *modelInfo {
	podInfo := sender.genPodInfo(podNS, podName)
	return podInfo
}

func (sender *podModelJobSender) getLastPrediction(datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	pod *datahub_resources.Pod, granularity int64) ([]*datahub_predictions.ContainerPrediction, error) {
	podPredictRes, err := datahubServiceClnt.ListPodPredictions(context.Background(),
		&datahub_predictions.ListPodPredictionsRequest{
			Granularity:    granularity,
			NamespacedName: pod.GetNamespacedName(),
			QueryCondition: &datahub_common.QueryCondition{
				Limit: 1,
				Order: datahub_common.QueryCondition_DESC,
				TimeRange: &datahub_common.TimeRange{
					Step: &duration.Duration{
						Seconds: granularity,
					},
				},
			},
		})
	if err != nil {
		return nil, err
	}
	if len(podPredictRes.GetPodPredictions()) > 0 {
		lastPodPrediction := podPredictRes.GetPodPredictions()[0]
		return lastPodPrediction.GetContainerPredictions(), nil
	}
	return nil, nil
}

func (sender *podModelJobSender) getQueryMetricStartTime(descPodPredictions []*datahub_predictions.PodPrediction) int64 {
	if len(descPodPredictions) > 0 {
		ctPdMDs := descPodPredictions[len(descPodPredictions)-1].GetContainerPredictions()
		for _, ctPdMD := range ctPdMDs {
			pdMDs := ctPdMD.GetPredictedRawData()
			for _, pdMD := range pdMDs {
				mD := pdMD.GetData()
				if len(mD) > 0 {
					return mD[len(mD)-1].GetTime().GetSeconds()
				}
			}
		}
	}
	return 0
}

func (sender *podModelJobSender) sendJobByMetrics(pod *datahub_resources.Pod, queueSender queue.QueueSender,
	pdUnit string, granularity int64, predictionStep int64, datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	lastPredictionContainers []*datahub_predictions.ContainerPrediction) {
	podNS := pod.GetNamespacedName().GetNamespace()
	podName := pod.GetNamespacedName().GetName()
	dataGranularity := queue.GetGranularityStr(granularity)
	queryCondition := &datahub_common.QueryCondition{
		Order: datahub_common.QueryCondition_DESC,
		TimeRange: &datahub_common.TimeRange{
			Step: &duration.Duration{
				Seconds: granularity,
			},
		},
	}

	nowSeconds := time.Now().Unix()

	if len(lastPredictionContainers) == 0 {
		podInfo := sender.genPodInfoWithAllMetrics(podNS, podName, pod)
		sender.sendJob(pod, queueSender, pdUnit, granularity, podInfo)
		scope.Infof("No last prediction metric found of pod (%s/%s)",
			podNS, podName)
		return
	}
	for _, lastPredictionContainer := range lastPredictionContainers {
		for _, lastPredictionMetric := range lastPredictionContainer.GetPredictedRawData() {
			if len(lastPredictionMetric.GetData()) == 0 {
				containers := []*container{}
				for _, ct := range pod.GetContainers() {
					containers = append(containers, &container{
						Name: ct.GetName(),
						ModelMetrics: []datahub_common.MetricType{
							datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
							datahub_common.MetricType_MEMORY_USAGE_BYTES,
						},
					})
				}
				podInfo := sender.genPodInfo(podNS, podName)
				podInfo.Containers = containers
				sender.sendJob(pod, queueSender, pdUnit, granularity, podInfo)
				scope.Infof("No last prediction metric %s found of pod (%s/%s)",
					lastPredictionMetric.GetMetricType().String(), podNS, podName)
				return
			} else {
				lastPrediction := lastPredictionMetric.GetData()[0]
				lastPredictionTime := lastPredictionMetric.GetData()[0].GetTime().GetSeconds()
				if lastPrediction != nil && lastPredictionTime <= nowSeconds {
					scope.Infof("pod (%s/%s) prediction is out of date due to last predict time is %v (current: %v)",
						podNS, podName, lastPredictionTime, nowSeconds)
				}
				if lastPrediction != nil && lastPredictionTime <= time.Now().Unix() {
					containers := []*container{}
					for _, ct := range pod.GetContainers() {
						containers = append(containers, &container{
							Name: ct.GetName(),
							ModelMetrics: []datahub_common.MetricType{
								datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
								datahub_common.MetricType_MEMORY_USAGE_BYTES,
							},
						})
					}
					podInfo := sender.genPodInfo(podNS, podName)
					podInfo.Containers = containers
					scope.Infof("send pod (%s/%s) model job due to no predict found or predict is out of date",
						podNS, podName)
					sender.sendJob(pod, queueSender, pdUnit, granularity, podInfo)
					return
				}

				podPredictRes, err := datahubServiceClnt.ListPodPredictions(context.Background(),
					&datahub_predictions.ListPodPredictionsRequest{
						NamespacedName: pod.GetNamespacedName(),
						ModelId:        lastPrediction.GetModelId(),
						Granularity:    granularity,
						QueryCondition: queryCondition,
					})
				podPredictions := podPredictRes.GetPodPredictions()
				queryStartTime := time.Now().Unix() - predictionStep*granularity
				firstPDTime := sender.getQueryMetricStartTime(podPredictions)
				if firstPDTime > 0 {
					queryStartTime = firstPDTime
				}

				podMetricsRes, err := datahubServiceClnt.ListPodMetrics(context.Background(),
					&datahub_metrics.ListPodMetricsRequest{
						QueryCondition: &datahub_common.QueryCondition{
							Order: datahub_common.QueryCondition_DESC,
							TimeRange: &datahub_common.TimeRange{
								StartTime: &timestamp.Timestamp{
									Seconds: queryStartTime,
								},
								Step: &duration.Duration{
									Seconds: granularity,
								},
							},
						},
						NamespacedName: pod.GetNamespacedName(),
					})

				if err != nil {
					scope.Errorf("List pods (%s/%s) metric with granularity %v for sending model job failed: %s",
						podNS, podName, granularity, err.Error())
					continue
				}
				podMetrics := podMetricsRes.GetPodMetrics()
				for _, podMetric := range podMetrics {
					podInfo := sender.genPodInfo(podNS, podName)
					containers := []*container{}
					containerMetrics := podMetric.GetContainerMetrics()
					for _, containerMetric := range containerMetrics {
						containerName := containerMetric.GetName()
						ct := &container{
							Name: containerName,
						}
						metricData := containerMetric.GetMetricData()
						modelMetrics := []datahub_common.MetricType{}
						for _, metricDatum := range metricData {
							mData := metricDatum.GetData()
							pData := []*datahub_predictions.Sample{}
							for _, podPrediction := range podPredictions {
								containerPredictions := podPrediction.GetContainerPredictions()
								for _, containerPrediction := range containerPredictions {
									predictRawData := containerPrediction.GetPredictedRawData()
									for _, predictRawDatum := range predictRawData {
										if metricDatum.GetMetricType() == predictRawDatum.GetMetricType() {
											pData = append(pData, predictRawDatum.GetData()...)
										}
									}
								}
							}
							metricsNeedToModel, drift := DriftEvaluation(UnitTypePod, metricDatum.GetMetricType(), granularity, mData, pData, map[string]string{
								"podNS":             podNS,
								"podName":           podName,
								"containerName":     containerName,
								"targetDisplayName": fmt.Sprintf("pod %s/%s container %s", podNS, podName, containerName),
							}, sender.metricExporter)
							if drift {
								scope.Infof("export pod %s/%s drift counter with granularity %s",
									podNS, podName, dataGranularity)
								sender.metricExporter.AddPodMetricDrift(podNS, podName,
									queue.GetGranularityStr(granularity), 1.0)
							}
							modelMetrics = append(modelMetrics, metricsNeedToModel...)
						}
						ct.ModelMetrics = modelMetrics
						containers = append(containers, ct)
					}
					podInfo.Containers = containers
					isModeling := sender.modelMapper.IsModeling(pdUnit, dataGranularity, podInfo)
					if !isModeling || (isModeling && sender.modelMapper.IsModelTimeout(pdUnit, dataGranularity, podInfo)) {
						sender.sendJob(pod, queueSender, pdUnit, granularity, podInfo)
						return
					}
				}
			}
		}
	}
}
