package dispatcher

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/containers-ai/alameda/ai-dispatcher/consts"
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
	for _, pod := range pods {
		sender.sendPodModelJobs(pod, queueSender, pdUnit, granularity, predictionStep, &wg)
	}
}

func (sender *podModelJobSender) sendPodModelJobs(pod *datahub_resources.Pod, queueSender queue.QueueSender,
	pdUnit string, granularity int64, predictionStep int64, wg *sync.WaitGroup) {
	datahubServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(sender.datahubGrpcCn)
	dataGranularity := queue.GetGranularityStr(granularity)

	podNS := pod.GetObjectMeta().GetNamespace()
	podName := pod.GetObjectMeta().GetName()
	lastPredictionContainers, err := sender.getLastMIdPrediction(datahubServiceClnt, pod, granularity)
	if err != nil {
		scope.Errorf("[POD][%s][%s/%s] Get last prediction failed: %s",
			dataGranularity, podNS, podName, err.Error())
		return
	}
	sender.sendJobByMetrics(pod, queueSender, pdUnit, granularity, predictionStep,
		datahubServiceClnt, lastPredictionContainers)
}

func (sender *podModelJobSender) sendJob(pod *datahub_resources.Pod, queueSender queue.QueueSender, pdUnit string,
	granularity int64, ctName string, metricType datahub_common.MetricType) {
	marshaler := jsonpb.Marshaler{}
	clusterID := pod.GetObjectMeta().GetClusterName()
	podNS := pod.GetObjectMeta().GetNamespace()
	podName := pod.GetObjectMeta().GetName()
	dataGranularity := queue.GetGranularityStr(granularity)
	podStr, err := marshaler.MarshalToString(pod)

	if err != nil {
		scope.Errorf("[POD][%s][%s/%s/%s] Encode pb message failed. %s",
			dataGranularity, podNS, podName, ctName, err.Error())
		return
	}

	jb := queue.NewJobBuilder(clusterID, pdUnit, granularity, metricType, podStr, map[string]string{
		"containerName": ctName,
	})
	jobJSONStr, err := jb.GetJobJSONString()
	if err != nil {
		scope.Errorf("[POD][%s][%s/%s/%s] Prepare model job payload failed. %s",
			dataGranularity, podNS, podName, ctName, err.Error())
		return
	}

	podJobStr := fmt.Sprintf("%s/%s/%s/%s/%s/%v/%s", consts.UnitTypePod,
		clusterID, podNS, podName, ctName, granularity, metricType)
	scope.Infof("[POD][%s][%s/%s/%s] Try to send pod model job: %s", dataGranularity, podNS, podName, ctName, podJobStr)
	err = queueSender.SendJsonString(modelQueueName, jobJSONStr, podJobStr, granularity)
	if err == nil {
		sender.modelMapper.AddModelInfo(clusterID, pdUnit, dataGranularity, metricType.String(), map[string]string{
			"namespace":     podNS,
			"name":          podName,
			"containerName": ctName,
		})
	} else {
		scope.Errorf("[POD][%s][%s/%s/%s] Send model job payload failed. %s",
			dataGranularity, podNS, podName, ctName, err.Error())
	}

}

func (sender *podModelJobSender) getLastMIdPrediction(datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	pod *datahub_resources.Pod, granularity int64) ([]*datahub_predictions.ContainerPrediction, error) {
	containerPredictions := []*datahub_predictions.ContainerPrediction{}
	dataGranularity := queue.GetGranularityStr(granularity)
	podNS := pod.GetObjectMeta().GetNamespace()
	podName := pod.GetObjectMeta().GetName()
	podPredictRes, err := datahubServiceClnt.ListPodPredictions(context.Background(),
		&datahub_predictions.ListPodPredictionsRequest{
			Granularity: granularity,
			ObjectMeta: []*datahub_resources.ObjectMeta{
				&datahub_resources.ObjectMeta{
					Name:      podName,
					Namespace: podNS,
				},
			},
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
		return containerPredictions, err
	}

	lastMid := ""
	if len(podPredictRes.GetPodPredictions()) == 0 {
		return []*datahub_predictions.ContainerPrediction{}, nil
	}

	lastPodPrediction := podPredictRes.GetPodPredictions()[0]

	for _, lctPrediction := range lastPodPrediction.GetContainerPredictions() {
		lctPDRData := lctPrediction.GetPredictedRawData()
		if lctPDRData == nil {
			continue
		}
		for _, pdRD := range lctPDRData {
			for _, theData := range pdRD.GetData() {
				lastMid = theData.GetModelId()
				break
			}

			if lastMid == "" {
				scope.Warnf("[POD][%s][%s/%s/%s] Query last model id for metric %s is empty",
					dataGranularity, podNS, podName, lctPrediction.GetName(), pdRD.GetMetricType())
			}
			podPredictRes, err = datahubServiceClnt.ListPodPredictions(context.Background(),
				&datahub_predictions.ListPodPredictionsRequest{
					Granularity: granularity,
					ObjectMeta: []*datahub_resources.ObjectMeta{
						&datahub_resources.ObjectMeta{
							Name:      podName,
							Namespace: podNS,
						},
					},
					QueryCondition: &datahub_common.QueryCondition{
						Order: datahub_common.QueryCondition_DESC,
						TimeRange: &datahub_common.TimeRange{
							Step: &duration.Duration{
								Seconds: granularity,
							},
						},
					},
					ModelId: lastMid,
				})
			if err != nil {
				scope.Errorf("[POD][%s][%s/%s/%s] Query last model id %v for metric %s failed",
					dataGranularity, podNS, podName, lctPrediction.GetName(), lastMid, pdRD.GetMetricType())
				continue
			}

			if len(podPredictRes.GetPodPredictions()) == 0 {
				scope.Errorf("[POD][%s][%s/%s/%s] No prediction found for last model id %v for metric %s",
					dataGranularity, podNS, podName, lctPrediction.GetName(), lastMid, pdRD.GetMetricType())
				continue
			}

			lastPodPrediction := podPredictRes.GetPodPredictions()[0]

			for _, lctPrediction := range lastPodPrediction.GetContainerPredictions() {
				cpIdx := -1
				for idx, containerPrediction := range containerPredictions {
					if containerPrediction.GetName() == lctPrediction.GetName() {
						cpIdx = idx
						break
					}
				}
				if cpIdx == -1 {
					containerPredictions = append(containerPredictions, lctPrediction)
					continue
				}

				for _, pdRD := range lctPrediction.GetPredictedRawData() {
					metricFound := false
					for _, resultPDRD := range containerPredictions[cpIdx].GetPredictedRawData() {
						if pdRD.GetMetricType() == resultPDRD.GetMetricType() {
							metricFound = true
							break
						}
					}
					if !metricFound {
						containerPredictions[cpIdx].PredictedRawData = append(containerPredictions[cpIdx].PredictedRawData, pdRD)
					}
				}
				for _, pdRD := range lctPrediction.GetPredictedLowerboundData() {
					metricFound := false
					for _, resultPDRD := range containerPredictions[cpIdx].GetPredictedLowerboundData() {
						if pdRD.GetMetricType() == resultPDRD.GetMetricType() {
							metricFound = true
							break
						}
					}
					if !metricFound {
						containerPredictions[cpIdx].PredictedLowerboundData = append(containerPredictions[cpIdx].PredictedLowerboundData, pdRD)
					}
				}
				for _, pdRD := range lctPrediction.GetPredictedUpperboundData() {
					metricFound := false
					for _, resultPDRD := range containerPredictions[cpIdx].GetPredictedUpperboundData() {
						if pdRD.GetMetricType() == resultPDRD.GetMetricType() {
							metricFound = true
							break
						}
					}
					if !metricFound {
						containerPredictions[cpIdx].PredictedUpperboundData = append(containerPredictions[cpIdx].PredictedUpperboundData, pdRD)
					}
				}
			}
		}
	}

	return containerPredictions, nil
}

func (sender *podModelJobSender) getQueryMetricStartTime(metricData *datahub_predictions.MetricData) int64 {
	mD := metricData.GetData()
	if len(mD) > 0 {
		return mD[len(mD)-1].GetTime().GetSeconds()
	}
	return 0
}

func (sender *podModelJobSender) sendJobByMetrics(pod *datahub_resources.Pod, queueSender queue.QueueSender,
	pdUnit string, granularity int64, predictionStep int64, datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	lastPredictionContainers []*datahub_predictions.ContainerPrediction) {
	clusterID := pod.GetObjectMeta().GetClusterName()
	podNS := pod.GetObjectMeta().GetNamespace()
	podName := pod.GetObjectMeta().GetName()
	dataGranularity := queue.GetGranularityStr(granularity)

	nowSeconds := time.Now().Unix()

	if len(lastPredictionContainers) == 0 {
		for _, ct := range pod.GetContainers() {
			for _, metricType := range []datahub_common.MetricType{
				datahub_common.MetricType_MEMORY_USAGE_BYTES,
				datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
			} {
				scope.Infof("[POD][%s][%s/%s/%s] No last prediction found, send model jobs",
					dataGranularity, podNS, podName, ct.GetName())
				sender.sendJob(pod, queueSender, pdUnit, granularity, ct.GetName(), metricType)
			}
		}
		return
	}

	for _, lastPredictionContainer := range lastPredictionContainers {
		lastPredictionMetrics := lastPredictionContainer.GetPredictedRawData()
		if len(lastPredictionMetrics) == 0 {
			scope.Infof("[POD][%s][%s/%s/%s] No any last metric prediction found, send model jobs",
				dataGranularity, podNS, podName, lastPredictionContainer.GetName())
			for _, metricType := range []datahub_common.MetricType{
				datahub_common.MetricType_MEMORY_USAGE_BYTES,
				datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
			} {
				sender.sendJob(pod, queueSender, pdUnit, granularity, lastPredictionContainer.GetName(), metricType)
			}
			continue
		}

		for _, lastPredictionMetric := range lastPredictionMetrics {
			if len(lastPredictionMetric.GetData()) == 0 {
				scope.Infof("[POD][%s][%s/%s/%s] No last prediction metric %s found, send model jobs",
					dataGranularity, podNS, podName, lastPredictionContainer.GetName(), lastPredictionMetric.GetMetricType().String())
				sender.sendJob(pod, queueSender, pdUnit, granularity, lastPredictionContainer.GetName(), lastPredictionMetric.GetMetricType())
				continue
			} else {
				lastPrediction := lastPredictionMetric.GetData()[0]
				lastPredictionTime := lastPrediction.GetTime().GetSeconds()
				if lastPrediction != nil && lastPredictionTime <= nowSeconds {
					scope.Infof("[POD][%s][%s/%s/%s] send model job due to no predict metric %s found or is out of date",
						dataGranularity, podNS, podName, lastPredictionContainer.GetName(), lastPredictionMetric.GetMetricType().String())
					sender.sendJob(pod, queueSender, pdUnit, granularity, lastPredictionContainer.GetName(), lastPredictionMetric.GetMetricType())
					continue
				}

				queryStartTime := time.Now().Unix() - predictionStep*granularity
				firstPDTime := sender.getQueryMetricStartTime(lastPredictionMetric)
				if firstPDTime > 0 && firstPDTime <= time.Now().Unix() {
					queryStartTime = firstPDTime
				}
				aggFun := datahub_common.TimeRange_AVG
				if granularity == 30 {
					aggFun = datahub_common.TimeRange_MAX
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
								AggregateFunction: aggFun,
							},
						},
						ObjectMeta: []*datahub_resources.ObjectMeta{
							&datahub_resources.ObjectMeta{
								Name:      pod.GetObjectMeta().GetName(),
								Namespace: pod.GetObjectMeta().GetNamespace(),
							},
						},
						MetricTypes: []datahub_common.MetricType{
							lastPredictionMetric.GetMetricType(),
						},
					})

				if err != nil {
					scope.Errorf("[POD][%s][%s/%s] List metric for sending model job failed: %s",
						dataGranularity, podNS, podName, err.Error())
					continue
				}
				podMetrics := podMetricsRes.GetPodMetrics()

				predictRawData := lastPredictionMetrics
				for _, predictRawDatum := range predictRawData {
					for _, podMetric := range podMetrics {
						containerMetrics := podMetric.GetContainerMetrics()
						for _, containerMetric := range containerMetrics {
							containerName := containerMetric.GetName()
							metricData := containerMetric.GetMetricData()

							for _, metricDatum := range metricData {
								mData := metricDatum.GetData()
								pData := []*datahub_predictions.Sample{}
								if lastPredictionContainer.GetName() == containerMetric.GetName() &&
									metricDatum.GetMetricType() == predictRawDatum.GetMetricType() {
									pData = append(pData, predictRawDatum.GetData()...)

									metricsNeedToModel, drift := DriftEvaluation(consts.UnitTypePod, predictRawDatum.GetMetricType(), granularity, mData, pData, map[string]string{
										"clusterID":         clusterID,
										"podNS":             podNS,
										"podName":           podName,
										"containerName":     containerName,
										"targetDisplayName": fmt.Sprintf("[POD][%s][%s/%s/%s]", dataGranularity, podNS, podName, containerName),
									}, sender.metricExporter)

									for _, mntm := range metricsNeedToModel {
										if drift {
											scope.Infof("[POD][%s][%s/%s/%s] Export metric %s drift counter",
												dataGranularity, podNS, podName, containerName, mntm)
											sender.metricExporter.AddContainerMetricDrift(clusterID, podNS, podName, containerName,
												queue.GetGranularityStr(granularity), mntm.String(), time.Now().Unix(), 1.0)
										}
										isModeling := sender.modelMapper.IsModeling(clusterID, pdUnit, dataGranularity, mntm.String(), map[string]string{
											"namespace":     podNS,
											"name":          podName,
											"containerName": containerName,
										})
										if !isModeling || (isModeling && sender.modelMapper.IsModelTimeout(
											clusterID, pdUnit, dataGranularity, mntm.String(), map[string]string{
												"namespace":     podNS,
												"name":          podName,
												"containerName": containerName,
											})) {
											sender.sendJob(pod, queueSender, pdUnit, granularity, containerName, mntm)
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
}
