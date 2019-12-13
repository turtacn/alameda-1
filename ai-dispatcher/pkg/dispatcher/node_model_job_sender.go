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
		go sender.sendNodeModelJobs(node, queueSender, pdUnit, granularity, predictionStep)
	}
}

func (sender *nodeModelJobSender) sendNodeModelJobs(node *datahub_resources.Node,
	queueSender queue.QueueSender, pdUnit string, granularity int64, predictionStep int64) {
	dataGranularity := queue.GetGranularityStr(granularity)
	datahubServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(sender.datahubGrpcCn)

	nodeName := node.GetObjectMeta().GetName()
	lastPredictionMetrics, err := sender.getLastMIdPrediction(datahubServiceClnt, node, granularity)
	if err != nil {
		scope.Infof("[NODE][%s][%s] Get last prediction failed: %s",
			dataGranularity, nodeName, err.Error())
		return
	}
	if lastPredictionMetrics == nil && err == nil {
		scope.Infof("[NODE][%s][%s] No prediction found",
			dataGranularity, nodeName)
	}

	sender.sendJobByMetrics(node, queueSender, pdUnit, granularity, predictionStep,
		datahubServiceClnt, lastPredictionMetrics)
}

func (sender *nodeModelJobSender) sendJob(node *datahub_resources.Node, queueSender queue.QueueSender, pdUnit string,
	granularity int64, nodeInfo *modelInfo) {
	nodeName := node.GetObjectMeta().GetName()
	dataGranularity := queue.GetGranularityStr(granularity)
	marshaler := jsonpb.Marshaler{}
	nodeStr, err := marshaler.MarshalToString(node)
	if err != nil {
		scope.Errorf("[NODE][%s][%s] Encode pb message failed. %s",
			dataGranularity, nodeName, err.Error())
		return
	}
	if len(nodeInfo.ModelMetrics) > 0 && nodeStr != "" {
		jb := queue.NewJobBuilder(pdUnit, granularity, nodeStr)
		jobJSONStr, err := jb.GetJobJSONString()
		if err != nil {
			scope.Errorf(
				"[NODE][%s][%s] Prepare model job payload failed. %s",
				dataGranularity, nodeName, err.Error())
			return
		}

		nodeJobStr := fmt.Sprintf("%s/%v", nodeName, granularity)
		scope.Infof("[NODE][%s][%s] Try to send node model job: %s", dataGranularity, nodeName, nodeJobStr)
		err = queueSender.SendJsonString(modelQueueName, jobJSONStr, nodeJobStr, granularity)
		if err == nil {
			sender.modelMapper.AddModelInfo(pdUnit, dataGranularity, nodeInfo)
		} else {
			scope.Errorf(
				"[NODE][%s][%s] Send model job payload failed. %s",
				dataGranularity, nodeName, err.Error())
		}
	}
}

func (sender *nodeModelJobSender) genNodeInfo(nodeName string,
	modelMetrics ...datahub_common.MetricType) *modelInfo {
	nodeInfo := new(modelInfo)
	nodeInfo.Name = nodeName
	nodeInfo.ModelMetrics = modelMetrics
	nodeInfo.SetTimeStamp(time.Now().Unix())
	return nodeInfo
}

func (sender *nodeModelJobSender) getLastMIdPrediction(datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	node *datahub_resources.Node, granularity int64) ([]*datahub_predictions.MetricData, error) {
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
		return nil, err
	}

	lastPid := ""
	if len(nodePredictRes.GetNodePredictions()) > 0 {
		lastNodePrediction := nodePredictRes.GetNodePredictions()[0]
		lnsPDRData := lastNodePrediction.GetPredictedRawData()
		if lnsPDRData == nil {
			lnsPDRData = lastNodePrediction.GetPredictedLowerboundData()
		}
		if lnsPDRData == nil {
			lnsPDRData = lastNodePrediction.GetPredictedUpperboundData()
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
		return nil, fmt.Errorf("[NODE][%s][%s] Query last prediction id failed",
			dataGranularity, nodeName)
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
			PredictionId: lastPid,
		})
	if err != nil {
		return nil, err
	}

	if len(nodePredictRes.GetNodePredictions()) > 0 {
		metricData := []*datahub_predictions.MetricData{}
		for _, nodePrediction := range nodePredictRes.GetNodePredictions() {
			for _, pdRD := range nodePrediction.GetPredictedRawData() {
				for _, pdD := range pdRD.GetData() {
					modelID := pdD.GetModelId()
					if modelID != "" {
						mIDNodePrediction, err := sender.getPredictionByMId(datahubServiceClnt, node, granularity, modelID)
						if err != nil {
							scope.Errorf("[NODE][%s][%s] Query prediction with model Id %s failed. %s",
								dataGranularity, nodeName, modelID, err.Error())
						}
						for _, mIDNodePD := range mIDNodePrediction {
							metricData = append(metricData, mIDNodePD.GetPredictedRawData()...)
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

func (sender *nodeModelJobSender) getPredictionByMId(datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	node *datahub_resources.Node, granularity int64, modelID string) ([]*datahub_predictions.NodePrediction, error) {
	nodePredictRes, err := datahubServiceClnt.ListNodePredictions(context.Background(),
		&datahub_predictions.ListNodePredictionsRequest{
			Granularity: granularity,
			ObjectMeta: []*datahub_resources.ObjectMeta{
				&datahub_resources.ObjectMeta{
					Name: node.GetObjectMeta().GetName(),
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
	return nodePredictRes.GetNodePredictions(), err
}

func (sender *nodeModelJobSender) getQueryMetricStartTime(descNodePredictions []*datahub_predictions.NodePrediction) int64 {
	if len(descNodePredictions) > 0 {
		pdMDs := descNodePredictions[len(descNodePredictions)-1].GetPredictedRawData()
		for _, pdMD := range pdMDs {
			mD := pdMD.GetData()
			if len(mD) > 0 {
				return mD[len(mD)-1].GetTime().GetSeconds()
			}
		}
	}
	return 0
}

func (sender *nodeModelJobSender) sendJobByMetrics(node *datahub_resources.Node, queueSender queue.QueueSender,
	pdUnit string, granularity int64, predictionStep int64, datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	lastPredictionMetrics []*datahub_predictions.MetricData) {
	nodeName := node.GetObjectMeta().GetName()
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
		nodeInfo := sender.genNodeInfo(nodeName,
			datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
			datahub_common.MetricType_MEMORY_USAGE_BYTES)
		sender.sendJob(node, queueSender, pdUnit, granularity, nodeInfo)
		scope.Infof("[NODE][%s][%s] No prediction metrics found, send model jobs",
			dataGranularity, nodeName)
		return
	}

	nodeInfo := sender.genNodeInfo(nodeName)
	for _, lastPredictionMetric := range lastPredictionMetrics {
		if len(lastPredictionMetric.GetData()) == 0 {
			nodeInfo := sender.genNodeInfo(nodeName,
				datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
				datahub_common.MetricType_MEMORY_USAGE_BYTES)
			sender.sendJob(node, queueSender, pdUnit, granularity, nodeInfo)
			scope.Infof("[NODE][%s][%s] No prediction metric %s found, send model jobs",
				dataGranularity, nodeName, lastPredictionMetric.GetMetricType().String())
			return
		} else {
			lastPrediction := lastPredictionMetric.GetData()[0]
			lastPredictionTime := lastPredictionMetric.GetData()[0].GetTime().GetSeconds()
			if lastPrediction != nil && lastPredictionTime <= nowSeconds {
				nodeInfo := sender.genNodeInfo(nodeName,
					datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
					datahub_common.MetricType_MEMORY_USAGE_BYTES)
				scope.Infof("[NODE][%s][%s] Send model job due to no predict found or predict is out of date",
					dataGranularity, nodeName)
				sender.sendJob(node, queueSender, pdUnit, granularity, nodeInfo)
				return
			}

			nodePredictRes, err := datahubServiceClnt.ListNodePredictions(context.Background(),
				&datahub_predictions.ListNodePredictionsRequest{
					ObjectMeta: []*datahub_resources.ObjectMeta{
						&datahub_resources.ObjectMeta{
							Name: nodeName,
						},
					},
					ModelId:        lastPrediction.GetModelId(),
					Granularity:    granularity,
					QueryCondition: queryCondition,
				})
			if err != nil {
				scope.Errorf("[NODE][%s][%s] Get prediction for sending model job failed: %s",
					dataGranularity, nodeName, err.Error())
				continue
			}
			nodePredictions := nodePredictRes.GetNodePredictions()
			queryStartTime := time.Now().Unix() - predictionStep*granularity
			firstPDTime := sender.getQueryMetricStartTime(nodePredictions)
			if firstPDTime > 0 {
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
				})
			if err != nil {
				scope.Errorf("[NODE][%s][%s] List metric for sending model job failed: %s",
					dataGranularity, nodeName, err.Error())
				continue
			}
			nodeMetrics := nodeMetricsRes.GetNodeMetrics()
			for _, nodePrediction := range nodePredictions {
				predictRawData := nodePrediction.GetPredictedRawData()
				for _, predictRawDatum := range predictRawData {
					for _, nodeMetric := range nodeMetrics {
						metricData := nodeMetric.GetMetricData()
						for _, metricDatum := range metricData {
							mData := metricDatum.GetData()
							pData := []*datahub_predictions.Sample{}

							if metricDatum.GetMetricType() == predictRawDatum.GetMetricType() {
								pData = append(pData, predictRawDatum.GetData()...)
								metricsNeedToModel, drift := DriftEvaluation(UnitTypeNode, predictRawDatum.GetMetricType(), granularity, mData, pData, map[string]string{
									"nodeName":          nodeName,
									"targetDisplayName": fmt.Sprintf("[NODE][%s][%s]", dataGranularity, nodeName),
								}, sender.metricExporter)
								if drift {
									scope.Infof("[NODE][%s][%s] Export drift counter",
										dataGranularity, nodeName)
									sender.metricExporter.AddNodeMetricDrift(nodeName, queue.GetGranularityStr(granularity),
										time.Now().Unix(), 1.0)
								}
								nodeInfo.ModelMetrics = append(nodeInfo.ModelMetrics, metricsNeedToModel...)
							}
						}
					}

				}
			}
		}
	}
	isModeling := sender.modelMapper.IsModeling(pdUnit, dataGranularity, nodeInfo)
	if !isModeling || (isModeling && sender.modelMapper.IsModelTimeout(
		pdUnit, dataGranularity, nodeInfo)) {
		sender.sendJob(node, queueSender, pdUnit, granularity, nodeInfo)
	}
}
