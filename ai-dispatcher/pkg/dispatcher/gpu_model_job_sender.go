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

	datahubServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(sender.datahubGrpcCn)
	for _, gpu := range gpus {
		gpuHost := gpu.GetMetadata().GetHost()
		gpuMinorNumber := gpu.GetMetadata().GetMinorNumber()

		lastPredictionMetrics, err := sender.getLastPrediction(datahubServiceClnt, gpu, granularity)
		if err != nil {
			scope.Infof("Get gpu (host: %s, minor number: %s) last prediction failed: %s",
				gpuHost, gpuMinorNumber, err.Error())
			continue
		}
		if lastPredictionMetrics == nil && err == nil {
			scope.Infof("No prediction found of gpu (host: %s, minor number: %s)",
				gpuHost, gpuMinorNumber)
		}
		sender.sendJobByMetrics(gpu, queueSender, pdUnit, granularity, predictionStep,
			datahubServiceClnt, lastPredictionMetrics)
	}
}

func (sender *gpuModelJobSender) sendJob(gpu *datahub_gpu.Gpu, queueSender queue.QueueSender, pdUnit string,
	granularity int64, gpuInfo *modelInfo) {
	marshaler := jsonpb.Marshaler{}
	dataGranularity := queue.GetGranularityStr(granularity)
	gpuHost := gpu.GetMetadata().GetHost()
	gpuMinorNumber := gpu.GetMetadata().GetMinorNumber()
	gpuStr, err := marshaler.MarshalToString(gpu)
	if err != nil {
		scope.Errorf("Encode pb message failed for (gpu host: %s minor number: %s) with granularity seconds %v. %s",
			gpuHost, gpuMinorNumber, granularity, err.Error())
		return
	}
	if len(gpuInfo.ModelMetrics) > 0 && gpuStr != "" {
		jb := queue.NewJobBuilder(pdUnit, granularity, gpuStr)
		jobJSONStr, err := jb.GetJobJSONString()
		if err != nil {
			scope.Errorf(
				"Prepare model job payload failed for (gpu host: %s, minor number: %s) with granularity seconds %v. %s",
				gpuHost, gpuMinorNumber, granularity, err.Error())
			return
		}

		gpuJobStr := fmt.Sprintf("%s/%s/%v", gpuHost, gpuMinorNumber, granularity)
		scope.Infof("Try to send gpu model job: %s", gpuJobStr)
		err = queueSender.SendJsonString(modelQueueName, jobJSONStr, gpuJobStr)
		if err == nil {
			sender.modelMapper.AddModelInfo(pdUnit, dataGranularity, gpuInfo)
		} else {
			scope.Errorf(
				"Send model job payload failed for (gpu host: %s minor number: %s) with granularity seconds %v. %s",
				gpuHost, gpuMinorNumber, granularity, err.Error())
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

func (sender *gpuModelJobSender) getLastPrediction(datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	gpu *datahub_gpu.Gpu, granularity int64) ([]*datahub_predictions.MetricData, error) {
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

	if len(gpuPredictRes.GetGpuPredictions()) > 0 {
		lastGpuPrediction := gpuPredictRes.GetGpuPredictions()[0]
		if lastGpuPrediction.GetPredictedRawData() != nil {
			return lastGpuPrediction.GetPredictedRawData(), nil
		} else if lastGpuPrediction.GetPredictedLowerboundData() != nil {
			return lastGpuPrediction.GetPredictedLowerboundData(), nil
		} else if lastGpuPrediction.GetPredictedUpperboundData() != nil {
			return lastGpuPrediction.GetPredictedUpperboundData(), nil
		}
	}
	return nil, nil
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
		scope.Infof("No prediction metric found of gpu (host: %s, minor number: %s), send model jobs with granularity %v",
			gpuHost, gpuMinorNumber, granularity)
		return
	}
	for _, lastPredictionMetric := range lastPredictionMetrics {
		if len(lastPredictionMetric.GetData()) == 0 {
			gpuInfo := sender.genGPUInfo(gpuHost, gpuMinorNumber,
				datahub_common.MetricType_MEMORY_USAGE_BYTES,
				datahub_common.MetricType_DUTY_CYCLE)
			sender.sendJob(gpu, queueSender, pdUnit, granularity, gpuInfo)
			scope.Infof("No prediction metric %s found of gpu (host: %s, minor number: %s), send model jobs with granularity %v",
				lastPredictionMetric.GetMetricType().String(), gpuHost, gpuMinorNumber, granularity)
			return
		} else {
			lastPrediction := lastPredictionMetric.GetData()[0]
			lastPredictionTime := lastPredictionMetric.GetData()[0].GetTime().GetSeconds()
			if lastPrediction != nil && lastPredictionTime <= nowSeconds {
				scope.Infof("gpu prediction (host: %s, minor number: %s) is out of date due to last predict time is %v (current: %v)",
					gpuHost, gpuMinorNumber, lastPredictionTime, nowSeconds)
			}

			if lastPrediction != nil && lastPredictionTime <= nowSeconds {
				gpuInfo := sender.genGPUInfo(gpuHost, gpuMinorNumber,
					datahub_common.MetricType_MEMORY_USAGE_BYTES,
					datahub_common.MetricType_DUTY_CYCLE)
				scope.Infof("send gpu (host: %s, minor number: %s) model job due to no predict found or predict is out of date, send model jobs with granularity %v",
					gpuHost, gpuMinorNumber, granularity)
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
				scope.Errorf("Get (gpu host: %s minor number: %s) Prediction with granularity %v for sending model job failed: %s",
					gpuHost, gpuMinorNumber, granularity, err.Error())
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
				scope.Errorf("List gpu (gpu host: %s minor number: %s) metric with granularity %v for sending model job failed: %s",
					gpuHost, gpuMinorNumber, granularity, err.Error())
				continue
			}
			gpuMetrics := gpuMetricsRes.GetGpuMetrics()
			// gpu tags are host, minor number, pid, mid
			for _, gpuMetric := range gpuMetrics {
				metricData := gpuMetric.GetMetricData()
				for _, metricDatum := range metricData {
					mData := metricDatum.GetData()
					pData := []*datahub_predictions.Sample{}
					gpuInfo := sender.genGPUInfo(gpuHost, gpuMinorNumber)
					for _, gpuPrediction := range gpuPredictions {
						predictRawData := gpuPrediction.GetPredictedRawData()
						for _, predictRawDatum := range predictRawData {
							if metricDatum.GetMetricType() == predictRawDatum.GetMetricType() {
								pData = append(pData, predictRawDatum.GetData()...)
							}
						}
					}
					metricsNeedToModel, drift := DriftEvaluation(UnitTypeGPU, metricDatum.GetMetricType(), granularity, mData, pData, map[string]string{
						"gpuHost":           gpuHost,
						"gpuMinorNumber":    gpuMinorNumber,
						"targetDisplayName": fmt.Sprintf("gpu host: %s minor number: %s", gpuHost, gpuMinorNumber),
					}, sender.metricExporter)
					if drift {
						scope.Infof("export (gpu host: %s minor number: %s) drift counter with granularity %s",
							gpuHost, gpuMinorNumber, dataGranularity)
						sender.metricExporter.AddGPUMetricDrift(gpuHost, gpuMinorNumber,
							queue.GetGranularityStr(granularity), 1.0)
					}
					gpuInfo.ModelMetrics = append(gpuInfo.ModelMetrics, metricsNeedToModel...)
					isModeling := sender.modelMapper.IsModeling(pdUnit, dataGranularity, gpuInfo)
					if !isModeling || (isModeling && sender.modelMapper.IsModelTimeout(
						pdUnit, dataGranularity, gpuInfo)) {
						sender.sendJob(gpu, queueSender, pdUnit, granularity, gpuInfo)
						return
					}
				}
			}
		}
	}
}
