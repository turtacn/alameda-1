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
	for _, pod := range pods {
		go sender.sendPodModelJobs(pod, queueSender, pdUnit, granularity, predictionStep)
	}
}

func (sender *podModelJobSender) sendPodModelJobs(pod *datahub_resources.Pod, queueSender queue.QueueSender,
	pdUnit string, granularity int64, predictionStep int64) {
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
	if lastPredictionContainers == nil && err == nil {
		scope.Infof("[POD][%s][%s/%s] No prediction found", dataGranularity, podNS, podName)
	}
	sender.sendJobByMetrics(pod, queueSender, pdUnit, granularity, predictionStep,
		datahubServiceClnt, lastPredictionContainers)
}

func (sender *podModelJobSender) sendJob(pod *datahub_resources.Pod, queueSender queue.QueueSender, pdUnit string,
	granularity int64, podInfo *modelInfo) {
	marshaler := jsonpb.Marshaler{}
	podNS := pod.GetObjectMeta().GetNamespace()
	podName := pod.GetObjectMeta().GetName()
	dataGranularity := queue.GetGranularityStr(granularity)
	podStr, err := marshaler.MarshalToString(pod)

	if err != nil {
		scope.Errorf("[POD][%s][%s/%s] Encode pb message failed. %s",
			dataGranularity, podNS, podName, err.Error())
		return
	}
	for _, ct := range podInfo.Containers {
		if len(ct.ModelMetrics) > 0 && podStr != "" {
			jb := queue.NewJobBuilder(pdUnit, granularity, podStr)
			jobJSONStr, err := jb.GetJobJSONString()
			if err != nil {
				scope.Errorf("[POD][%s][%s/%s] Prepare model job payload failed. %s",
					dataGranularity, podNS, podName, err.Error())
				return
			}

			podJobStr := fmt.Sprintf("%s/%s/%v", podNS, podName, granularity)
			scope.Infof("[POD][%s][%s/%s] Try to send pod model job: %s", dataGranularity, podNS, podName, podJobStr)
			err = queueSender.SendJsonString(modelQueueName, jobJSONStr, podJobStr, granularity)
			if err == nil {
				sender.modelMapper.AddModelInfo(pdUnit, dataGranularity, podInfo)
			} else {
				scope.Errorf("[POD][%s][%s/%s] Send model job payload failed. %s",
					dataGranularity, podNS, podName, err.Error())
			}
			break
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
	podInfo.Containers = containers
	return podInfo
}

func (sender *podModelJobSender) getLastMIdPrediction(datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	pod *datahub_resources.Pod, granularity int64) ([]*datahub_predictions.ContainerPrediction, error) {
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
		return nil, err
	}

	lastPid := ""
	if len(podPredictRes.GetPodPredictions()) > 0 {
		lastPodPrediction := podPredictRes.GetPodPredictions()[0]

		for _, lctPrediction := range lastPodPrediction.GetContainerPredictions() {
			lctPDRData := lctPrediction.GetPredictedRawData()
			if lctPDRData == nil {
				lctPDRData = lctPrediction.GetPredictedUpperboundData()
			}
			if lctPDRData == nil {
				lctPDRData = lctPrediction.GetPredictedLowerboundData()
			}
			for _, pdRD := range lctPDRData {
				for _, theData := range pdRD.GetData() {
					lastPid = theData.GetPredictionId()
					break
				}
				if lastPid != "" {
					break
				}
			}
			if lastPid != "" {
				break
			}
		}
	} else {
		return []*datahub_predictions.ContainerPrediction{}, nil
	}
	if lastPid == "" {
		return nil, fmt.Errorf("[POD][%s][%s/%s] Query last prediction id failed",
			dataGranularity, podNS, podName)
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
			PredictionId: lastPid,
		})
	if err != nil {
		return nil, err
	}

	if len(podPredictRes.GetPodPredictions()) > 0 {
		containerPredictions := []*datahub_predictions.ContainerPrediction{}

		lastPodPrediction := podPredictRes.GetPodPredictions()[0]

		for _, lctPrediction := range lastPodPrediction.GetContainerPredictions() {
			for _, pdRD := range lctPrediction.GetPredictedRawData() {
				for _, pdD := range pdRD.GetData() {
					modelID := pdD.GetModelId()
					if modelID != "" {
						mIDPodPrediction, err := sender.getPredictionByMId(datahubServiceClnt, pod, granularity, modelID)
						if err != nil {
							scope.Errorf("[POD][%s][%s/%s] Query prediction with model Id %s failed. %s",
								dataGranularity, podNS, podName, modelID, err.Error())
						}
						for _, podPrediction := range mIDPodPrediction {
							for _, midCtPrediction := range podPrediction.GetContainerPredictions() {
								ctFound := false
								for _, containerPrediction := range containerPredictions {
									if containerPrediction.GetName() == midCtPrediction.GetName() {
										containerPrediction.PredictedRawData = append(containerPrediction.PredictedRawData,
											midCtPrediction.GetPredictedRawData()...)
										containerPrediction.PredictedUpperboundData = append(containerPrediction.PredictedUpperboundData,
											midCtPrediction.GetPredictedUpperboundData()...)
										containerPrediction.PredictedLowerboundData = append(containerPrediction.PredictedLowerboundData,
											midCtPrediction.GetPredictedLowerboundData()...)
										ctFound = true
										break
									}
								}
								if !ctFound {
									containerPredictions = append(containerPredictions, midCtPrediction)
								}
							}
						}
						break
					}
				}
			}
		}
		return containerPredictions, nil
	}
	return nil, nil
}

func (sender *podModelJobSender) getPredictionByMId(datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	pod *datahub_resources.Pod, granularity int64, modelID string) ([]*datahub_predictions.PodPrediction, error) {
	podPredictRes, err := datahubServiceClnt.ListPodPredictions(context.Background(),
		&datahub_predictions.ListPodPredictionsRequest{
			Granularity: granularity,
			ObjectMeta: []*datahub_resources.ObjectMeta{
				&datahub_resources.ObjectMeta{
					Name:      pod.GetObjectMeta().GetName(),
					Namespace: pod.GetObjectMeta().GetNamespace(),
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
			ModelId: modelID,
		})
	return podPredictRes.GetPodPredictions(), err
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
	podNS := pod.GetObjectMeta().GetNamespace()
	podName := pod.GetObjectMeta().GetName()
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
		scope.Infof("[POD][%s][%s/%s] No last prediction found, send model jobs",
			dataGranularity, podNS, podName)
		return
	}

	podInfo := sender.genPodInfo(podNS, podName)
	containers := []*container{}
	for _, lastPredictionContainer := range lastPredictionContainers {
		containerInfo := &container{
			Name:         lastPredictionContainer.GetName(),
			ModelMetrics: []datahub_common.MetricType{},
		}
		lastPredictionMetrics := []*datahub_predictions.MetricData{}
		if len(lastPredictionContainer.GetPredictedRawData()) > 0 {
			lastPredictionMetrics = lastPredictionContainer.GetPredictedRawData()
		} else if len(lastPredictionContainer.GetPredictedLowerboundData()) > 0 {
			lastPredictionMetrics = lastPredictionContainer.GetPredictedLowerboundData()
		} else if len(lastPredictionContainer.GetPredictedUpperboundData()) > 0 {
			lastPredictionMetrics = lastPredictionContainer.GetPredictedUpperboundData()
		} else {
			podInfo := sender.genPodInfoWithAllMetrics(podNS, podName, pod)
			sender.sendJob(pod, queueSender, pdUnit, granularity, podInfo)
			scope.Infof("[POD][%s][%s/%s] No any last container metric prediction %s found, send model jobs",
				dataGranularity, podNS, podName, lastPredictionContainer.GetName())
			return
		}

		for _, lastPredictionMetric := range lastPredictionMetrics {
			if len(lastPredictionMetric.GetData()) == 0 {
				podInfo := sender.genPodInfoWithAllMetrics(podNS, podName, pod)
				sender.sendJob(pod, queueSender, pdUnit, granularity, podInfo)
				scope.Infof("[POD][%s][%s/%s] No last prediction metric %s found, send model jobs",
					dataGranularity, lastPredictionMetric.GetMetricType().String(), podNS, podName)
				return
			} else {
				lastPrediction := lastPredictionMetric.GetData()[0]
				lastPredictionTime := lastPrediction.GetTime().GetSeconds()
				if lastPrediction != nil && lastPredictionTime <= nowSeconds {
					podInfo := sender.genPodInfoWithAllMetrics(podNS, podName, pod)
					scope.Infof("[POD][%s][%s/%s] send model job due to no predict found or predict is out of date",
						dataGranularity, podNS, podName)
					sender.sendJob(pod, queueSender, pdUnit, granularity, podInfo)
					return
				}

				// one container metric prediction series
				podPredictRes, err := datahubServiceClnt.ListPodPredictions(context.Background(),
					&datahub_predictions.ListPodPredictionsRequest{
						ObjectMeta: []*datahub_resources.ObjectMeta{
							&datahub_resources.ObjectMeta{
								Name:      pod.GetObjectMeta().GetName(),
								Namespace: pod.GetObjectMeta().GetNamespace(),
							},
						},
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
					})

				if err != nil {
					scope.Errorf("[POD][%s][%s/%s] List metric for sending model job failed: %s",
						dataGranularity, podNS, podName, err.Error())
					continue
				}
				podMetrics := podMetricsRes.GetPodMetrics()

				for _, podPrediction := range podPredictions {
					containerPredictions := podPrediction.GetContainerPredictions()
					for _, containerPrediction := range containerPredictions {
						predictRawData := containerPrediction.GetPredictedRawData()
						for _, predictRawDatum := range predictRawData {
							for _, podMetric := range podMetrics {
								containerMetrics := podMetric.GetContainerMetrics()
								for _, containerMetric := range containerMetrics {
									containerName := containerMetric.GetName()
									metricData := containerMetric.GetMetricData()

									for _, metricDatum := range metricData {
										mData := metricDatum.GetData()
										pData := []*datahub_predictions.Sample{}
										if containerPrediction.GetName() == containerMetric.GetName() &&
											metricDatum.GetMetricType() == predictRawDatum.GetMetricType() {
											pData = append(pData, predictRawDatum.GetData()...)

											metricsNeedToModel, drift := DriftEvaluation(UnitTypePod, predictRawDatum.GetMetricType(), granularity, mData, pData, map[string]string{
												"podNS":             podNS,
												"podName":           podName,
												"containerName":     containerName,
												"targetDisplayName": fmt.Sprintf("[POD][%s][%s/%s]", dataGranularity, podNS, podName),
											}, sender.metricExporter)
											if drift {
												scope.Infof("[POD][%s][%s/%s] Export drift counter",
													dataGranularity, podNS, podName)
												sender.metricExporter.AddPodMetricDrift(podNS, podName,
													queue.GetGranularityStr(granularity), time.Now().Unix(), 1.0)
											}
											containerInfo.ModelMetrics = append(containerInfo.ModelMetrics, metricsNeedToModel...)
										}
									}
								}
							}
						}
					}
				}
			}
		}
		containers = append(containers, containerInfo)
	}
	podInfo.Containers = containers
	isModeling := sender.modelMapper.IsModeling(pdUnit, dataGranularity, podInfo)
	if !isModeling || (isModeling && sender.modelMapper.IsModelTimeout(pdUnit, dataGranularity, podInfo)) {
		sender.sendJob(pod, queueSender, pdUnit, granularity, podInfo)
		return
	}
}
