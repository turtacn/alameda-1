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
		sender.sendApplicationModelJobs(application,
			queueSender, pdUnit, granularity, predictionStep, &wg)
	}
}

func (sender *applicationModelJobSender) sendApplicationModelJobs(application *datahub_resources.Application,
	queueSender queue.QueueSender, pdUnit string, granularity int64, predictionStep int64, wg *sync.WaitGroup) {
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
	sender.sendJobByMetrics(application, queueSender, pdUnit, granularity, predictionStep,
		datahubServiceClnt, lastPredictionMetrics)
}

func (sender *applicationModelJobSender) sendJob(application *datahub_resources.Application,
	queueSender queue.QueueSender, pdUnit string, granularity int64,
	metricType datahub_common.MetricType) {

	clusterID := application.GetObjectMeta().GetClusterName()
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

	jb := queue.NewJobBuilder(clusterID, pdUnit, granularity, metricType, applicationStr, nil)
	jobJSONStr, err := jb.GetJobJSONString()
	if err != nil {
		scope.Errorf(
			"[APPLICATION][%s][%s/%s] Prepare model job payload failed. %s",
			dataGranularity, applicationNS, applicationName, err.Error())
		return
	}

	appJobStr := fmt.Sprintf("%s/%s/%s/%s/%v/%s", consts.UnitTypeApplication,
		clusterID, applicationNS, applicationName, granularity, metricType)
	scope.Infof("[APPLICATION][%s][%s/%s] Try to send application model job: %s",
		dataGranularity, applicationNS, applicationName, appJobStr)
	err = queueSender.SendJsonString(modelQueueName, jobJSONStr, appJobStr, granularity)
	if err == nil {
		sender.modelMapper.AddModelInfo(clusterID, pdUnit, dataGranularity, metricType.String(), map[string]string{
			"namespace": applicationNS,
			"name":      applicationName,
		})
	} else {
		scope.Errorf(
			"[APPLICATION][%s][%s/%s] Send model job payload failed. %s",
			dataGranularity, applicationNS, applicationName, err.Error())
	}
}

func (sender *applicationModelJobSender) getLastMIdPrediction(datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	application *datahub_resources.Application, granularity int64) ([]*datahub_predictions.MetricData, error) {

	metricData := []*datahub_predictions.MetricData{}
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
		return metricData, err
	}

	lastMid := ""
	if len(applicationPredictRes.GetApplicationPredictions()) == 0 {
		return []*datahub_predictions.MetricData{}, nil
	}

	lastApplicationPrediction := applicationPredictRes.GetApplicationPredictions()[0]
	lctrlPDRData := lastApplicationPrediction.GetPredictedRawData()
	if lctrlPDRData == nil {
		return metricData, nil
	}

	for _, pdRD := range lctrlPDRData {
		for _, theData := range pdRD.GetData() {
			lastMid = theData.GetModelId()
			break
		}
		if lastMid == "" {
			scope.Warnf("[APPLICATION][%s][%s/%s] Query last model id for metric %s is empty",
				dataGranularity, applicationNS, applicationName, pdRD.GetMetricType())
			continue
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
				ModelId: lastMid,
			})
		if err != nil {
			scope.Errorf("[APPLICATION][%s][%s/%s] Query last model id %v for metric %s failed",
				dataGranularity, applicationNS, applicationName, lastMid, pdRD.GetMetricType())
			continue
		}

		for _, appPrediction := range applicationPredictRes.GetApplicationPredictions() {
			for _, lMIDPdRD := range appPrediction.GetPredictedRawData() {
				if lMIDPdRD.GetMetricType() == pdRD.GetMetricType() {
					metricData = append(metricData, lMIDPdRD)
				}
			}
		}
	}
	return metricData, nil
}

func (sender *applicationModelJobSender) getQueryMetricStartTime(
	metricData *datahub_predictions.MetricData) int64 {
	mD := metricData.GetData()
	if len(mD) > 0 {
		return mD[len(mD)-1].GetTime().GetSeconds()
	}
	return 0
}

func (sender *applicationModelJobSender) sendJobByMetrics(application *datahub_resources.Application, queueSender queue.QueueSender,
	pdUnit string, granularity int64, predictionStep int64, datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	lastPredictionMetrics []*datahub_predictions.MetricData) {

	dataGranularity := queue.GetGranularityStr(granularity)
	clusterID := application.GetObjectMeta().GetClusterName()
	applicationNS := application.GetObjectMeta().GetNamespace()
	applicationName := application.GetObjectMeta().GetName()
	nowSeconds := time.Now().Unix()

	if len(lastPredictionMetrics) == 0 {
		scope.Infof("[APPLICATION][%s][%s/%s] No prediction metric found, send model jobs",
			dataGranularity, applicationNS, applicationName)
		for _, metricType := range []datahub_common.MetricType{
			datahub_common.MetricType_MEMORY_USAGE_BYTES,
			datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
		} {
			sender.sendJob(application, queueSender, pdUnit, granularity, metricType)
		}
		return
	}

	for _, lastPredictionMetric := range lastPredictionMetrics {
		if len(lastPredictionMetric.GetData()) == 0 {
			scope.Infof("[APPLICATION][%s][%s/%s] No prediction metric %s found, send model jobs",
				dataGranularity, applicationNS, applicationName, lastPredictionMetric.GetMetricType().String())
			sender.sendJob(application, queueSender, pdUnit, granularity, lastPredictionMetric.GetMetricType())
			continue
		} else {
			lastPrediction := lastPredictionMetric.GetData()[0]
			lastPredictionTime := lastPredictionMetric.GetData()[0].GetTime().GetSeconds()

			if lastPrediction != nil && lastPredictionTime <= nowSeconds {
				scope.Infof("[APPLICATION][%s][%s/%s] Send model job due to no predict metric %s found or is out of date",
					dataGranularity, applicationNS, applicationName, lastPredictionMetric.GetMetricType().String())
				sender.sendJob(application, queueSender, pdUnit, granularity, lastPredictionMetric.GetMetricType())
				continue
			}

			queryStartTime := time.Now().Unix() - predictionStep*granularity
			firstPDTime := sender.getQueryMetricStartTime(lastPredictionMetric)
			if firstPDTime > 0 && firstPDTime <= time.Now().Unix() {
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
					MetricTypes: []datahub_common.MetricType{
						lastPredictionMetric.GetMetricType(),
					},
				})

			if err != nil {
				scope.Errorf("[APPLICATION][%s][%s/%s] List metric for sending model job failed: %s",
					dataGranularity, applicationNS, applicationName, err.Error())
				continue
			}
			applicationMetrics := applicationMetricsRes.GetApplicationMetrics()
			predictRawData := lastPredictionMetrics
			for _, predictRawDatum := range predictRawData {
				for _, applicationMetric := range applicationMetrics {
					metricData := applicationMetric.GetMetricData()
					for _, metricDatum := range metricData {
						mData := metricDatum.GetData()
						pData := []*datahub_predictions.Sample{}
						if metricDatum.GetMetricType() == predictRawDatum.GetMetricType() {
							pData = append(pData, predictRawDatum.GetData()...)
							metricsNeedToModel, drift := DriftEvaluation(consts.UnitTypeApplication, predictRawDatum.GetMetricType(), granularity, mData, pData, map[string]string{
								"clusterID":         clusterID,
								"applicationNS":     applicationNS,
								"applicationName":   applicationName,
								"targetDisplayName": fmt.Sprintf("[APPLICATION][%s][%s/%s]", dataGranularity, applicationNS, applicationName),
							}, sender.metricExporter)

							for _, mntm := range metricsNeedToModel {
								if drift {
									scope.Infof("[APPLICATION][%s][%s/%s] Export metric %s drift counter",
										dataGranularity, applicationNS, applicationName, mntm)
									sender.metricExporter.AddApplicationMetricDrift(clusterID, applicationNS, applicationName,
										queue.GetGranularityStr(granularity), mntm.String(), time.Now().Unix(), 1.0)
								}
								isModeling := sender.modelMapper.IsModeling(clusterID, pdUnit, dataGranularity, mntm.String(), map[string]string{
									"namespace": applicationNS,
									"name":      applicationName,
								})
								if !isModeling || (isModeling && sender.modelMapper.IsModelTimeout(
									clusterID, pdUnit, dataGranularity, mntm.String(), map[string]string{
										"namespace": applicationNS,
										"name":      applicationName,
									})) {
									sender.sendJob(application, queueSender, pdUnit, granularity, mntm)
								}
							}
						}
					}
				}
			}
		}
	}
}
