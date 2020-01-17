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

type nodeModelJobSender struct {
	datahubGrpcCn  *grpc.ClientConn
	modelMapper    *ModelMapper
	metricExporter *metrics.Exporter
}

func NewNodeModelJobSender(datahubGrpcCn *grpc.ClientConn, modelMapper *ModelMapper,
	metricExporter *metrics.Exporter) *nodeModelJobSender {
	return &nodeModelJobSender{
		datahubGrpcCn:  datahubGrpcCn,
		modelMapper:    modelMapper,
		metricExporter: metricExporter,
	}
}

func (sender *nodeModelJobSender) sendModelJobs(nodes []*datahub_resources.Node,
	queueSender queue.QueueSender, pdUnit string, granularity int64, predictionStep int64) {
	for _, node := range nodes {
		sender.sendNodeModelJobs(node, queueSender, pdUnit, granularity, predictionStep, &wg)
	}
}

func (sender *nodeModelJobSender) sendNodeModelJobs(node *datahub_resources.Node,
	queueSender queue.QueueSender, pdUnit string, granularity int64, predictionStep int64, wg *sync.WaitGroup) {
	dataGranularity := queue.GetGranularityStr(granularity)
	datahubServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(sender.datahubGrpcCn)

	nodeName := node.GetObjectMeta().GetName()
	lastPredictionMetrics, err := sender.getLastMIdPrediction(datahubServiceClnt, node, granularity)
	if err != nil {
		scope.Infof("[NODE][%s][%s] Get last prediction failed: %s",
			dataGranularity, nodeName, err.Error())
		return
	}

	sender.sendJobByMetrics(node, queueSender, pdUnit, granularity, predictionStep,
		datahubServiceClnt, lastPredictionMetrics)
}

func (sender *nodeModelJobSender) sendJob(node *datahub_resources.Node, queueSender queue.QueueSender, pdUnit string,
	granularity int64, metricType datahub_common.MetricType) {
	clusterID := node.GetObjectMeta().GetClusterName()
	nodeName := node.GetObjectMeta().GetName()
	dataGranularity := queue.GetGranularityStr(granularity)
	marshaler := jsonpb.Marshaler{}
	nodeStr, err := marshaler.MarshalToString(node)
	if err != nil {
		scope.Errorf("[NODE][%s][%s] Encode pb message failed. %s",
			dataGranularity, nodeName, err.Error())
		return
	}

	jb := queue.NewJobBuilder(clusterID, pdUnit, granularity, metricType, nodeStr, nil)
	jobJSONStr, err := jb.GetJobJSONString()
	if err != nil {
		scope.Errorf(
			"[NODE][%s][%s] Prepare model job payload failed. %s",
			dataGranularity, nodeName, err.Error())
		return
	}

	nodeJobStr := fmt.Sprintf("%s/%s/%s/%v/%s", consts.UnitTypeNode, clusterID, nodeName, granularity, metricType)
	scope.Infof("[NODE][%s][%s] Try to send node model job: %s", dataGranularity, nodeName, nodeJobStr)
	err = queueSender.SendJsonString(modelQueueName, jobJSONStr, nodeJobStr, granularity)
	if err == nil {
		sender.modelMapper.AddModelInfo(clusterID, pdUnit, dataGranularity, metricType.String(), map[string]string{
			"name": nodeName,
		})
	} else {
		scope.Errorf(
			"[NODE][%s][%s] Send model job payload failed. %s",
			dataGranularity, nodeName, err.Error())
	}
}

func (sender *nodeModelJobSender) getLastMIdPrediction(datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	node *datahub_resources.Node, granularity int64) ([]*datahub_predictions.MetricData, error) {

	metricData := []*datahub_predictions.MetricData{}
	dataGranularity := queue.GetGranularityStr(granularity)
	nodeName := node.GetObjectMeta().GetName()
	nodePredictRes, err := datahubServiceClnt.ListNodePredictions(context.Background(),
		&datahub_predictions.ListNodePredictionsRequest{
			ObjectMeta: []*datahub_resources.ObjectMeta{
				&datahub_resources.ObjectMeta{
					Name: nodeName,
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
	if len(nodePredictRes.GetNodePredictions()) == 0 {
		return []*datahub_predictions.MetricData{}, nil
	}

	lastNodePrediction := nodePredictRes.GetNodePredictions()[0]
	lnsPDRData := lastNodePrediction.GetPredictedRawData()
	if lnsPDRData == nil {
		return metricData, nil
	}

	for _, pdRD := range lnsPDRData {
		for _, theData := range pdRD.GetData() {
			lastMid = theData.GetModelId()
			break
		}
		if lastMid == "" {
			scope.Warnf("[NODE][%s][%s] Query last model id for metric %s is empty",
				dataGranularity, nodeName, pdRD.GetMetricType())
		}

		nodePredictRes, err = datahubServiceClnt.ListNodePredictions(context.Background(),
			&datahub_predictions.ListNodePredictionsRequest{
				ObjectMeta: []*datahub_resources.ObjectMeta{
					&datahub_resources.ObjectMeta{
						Name: nodeName,
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
			scope.Errorf("[NODE][%s][%s] Query last model id %v for metric %s failed",
				dataGranularity, nodeName, lastMid, pdRD.GetMetricType())
			continue
		}

		for _, nodePrediction := range nodePredictRes.GetNodePredictions() {
			for _, lMIDPdRD := range nodePrediction.GetPredictedRawData() {
				if lMIDPdRD.GetMetricType() == pdRD.GetMetricType() {
					metricData = append(metricData, lMIDPdRD)
				}
			}
		}
	}
	return metricData, nil
}

func (sender *nodeModelJobSender) getQueryMetricStartTime(metricData *datahub_predictions.MetricData) int64 {
	mD := metricData.GetData()
	if len(mD) > 0 {
		return mD[len(mD)-1].GetTime().GetSeconds()
	}
	return 0
}

func (sender *nodeModelJobSender) sendJobByMetrics(node *datahub_resources.Node, queueSender queue.QueueSender,
	pdUnit string, granularity int64, predictionStep int64, datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	lastPredictionMetrics []*datahub_predictions.MetricData) {

	clusterID := node.GetObjectMeta().GetClusterName()
	nodeName := node.GetObjectMeta().GetName()
	dataGranularity := queue.GetGranularityStr(granularity)
	nowSeconds := time.Now().Unix()

	if len(lastPredictionMetrics) == 0 {
		scope.Infof("[NODE][%s][%s] No prediction metrics found, send model jobs",
			dataGranularity, nodeName)
		for _, metricType := range []datahub_common.MetricType{
			datahub_common.MetricType_MEMORY_USAGE_BYTES,
			datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
		} {
			sender.sendJob(node, queueSender, pdUnit, granularity, metricType)
		}
		return
	}

	for _, lastPredictionMetric := range lastPredictionMetrics {
		if len(lastPredictionMetric.GetData()) == 0 {
			scope.Infof("[NODE][%s][%s] No prediction metric %s found, send model jobs",
				dataGranularity, nodeName, lastPredictionMetric.GetMetricType().String())
			sender.sendJob(node, queueSender, pdUnit, granularity, lastPredictionMetric.GetMetricType())
			continue
		} else {
			lastPrediction := lastPredictionMetric.GetData()[0]
			lastPredictionTime := lastPredictionMetric.GetData()[0].GetTime().GetSeconds()
			if lastPrediction != nil && lastPredictionTime <= nowSeconds {
				scope.Infof("[NODE][%s][%s] Send model job due to no predict metric %s found or is out of date",
					dataGranularity, nodeName, lastPredictionMetric.GetMetricType().String())
				sender.sendJob(node, queueSender, pdUnit, granularity, lastPredictionMetric.GetMetricType())
				continue
			}

			queryStartTime := time.Now().Unix() - predictionStep*granularity
			firstPDTime := sender.getQueryMetricStartTime(lastPredictionMetric)
			if firstPDTime > 0 && firstPDTime <= time.Now().Unix() {
				queryStartTime = firstPDTime
			}
			nodeMetricsRes, err := datahubServiceClnt.ListNodeMetrics(context.Background(),
				&datahub_metrics.ListNodeMetricsRequest{
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
							Name: nodeName,
						},
					},
					MetricTypes: []datahub_common.MetricType{
						lastPredictionMetric.GetMetricType(),
					},
				})
			if err != nil {
				scope.Errorf("[NODE][%s][%s] List metric for sending model job failed: %s",
					dataGranularity, nodeName, err.Error())
				continue
			}
			nodeMetrics := nodeMetricsRes.GetNodeMetrics()
			predictRawData := lastPredictionMetrics
			for _, predictRawDatum := range predictRawData {
				for _, nodeMetric := range nodeMetrics {
					metricData := nodeMetric.GetMetricData()
					for _, metricDatum := range metricData {
						mData := metricDatum.GetData()
						pData := []*datahub_predictions.Sample{}

						if metricDatum.GetMetricType() == predictRawDatum.GetMetricType() {
							pData = append(pData, predictRawDatum.GetData()...)
							metricsNeedToModel, drift := DriftEvaluation(consts.UnitTypeNode, predictRawDatum.GetMetricType(), granularity, mData, pData, map[string]string{
								"clusterID":         clusterID,
								"nodeName":          nodeName,
								"targetDisplayName": fmt.Sprintf("[NODE][%s][%s]", dataGranularity, nodeName),
							}, sender.metricExporter)

							for _, mntm := range metricsNeedToModel {
								if drift {
									scope.Infof("[NODE][%s][%s] Export metric %s drift counter",
										dataGranularity, nodeName, mntm)
									sender.metricExporter.AddNodeMetricDrift(clusterID, nodeName, queue.GetGranularityStr(granularity),
										mntm.String(), time.Now().Unix(), 1.0)
								}
								isModeling := sender.modelMapper.IsModeling(clusterID, pdUnit, dataGranularity, mntm.String(), map[string]string{
									"name": nodeName,
								})
								if !isModeling || (isModeling && sender.modelMapper.IsModelTimeout(
									clusterID, pdUnit, dataGranularity, mntm.String(), map[string]string{
										"name": nodeName,
									})) {
									sender.sendJob(node, queueSender, pdUnit, granularity, mntm)
								}
							}
						}
					}
				}
			}
		}
	}
}
