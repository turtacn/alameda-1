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

type applicationModelJobSender struct {
	datahubGrpcCn  *grpc.ClientConn
	modelMapper    *ModelMapper
	metricExporter *metrics.Exporter
}

func NewApplicationModelJobSender(datahubGrpcCn *grpc.ClientConn, modelMapper *ModelMapper,
	metricExporter *metrics.Exporter) *applicationModelJobSender {
	return &applicationModelJobSender{
		datahubGrpcCn:  datahubGrpcCn,
		modelMapper:    modelMapper,
		metricExporter: metricExporter,
	}
}

func (sender *applicationModelJobSender) sendModelJobs(applications []*datahub_resources.Application,
	queueSender queue.QueueSender, pdUnit string, granularity int64, predictionStep int64) {
	for _, application := range applications {
		go sender.sendApplicationModelJobs(application,
			queueSender, pdUnit, granularity, predictionStep)
	}
}

func (sender *applicationModelJobSender) sendApplicationModelJobs(application *datahub_resources.Application,
	queueSender queue.QueueSender, pdUnit string, granularity int64, predictionStep int64) {
	dataGranularity := queue.GetGranularityStr(granularity)
	datahubServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(sender.datahubGrpcCn)

	applicationNS := application.GetObjectMeta().GetNamespace()
	applicationName := application.GetObjectMeta().GetName()

	lastPredictionMetrics, err := sender.getLastMIdPrediction(datahubServiceClnt, application, granularity)
	if err != nil {
		scope.Infof("[APPLICATION][%s][%s/%s] Get last prediction failed: %s",
			dataGranularity, applicationNS, applicationName, err.Error())
		return
	}
	if lastPredictionMetrics == nil && err == nil {
		scope.Infof("[APPLICATION][%s][%s/%s] No prediction found", dataGranularity,
			applicationNS, applicationName)
	}
	sender.sendJobByMetrics(application, queueSender, pdUnit, granularity, predictionStep,
		datahubServiceClnt, lastPredictionMetrics)
}

func (sender *applicationModelJobSender) sendJob(application *datahub_resources.Application, queueSender queue.QueueSender, pdUnit string,
	granularity int64, applicationInfo *modelInfo) {
	marshaler := jsonpb.Marshaler{}
	dataGranularity := queue.GetGranularityStr(granularity)
	applicationNS := application.GetObjectMeta().GetNamespace()
	applicationName := application.GetObjectMeta().GetName()
	applicationStr, err := marshaler.MarshalToString(application)
	if err != nil {
		scope.Errorf("[APPLICATION][%s][%s/%s] Encode pb message failed. %s",
			dataGranularity, applicationNS, applicationName, err.Error())
		return
	}
	if len(applicationInfo.ModelMetrics) > 0 && applicationStr != "" {
		jb := queue.NewJobBuilder(pdUnit, granularity, applicationStr)
		jobJSONStr, err := jb.GetJobJSONString()
		if err != nil {
			scope.Errorf(
				"[APPLICATION][%s][%s/%s] Prepare model job payload failed. %s",
				dataGranularity, applicationNS, applicationName, err.Error())
			return
		}

		appJobStr := fmt.Sprintf("%s/%s/%v", applicationNS, applicationName, granularity)
		scope.Infof("[APPLICATION][%s][%s/%s] Try to send application model job: %s",
			dataGranularity, applicationNS, applicationName, appJobStr)
		err = queueSender.SendJsonString(modelQueueName, jobJSONStr, appJobStr, granularity)
		if err == nil {
			sender.modelMapper.AddModelInfo(pdUnit, dataGranularity, applicationInfo)
		} else {
			scope.Errorf(
				"[APPLICATION][%s][%s/%s] Send model job payload failed. %s",
				dataGranularity, applicationNS, applicationName, err.Error())
		}
	}
}

func (sender *applicationModelJobSender) genApplicationInfo(applicationNS,
	applicationName string, modelMetrics ...datahub_common.MetricType) *modelInfo {
	applicationInfo := new(modelInfo)
	applicationInfo.NamespacedName = &namespacedName{
		Namespace: applicationNS,
		Name:      applicationName,
	}
	applicationInfo.ModelMetrics = modelMetrics
	applicationInfo.SetTimeStamp(time.Now().Unix())
	return applicationInfo
}

func (sender *applicationModelJobSender) getLastMIdPrediction(datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	application *datahub_resources.Application, granularity int64) ([]*datahub_predictions.MetricData, error) {
	dataGranularity := queue.GetGranularityStr(granularity)
	applicationNS := application.GetObjectMeta().GetNamespace()
	applicationName := application.GetObjectMeta().GetName()
	applicationPredictRes, err := datahubServiceClnt.ListApplicationPredictions(context.Background(),
		&datahub_predictions.ListApplicationPredictionsRequest{
			ObjectMeta: []*datahub_resources.ObjectMeta{
				&datahub_resources.ObjectMeta{
					Namespace: applicationNS,
					Name:      applicationName,
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
	if len(applicationPredictRes.GetApplicationPredictions()) > 0 {
		lastApplicationPrediction := applicationPredictRes.GetApplicationPredictions()[0]
		lctrlPDRData := lastApplicationPrediction.GetPredictedRawData()
		if lctrlPDRData == nil {
			lctrlPDRData = lastApplicationPrediction.GetPredictedLowerboundData()
		}
		if lctrlPDRData == nil {
			lctrlPDRData = lastApplicationPrediction.GetPredictedUpperboundData()
		}
		for _, pdRD := range lctrlPDRData {
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
		return nil, fmt.Errorf("[APPLICATION][%s][%s/%s] Query last prediction id failed",
			dataGranularity, applicationNS, applicationName)
	}

	applicationPredictRes, err = datahubServiceClnt.ListApplicationPredictions(context.Background(),
		&datahub_predictions.ListApplicationPredictionsRequest{
			ObjectMeta: []*datahub_resources.ObjectMeta{
				&datahub_resources.ObjectMeta{
					Namespace: applicationNS,
					Name:      applicationName,
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
	if len(applicationPredictRes.GetApplicationPredictions()) > 0 {
		metricData := []*datahub_predictions.MetricData{}
		for _, appPrediction := range applicationPredictRes.GetApplicationPredictions() {
			for _, pdRD := range appPrediction.GetPredictedRawData() {
				for _, pdD := range pdRD.GetData() {
					modelID := pdD.GetModelId()
					if modelID != "" {
						mIDAppPrediction, err := sender.getPredictionByMId(datahubServiceClnt, application, granularity, modelID)
						if err != nil {
							scope.Errorf("[APPLICATION][%s][%s/%s] Query prediction with model Id %s failed. %s",
								dataGranularity, applicationNS, applicationName, modelID, err.Error())
						}
						for _, mIDAppPD := range mIDAppPrediction {
							metricData = append(metricData, mIDAppPD.GetPredictedRawData()...)
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

func (sender *applicationModelJobSender) getPredictionByMId(datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	application *datahub_resources.Application, granularity int64, modelID string) ([]*datahub_predictions.ApplicationPrediction, error) {
	appPredictRes, err := datahubServiceClnt.ListApplicationPredictions(context.Background(),
		&datahub_predictions.ListApplicationPredictionsRequest{
			Granularity: granularity,
			ObjectMeta: []*datahub_resources.ObjectMeta{
				&datahub_resources.ObjectMeta{
					Name:      application.GetObjectMeta().GetName(),
					Namespace: application.GetObjectMeta().GetNamespace(),
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
	return appPredictRes.GetApplicationPredictions(), err
}

func (sender *applicationModelJobSender) getQueryMetricStartTime(
	descApplicationPredictions []*datahub_predictions.ApplicationPrediction) int64 {
	if len(descApplicationPredictions) > 0 {
		pdMDs := descApplicationPredictions[len(descApplicationPredictions)-1].GetPredictedRawData()
		for _, pdMD := range pdMDs {
			mD := pdMD.GetData()
			if len(mD) > 0 {
				return mD[len(mD)-1].GetTime().GetSeconds()
			}
		}
	}
	return 0
}

func (sender *applicationModelJobSender) sendJobByMetrics(application *datahub_resources.Application, queueSender queue.QueueSender,
	pdUnit string, granularity int64, predictionStep int64, datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	lastPredictionMetrics []*datahub_predictions.MetricData) {
	dataGranularity := queue.GetGranularityStr(granularity)
	queryCondition := &datahub_common.QueryCondition{
		Order: datahub_common.QueryCondition_DESC,
		TimeRange: &datahub_common.TimeRange{
			Step: &duration.Duration{
				Seconds: granularity,
			},
		},
	}
	applicationNS := application.GetObjectMeta().GetNamespace()
	applicationName := application.GetObjectMeta().GetName()
	nowSeconds := time.Now().Unix()

	if len(lastPredictionMetrics) == 0 {
		applicationInfo := sender.genApplicationInfo(applicationNS, applicationName,
			datahub_common.MetricType_MEMORY_USAGE_BYTES,
			datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
		)
		sender.sendJob(application, queueSender, pdUnit, granularity, applicationInfo)
		scope.Infof("[APPLICATION][%s][%s/%s] No prediction metric found, send model jobs",
			dataGranularity, applicationNS, applicationName)
		return
	}

	applicationInfo := sender.genApplicationInfo(applicationNS, applicationName)
	for _, lastPredictionMetric := range lastPredictionMetrics {
		if len(lastPredictionMetric.GetData()) == 0 {
			applicationInfo := sender.genApplicationInfo(applicationNS, applicationName,
				datahub_common.MetricType_MEMORY_USAGE_BYTES,
				datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE)
			sender.sendJob(application, queueSender, pdUnit, granularity, applicationInfo)
			scope.Infof("[APPLICATION][%s][%s/%s] No prediction metric %s found, send model jobs",
				dataGranularity, applicationNS, applicationName, lastPredictionMetric.GetMetricType().String())
			return
		} else {
			lastPrediction := lastPredictionMetric.GetData()[0]
			lastPredictionTime := lastPredictionMetric.GetData()[0].GetTime().GetSeconds()

			if lastPrediction != nil && lastPredictionTime <= nowSeconds {
				applicationInfo := sender.genApplicationInfo(applicationNS, applicationName,
					datahub_common.MetricType_MEMORY_USAGE_BYTES,
					datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE)
				scope.Infof("[APPLICATION][%s][%s/%s] Send model job due to no predict found or predict is out of date",
					dataGranularity, applicationNS, applicationName)
				sender.sendJob(application, queueSender, pdUnit, granularity, applicationInfo)
				return
			}
			applicationPredictRes, err := datahubServiceClnt.ListApplicationPredictions(context.Background(),
				&datahub_predictions.ListApplicationPredictionsRequest{
					ObjectMeta: []*datahub_resources.ObjectMeta{
						&datahub_resources.ObjectMeta{
							Namespace: applicationNS,
							Name:      applicationName,
						},
					},
					Granularity:    granularity,
					ModelId:        lastPrediction.GetModelId(),
					QueryCondition: queryCondition,
				})
			if err != nil {
				scope.Errorf("[APPLICATION][%s][%s/%s] Get prediction for sending model job failed: %s",
					dataGranularity, applicationNS, applicationName, err.Error())
				continue
			}
			applicationPredictions := applicationPredictRes.GetApplicationPredictions()
			queryStartTime := time.Now().Unix() - predictionStep*granularity
			firstPDTime := sender.getQueryMetricStartTime(applicationPredictions)
			if firstPDTime > 0 {
				queryStartTime = firstPDTime
			}
			applicationMetricsRes, err := datahubServiceClnt.ListApplicationMetrics(context.Background(),
				&datahub_metrics.ListApplicationMetricsRequest{
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
							Namespace: applicationNS,
							Name:      applicationName,
						},
					},
				})

			if err != nil {
				scope.Errorf("[APPLICATION][%s][%s/%s] List metric for sending model job failed: %s",
					dataGranularity, applicationNS, applicationName, err.Error())
				continue
			}
			applicationMetrics := applicationMetricsRes.GetApplicationMetrics()
			for _, applicationPrediction := range applicationPredictions {
				predictRawData := applicationPrediction.GetPredictedRawData()
				for _, predictRawDatum := range predictRawData {
					for _, applicationMetric := range applicationMetrics {
						metricData := applicationMetric.GetMetricData()
						for _, metricDatum := range metricData {
							mData := metricDatum.GetData()
							pData := []*datahub_predictions.Sample{}
							if metricDatum.GetMetricType() == predictRawDatum.GetMetricType() {
								pData = append(pData, predictRawDatum.GetData()...)
								metricsNeedToModel, drift := DriftEvaluation(UnitTypeApplication, predictRawDatum.GetMetricType(), granularity, mData, pData, map[string]string{
									"applicationNS":     applicationNS,
									"applicationName":   applicationName,
									"targetDisplayName": fmt.Sprintf("[APPLICATION][%s][%s/%s]", dataGranularity, applicationNS, applicationName),
								}, sender.metricExporter)
								if drift {
									scope.Infof("[APPLICATION][%s][%s/%s] Export drift counter",
										dataGranularity, applicationNS, applicationName)
									sender.metricExporter.AddApplicationMetricDrift(applicationNS, applicationName,
										queue.GetGranularityStr(granularity), time.Now().Unix(), 1.0)
								}
								applicationInfo.ModelMetrics = append(applicationInfo.ModelMetrics, metricsNeedToModel...)
							}
						}
					}
				}
			}
		}
	}
	isModeling := sender.modelMapper.IsModeling(pdUnit, dataGranularity, applicationInfo)
	if !isModeling || (isModeling && sender.modelMapper.IsModelTimeout(
		pdUnit, dataGranularity, applicationInfo)) {
		sender.sendJob(application, queueSender, pdUnit, granularity, applicationInfo)
		return
	}
}
