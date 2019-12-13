package dispatcher

import (
	"context"
	"fmt"
	"time"

	"github.com/containers-ai/alameda/ai-dispatcher/pkg/metrics"
	"github.com/containers-ai/alameda/ai-dispatcher/pkg/queue"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	datahub_common "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	datahub_gpu "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/gpu"
	datahub_predictions "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/predictions"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/grpc"
)

type gpuModelJobSender struct {
	datahubGrpcCn  *grpc.ClientConn
	modelMapper    *ModelMapper
	metricExporter *metrics.Exporter
}

func NewGPUModelJobSender(datahubGrpcCn *grpc.ClientConn, modelMapper *ModelMapper,
	metricExporter *metrics.Exporter) *gpuModelJobSender {
	return &gpuModelJobSender{
		datahubGrpcCn:  datahubGrpcCn,
		modelMapper:    modelMapper,
		metricExporter: metricExporter,
	}
}

func (sender *gpuModelJobSender) sendModelJobs(gpus []*datahub_gpu.Gpu,
	queueSender queue.QueueSender, pdUnit string, granularity int64, predictionStep int64) {
	for _, gpu := range gpus {
		go sender.sendGpuModelJobs(gpu, queueSender, pdUnit, granularity, predictionStep)
	}
}

func (sender *gpuModelJobSender) sendGpuModelJobs(gpu *datahub_gpu.Gpu,
	queueSender queue.QueueSender, pdUnit string, granularity int64, predictionStep int64) {
	dataGranularity := queue.GetGranularityStr(granularity)
	datahubServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(sender.datahubGrpcCn)

	gpuHost := gpu.GetMetadata().GetHost()
	gpuMinorNumber := gpu.GetMetadata().GetMinorNumber()

	lastPredictionMetrics, err := sender.getLastMIdPrediction(datahubServiceClnt, gpu, granularity)
	if err != nil {
		scope.Infof("[GPU][%s][%s/%s] Get last prediction failed: %s",
			dataGranularity, gpuHost, gpuMinorNumber, err.Error())
		return
	}
	if lastPredictionMetrics == nil && err == nil {
		scope.Infof("[GPU][%s][%s/%s] No prediction found",
			dataGranularity, gpuHost, gpuMinorNumber)
	}
	sender.sendJobByMetrics(gpu, queueSender, pdUnit, granularity, predictionStep,
		datahubServiceClnt, lastPredictionMetrics)
}

func (sender *gpuModelJobSender) sendJob(gpu *datahub_gpu.Gpu, queueSender queue.QueueSender, pdUnit string,
	granularity int64, gpuInfo *modelInfo) {
	marshaler := jsonpb.Marshaler{}
	dataGranularity := queue.GetGranularityStr(granularity)
	gpuHost := gpu.GetMetadata().GetHost()
	gpuMinorNumber := gpu.GetMetadata().GetMinorNumber()
	gpuStr, err := marshaler.MarshalToString(gpu)
	if err != nil {
		scope.Errorf("[GPU][%s][%s/%s] Encode pb message failed. %s",
			dataGranularity, gpuHost, gpuMinorNumber, err.Error())
		return
	}
	if len(gpuInfo.ModelMetrics) > 0 && gpuStr != "" {
		jb := queue.NewJobBuilder(pdUnit, granularity, gpuStr)
		jobJSONStr, err := jb.GetJobJSONString()
		if err != nil {
			scope.Errorf(
				"[GPU][%s][%s/%s] Prepare model job payload failed. %s",
				dataGranularity, gpuHost, gpuMinorNumber, err.Error())
			return
		}

		gpuJobStr := fmt.Sprintf("%s/%s/%v", gpuHost, gpuMinorNumber, granularity)
		scope.Infof("[GPU][%s][%s/%s] Try to send gpu model job: %s", dataGranularity, gpuHost, gpuMinorNumber, gpuJobStr)
		err = queueSender.SendJsonString(modelQueueName, jobJSONStr, gpuJobStr, granularity)
		if err == nil {
			sender.modelMapper.AddModelInfo(pdUnit, dataGranularity, gpuInfo)
		} else {
			scope.Errorf(
				"[GPU][%s][%s/%s] Send model job payload failed. %s",
				dataGranularity, gpuHost, gpuMinorNumber, err.Error())
		}
	}
}

func (sender *gpuModelJobSender) genGPUInfo(gpuHost,
	gpuMinorNumber string, modelMetrics ...datahub_common.MetricType) *modelInfo {
	gpuInfo := new(modelInfo)
	gpuInfo.Host = gpuHost
	gpuInfo.MinorNumber = gpuMinorNumber
	gpuInfo.ModelMetrics = modelMetrics
	gpuInfo.SetTimeStamp(time.Now().Unix())
	return gpuInfo
}

func (sender *gpuModelJobSender) getLastMIdPrediction(datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	gpu *datahub_gpu.Gpu, granularity int64) ([]*datahub_predictions.MetricData, error) {
	dataGranularity := queue.GetGranularityStr(granularity)
	gpuHost := gpu.GetMetadata().GetHost()
	gpuMinorNumber := gpu.GetMetadata().GetMinorNumber()
	gpuPredictRes, err := datahubServiceClnt.ListGpuPredictions(context.Background(),
		&datahub_gpu.ListGpuPredictionsRequest{
			Host:        gpuHost,
			MinorNumber: gpuMinorNumber,
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
	if len(gpuPredictRes.GetGpuPredictions()) > 0 {
		lastGpuPrediction := gpuPredictRes.GetGpuPredictions()[0]
		lgpuPDRData := lastGpuPrediction.GetPredictedRawData()
		if lgpuPDRData == nil {
			lgpuPDRData = lastGpuPrediction.GetPredictedLowerboundData()
		}
		if lgpuPDRData == nil {
			lgpuPDRData = lastGpuPrediction.GetPredictedUpperboundData()
		}
		for _, pdRD := range lgpuPDRData {
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
		return nil, fmt.Errorf("[GPU][%s][%s/%s] Query last prediction id failed",
			dataGranularity, gpuHost, gpuMinorNumber)
	}

	gpuPredictRes, err = datahubServiceClnt.ListGpuPredictions(context.Background(),
		&datahub_gpu.ListGpuPredictionsRequest{
			Host:        gpuHost,
			MinorNumber: gpuMinorNumber,
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
	if len(gpuPredictRes.GetGpuPredictions()) > 0 {
		metricData := []*datahub_predictions.MetricData{}
		for _, gpuPrediction := range gpuPredictRes.GetGpuPredictions() {
			for _, pdRD := range gpuPrediction.GetPredictedRawData() {
				for _, pdD := range pdRD.GetData() {
					modelID := pdD.GetModelId()
					if modelID != "" {
						mIDCtrlPrediction, err := sender.getPredictionByMId(datahubServiceClnt, gpu, granularity, modelID)
						if err != nil {
							scope.Errorf("[GPU][%s][%s/%s] Query prediction with model Id %s failed. %s",
								dataGranularity, gpuHost, gpuMinorNumber, modelID, err.Error())
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

func (sender *gpuModelJobSender) getPredictionByMId(datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	gpu *datahub_gpu.Gpu, granularity int64, modelID string) ([]*datahub_gpu.GpuPrediction, error) {
	gpuHost := gpu.GetMetadata().GetHost()
	gpuMinorNumber := gpu.GetMetadata().GetMinorNumber()
	gpuPredictRes, err := datahubServiceClnt.ListGpuPredictions(context.Background(),
		&datahub_gpu.ListGpuPredictionsRequest{
			Granularity: granularity,
			Host:        gpuHost,
			MinorNumber: gpuMinorNumber,
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
	return gpuPredictRes.GetGpuPredictions(), err
}

func (sender *gpuModelJobSender) getQueryMetricStartTime(descGpuPredictions []*datahub_gpu.GpuPrediction) int64 {
	if len(descGpuPredictions) > 0 {
		pdMDs := descGpuPredictions[len(descGpuPredictions)-1].GetPredictedRawData()
		for _, pdMD := range pdMDs {
			mD := pdMD.GetData()
			if len(mD) > 0 {
				return mD[len(mD)-1].GetTime().GetSeconds()
			}
		}
	}
	return 0
}

func (sender *gpuModelJobSender) sendJobByMetrics(gpu *datahub_gpu.Gpu, queueSender queue.QueueSender,
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
	gpuHost := gpu.GetMetadata().GetHost()
	gpuMinorNumber := gpu.GetMetadata().GetMinorNumber()
	nowSeconds := time.Now().Unix()

	if len(lastPredictionMetrics) == 0 {
		gpuInfo := sender.genGPUInfo(gpuHost, gpuMinorNumber,
			datahub_common.MetricType_MEMORY_USAGE_BYTES,
			datahub_common.MetricType_DUTY_CYCLE,
		)
		sender.sendJob(gpu, queueSender, pdUnit, granularity, gpuInfo)
		scope.Infof("[GPU][%s][%s/%s] No prediction metric found, send model jobs",
			dataGranularity, gpuHost, gpuMinorNumber)
		return
	}

	gpuInfo := sender.genGPUInfo(gpuHost, gpuMinorNumber)
	for _, lastPredictionMetric := range lastPredictionMetrics {
		if len(lastPredictionMetric.GetData()) == 0 {
			gpuInfo := sender.genGPUInfo(gpuHost, gpuMinorNumber,
				datahub_common.MetricType_MEMORY_USAGE_BYTES,
				datahub_common.MetricType_DUTY_CYCLE)
			sender.sendJob(gpu, queueSender, pdUnit, granularity, gpuInfo)
			scope.Infof("[GPU][%s][%s/%s] No prediction metric %s found, send model jobs",
				dataGranularity, gpuHost, gpuMinorNumber, lastPredictionMetric.GetMetricType().String())
			return
		} else {
			lastPrediction := lastPredictionMetric.GetData()[0]
			lastPredictionTime := lastPredictionMetric.GetData()[0].GetTime().GetSeconds()
			if lastPrediction != nil && lastPredictionTime <= nowSeconds {
				gpuInfo := sender.genGPUInfo(gpuHost, gpuMinorNumber,
					datahub_common.MetricType_MEMORY_USAGE_BYTES,
					datahub_common.MetricType_DUTY_CYCLE)
				scope.Infof("[GPU][%s][%s/%s] Send model job due to no predict found or predict is out of date",
					dataGranularity, gpuHost, gpuMinorNumber)
				sender.sendJob(gpu, queueSender, pdUnit, granularity, gpuInfo)
				return
			}
			gpuPredictRes, err := datahubServiceClnt.ListGpuPredictions(context.Background(),
				&datahub_gpu.ListGpuPredictionsRequest{
					Host:           gpuHost,
					MinorNumber:    gpuMinorNumber,
					Granularity:    granularity,
					ModelId:        lastPrediction.GetModelId(),
					QueryCondition: queryCondition,
				})
			if err != nil {
				scope.Errorf("[GPU][%s][%s/%s] Get prediction for sending model job failed: %s",
					dataGranularity, gpuHost, gpuMinorNumber, err.Error())
				continue
			}
			gpuPredictions := gpuPredictRes.GetGpuPredictions()
			queryStartTime := time.Now().Unix() - predictionStep*granularity
			firstPDTime := sender.getQueryMetricStartTime(gpuPredictions)
			if firstPDTime > 0 {
				queryStartTime = firstPDTime
			}
			gpuMetricsRes, err := datahubServiceClnt.ListGpuMetrics(context.Background(),
				&datahub_gpu.ListGpuMetricsRequest{
					QueryCondition: &datahub_common.QueryCondition{
						Order: datahub_common.QueryCondition_DESC,
						TimeRange: &datahub_common.TimeRange{
							StartTime: &timestamp.Timestamp{
								Seconds: queryStartTime,
							},
							Step: &duration.Duration{
								Seconds: granularity,
							},
							AggregateFunction: datahub_common.TimeRange_MAX,
						},
					},
					Host:        gpuHost,
					MinorNumber: gpuMinorNumber,
				})

			if err != nil {
				scope.Errorf("[GPU][%s][%s/%s] List gpu metric for sending model job failed: %s",
					dataGranularity, gpuHost, gpuMinorNumber, err.Error())
				continue
			}
			gpuMetrics := gpuMetricsRes.GetGpuMetrics()
			// gpu tags are host, minor number, pid, mid
			for _, gpuPrediction := range gpuPredictions {
				predictRawData := gpuPrediction.GetPredictedRawData()
				for _, predictRawDatum := range predictRawData {
					for _, gpuMetric := range gpuMetrics {
						metricData := gpuMetric.GetMetricData()
						for _, metricDatum := range metricData {
							mData := metricDatum.GetData()
							pData := []*datahub_predictions.Sample{}
							if metricDatum.GetMetricType() == predictRawDatum.GetMetricType() {
								pData = append(pData, predictRawDatum.GetData()...)
								metricsNeedToModel, drift := DriftEvaluation(UnitTypeGPU, predictRawDatum.GetMetricType(), granularity, mData, pData, map[string]string{
									"gpuHost":           gpuHost,
									"gpuMinorNumber":    gpuMinorNumber,
									"targetDisplayName": fmt.Sprintf("[GPU][%s][%s/%s]", dataGranularity, gpuHost, gpuMinorNumber),
								}, sender.metricExporter)
								if drift {
									scope.Infof("[GPU][%s][%s/%s] Export drift counter",
										dataGranularity, gpuHost, gpuMinorNumber)
									sender.metricExporter.AddGPUMetricDrift(gpuHost, gpuMinorNumber,
										queue.GetGranularityStr(granularity), time.Now().Unix(), 1.0)
								}
								gpuInfo.ModelMetrics = append(gpuInfo.ModelMetrics, metricsNeedToModel...)
							}
						}
					}
				}
			}
		}
	}
	isModeling := sender.modelMapper.IsModeling(pdUnit, dataGranularity, gpuInfo)
	if !isModeling || (isModeling && sender.modelMapper.IsModelTimeout(
		pdUnit, dataGranularity, gpuInfo)) {
		sender.sendJob(gpu, queueSender, pdUnit, granularity, gpuInfo)
		return
	}
}
