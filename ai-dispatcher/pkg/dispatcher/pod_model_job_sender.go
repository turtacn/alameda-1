package dispatcher

import (
	"context"
	"fmt"
	"time"

	"github.com/containers-ai/alameda/ai-dispatcher/pkg/metrics"
	"github.com/containers-ai/alameda/ai-dispatcher/pkg/queue"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
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

func (sender *podModelJobSender) sendModelJobs(pods []*datahub_v1alpha1.Pod, queueSender queue.QueueSender,
	pdUnit string, granularity int64, predictionStep int64) {

	dataGranularity := queue.GetGranularityStr(granularity)
	datahubServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(sender.datahubGrpcCn)
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
		shouldDrift := false

		lastPrediction, lastPredictionTime, err := sender.getLastPrediction(datahubServiceClnt, pod, granularity)
		if err != nil {
			scope.Errorf("Get pod (%s/%s) last Prediction failed: %s",
				podNS, podName, err.Error())
			continue
		}
		if lastPrediction == nil && err == nil {
			scope.Infof("No prediction found of pod (%s/%s)", podNS, podName)
		}
		nowSeconds := time.Now().Unix()
		if lastPrediction != nil && lastPredictionTime <= nowSeconds {
			scope.Infof("pod (%s/%s) prediction is out of date due to last predict time is %v (current: %v)",
				podNS, podName, lastPredictionTime, nowSeconds)
		}
		if (lastPrediction == nil && err == nil) || (lastPrediction != nil && lastPredictionTime <= time.Now().Unix()) {
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
			podInfo := sender.genPodInfo(podNS, podName)
			podInfo.Containers = containers
			scope.Infof("send pod (%s/%s) model job due to no predict found or predict is out of date",
				podNS, podName)
			sender.sendJob(pod, queueSender, pdUnit, granularity, podInfo)
		}

		podPredictRes, err := datahubServiceClnt.ListPodPredictions(context.Background(),
			&datahub_v1alpha1.ListPodPredictionsRequest{
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
			&datahub_v1alpha1.ListPodMetricsRequest{
				QueryCondition: &datahub_v1alpha1.QueryCondition{
					Order: datahub_v1alpha1.QueryCondition_DESC,
					TimeRange: &datahub_v1alpha1.TimeRange{
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
			containerMetrics := podMetric.GetContainerMetrics()
			for _, containerMetric := range containerMetrics {
				containerName := containerMetric.GetName()
				metricData := containerMetric.GetMetricData()
				modelMetrics := []datahub_v1alpha1.MetricType{}
				podInfo := sender.genPodInfo(podNS, podName)
				for _, metricDatum := range metricData {
					mData := metricDatum.GetData()
					pData := []*datahub_v1alpha1.Sample{}
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
						shouldDrift = drift
					}
					modelMetrics = append(modelMetrics, metricsNeedToModel...)
					isModeling := sender.modelMapper.IsModeling(pdUnit, dataGranularity, podInfo)
					if !isModeling || (isModeling && sender.modelMapper.IsModelTimeout(pdUnit, dataGranularity, podInfo)) {
						sender.sendJob(pod, queueSender, pdUnit, granularity, podInfo)
					}
				}
			}
		}
		if shouldDrift {
			scope.Infof("export pod %s/%s drift counter with granularity %s",
				podNS, podName, dataGranularity)
			sender.metricExporter.AddPodMetricDrift(podNS, podName,
				queue.GetGranularityStr(granularity), 1.0)
		}
	}
}

func (sender *podModelJobSender) sendJob(pod *datahub_v1alpha1.Pod, queueSender queue.QueueSender, pdUnit string,
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

func (sender *podModelJobSender) getLastPrediction(datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	pod *datahub_v1alpha1.Pod, granularity int64) (*datahub_v1alpha1.PodPrediction, int64, error) {
	podPredictRes, err := datahubServiceClnt.ListPodPredictions(context.Background(),
		&datahub_v1alpha1.ListPodPredictionsRequest{
			Granularity:    granularity,
			NamespacedName: pod.GetNamespacedName(),
			QueryCondition: &datahub_v1alpha1.QueryCondition{
				Limit: 1,
				Order: datahub_v1alpha1.QueryCondition_DESC,
				TimeRange: &datahub_v1alpha1.TimeRange{
					Step: &duration.Duration{
						Seconds: granularity,
					},
				},
			},
		})
	if err != nil {
		return nil, 0, err
	}
	if len(podPredictRes.GetPodPredictions()) > 0 {
		lastPodPrediction := podPredictRes.GetPodPredictions()[0]
		for _, ctPd := range lastPodPrediction.GetContainerPredictions() {
			for _, metricPd := range ctPd.GetPredictedRawData() {
				for _, metricPdSample := range metricPd.GetData() {
					return lastPodPrediction, metricPdSample.GetTime().GetSeconds(), nil
				}
			}
		}
	}
	return nil, 0, nil
}

func (sender *podModelJobSender) getQueryMetricStartTime(descPodPredictions []*datahub_v1alpha1.PodPrediction) int64 {
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
