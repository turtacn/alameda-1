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

	datahubServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(sender.datahubGrpcCn)
	for _, namespace := range namespaces {
		namespaceName := namespace.GetObjectMeta().GetName()
		lastPredictionMetrics, err := sender.getLastMIdPrediction(datahubServiceClnt, namespace, granularity)
		if err != nil {
			scope.Infof("Get namespace %s last prediction failed: %s",
				namespaceName, err.Error())
			continue
		}
		if lastPredictionMetrics == nil && err == nil {
			scope.Infof("No prediction found of namespace %s",
				namespaceName)
		}

		sender.sendJobByMetrics(namespace, queueSender, pdUnit, granularity, predictionStep,
			datahubServiceClnt, lastPredictionMetrics)
	}
}

func (sender *namespaceModelJobSender) sendJob(namespace *datahub_resources.Namespace, queueSender queue.QueueSender, pdUnit string,
	granularity int64, namespaceInfo *modelInfo) {
	namespaceName := namespace.GetObjectMeta().GetName()
	dataGranularity := queue.GetGranularityStr(granularity)
	marshaler := jsonpb.Marshaler{}
	namespaceStr, err := marshaler.MarshalToString(namespace)
	if err != nil {
		scope.Errorf("Encode pb message failed for namespace %s with granularity seconds %v. %s",
			namespaceName, granularity, err.Error())
		return
	}
	if len(namespaceInfo.ModelMetrics) > 0 && namespaceStr != "" {
		jb := queue.NewJobBuilder(pdUnit, granularity, namespaceStr)
		jobJSONStr, err := jb.GetJobJSONString()
		if err != nil {
			scope.Errorf(
				"Prepare model job payload failed for namespace %s with granularity seconds %v. %s",
				namespaceName, granularity, err.Error())
			return
		}

		nsJobStr := fmt.Sprintf("%s/%v", namespaceName, granularity)
		scope.Infof("Try to send namespace model job: %s", nsJobStr)
		err = queueSender.SendJsonString(modelQueueName, jobJSONStr, nsJobStr)
		if err == nil {
			sender.modelMapper.AddModelInfo(pdUnit, dataGranularity, namespaceInfo)
		} else {
			scope.Errorf(
				"Send model job payload failed for namespace %s with granularity seconds %v. %s",
				namespaceName, granularity, err.Error())
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
		return nil, fmt.Errorf("Query namespace %s last prediction id with granularity %v seconds failed",
			namespaceName, granularity)
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
					modelId := pdD.GetModelId()
					if modelId != "" {
						mIdNSPrediction, err := sender.getPredictionByMId(datahubServiceClnt, namespace, granularity, modelId)
						if err != nil {
							scope.Errorf("Query namespace %s prediction with granularity %v seconds and model Id %s failed. %s",
								namespaceName, granularity, modelId, err.Error())
						}
						for _, mIdNSPD := range mIdNSPrediction {
							metricData = append(metricData, mIdNSPD.GetPredictedRawData()...)
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
	namespace *datahub_resources.Namespace, granularity int64, modelId string) ([]*datahub_predictions.NamespacePrediction, error) {
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
			ModelId: modelId,
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
		scope.Infof("No prediction metrics found of namespace %s, send model jobs with granularity %v",
			namespaceName, granularity)
		return
	}

	namespaceInfo := sender.genNamespaceInfo(namespaceName)
	for _, lastPredictionMetric := range lastPredictionMetrics {
		if len(lastPredictionMetric.GetData()) == 0 {
			namespaceInfo := sender.genNamespaceInfo(namespaceName,
				datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
				datahub_common.MetricType_MEMORY_USAGE_BYTES)
			sender.sendJob(namespace, queueSender, pdUnit, granularity, namespaceInfo)
			scope.Infof("No prediction metric %s found of namespace %s, send model jobs with granularity %v",
				lastPredictionMetric.GetMetricType().String(), namespaceName, granularity)
			return
		} else {
			lastPrediction := lastPredictionMetric.GetData()[0]
			lastPredictionTime := lastPredictionMetric.GetData()[0].GetTime().GetSeconds()
			if lastPrediction != nil && lastPredictionTime <= nowSeconds {
				scope.Infof("namespace prediction %s is out of date due to last predict time is %v (current: %v)",
					namespaceName, lastPredictionTime, nowSeconds)
			}
			if lastPrediction != nil && lastPredictionTime <= nowSeconds {
				namespaceInfo := sender.genNamespaceInfo(namespaceName,
					datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
					datahub_common.MetricType_MEMORY_USAGE_BYTES)
				scope.Infof("send namespace %s model job due to no predict found or predict is out of date, send model jobs with granularity %v",
					namespaceName, granularity)
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
				scope.Errorf("Get namespace %s Prediction with granularity %v for sending model job failed: %s",
					namespaceName, granularity, err.Error())
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
				scope.Errorf("List namespaces %s metric with granularity %v for sending model job failed: %s",
					namespaceName, granularity, err.Error())
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
								metricsNeedToModel, drift := DriftEvaluation(UnitTypeNamespace, predictRawDatum.GetMetricType(), granularity, mData, pData, map[string]string{
									"namespaceName":     namespaceName,
									"targetDisplayName": fmt.Sprintf("namespace %s", namespaceName),
								}, sender.metricExporter)
								if drift {
									scope.Infof("export namespace %s drift counter with granularity %s",
										namespaceName, dataGranularity)
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
