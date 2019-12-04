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

	datahubServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(sender.datahubGrpcCn)
	for _, application := range applications {
		applicationNS := application.GetObjectMeta().GetNamespace()
		applicationName := application.GetObjectMeta().GetName()

		lastPredictionMetrics, err := sender.getLastPrediction(datahubServiceClnt, application, granularity)
		if err != nil {
			scope.Infof("Get application %s/%s last prediction failed: %s",
				applicationNS, applicationName, err.Error())
			continue
		}
		if lastPredictionMetrics == nil && err == nil {
			scope.Infof("No prediction found of application %s/%s",
				applicationNS, applicationName)
		}
		sender.sendJobByMetrics(application, queueSender, pdUnit, granularity, predictionStep,
			datahubServiceClnt, lastPredictionMetrics)
	}
}

func (sender *applicationModelJobSender) sendJob(application *datahub_resources.Application, queueSender queue.QueueSender, pdUnit string,
	granularity int64, applicationInfo *modelInfo) {
	marshaler := jsonpb.Marshaler{}
	dataGranularity := queue.GetGranularityStr(granularity)
	applicationNS := application.GetObjectMeta().GetNamespace()
	applicationName := application.GetObjectMeta().GetName()
	applicationStr, err := marshaler.MarshalToString(application)
	if err != nil {
		scope.Errorf("Encode pb message failed for %s/%s with granularity seconds %v. %s",
			applicationNS, applicationName, granularity, err.Error())
		return
	}
	if len(applicationInfo.ModelMetrics) > 0 && applicationStr != "" {
		jb := queue.NewJobBuilder(pdUnit, granularity, applicationStr)
		jobJSONStr, err := jb.GetJobJSONString()
		if err != nil {
			scope.Errorf(
				"Prepare model job payload failed for %s/%s with granularity seconds %v. %s",
				applicationNS, applicationName, granularity, err.Error())
			return
		}

		appJobStr := fmt.Sprintf("%s/%s/%v", applicationNS, applicationName, granularity)
		scope.Infof("Try to send application model job: %s", appJobStr)
		err = queueSender.SendJsonString(modelQueueName, jobJSONStr, appJobStr)
		if err == nil {
			sender.modelMapper.AddModelInfo(pdUnit, dataGranularity, applicationInfo)
		} else {
			scope.Errorf(
				"Send model job payload failed for %s/%s with granularity seconds %v. %s",
				applicationNS, applicationName, granularity, err.Error())
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

func (sender *applicationModelJobSender) getLastPrediction(datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	application *datahub_resources.Application, granularity int64) ([]*datahub_predictions.MetricData, error) {
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

	if len(applicationPredictRes.GetApplicationPredictions()) > 0 {
		lastApplicationPrediction := applicationPredictRes.GetApplicationPredictions()[0]
		if lastApplicationPrediction.GetPredictedRawData() != nil {
			return lastApplicationPrediction.GetPredictedRawData(), nil
		} else if lastApplicationPrediction.GetPredictedLowerboundData() != nil {
			return lastApplicationPrediction.GetPredictedLowerboundData(), nil
		} else if lastApplicationPrediction.GetPredictedUpperboundData() != nil {
			return lastApplicationPrediction.GetPredictedUpperboundData(), nil
		}
	}
	return nil, nil
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
	applicationName := application.GetObjectMeta().GetNamespace()
	nowSeconds := time.Now().Unix()

	if len(lastPredictionMetrics) == 0 {
		applicationInfo := sender.genApplicationInfo(applicationNS, applicationName,
			datahub_common.MetricType_MEMORY_USAGE_BYTES,
			datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
		)
		sender.sendJob(application, queueSender, pdUnit, granularity, applicationInfo)
		scope.Infof("No prediction metric found of application %s/%s, send model jobs with granularity %v",
			applicationNS, applicationName, granularity)
		return
	}
	for _, lastPredictionMetric := range lastPredictionMetrics {
		if len(lastPredictionMetric.GetData()) == 0 {
			applicationInfo := sender.genApplicationInfo(applicationNS, applicationName,
				datahub_common.MetricType_MEMORY_USAGE_BYTES,
				datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE)
			sender.sendJob(application, queueSender, pdUnit, granularity, applicationInfo)
			scope.Infof("No prediction metric %s found of application %s/%s, send model jobs with granularity %v",
				lastPredictionMetric.GetMetricType().String(), applicationNS, applicationName, granularity)
			return
		} else {
			lastPrediction := lastPredictionMetric.GetData()[0]
			lastPredictionTime := lastPredictionMetric.GetData()[0].GetTime().GetSeconds()
			if lastPrediction != nil && lastPredictionTime <= nowSeconds {
				scope.Infof("application prediction %s/%s is out of date due to last predict time is %v (current: %v)",
					applicationNS, applicationName, lastPredictionTime, nowSeconds)
			}

			if lastPrediction != nil && lastPredictionTime <= nowSeconds {
				applicationInfo := sender.genApplicationInfo(applicationNS, applicationName,
					datahub_common.MetricType_MEMORY_USAGE_BYTES,
					datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE)
				scope.Infof("send application %s/%s model job due to no predict found or predict is out of date, send model jobs with granularity %v",
					applicationNS, applicationName, granularity)
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
				scope.Errorf("Get application %s/%s Prediction with granularity %v for sending model job failed: %s",
					applicationNS, applicationName, granularity, err.Error())
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
				scope.Errorf("List application %s/%s metric with granularity %v for sending model job failed: %s",
					applicationNS, applicationName, granularity, err.Error())
				continue
			}
			applicationMetrics := applicationMetricsRes.GetApplicationMetrics()

			for _, applicationMetric := range applicationMetrics {
				metricData := applicationMetric.GetMetricData()
				for _, metricDatum := range metricData {
					mData := metricDatum.GetData()
					pData := []*datahub_predictions.Sample{}
					applicationInfo := sender.genApplicationInfo(applicationNS, applicationName)
					for _, applicationPrediction := range applicationPredictions {
						predictRawData := applicationPrediction.GetPredictedRawData()
						for _, predictRawDatum := range predictRawData {
							if metricDatum.GetMetricType() == predictRawDatum.GetMetricType() {
								pData = append(pData, predictRawDatum.GetData()...)
							}
						}
					}
					metricsNeedToModel, drift := DriftEvaluation(UnitTypeApplication, metricDatum.GetMetricType(), granularity, mData, pData, map[string]string{
						"applicationNS":     applicationNS,
						"applicationName":   applicationName,
						"targetDisplayName": fmt.Sprintf("application %s/%s", applicationNS, applicationName),
					}, sender.metricExporter)
					if drift {
						scope.Infof("export application %s/%s drift counter with granularity %s",
							applicationNS, applicationName, dataGranularity)
						sender.metricExporter.AddApplicationMetricDrift(applicationNS, applicationName,
							queue.GetGranularityStr(granularity), 1.0)
					}
					applicationInfo.ModelMetrics = append(applicationInfo.ModelMetrics, metricsNeedToModel...)
					isModeling := sender.modelMapper.IsModeling(pdUnit, dataGranularity, applicationInfo)
					if !isModeling || (isModeling && sender.modelMapper.IsModelTimeout(
						pdUnit, dataGranularity, applicationInfo)) {
						sender.sendJob(application, queueSender, pdUnit, granularity, applicationInfo)
						return
					}
				}
			}
		}
	}
}
