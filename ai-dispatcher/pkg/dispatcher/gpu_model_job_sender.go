package dispatcher

import (
	"context"
	"fmt"
	"time"

	"github.com/containers-ai/alameda/ai-dispatcher/pkg/metrics"
	"github.com/containers-ai/alameda/ai-dispatcher/pkg/queue"
	"github.com/containers-ai/alameda/ai-dispatcher/pkg/stats"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

type gpuModelJobSender struct {
	datahubGrpcCn  *grpc.ClientConn
	modelThreshold float64
	modelMapper    *ModelMapper
	metricExporter *metrics.Exporter
}

func NewGPUModelJobSender(datahubGrpcCn *grpc.ClientConn, modelMapper *ModelMapper,
	metricExporter *metrics.Exporter) *gpuModelJobSender {
	return &gpuModelJobSender{
		datahubGrpcCn:  datahubGrpcCn,
		modelThreshold: viper.GetFloat64("model.threshold"),
		modelMapper:    modelMapper,
		metricExporter: metricExporter,
	}
}

func (sender *gpuModelJobSender) sendModelJobs(gpus []*datahub_v1alpha1.Gpu,
	queueSender queue.QueueSender, pdUnit string, granularity int64, predictionStep int64) {

	dataGranularity := queue.GetGranularityStr(granularity)
	datahubServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(sender.datahubGrpcCn)
	queryCondition := &datahub_v1alpha1.QueryCondition{
		Order: datahub_v1alpha1.QueryCondition_DESC,
		TimeRange: &datahub_v1alpha1.TimeRange{
			StartTime: &timestamp.Timestamp{
				Seconds: time.Now().Unix() - predictionStep*granularity,
			},
			Step: &duration.Duration{
				Seconds: granularity,
			},
		},
	}
	for _, gpu := range gpus {
		gpuHost := gpu.GetMetadata().GetHost()
		gpuMinorNumber := gpu.GetMetadata().GetMinorNumber()

		_, err := sender.getLastPrediction(datahubServiceClnt, gpu, granularity)
		if err != nil {
			scope.Infof("Get gpu last prediction failed: %s",
				err.Error())
			gpuInfo := sender.genGPUInfo(gpuHost, gpuMinorNumber)
			gpuInfo.ModelMetrics = []datahub_v1alpha1.MetricType{
				datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES,
				datahub_v1alpha1.MetricType_DUTY_CYCLE,
			}
			sender.sendJob(gpu, queueSender, pdUnit, granularity, gpuInfo)
			continue
		}

		//TODO: use mid to query
		gpuPredictRes, err := datahubServiceClnt.ListGpuPredictions(context.Background(),
			&datahub_v1alpha1.ListGpuPredictionsRequest{
				Host:           gpuHost,
				MinorNumber:    gpuMinorNumber,
				Granularity:    granularity,
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
			&datahub_v1alpha1.ListGpuMetricsRequest{
				QueryCondition: &datahub_v1alpha1.QueryCondition{
					Order: datahub_v1alpha1.QueryCondition_DESC,
					TimeRange: &datahub_v1alpha1.TimeRange{
						StartTime: &timestamp.Timestamp{
							Seconds: queryStartTime,
						},
						Step: &duration.Duration{
							Seconds: granularity,
						},
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
				pData := []*datahub_v1alpha1.Sample{}
				gpuInfo := sender.genGPUInfo(gpuHost, gpuMinorNumber)
				for _, gpuPrediction := range gpuPredictions {
					predictRawData := gpuPrediction.GetPredictedRawData()
					for _, predictRawDatum := range predictRawData {
						if metricDatum.GetMetricType() == predictRawDatum.GetMetricType() {
							pData = append(pData, predictRawDatum.GetData()...)
						}
					}
				}
				if len(pData) > 0 {
					metricType := metricDatum.GetMetricType()
					scope.Infof("start MAPE calculation for (gpu host: %s minor number: %s) metric %v with granularity %v",
						gpuHost, gpuMinorNumber, metricType, granularity)
					measurementDataSet := stats.NewMeasurementDataSet(mData, pData, granularity)
					mape, mapeErr := stats.MAPE(measurementDataSet)
					if mapeErr == nil {
						scope.Infof("export MAPE value %v for (gpu host: %s minor number: %s) metric %v with granularity %v", mape,
							gpuHost, gpuMinorNumber, metricType, granularity)
						sender.metricExporter.SetGPUMetricMAPE(gpuHost, gpuMinorNumber,
							queue.GetMetricLabel(metricDatum.GetMetricType()), queue.GetGranularityStr(granularity), mape)
					}
					if err != nil {
						gpuInfo.ModelMetrics = append(gpuInfo.ModelMetrics, metricType)
						scope.Infof(
							"model job for (gpu host: %s minor number: %s) metric %v with granularity %v should be sent due to MAPE calculation failed: %s",
							gpuHost, gpuMinorNumber, metricType, granularity, mapeErr.Error())
					} else if mape > sender.modelThreshold {
						gpuInfo.ModelMetrics = append(gpuInfo.ModelMetrics, metricType)
						scope.Infof("model job (gpu host: %s minor number: %s) metric %v with granularity %v should be sent due to MAPE %v > %v",
							gpuHost, gpuMinorNumber, metricType, granularity, mape, sender.modelThreshold)
					} else {
						scope.Infof("(gpu host: %s minor number: %s) metric %v with granularity %v MAPE %v <= %v, skip sending this model metric",
							gpuHost, gpuMinorNumber, metricType, granularity, mape, sender.modelThreshold)
					}
				}
				isModeling := sender.modelMapper.IsModeling(pdUnit, dataGranularity, gpuInfo)
				if !isModeling || (isModeling && sender.modelMapper.IsModelTimeout(
					pdUnit, dataGranularity, gpuInfo)) {
					sender.sendJob(gpu, queueSender, pdUnit, granularity, gpuInfo)
				}
			}
		}
	}
}

func (sender *gpuModelJobSender) sendJob(gpu *datahub_v1alpha1.Gpu, queueSender queue.QueueSender, pdUnit string,
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

		scope.Infof("export (gpu host: %s minor number: %s) drift counter with granularity %s",
			gpuHost, gpuMinorNumber, dataGranularity)
		sender.metricExporter.AddGPUMetricDrift(gpuHost, gpuMinorNumber,
			queue.GetGranularityStr(granularity), 1.0)

		err = queueSender.SendJsonString(modelQueueName, jobJSONStr,
			fmt.Sprintf("%s/%s", gpuHost, gpuMinorNumber))
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
	gpuMinorNumber string) *modelInfo {
	gpuInfo := new(modelInfo)
	gpuInfo.Host = gpuHost
	gpuInfo.MinorNumber = gpuMinorNumber
	gpuInfo.ModelMetrics = []datahub_v1alpha1.MetricType{}
	gpuInfo.SetTimeStamp(time.Now().Unix())
	return gpuInfo
}

func (sender *gpuModelJobSender) getLastPrediction(datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	gpu *datahub_v1alpha1.Gpu, granularity int64) (*datahub_v1alpha1.GpuPrediction, error) {
	gpuHost := gpu.GetMetadata().GetHost()
	gpuMinorNumber := gpu.GetMetadata().GetMinorNumber()
	gpuPredictRes, err := datahubServiceClnt.ListGpuPredictions(context.Background(),
		&datahub_v1alpha1.ListGpuPredictionsRequest{
			Host:        gpuHost,
			MinorNumber: gpuMinorNumber,
			Granularity: granularity,
			QueryCondition: &datahub_v1alpha1.QueryCondition{
				Limit: 1,
				Order: datahub_v1alpha1.QueryCondition_DESC,
				TimeRange: &datahub_v1alpha1.TimeRange{
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
		return gpuPredictRes.GetGpuPredictions()[0], nil
	}
	return nil, fmt.Errorf("No gpu (host: %s, minor number: %s) prediction found",
		gpuHost, gpuMinorNumber)
}

func (sender *gpuModelJobSender) getQueryMetricStartTime(descGpuPredictions []*datahub_v1alpha1.GpuPrediction) int64 {
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
