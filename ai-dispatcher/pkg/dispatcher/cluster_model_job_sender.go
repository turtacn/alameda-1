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
		sender.sendClusterModelJobs(cluster, queueSender, pdUnit, granularity, predictionStep, &wg)
	}
}

func (sender *clusterModelJobSender) sendClusterModelJobs(cluster *datahub_resources.Cluster,
	queueSender queue.QueueSender, pdUnit string, granularity int64, predictionStep int64, wg *sync.WaitGroup) {
	dataGranularity := queue.GetGranularityStr(granularity)
	datahubServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(sender.datahubGrpcCn)
	clusterName := cluster.GetObjectMeta().GetName()
	lastPredictionMetrics, err := sender.getLastMIdPrediction(datahubServiceClnt, cluster, granularity)
	if err != nil {
		scope.Infof("[CLUSTER][%s][%s] Get last prediction failed: %s",
			dataGranularity, clusterName, err.Error())
		return
	}

	sender.sendJobByMetrics(cluster, queueSender, pdUnit, granularity, predictionStep,
		datahubServiceClnt, lastPredictionMetrics)
}

func (sender *clusterModelJobSender) sendJob(cluster *datahub_resources.Cluster,
	queueSender queue.QueueSender, pdUnit string, granularity int64,
	metricType datahub_common.MetricType) {
	clusterID := cluster.GetObjectMeta().GetClusterName()
	clusterName := cluster.GetObjectMeta().GetName()
	dataGranularity := queue.GetGranularityStr(granularity)
	marshaler := jsonpb.Marshaler{}
	clusterStr, err := marshaler.MarshalToString(cluster)
	if err != nil {
		scope.Errorf("[CLUSTER][%s][%s] Encode pb message failed. %s",
			dataGranularity, clusterName, err.Error())
		return
	}

	jb := queue.NewJobBuilder(clusterID, pdUnit, granularity, metricType, clusterStr, nil)
	jobJSONStr, err := jb.GetJobJSONString()
	if err != nil {
		scope.Errorf(
			"[CLUSTER][%s][%s] Prepare model job payload failed. %s",
			dataGranularity, clusterName, err.Error())
		return
	}

	clusterJobStr := fmt.Sprintf("%s/%s/%v/%s", consts.UnitTypeCluster, clusterName, granularity, metricType)
	scope.Infof("[CLUSTER][%s][%s] Try to send cluster model job: %s", dataGranularity, clusterName, clusterJobStr)
	err = queueSender.SendJsonString(modelQueueName, jobJSONStr, clusterJobStr, granularity)
	if err == nil {
		sender.modelMapper.AddModelInfo(clusterID, pdUnit, dataGranularity, metricType.String(), map[string]string{
			"name": clusterName,
		})
	} else {
		scope.Errorf(
			"[CLUSTER][%s][%s] Send model job payload failed. %s",
			dataGranularity, clusterName, err.Error())
	}
}

func (sender *clusterModelJobSender) getLastMIdPrediction(datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	cluster *datahub_resources.Cluster, granularity int64) ([]*datahub_predictions.MetricData, error) {

	metricData := []*datahub_predictions.MetricData{}
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
		return metricData, err
	}
	lastMid := ""

	if len(clusterPredictRes.GetClusterPredictions()) == 0 {
		return []*datahub_predictions.MetricData{}, nil
	}
	lastClusterPrediction := clusterPredictRes.GetClusterPredictions()[0]
	lnsPDRData := lastClusterPrediction.GetPredictedRawData()
	if lnsPDRData == nil {
		return metricData, nil
	}
	for _, pdRD := range lnsPDRData {
		for _, theData := range pdRD.GetData() {
			lastMid = theData.GetModelId()
			break
		}
		if lastMid == "" {
			scope.Warnf("[CLUSTER][%s][%s] Query last model id for metric %s is empty",
				dataGranularity, clusterName, pdRD.GetMetricType())
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
				ModelId: lastMid,
			})
		if err != nil {
			scope.Warnf("[CLUSTER][%s][%s] Query last model id %v for metric %s is empty",
				dataGranularity, clusterName, lastMid, pdRD.GetMetricType())
			continue
		}

		for _, clusterPrediction := range clusterPredictRes.GetClusterPredictions() {
			for _, lMIDPdRD := range clusterPrediction.GetPredictedRawData() {
				if lMIDPdRD.GetMetricType() == pdRD.GetMetricType() {
					metricData = append(metricData, lMIDPdRD)
				}
			}
		}
	}
	return metricData, nil
}

func (sender *clusterModelJobSender) getQueryMetricStartTime(metricData *datahub_predictions.MetricData) int64 {
	mD := metricData.GetData()
	if len(mD) > 0 {
		return mD[len(mD)-1].GetTime().GetSeconds()
	}
	return 0
}

func (sender *clusterModelJobSender) sendJobByMetrics(cluster *datahub_resources.Cluster, queueSender queue.QueueSender,
	pdUnit string, granularity int64, predictionStep int64, datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	lastPredictionMetrics []*datahub_predictions.MetricData) {
	clusterName := cluster.GetObjectMeta().GetName()
	clusterID := cluster.GetObjectMeta().GetClusterName()
	dataGranularity := queue.GetGranularityStr(granularity)
	nowSeconds := time.Now().Unix()

	if len(lastPredictionMetrics) == 0 {
		scope.Infof("[CLUSTER][%s][%s] No prediction metrics found, send model jobs",
			dataGranularity, clusterName)
		for _, metricType := range []datahub_common.MetricType{
			datahub_common.MetricType_MEMORY_USAGE_BYTES,
			datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
		} {
			sender.sendJob(cluster, queueSender, pdUnit, granularity, metricType)
		}
		return
	}

	for _, lastPredictionMetric := range lastPredictionMetrics {
		if len(lastPredictionMetric.GetData()) == 0 {
			scope.Infof("[CLUSTER][%s][%s] No prediction metric %s found, send model jobs",
				dataGranularity, clusterName, lastPredictionMetric.GetMetricType().String())
			sender.sendJob(cluster, queueSender, pdUnit, granularity, lastPredictionMetric.GetMetricType())
			continue
		} else {
			lastPrediction := lastPredictionMetric.GetData()[0]
			lastPredictionTime := lastPredictionMetric.GetData()[0].GetTime().GetSeconds()
			if lastPrediction != nil && lastPredictionTime <= nowSeconds {
				scope.Infof("[CLUSTER][%s][%s] send model job due to no predict metric %s found or is out of date, send model jobs",
					dataGranularity, clusterName, lastPredictionMetric.GetMetricType().String())
				sender.sendJob(cluster, queueSender, pdUnit, granularity, lastPredictionMetric.GetMetricType())
				continue
			}

			queryStartTime := time.Now().Unix() - predictionStep*granularity
			firstPDTime := sender.getQueryMetricStartTime(lastPredictionMetric)
			if firstPDTime > 0 && firstPDTime <= time.Now().Unix() {
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
					MetricTypes: []datahub_common.MetricType{
						lastPredictionMetric.GetMetricType(),
					},
				})
			if err != nil {
				scope.Errorf("[CLUSTER][%s][%s] List metric for sending model job failed: %s",
					dataGranularity, clusterName, err.Error())
				continue
			}
			clusterMetrics := clusterMetricsRes.GetClusterMetrics()
			predictRawData := lastPredictionMetrics
			for _, predictRawDatum := range predictRawData {
				for _, clusterMetric := range clusterMetrics {
					metricData := clusterMetric.GetMetricData()
					for _, metricDatum := range metricData {
						mData := metricDatum.GetData()
						pData := []*datahub_predictions.Sample{}
						if metricDatum.GetMetricType() == predictRawDatum.GetMetricType() {
							pData = append(pData, predictRawDatum.GetData()...)
							metricsNeedToModel, drift := DriftEvaluation(consts.UnitTypeCluster, predictRawDatum.GetMetricType(), granularity, mData, pData, map[string]string{
								"clusterName":       clusterName,
								"targetDisplayName": fmt.Sprintf("[CLUSTER][%s][%s]", dataGranularity, clusterName),
							}, sender.metricExporter)

							for _, mntm := range metricsNeedToModel {
								if drift {
									scope.Infof("[CLUSTER][%s][%s] Export metric %s drift counter",
										dataGranularity, clusterName, mntm)
									sender.metricExporter.AddClusterMetricDrift(clusterName, queue.GetGranularityStr(granularity), mntm.String(),
										time.Now().Unix(), 1.0)
								}
								isModeling := sender.modelMapper.IsModeling(clusterID, pdUnit, dataGranularity, mntm.String(), map[string]string{
									"name": clusterName,
								})
								if !isModeling || (isModeling && sender.modelMapper.IsModelTimeout(
									clusterID, pdUnit, dataGranularity, mntm.String(), map[string]string{
										"name": clusterName,
									})) {
									sender.sendJob(cluster, queueSender, pdUnit, granularity, mntm)
								}
							}
						}
					}
				}
			}
		}
	}
}
