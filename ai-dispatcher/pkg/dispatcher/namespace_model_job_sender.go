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

type namespaceModelJobSender struct {
	datahubGrpcCn  *grpc.ClientConn
	modelMapper    *ModelMapper
	metricExporter *metrics.Exporter
}

func NewNamespaceModelJobSender(datahubGrpcCn *grpc.ClientConn, modelMapper *ModelMapper,
	metricExporter *metrics.Exporter) *namespaceModelJobSender {
	return &namespaceModelJobSender{
		datahubGrpcCn:  datahubGrpcCn,
		modelMapper:    modelMapper,
		metricExporter: metricExporter,
	}
}

func (sender *namespaceModelJobSender) sendModelJobs(namespaces []*datahub_resources.Namespace,
	queueSender queue.QueueSender, pdUnit string, granularity int64, predictionStep int64) {
	for _, namespace := range namespaces {
		go sender.sendNamespaceModelJobs(namespace, queueSender, pdUnit, granularity, predictionStep)
	}
}

func (sender *namespaceModelJobSender) sendNamespaceModelJobs(namespace *datahub_resources.Namespace,
	queueSender queue.QueueSender, pdUnit string, granularity int64, predictionStep int64) {
	dataGranularity := queue.GetGranularityStr(granularity)
	datahubServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(sender.datahubGrpcCn)

	namespaceName := namespace.GetObjectMeta().GetName()
	lastPredictionMetrics, err := sender.getLastMIdPrediction(datahubServiceClnt, namespace, granularity)
	if err != nil {
		scope.Infof("[NAMESPACE][%s][%s] Get last prediction failed: %s",
			dataGranularity, namespaceName, err.Error())
		return
	}
	if lastPredictionMetrics == nil && err == nil {
		scope.Infof("[NAMESPACE][%s][%s] No prediction found",
			dataGranularity, namespaceName)
	}

	sender.sendJobByMetrics(namespace, queueSender, pdUnit, granularity, predictionStep,
		datahubServiceClnt, lastPredictionMetrics)
}

func (sender *namespaceModelJobSender) sendJob(namespace *datahub_resources.Namespace,
	queueSender queue.QueueSender, pdUnit string, granularity int64, namespaceInfo *modelInfo) {
	namespaceName := namespace.GetObjectMeta().GetName()
	dataGranularity := queue.GetGranularityStr(granularity)
	marshaler := jsonpb.Marshaler{}
	namespaceStr, err := marshaler.MarshalToString(namespace)
	if err != nil {
		scope.Errorf("[NAMESPACE][%s][%s] Encode pb message failed. %s",
			dataGranularity, namespaceName, err.Error())
		return
	}
	if len(namespaceInfo.ModelMetrics) > 0 && namespaceStr != "" {
		jb := queue.NewJobBuilder(pdUnit, granularity, namespaceStr)
		jobJSONStr, err := jb.GetJobJSONString()
		if err != nil {
			scope.Errorf(
				"[NAMESPACE][%s][%s] Prepare model job payload failed. %s",
				dataGranularity, namespaceName, err.Error())
			return
		}

		nsJobStr := fmt.Sprintf("%s/%v", namespaceName, granularity)
		scope.Infof("[NAMESPACE][%s][%s] Try to send namespace model job: %s",
			dataGranularity, namespaceName, nsJobStr)
		err = queueSender.SendJsonString(modelQueueName, jobJSONStr, nsJobStr, granularity)
		if err == nil {
			sender.modelMapper.AddModelInfo(pdUnit, dataGranularity, namespaceInfo)
		} else {
			scope.Errorf(
				"[NAMESPACE][%s][%s] Send model job payload failed. %s",
				dataGranularity, namespaceName, err.Error())
		}
	}
}

func (sender *namespaceModelJobSender) genNamespaceInfo(namespaceName string,
	modelMetrics ...datahub_common.MetricType) *modelInfo {
	namespaceInfo := new(modelInfo)
	namespaceInfo.Name = namespaceName
	namespaceInfo.ModelMetrics = modelMetrics
	namespaceInfo.SetTimeStamp(time.Now().Unix())
	return namespaceInfo
}

func (sender *namespaceModelJobSender) getLastMIdPrediction(datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	namespace *datahub_resources.Namespace, granularity int64) ([]*datahub_predictions.MetricData, error) {
	dataGranularity := queue.GetGranularityStr(granularity)
	namespaceName := namespace.GetObjectMeta().GetName()
	namespacePredictRes, err := datahubServiceClnt.ListNamespacePredictions(context.Background(),
		&datahub_predictions.ListNamespacePredictionsRequest{
			ObjectMeta: []*datahub_resources.ObjectMeta{
				&datahub_resources.ObjectMeta{
					Name: namespaceName,
				},
			},
			Granularity: granularity,
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
	if len(namespacePredictRes.GetNamespacePredictions()) > 0 {
		lastNamespacePrediction := namespacePredictRes.GetNamespacePredictions()[0]
		lnsPDRData := lastNamespacePrediction.GetPredictedRawData()
		if lnsPDRData == nil {
			lnsPDRData = lastNamespacePrediction.GetPredictedLowerboundData()
		}
		if lnsPDRData == nil {
			lnsPDRData = lastNamespacePrediction.GetPredictedUpperboundData()
		}
		for _, pdRD := range lnsPDRData {
			for _, theData := range pdRD.GetData() {
				lastPid = theData.GetPredictionId()
				break
			}
			if lastPid != "" {
				break
			}
		}
	} else {
		return []*datahub_predictions.MetricData{}, nil
	}

	if lastPid == "" {
		return nil, fmt.Errorf("[NAMESPACE][%s][%s] Query last prediction id failed",
			dataGranularity, namespaceName)
	}
	namespacePredictRes, err = datahubServiceClnt.ListNamespacePredictions(context.Background(),
		&datahub_predictions.ListNamespacePredictionsRequest{
			ObjectMeta: []*datahub_resources.ObjectMeta{
				&datahub_resources.ObjectMeta{
					Name: namespaceName,
				},
			},
			Granularity: granularity,
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

	if len(namespacePredictRes.GetNamespacePredictions()) > 0 {
		metricData := []*datahub_predictions.MetricData{}
		for _, nsPrediction := range namespacePredictRes.GetNamespacePredictions() {
			for _, pdRD := range nsPrediction.GetPredictedRawData() {
				for _, pdD := range pdRD.GetData() {
					modelID := pdD.GetModelId()
					if modelID != "" {
						mIDNSPrediction, err := sender.getPredictionByMId(datahubServiceClnt, namespace, granularity, modelID)
						if err != nil {
							scope.Errorf("[NAMESPACE][%s][%s] Query prediction with model Id %s failed. %s",
								dataGranularity, namespaceName, modelID, err.Error())
						}
						for _, mIDNSPD := range mIDNSPrediction {
							metricData = append(metricData, mIDNSPD.GetPredictedRawData()...)
						}
						break
					}
				}
			}
		}
		return metricData, nil
	}
	return nil, nil
}

func (sender *namespaceModelJobSender) getPredictionByMId(datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	namespace *datahub_resources.Namespace, granularity int64, modelID string) ([]*datahub_predictions.NamespacePrediction, error) {
	namespacePredictRes, err := datahubServiceClnt.ListNamespacePredictions(context.Background(),
		&datahub_predictions.ListNamespacePredictionsRequest{
			Granularity: granularity,
			ObjectMeta: []*datahub_resources.ObjectMeta{
				&datahub_resources.ObjectMeta{
					Name: namespace.GetObjectMeta().GetName(),
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
	return namespacePredictRes.GetNamespacePredictions(), err
}

func (sender *namespaceModelJobSender) getQueryMetricStartTime(descNamespacePredictions []*datahub_predictions.NamespacePrediction) int64 {
	if len(descNamespacePredictions) > 0 {
		pdMDs := descNamespacePredictions[len(descNamespacePredictions)-1].GetPredictedRawData()
		for _, pdMD := range pdMDs {
			mD := pdMD.GetData()
			if len(mD) > 0 {
				return mD[len(mD)-1].GetTime().GetSeconds()
			}
		}
	}
	return 0
}

func (sender *namespaceModelJobSender) sendJobByMetrics(namespace *datahub_resources.Namespace, queueSender queue.QueueSender,
	pdUnit string, granularity int64, predictionStep int64, datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	lastPredictionMetrics []*datahub_predictions.MetricData) {
	namespaceName := namespace.GetObjectMeta().GetName()
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

	if len(lastPredictionMetrics) == 0 {
		namespaceInfo := sender.genNamespaceInfo(namespaceName,
			datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
			datahub_common.MetricType_MEMORY_USAGE_BYTES)
		sender.sendJob(namespace, queueSender, pdUnit, granularity, namespaceInfo)
		scope.Infof("[NAMESPACE][%s][%s] No prediction metrics found, send model jobs.",
			dataGranularity, namespaceName)
		return
	}

	namespaceInfo := sender.genNamespaceInfo(namespaceName)
	for _, lastPredictionMetric := range lastPredictionMetrics {
		if len(lastPredictionMetric.GetData()) == 0 {
			namespaceInfo := sender.genNamespaceInfo(namespaceName,
				datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
				datahub_common.MetricType_MEMORY_USAGE_BYTES)
			sender.sendJob(namespace, queueSender, pdUnit, granularity, namespaceInfo)
			scope.Infof("[NAMESPACE][%s][%s] No prediction metric %s found, send model jobs",
				dataGranularity, namespaceName, lastPredictionMetric.GetMetricType().String())
			return
		} else {
			lastPrediction := lastPredictionMetric.GetData()[0]
			lastPredictionTime := lastPredictionMetric.GetData()[0].GetTime().GetSeconds()
			if lastPrediction != nil && lastPredictionTime <= nowSeconds {
				namespaceInfo := sender.genNamespaceInfo(namespaceName,
					datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
					datahub_common.MetricType_MEMORY_USAGE_BYTES)
				scope.Infof("[NAMESPACE][%s][%s] Send model job due to no predict found or predict is out of date",
					dataGranularity, namespaceName)
				sender.sendJob(namespace, queueSender, pdUnit, granularity, namespaceInfo)
				return
			}

			namespacePredictRes, err := datahubServiceClnt.ListNamespacePredictions(context.Background(),
				&datahub_predictions.ListNamespacePredictionsRequest{
					ObjectMeta: []*datahub_resources.ObjectMeta{
						&datahub_resources.ObjectMeta{
							Name: namespaceName,
						},
					},
					ModelId:        lastPrediction.GetModelId(),
					Granularity:    granularity,
					QueryCondition: queryCondition,
				})
			if err != nil {
				scope.Errorf("[NAMESPACE][%s][%s] Get prediction for sending model job failed: %s",
					dataGranularity, namespaceName, err.Error())
				continue
			}
			namespacePredictions := namespacePredictRes.GetNamespacePredictions()
			queryStartTime := time.Now().Unix() - predictionStep*granularity
			firstPDTime := sender.getQueryMetricStartTime(namespacePredictions)
			if firstPDTime > 0 {
				queryStartTime = firstPDTime
			}
			namespaceMetricsRes, err := datahubServiceClnt.ListNamespaceMetrics(context.Background(),
				&datahub_metrics.ListNamespaceMetricsRequest{
					QueryCondition: &datahub_common.QueryCondition{
						Order: datahub_common.QueryCondition_DESC,
						TimeRange: &datahub_common.TimeRange{
							StartTime: &timestamp.Timestamp{
								Seconds: queryStartTime,
							},
							Step: &duration.Duration{
								Seconds: granularity,
							},
							AggregateFunction: datahub_common.TimeRange_AVG,
						},
					},
					ObjectMeta: []*datahub_resources.ObjectMeta{
						&datahub_resources.ObjectMeta{
							Name: namespaceName,
						},
					},
				})
			if err != nil {
				scope.Errorf("[NAMESPACE][%s][%s] List metric for sending model job failed: %s",
					dataGranularity, namespaceName, err.Error())
				continue
			}
			namespaceMetrics := namespaceMetricsRes.GetNamespaceMetrics()
			for _, namespacePrediction := range namespacePredictions {
				predictRawData := namespacePrediction.GetPredictedRawData()
				for _, predictRawDatum := range predictRawData {
					for _, namespaceMetric := range namespaceMetrics {
						metricData := namespaceMetric.GetMetricData()
						for _, metricDatum := range metricData {
							mData := metricDatum.GetData()
							pData := []*datahub_predictions.Sample{}
							if metricDatum.GetMetricType() == predictRawDatum.GetMetricType() {
								pData = append(pData, predictRawDatum.GetData()...)
								metricsNeedToModel, drift := DriftEvaluation(UnitTypeNamespace, predictRawDatum.GetMetricType(),
									granularity, mData, pData, map[string]string{
										"namespaceName":     namespaceName,
										"targetDisplayName": fmt.Sprintf("[NAMESPACE][%s][%s]", dataGranularity, namespaceName),
									}, sender.metricExporter)
								if drift {
									scope.Infof("[NAMESPACE][%s][%s] Export drift counter",
										dataGranularity, namespaceName)
									sender.metricExporter.AddNamespaceMetricDrift(namespaceName, queue.GetGranularityStr(granularity),
										time.Now().Unix(), 1.0)
								}
								namespaceInfo.ModelMetrics = append(namespaceInfo.ModelMetrics, metricsNeedToModel...)
							}
						}
					}
				}
			}
		}
	}
	isModeling := sender.modelMapper.IsModeling(pdUnit, dataGranularity, namespaceInfo)
	if !isModeling || (isModeling && sender.modelMapper.IsModelTimeout(
		pdUnit, dataGranularity, namespaceInfo)) {
		sender.sendJob(namespace, queueSender, pdUnit, granularity, namespaceInfo)
	}
}
