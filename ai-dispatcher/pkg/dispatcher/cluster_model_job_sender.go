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
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

type clusterModelJobSender struct {
	datahubGrpcCn  *grpc.ClientConn
	modelMapper    *ModelMapper
	metricExporter *metrics.Exporter
}

func NewClusterModelJobSender(datahubGrpcCn *grpc.ClientConn, modelMapper *ModelMapper,
	metricExporter *metrics.Exporter) *clusterModelJobSender {
	return &clusterModelJobSender{
		datahubGrpcCn:  datahubGrpcCn,
		modelMapper:    modelMapper,
		metricExporter: metricExporter,
	}
}

func (sender *clusterModelJobSender) sendModelJobs(clusters []*datahub_resources.Cluster,
	queueSender queue.QueueSender, pdUnit string, granularity int64, predictionStep int64) {

	datahubServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(sender.datahubGrpcCn)
	for _, cluster := range clusters {
		if granularity == 30 && !viper.GetBool("hourlyPredict") {
			continue
		}

		clusterName := cluster.GetObjectMeta().GetName()
		lastPredictionMetrics, err := sender.getLastPrediction(datahubServiceClnt, cluster, granularity)
		if err != nil {
			scope.Infof("Get cluster %s last prediction failed: %s",
				clusterName, err.Error())
			continue
		}
		if lastPredictionMetrics == nil && err == nil {
			scope.Infof("No prediction found of cluster %s",
				clusterName)
		}

		sender.sendJobByMetrics(cluster, queueSender, pdUnit, granularity, predictionStep,
			datahubServiceClnt, lastPredictionMetrics)
	}
}

func (sender *clusterModelJobSender) sendJob(cluster *datahub_resources.Cluster, queueSender queue.QueueSender, pdUnit string,
	granularity int64, clusterInfo *modelInfo) {
	clusterName := cluster.GetObjectMeta().GetName()
	dataGranularity := queue.GetGranularityStr(granularity)
	marshaler := jsonpb.Marshaler{}
	clusterStr, err := marshaler.MarshalToString(cluster)
	if err != nil {
		scope.Errorf("Encode pb message failed for cluster %s with granularity seconds %v. %s",
			clusterName, granularity, err.Error())
		return
	}
	if len(clusterInfo.ModelMetrics) > 0 && clusterStr != "" {
		jb := queue.NewJobBuilder(pdUnit, granularity, clusterStr)
		jobJSONStr, err := jb.GetJobJSONString()
		if err != nil {
			scope.Errorf(
				"Prepare model job payload failed for cluster %s with granularity seconds %v. %s",
				clusterName, granularity, err.Error())
			return
		}

		err = queueSender.SendJsonString(modelQueueName, jobJSONStr,
			fmt.Sprintf("%s/%v", clusterName, granularity))
		if err == nil {
			sender.modelMapper.AddModelInfo(pdUnit, dataGranularity, clusterInfo)
		} else {
			scope.Errorf(
				"Send model job payload failed for cluster %s with granularity seconds %v. %s",
				clusterName, granularity, err.Error())
		}
	}
}

func (sender *clusterModelJobSender) genClusterInfo(clusterName string,
	modelMetrics ...datahub_common.MetricType) *modelInfo {
	clusterInfo := new(modelInfo)
	clusterInfo.Name = clusterName
	clusterInfo.ModelMetrics = modelMetrics
	clusterInfo.SetTimeStamp(time.Now().Unix())
	return clusterInfo
}

func (sender *clusterModelJobSender) getLastPrediction(datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	cluster *datahub_resources.Cluster, granularity int64) ([]*datahub_predictions.MetricData, error) {
	clusterName := cluster.ObjectMeta.GetName()
	clusterPredictRes, err := datahubServiceClnt.ListClusterPredictions(context.Background(),
		&datahub_predictions.ListClusterPredictionsRequest{
			ObjectMeta: []*datahub_resources.ObjectMeta{
				&datahub_resources.ObjectMeta{
					Name: clusterName,
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
	if len(clusterPredictRes.GetClusterPredictions()) > 0 {
		lastClusterPrediction := clusterPredictRes.GetClusterPredictions()[0]
		if lastClusterPrediction.GetPredictedRawData() != nil {
			return lastClusterPrediction.GetPredictedRawData(), nil
		} else if lastClusterPrediction.GetPredictedLowerboundData() != nil {
			return lastClusterPrediction.GetPredictedLowerboundData(), nil
		} else if lastClusterPrediction.GetPredictedUpperboundData() != nil {
			return lastClusterPrediction.GetPredictedUpperboundData(), nil
		}
	}
	return nil, nil
}

func (sender *clusterModelJobSender) getQueryMetricStartTime(descClusterPredictions []*datahub_predictions.ClusterPrediction) int64 {
	if len(descClusterPredictions) > 0 {
		pdMDs := descClusterPredictions[len(descClusterPredictions)-1].GetPredictedRawData()
		for _, pdMD := range pdMDs {
			mD := pdMD.GetData()
			if len(mD) > 0 {
				return mD[len(mD)-1].GetTime().GetSeconds()
			}
		}
	}
	return 0
}

func (sender *clusterModelJobSender) sendJobByMetrics(cluster *datahub_resources.Cluster, queueSender queue.QueueSender,
	pdUnit string, granularity int64, predictionStep int64, datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	lastPredictionMetrics []*datahub_predictions.MetricData) {
	clusterName := cluster.ObjectMeta.GetName()
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
		clusterInfo := sender.genClusterInfo(clusterName,
			datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
			datahub_common.MetricType_MEMORY_USAGE_BYTES)
		sender.sendJob(cluster, queueSender, pdUnit, granularity, clusterInfo)
		scope.Infof("No prediction metrics found of cluster %s, send model jobs with granularity %v",
			clusterName, granularity)
		return
	}

	for _, lastPredictionMetric := range lastPredictionMetrics {
		if len(lastPredictionMetric.GetData()) == 0 {
			clusterInfo := sender.genClusterInfo(clusterName,
				datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
				datahub_common.MetricType_MEMORY_USAGE_BYTES)
			sender.sendJob(cluster, queueSender, pdUnit, granularity, clusterInfo)
			scope.Infof("No prediction metric %s found of cluster %s, send model jobs with granularity %v",
				lastPredictionMetric.GetMetricType().String(), clusterName, granularity)
			return
		} else {
			lastPrediction := lastPredictionMetric.GetData()[0]
			lastPredictionTime := lastPredictionMetric.GetData()[0].GetTime().GetSeconds()
			if lastPrediction != nil && lastPredictionTime <= nowSeconds {
				scope.Infof("cluster prediction %s is out of date due to last predict time is %v (current: %v)",
					clusterName, lastPredictionTime, nowSeconds)
			}
			if lastPrediction != nil && lastPredictionTime <= nowSeconds {
				clusterInfo := sender.genClusterInfo(clusterName,
					datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
					datahub_common.MetricType_MEMORY_USAGE_BYTES)
				scope.Infof("send cluster %s model job due to no predict found or predict is out of date, send model jobs with granularity %v",
					clusterName, granularity)
				sender.sendJob(cluster, queueSender, pdUnit, granularity, clusterInfo)
				return
			}

			clusterPredictRes, err := datahubServiceClnt.ListClusterPredictions(context.Background(),
				&datahub_predictions.ListClusterPredictionsRequest{
					ObjectMeta: []*datahub_resources.ObjectMeta{
						&datahub_resources.ObjectMeta{
							Name: clusterName,
						},
					},
					ModelId:        lastPrediction.GetModelId(),
					Granularity:    granularity,
					QueryCondition: queryCondition,
				})
			if err != nil {
				scope.Errorf("Get cluster %s Prediction with granularity %v for sending model job failed: %s",
					clusterName, granularity, err.Error())
				continue
			}
			clusterPredictions := clusterPredictRes.GetClusterPredictions()
			queryStartTime := time.Now().Unix() - predictionStep*granularity
			firstPDTime := sender.getQueryMetricStartTime(clusterPredictions)
			if firstPDTime > 0 {
				queryStartTime = firstPDTime
			}
			clusterMetricsRes, err := datahubServiceClnt.ListClusterMetrics(context.Background(),
				&datahub_metrics.ListClusterMetricsRequest{
					QueryCondition: &datahub_common.QueryCondition{
						Order: datahub_common.QueryCondition_DESC,
						TimeRange: &datahub_common.TimeRange{
							StartTime: &timestamp.Timestamp{
								Seconds: queryStartTime,
							},
							Step: &duration.Duration{
								Seconds: granularity,
							},
						},
					},
					ObjectMeta: []*datahub_resources.ObjectMeta{
						&datahub_resources.ObjectMeta{
							Name: clusterName,
						},
					},
				})
			if err != nil {
				scope.Errorf("List clusters %s metric with granularity %v for sending model job failed: %s",
					clusterName, granularity, err.Error())
				continue
			}
			clusterMetrics := clusterMetricsRes.GetClusterMetrics()

			for _, clusterMetric := range clusterMetrics {
				metricData := clusterMetric.GetMetricData()
				for _, metricDatum := range metricData {
					mData := metricDatum.GetData()
					pData := []*datahub_predictions.Sample{}
					clusterInfo := sender.genClusterInfo(clusterName)
					for _, clusterPrediction := range clusterPredictions {
						predictRawData := clusterPrediction.GetPredictedRawData()
						for _, predictRawDatum := range predictRawData {
							if metricDatum.GetMetricType() == predictRawDatum.GetMetricType() {
								pData = append(pData, predictRawDatum.GetData()...)
							}
						}
					}
					metricsNeedToModel, drift := DriftEvaluation(UnitTypeCluster, metricDatum.GetMetricType(), granularity, mData, pData, map[string]string{
						"clusterName":       clusterName,
						"targetDisplayName": fmt.Sprintf("cluster %s", clusterName),
					}, sender.metricExporter)
					if drift {
						scope.Infof("export cluster %s drift counter with granularity %s",
							clusterName, dataGranularity)
						sender.metricExporter.AddClusterMetricDrift(clusterName, queue.GetGranularityStr(granularity), 1.0)
					}
					clusterInfo.ModelMetrics = append(clusterInfo.ModelMetrics, metricsNeedToModel...)
					isModeling := sender.modelMapper.IsModeling(pdUnit, dataGranularity, clusterInfo)
					if !isModeling || (isModeling && sender.modelMapper.IsModelTimeout(
						pdUnit, dataGranularity, clusterInfo)) {
						sender.sendJob(cluster, queueSender, pdUnit, granularity, clusterInfo)
					}
				}
			}
		}
	}
}
