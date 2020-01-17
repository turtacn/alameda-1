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
		sender.sendGpuModelJobs(gpu, queueSender, pdUnit, granularity, predictionStep, &wg)
	}
}

func (sender *gpuModelJobSender) sendGpuModelJobs(gpu *datahub_gpu.Gpu,
	queueSender queue.QueueSender, pdUnit string, granularity int64, predictionStep int64, wg *sync.WaitGroup) {
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

	sender.sendJobByMetrics(gpu, queueSender, pdUnit, granularity, predictionStep,
		datahubServiceClnt, lastPredictionMetrics)
}

func (sender *gpuModelJobSender) sendJob(gpu *datahub_gpu.Gpu,
	queueSender queue.QueueSender, pdUnit string, granularity int64,
	metricType datahub_common.MetricType) {
	marshaler := jsonpb.Marshaler{}
	clusterID := "GPU_CLUSTER_NAME"
	dataGranularity := queue.GetGranularityStr(granularity)
	gpuHost := gpu.GetMetadata().GetHost()
	gpuMinorNumber := gpu.GetMetadata().GetMinorNumber()
	gpuStr, err := marshaler.MarshalToString(gpu)
	if err != nil {
		scope.Errorf("[GPU][%s][%s/%s] Encode pb message failed. %s",
			dataGranularity, gpuHost, gpuMinorNumber, err.Error())
		return
	}

	jb := queue.NewJobBuilder(clusterID, pdUnit, granularity, metricType, gpuStr, nil)
	jobJSONStr, err := jb.GetJobJSONString()
	if err != nil {
		scope.Errorf(
			"[GPU][%s][%s/%s] Prepare model job payload failed. %s",
			dataGranularity, gpuHost, gpuMinorNumber, err.Error())
		return
	}

	gpuJobStr := fmt.Sprintf("%s/%s/%s/%s/%v/%s", consts.UnitTypeGPU, clusterID, gpuHost, gpuMinorNumber, granularity, metricType)
	scope.Infof("[GPU][%s][%s/%s] Try to send gpu model job: %s", dataGranularity, gpuHost, gpuMinorNumber, gpuJobStr)
	err = queueSender.SendJsonString(modelQueueName, jobJSONStr, gpuJobStr, granularity)
	if err == nil {
		sender.modelMapper.AddModelInfo(clusterID, pdUnit, dataGranularity, metricType.String(), map[string]string{
			"host":        gpuHost,
			"minorNumber": gpuMinorNumber,
		})
	} else {
		scope.Errorf(
			"[GPU][%s][%s/%s] Send model job payload failed. %s",
			dataGranularity, gpuHost, gpuMinorNumber, err.Error())
	}

}

func (sender *gpuModelJobSender) getLastMIdPrediction(datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	gpu *datahub_gpu.Gpu, granularity int64) ([]*datahub_predictions.MetricData, error) {

	metricData := []*datahub_predictions.MetricData{}
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
		return metricData, err
	}

	lastMid := ""
	if len(gpuPredictRes.GetGpuPredictions()) == 0 {
		return []*datahub_predictions.MetricData{}, nil
	}

	lastGpuPrediction := gpuPredictRes.GetGpuPredictions()[0]
	lgpuPDRData := lastGpuPrediction.GetPredictedRawData()
	if lgpuPDRData == nil {
		return metricData, nil
	}
	for _, pdRD := range lgpuPDRData {
		for _, theData := range pdRD.GetData() {
			lastMid = theData.GetModelId()
			break
		}

		if lastMid == "" {
			scope.Warnf("[GPU][%s][%s/%s] Query last model id for metric %s is empty",
				dataGranularity, gpuHost, gpuMinorNumber, pdRD.GetMetricType())
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
				ModelId: lastMid,
			})
		if err != nil {
			scope.Warnf("[GPU][%s][%s/%s] Query last model id %v for metric %s failed",
				dataGranularity, gpuHost, gpuMinorNumber, lastMid, pdRD.GetMetricType())
			continue
		}

		for _, gpuPrediction := range gpuPredictRes.GetGpuPredictions() {
			for _, lMIDPdRD := range gpuPrediction.GetPredictedRawData() {
				if lMIDPdRD.GetMetricType() == pdRD.GetMetricType() {
					metricData = append(metricData, lMIDPdRD)
				}
			}
		}
	}

	return metricData, nil
}

func (sender *gpuModelJobSender) getQueryMetricStartTime(metricData *datahub_predictions.MetricData) int64 {
	mD := metricData.GetData()
	if len(mD) > 0 {
		return mD[len(mD)-1].GetTime().GetSeconds()
	}
	return 0
}

func (sender *gpuModelJobSender) sendJobByMetrics(gpu *datahub_gpu.Gpu, queueSender queue.QueueSender,
	pdUnit string, granularity int64, predictionStep int64, datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	lastPredictionMetrics []*datahub_predictions.MetricData) {
	clusterID := "GPU_CLUSTER_NAME"
	dataGranularity := queue.GetGranularityStr(granularity)
	gpuHost := gpu.GetMetadata().GetHost()
	gpuMinorNumber := gpu.GetMetadata().GetMinorNumber()
	nowSeconds := time.Now().Unix()

	if len(lastPredictionMetrics) == 0 {
		scope.Infof("[GPU][%s][%s/%s] No prediction metric found, send model jobs",
			dataGranularity, gpuHost, gpuMinorNumber)
		for _, metricType := range []datahub_common.MetricType{
			datahub_common.MetricType_MEMORY_USAGE_BYTES,
			datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
		} {
			sender.sendJob(gpu, queueSender, pdUnit, granularity, metricType)
		}
		return
	}

	for _, lastPredictionMetric := range lastPredictionMetrics {
		if len(lastPredictionMetric.GetData()) == 0 {
			scope.Infof("[GPU][%s][%s/%s] No prediction metric %s found, send model jobs",
				dataGranularity, gpuHost, gpuMinorNumber, lastPredictionMetric.GetMetricType().String())
			sender.sendJob(gpu, queueSender, pdUnit, granularity, lastPredictionMetric.GetMetricType())
			continue
		} else {
			lastPrediction := lastPredictionMetric.GetData()[0]
			lastPredictionTime := lastPredictionMetric.GetData()[0].GetTime().GetSeconds()
			if lastPrediction != nil && lastPredictionTime <= nowSeconds {
				scope.Infof("[GPU][%s][%s/%s] Send model job due to no predict metric %s found or is out of date",
					dataGranularity, gpuHost, gpuMinorNumber, lastPredictionMetric.GetMetricType().String())
				sender.sendJob(gpu, queueSender, pdUnit, granularity, lastPredictionMetric.GetMetricType())
				continue
			}

			queryStartTime := time.Now().Unix() - predictionStep*granularity
			firstPDTime := sender.getQueryMetricStartTime(lastPredictionMetric)
			if firstPDTime > 0 && firstPDTime <= time.Now().Unix() {
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
					MetricTypes: []datahub_common.MetricType{
						lastPredictionMetric.GetMetricType(),
					},
				})

			if err != nil {
				scope.Errorf("[GPU][%s][%s/%s] List gpu metric for sending model job failed: %s",
					dataGranularity, gpuHost, gpuMinorNumber, err.Error())
				continue
			}
			gpuMetrics := gpuMetricsRes.GetGpuMetrics()
			// gpu tags are host, minor number, pid, mid

			predictRawData := lastPredictionMetrics
			for _, predictRawDatum := range predictRawData {
				for _, gpuMetric := range gpuMetrics {
					metricData := gpuMetric.GetMetricData()
					for _, metricDatum := range metricData {
						mData := metricDatum.GetData()
						pData := []*datahub_predictions.Sample{}
						if metricDatum.GetMetricType() == predictRawDatum.GetMetricType() {
							pData = append(pData, predictRawDatum.GetData()...)
							metricsNeedToModel, drift := DriftEvaluation(consts.UnitTypeGPU, predictRawDatum.GetMetricType(), granularity, mData, pData, map[string]string{
								"clusterID":         clusterID,
								"gpuHost":           gpuHost,
								"gpuMinorNumber":    gpuMinorNumber,
								"targetDisplayName": fmt.Sprintf("[GPU][%s][%s/%s]", dataGranularity, gpuHost, gpuMinorNumber),
							}, sender.metricExporter)

							for _, mntm := range metricsNeedToModel {
								if drift {
									scope.Infof("[GPU][%s][%s/%s] Export metric %s drift counter",
										dataGranularity, gpuHost, gpuMinorNumber, mntm)
									sender.metricExporter.AddGPUMetricDrift(clusterID, gpuHost, gpuMinorNumber,
										queue.GetGranularityStr(granularity), mntm.String(), time.Now().Unix(), 1.0)
								}
								isModeling := sender.modelMapper.IsModeling(clusterID, pdUnit, dataGranularity, mntm.String(), map[string]string{
									"host":        gpuHost,
									"minorNumber": gpuMinorNumber,
								})
								if !isModeling || (isModeling && sender.modelMapper.IsModelTimeout(
									clusterID, pdUnit, dataGranularity, mntm.String(), map[string]string{
										"host":        gpuHost,
										"minorNumber": gpuMinorNumber,
									})) {
									sender.sendJob(gpu, queueSender, pdUnit, granularity, mntm)
								}
							}
						}
					}
				}
			}
		}
	}
}
