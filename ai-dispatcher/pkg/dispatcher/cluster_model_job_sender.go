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
	for _, cluster := range clusters {
		go sender.sendClusterModelJobs(cluster, queueSender, pdUnit, granularity, predictionStep)
	}
}

func (sender *clusterModelJobSender) sendClusterModelJobs(cluster *datahub_resources.Cluster,
	queueSender queue.QueueSender, pdUnit string, granularity int64, predictionStep int64) {
	dataGranularity := queue.GetGranularityStr(granularity)
	datahubServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(sender.datahubGrpcCn)
	clusterName := cluster.GetObjectMeta().GetName()
	lastPredictionMetrics, err := sender.getLastMIdPrediction(datahubServiceClnt, cluster, granularity)
	if err != nil {
		scope.Infof("[CLUSTER][%s][%s] Get last prediction failed: %s",
			dataGranularity, clusterName, err.Error())
		return
	}
	if lastPredictionMetrics == nil && err == nil {
		scope.Infof("[CLUSTER][%s][%s] No prediction found",
			dataGranularity, clusterName)
	}

	sender.sendJobByMetrics(cluster, queueSender, pdUnit, granularity, predictionStep,
		datahubServiceClnt, lastPredictionMetrics)
}

func (sender *clusterModelJobSender) sendJob(cluster *datahub_resources.Cluster, queueSender queue.QueueSender, pdUnit string,
	granularity int64, clusterInfo *modelInfo) {
	clusterName := cluster.GetObjectMeta().GetName()
	dataGranularity := queue.GetGranularityStr(granularity)
	marshaler := jsonpb.Marshaler{}
	clusterStr, err := marshaler.MarshalToString(cluster)
	if err != nil {
		scope.Errorf("[CLUSTER][%s][%s] Encode pb message failed. %s",
			dataGranularity, clusterName, err.Error())
		return
	}
	if len(clusterInfo.ModelMetrics) > 0 && clusterStr != "" {
		jb := queue.NewJobBuilder(pdUnit, granularity, clusterStr)
		jobJSONStr, err := jb.GetJobJSONString()
		if err != nil {
			scope.Errorf(
				"[CLUSTER][%s][%s] Prepare model job payload failed. %s",
				dataGranularity, clusterName, err.Error())
			return
		}

		clusterJobStr := fmt.Sprintf("%s/%v", clusterName, granularity)
		scope.Infof("[CLUSTER][%s][%s] Try to send cluster model job: %s", dataGranularity, clusterName, clusterJobStr)
		err = queueSender.SendJsonString(modelQueueName, jobJSONStr, clusterJobStr, granularity)
		if err == nil {
			sender.modelMapper.AddModelInfo(pdUnit, dataGranularity, clusterInfo)
		} else {
			scope.Errorf(
				"[CLUSTER][%s][%s] Send model job payload failed. %s",
				dataGranularity, clusterName, err.Error())
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

func (sender *clusterModelJobSender) getLastMIdPrediction(datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	cluster *datahub_resources.Cluster, granularity int64) ([]*datahub_predictions.MetricData, error) {
	dataGranularity := queue.GetGranularityStr(granularity)
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
	lastPid := ""
	if len(clusterPredictRes.GetClusterPredictions()) > 0 {
		lastClusterPrediction := clusterPredictRes.GetClusterPredictions()[0]
		lnsPDRData := lastClusterPrediction.GetPredictedRawData()
		if lnsPDRData == nil {
			lnsPDRData = lastClusterPrediction.GetPredictedLowerboundData()
		}
		if lnsPDRData == nil {
			lnsPDRData = lastClusterPrediction.GetPredictedUpperboundData()
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
		return nil, fmt.Errorf("[CLUSTER][%s][%s] Query last prediction id failed",
			dataGranularity, clusterName)
	}
	clusterPredictRes, err = datahubServiceClnt.ListClusterPredictions(context.Background(),
		&datahub_predictions.ListClusterPredictionsRequest{
			ObjectMeta: []*datahub_resources.ObjectMeta{
				&datahub_resources.ObjectMeta{
					Name: clusterName,
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

	if len(clusterPredictRes.GetClusterPredictions()) > 0 {
		metricData := []*datahub_predictions.MetricData{}
		for _, clusterPrediction := range clusterPredictRes.GetClusterPredictions() {
			for _, pdRD := range clusterPrediction.GetPredictedRawData() {
				for _, pdD := range pdRD.GetData() {
					modelID := pdD.GetModelId()
					if modelID != "" {
						mIDClusterPrediction, err := sender.getPredictionByMId(datahubServiceClnt, cluster, granularity, modelID)
						if err != nil {
							scope.Errorf("[CLUSTER][%s][%s] Query prediction with model Id %s failed. %s",
								dataGranularity, clusterName, modelID, err.Error())
						}
						for _, mIDClusterPD := range mIDClusterPrediction {
							metricData = append(metricData, mIDClusterPD.GetPredictedRawData()...)
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

func (sender *clusterModelJobSender) getPredictionByMId(datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	cluster *datahub_resources.Cluster, granularity int64, modelID string) ([]*datahub_predictions.ClusterPrediction, error) {
	clusterPredictRes, err := datahubServiceClnt.ListClusterPredictions(context.Background(),
		&datahub_predictions.ListClusterPredictionsRequest{
			Granularity: granularity,
			ObjectMeta: []*datahub_resources.ObjectMeta{
				&datahub_resources.ObjectMeta{
					Name: cluster.GetObjectMeta().GetName(),
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
	return clusterPredictRes.GetClusterPredictions(), err
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
		scope.Infof("[CLUSTER][%s][%s] No prediction metrics found, send model jobs",
			dataGranularity, clusterName)
		return
	}

	clusterInfo := sender.genClusterInfo(clusterName)
	for _, lastPredictionMetric := range lastPredictionMetrics {
		if len(lastPredictionMetric.GetData()) == 0 {
			clusterInfo := sender.genClusterInfo(clusterName,
				datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
				datahub_common.MetricType_MEMORY_USAGE_BYTES)
			sender.sendJob(cluster, queueSender, pdUnit, granularity, clusterInfo)
			scope.Infof("[CLUSTER][%s][%s] No prediction metric %s found, send model jobs",
				dataGranularity, clusterName, lastPredictionMetric.GetMetricType().String())
			return
		} else {
			lastPrediction := lastPredictionMetric.GetData()[0]
			lastPredictionTime := lastPredictionMetric.GetData()[0].GetTime().GetSeconds()
			if lastPrediction != nil && lastPredictionTime <= nowSeconds {
				clusterInfo := sender.genClusterInfo(clusterName,
					datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
					datahub_common.MetricType_MEMORY_USAGE_BYTES)
				scope.Infof("[CLUSTER][%s][%s] send model job due to no predict found or predict is out of date, send model jobs",
					dataGranularity, clusterName)
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
				scope.Errorf("[CLUSTER][%s][%s] Get prediction for sending model job failed: %s",
					dataGranularity, clusterName, err.Error())
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
							AggregateFunction: datahub_common.TimeRange_AVG,
						},
					},
					ObjectMeta: []*datahub_resources.ObjectMeta{
						&datahub_resources.ObjectMeta{
							Name: clusterName,
						},
					},
				})
			if err != nil {
				scope.Errorf("[CLUSTER][%s][%s] List metric for sending model job failed: %s",
					dataGranularity, clusterName, err.Error())
				continue
			}
			clusterMetrics := clusterMetricsRes.GetClusterMetrics()
			for _, clusterPrediction := range clusterPredictions {
				predictRawData := clusterPrediction.GetPredictedRawData()
				for _, predictRawDatum := range predictRawData {

					for _, clusterMetric := range clusterMetrics {
						metricData := clusterMetric.GetMetricData()
						for _, metricDatum := range metricData {
							mData := metricDatum.GetData()
							pData := []*datahub_predictions.Sample{}
							if metricDatum.GetMetricType() == predictRawDatum.GetMetricType() {
								pData = append(pData, predictRawDatum.GetData()...)
								metricsNeedToModel, drift := DriftEvaluation(UnitTypeCluster, predictRawDatum.GetMetricType(), granularity, mData, pData, map[string]string{
									"clusterName":       clusterName,
									"targetDisplayName": fmt.Sprintf("[CLUSTER][%s][%s]", dataGranularity, clusterName),
								}, sender.metricExporter)
								if drift {
									scope.Infof("[CLUSTER][%s][%s] Export drift counter",
										dataGranularity, clusterName)
									sender.metricExporter.AddClusterMetricDrift(clusterName, queue.GetGranularityStr(granularity),
										time.Now().Unix(), 1.0)
								}
								clusterInfo.ModelMetrics = append(clusterInfo.ModelMetrics, metricsNeedToModel...)
							}
						}
					}
				}
			}
		}
	}
	isModeling := sender.modelMapper.IsModeling(pdUnit, dataGranularity, clusterInfo)
	if !isModeling || (isModeling && sender.modelMapper.IsModelTimeout(
		pdUnit, dataGranularity, clusterInfo)) {
		sender.sendJob(cluster, queueSender, pdUnit, granularity, clusterInfo)
	}
}
