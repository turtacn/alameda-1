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

type controllerModelJobSender struct {
	datahubGrpcCn  *grpc.ClientConn
	modelMapper    *ModelMapper
	metricExporter *metrics.Exporter
}

func NewControllerModelJobSender(datahubGrpcCn *grpc.ClientConn, modelMapper *ModelMapper,
	metricExporter *metrics.Exporter) *controllerModelJobSender {
	return &controllerModelJobSender{
		datahubGrpcCn:  datahubGrpcCn,
		modelMapper:    modelMapper,
		metricExporter: metricExporter,
	}
}

func (sender *controllerModelJobSender) sendModelJobs(controllers []*datahub_resources.Controller,
	queueSender queue.QueueSender, pdUnit string, granularity int64, predictionStep int64) {
	for _, controller := range controllers {
		go sender.sendControllerModelJobs(controller, queueSender, pdUnit, granularity, predictionStep)
	}
}

func (sender *controllerModelJobSender) sendControllerModelJobs(controller *datahub_resources.Controller,
	queueSender queue.QueueSender, pdUnit string, granularity int64, predictionStep int64) {
	dataGranularity := queue.GetGranularityStr(granularity)
	datahubServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(sender.datahubGrpcCn)

	controllerNS := controller.GetObjectMeta().GetNamespace()
	controllerName := controller.GetObjectMeta().GetName()

	lastPredictionMetrics, err := sender.getLastMIdPrediction(datahubServiceClnt, controller, granularity)
	if err != nil {
		scope.Infof("[CONTROLLER][%s][%s][%s/%s] Get last prediction failed: %s",
			controller.GetKind().String(), dataGranularity, controllerNS, controllerName, err.Error())
		return
	}
	if lastPredictionMetrics == nil && err == nil {
		scope.Infof("[CONTROLLER][%s][%s][%s/%s] No prediction found",
			controller.GetKind().String(), dataGranularity, controllerNS, controllerName)
	}
	sender.sendJobByMetrics(controller, queueSender, pdUnit, granularity, predictionStep,
		datahubServiceClnt, lastPredictionMetrics)
}

func (sender *controllerModelJobSender) sendJob(controller *datahub_resources.Controller, queueSender queue.QueueSender, pdUnit string,
	granularity int64, controllerInfo *modelInfo) {
	marshaler := jsonpb.Marshaler{}
	dataGranularity := queue.GetGranularityStr(granularity)
	controllerNS := controller.GetObjectMeta().GetNamespace()
	controllerName := controller.GetObjectMeta().GetName()
	controllerStr, err := marshaler.MarshalToString(controller)
	if err != nil {
		scope.Errorf("[CONTROLLER][%s][%s][%s/%s] Encode pb message failed. %s",
			controller.GetKind().String(), dataGranularity, controllerNS, controllerName, err.Error())
		return
	}
	if len(controllerInfo.ModelMetrics) > 0 && controllerStr != "" {
		jb := queue.NewJobBuilder(pdUnit, granularity, controllerStr)
		jobJSONStr, err := jb.GetJobJSONString()
		if err != nil {
			scope.Errorf(
				"[CONTROLLER][%s][%s][%s/%s] Prepare model job payload failed. %s",
				controller.GetKind().String(), dataGranularity, controllerNS, controllerName, err.Error())
			return
		}

		controllerJobStr := fmt.Sprintf("%s/%s/%v", controllerNS, controllerName, granularity)
		scope.Infof("[CONTROLLER][%s][%s][%s/%s] Try to send controller model job: %s",
			controller.GetKind().String(), dataGranularity, controllerNS, controllerName, controllerJobStr)
		err = queueSender.SendJsonString(modelQueueName, jobJSONStr, controllerJobStr, granularity)
		if err == nil {
			sender.modelMapper.AddModelInfo(pdUnit, dataGranularity, controllerInfo)
		} else {
			scope.Errorf(
				"[CONTROLLER][%s][%s][%s/%s] Send model job payload failed. %s",
				controller.GetKind().String(), dataGranularity, controllerNS, controllerName, err.Error())
		}
	}
}

func (sender *controllerModelJobSender) genControllerInfo(controllerNS,
	controllerName string, modelMetrics ...datahub_common.MetricType) *modelInfo {
	controllerInfo := new(modelInfo)
	controllerInfo.NamespacedName = &namespacedName{
		Namespace: controllerNS,
		Name:      controllerName,
	}
	controllerInfo.ModelMetrics = modelMetrics
	controllerInfo.SetTimeStamp(time.Now().Unix())
	return controllerInfo
}

func (sender *controllerModelJobSender) getLastMIdPrediction(datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	controller *datahub_resources.Controller, granularity int64) ([]*datahub_predictions.MetricData, error) {
	dataGranularity := queue.GetGranularityStr(granularity)
	controllerNS := controller.GetObjectMeta().GetNamespace()
	controllerName := controller.GetObjectMeta().GetName()
	controllerPredictRes, err := datahubServiceClnt.ListControllerPredictions(context.Background(),
		&datahub_predictions.ListControllerPredictionsRequest{
			ObjectMeta: []*datahub_resources.ObjectMeta{
				&datahub_resources.ObjectMeta{
					Namespace: controllerNS,
					Name:      controllerName,
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
	if len(controllerPredictRes.GetControllerPredictions()) > 0 {
		lastControllerPrediction := controllerPredictRes.GetControllerPredictions()[0]
		lctrlPDRData := lastControllerPrediction.GetPredictedRawData()
		if lctrlPDRData == nil {
			lctrlPDRData = lastControllerPrediction.GetPredictedLowerboundData()
		}
		if lctrlPDRData == nil {
			lctrlPDRData = lastControllerPrediction.GetPredictedUpperboundData()
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
		return nil, fmt.Errorf("[CONTROLLER][%s][%s][%s/%s] Query last prediction id failed",
			controller.GetKind().String(), dataGranularity, controllerNS, controllerName)
	}

	controllerPredictRes, err = datahubServiceClnt.ListControllerPredictions(context.Background(),
		&datahub_predictions.ListControllerPredictionsRequest{
			ObjectMeta: []*datahub_resources.ObjectMeta{
				&datahub_resources.ObjectMeta{
					Namespace: controllerNS,
					Name:      controllerName,
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
	if len(controllerPredictRes.GetControllerPredictions()) > 0 {
		metricData := []*datahub_predictions.MetricData{}
		for _, ctrlPrediction := range controllerPredictRes.GetControllerPredictions() {
			for _, pdRD := range ctrlPrediction.GetPredictedRawData() {
				for _, pdD := range pdRD.GetData() {
					modelID := pdD.GetModelId()
					if modelID != "" {
						mIDCtrlPrediction, err := sender.getPredictionByMId(datahubServiceClnt, controller, granularity, modelID)
						if err != nil {
							scope.Errorf("[CONTROLLER][%s][%s][%s/%s] Query prediction with model Id %s failed. %s",
								controller.GetKind().String(), dataGranularity, controllerNS, controllerName, modelID, err.Error())
						}
						for _, mIDCtrlPD := range mIDCtrlPrediction {
							metricData = append(metricData, mIDCtrlPD.GetPredictedRawData()...)
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

func (sender *controllerModelJobSender) getPredictionByMId(datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	controller *datahub_resources.Controller, granularity int64, modelID string) ([]*datahub_predictions.ControllerPrediction, error) {
	controllerPredictRes, err := datahubServiceClnt.ListControllerPredictions(context.Background(),
		&datahub_predictions.ListControllerPredictionsRequest{
			Granularity: granularity,
			ObjectMeta: []*datahub_resources.ObjectMeta{
				&datahub_resources.ObjectMeta{
					Name:      controller.GetObjectMeta().GetName(),
					Namespace: controller.GetObjectMeta().GetNamespace(),
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
	return controllerPredictRes.GetControllerPredictions(), err
}

func (sender *controllerModelJobSender) getQueryMetricStartTime(
	descControllerPredictions []*datahub_predictions.ControllerPrediction) int64 {
	if len(descControllerPredictions) > 0 {
		pdMDs := descControllerPredictions[len(descControllerPredictions)-1].GetPredictedRawData()
		for _, pdMD := range pdMDs {
			mD := pdMD.GetData()
			if len(mD) > 0 {
				return mD[len(mD)-1].GetTime().GetSeconds()
			}
		}
	}
	return 0
}

func (sender *controllerModelJobSender) sendJobByMetrics(controller *datahub_resources.Controller, queueSender queue.QueueSender,
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
	controllerNS := controller.GetObjectMeta().GetNamespace()
	controllerName := controller.GetObjectMeta().GetName()
	nowSeconds := time.Now().Unix()

	if len(lastPredictionMetrics) == 0 {
		controllerInfo := sender.genControllerInfo(controllerNS, controllerName,
			datahub_common.MetricType_MEMORY_USAGE_BYTES,
			datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
		)
		sender.sendJob(controller, queueSender, pdUnit, granularity, controllerInfo)
		scope.Infof("[CONTROLLER][%s][%s][%s/%s] No prediction metric found, send model jobs",
			controller.GetKind().String(), dataGranularity, controllerNS, controllerName)
		return
	}

	controllerInfo := sender.genControllerInfo(controllerNS, controllerName)
	for _, lastPredictionMetric := range lastPredictionMetrics {
		if len(lastPredictionMetric.GetData()) == 0 {
			controllerInfo := sender.genControllerInfo(controllerNS, controllerName,
				datahub_common.MetricType_MEMORY_USAGE_BYTES,
				datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE)
			sender.sendJob(controller, queueSender, pdUnit, granularity, controllerInfo)
			scope.Infof("[CONTROLLER][%s][%s][%s/%s] No prediction metric %s found, send model jobs",
				controller.GetKind().String(), dataGranularity, controllerNS, controllerName, lastPredictionMetric.GetMetricType().String())
			return
		} else {
			lastPrediction := lastPredictionMetric.GetData()[0]
			lastPredictionTime := lastPredictionMetric.GetData()[0].GetTime().GetSeconds()

			if lastPrediction != nil && lastPredictionTime <= nowSeconds {
				controllerInfo := sender.genControllerInfo(controllerNS, controllerName,
					datahub_common.MetricType_MEMORY_USAGE_BYTES,
					datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE)
				scope.Infof("[CONTROLLER][%s][%s][%s/%s] Send model job due to no predict found or predict is out of date",
					controller.GetKind().String(), dataGranularity, controllerNS, controllerName)
				sender.sendJob(controller, queueSender, pdUnit, granularity, controllerInfo)
				return
			}
			controllerPredictRes, err := datahubServiceClnt.ListControllerPredictions(context.Background(),
				&datahub_predictions.ListControllerPredictionsRequest{
					ObjectMeta: []*datahub_resources.ObjectMeta{
						&datahub_resources.ObjectMeta{
							Namespace: controllerNS,
							Name:      controllerName,
						},
					},
					Granularity:    granularity,
					ModelId:        lastPrediction.GetModelId(),
					QueryCondition: queryCondition,
				})
			if err != nil {
				scope.Errorf("[CONTROLLER][%s][%s][%s/%s] Get prediction for sending model job failed: %s",
					controller.GetKind().String(), dataGranularity, controllerNS, controllerName, err.Error())
				continue
			}
			controllerPredictions := controllerPredictRes.GetControllerPredictions()
			queryStartTime := time.Now().Unix() - predictionStep*granularity
			firstPDTime := sender.getQueryMetricStartTime(controllerPredictions)
			if firstPDTime > 0 {
				queryStartTime = firstPDTime
			}
			controllerMetricsRes, err := datahubServiceClnt.ListControllerMetrics(context.Background(),
				&datahub_metrics.ListControllerMetricsRequest{
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
							Namespace: controllerNS,
							Name:      controllerName,
						},
					},
				})

			if err != nil {
				scope.Errorf("[CONTROLLER][%s][%s][%s/%s] List metric for sending model job failed: %s",
					controller.GetKind().String(), dataGranularity, controllerNS, controllerName, err.Error())
				continue
			}
			controllerMetrics := controllerMetricsRes.GetControllerMetrics()
			for _, controllerPrediction := range controllerPredictions {
				predictRawData := controllerPrediction.GetPredictedRawData()
				for _, predictRawDatum := range predictRawData {
					for _, controllerMetric := range controllerMetrics {
						metricData := controllerMetric.GetMetricData()
						for _, metricDatum := range metricData {
							mData := metricDatum.GetData()
							pData := []*datahub_predictions.Sample{}
							if metricDatum.GetMetricType() == predictRawDatum.GetMetricType() {
								pData = append(pData, predictRawDatum.GetData()...)
								metricsNeedToModel, drift := DriftEvaluation(UnitTypeController, predictRawDatum.GetMetricType(), granularity, mData, pData, map[string]string{
									"controllerNS":   controllerNS,
									"controllerName": controllerName,
									"controllerKind": controller.GetKind().String(),
									"targetDisplayName": fmt.Sprintf("[CONTROLLER][%s][%s][%s/%s]",
										controller.GetKind().String(), dataGranularity, controllerNS, controllerName),
								}, sender.metricExporter)
								if drift {
									scope.Infof("[CONTROLLER][%s][%s][%s/%s] Export drift counter",
										controller.GetKind().String(), dataGranularity, controllerNS, controllerName)
									sender.metricExporter.AddControllerMetricDrift(controllerNS, controllerName,
										controller.GetKind().String(), queue.GetGranularityStr(granularity), time.Now().Unix(), 1.0)
								}
								controllerInfo.ModelMetrics = append(controllerInfo.ModelMetrics, metricsNeedToModel...)
							}
						}
					}
				}
			}
		}
	}
	isModeling := sender.modelMapper.IsModeling(pdUnit, dataGranularity, controllerInfo)
	if !isModeling || (isModeling && sender.modelMapper.IsModelTimeout(
		pdUnit, dataGranularity, controllerInfo)) {
		sender.sendJob(controller, queueSender, pdUnit, granularity, controllerInfo)
		return
	}
}
