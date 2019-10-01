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

type nodeModelJobSender struct {
	datahubGrpcCn  *grpc.ClientConn
	modelThreshold float64
	modelMapper    *ModelMapper
	metricExporter *metrics.Exporter
}

func NewNodeModelJobSender(datahubGrpcCn *grpc.ClientConn, modelMapper *ModelMapper,
	metricExporter *metrics.Exporter) *nodeModelJobSender {
	return &nodeModelJobSender{
		datahubGrpcCn:  datahubGrpcCn,
		modelThreshold: viper.GetFloat64("model.threshold"),
		modelMapper:    modelMapper,
		metricExporter: metricExporter,
	}
}

func (sender *nodeModelJobSender) sendModelJobs(nodes []*datahub_v1alpha1.Node,
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
	for _, node := range nodes {
		nodeName := node.GetName()
		shouldDrift := false

		lastPrediction, lastPredictionTime, err := sender.getLastPrediction(datahubServiceClnt, node, granularity)
		if err != nil {
			scope.Infof("Get node %s last prediction failed: %s",
				nodeName, err.Error())
			continue
		}
		if lastPrediction == nil && err == nil {
			scope.Infof("No prediction found of node %s",
				nodeName)
		}
		nowSeconds := time.Now().Unix()
		if lastPrediction != nil && lastPredictionTime <= nowSeconds {
			scope.Infof("node prediction %s is out of date due to last predict time is %v (current: %v)",
				nodeName, lastPredictionTime, nowSeconds)
		}
		if (lastPrediction == nil && err == nil) || (lastPrediction != nil && lastPredictionTime <= nowSeconds) {
			nodeInfo := sender.genNodeInfo(nodeName)
			nodeInfo.ModelMetrics = []datahub_v1alpha1.MetricType{
				datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
				datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES,
			}
			scope.Infof("send node %s model job due to no predict found or predict is out of date",
				nodeName)
			sender.sendJob(node, queueSender, pdUnit, granularity, nodeInfo)
		}

		nodePredictRes, err := datahubServiceClnt.ListNodePredictions(context.Background(),
			&datahub_v1alpha1.ListNodePredictionsRequest{
				NodeNames:      []string{nodeName},
				ModelId:        lastPrediction.GetModelId(),
				Granularity:    granularity,
				QueryCondition: queryCondition,
			})
		if err != nil {
			scope.Errorf("Get node %s Prediction with granularity %v for sending model job failed: %s",
				nodeName, granularity, err.Error())
			continue
		}
		nodePredictions := nodePredictRes.GetNodePredictions()
		queryStartTime := time.Now().Unix() - predictionStep*granularity
		firstPDTime := sender.getQueryMetricStartTime(nodePredictions)
		if firstPDTime > 0 {
			queryStartTime = firstPDTime
		}
		nodeMetricsRes, err := datahubServiceClnt.ListNodeMetrics(context.Background(),
			&datahub_v1alpha1.ListNodeMetricsRequest{
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
				NodeNames: []string{nodeName},
			})
		if err != nil {
			scope.Errorf("List nodes %s metric with granularity %v for sending model job failed: %s",
				nodeName, granularity, err.Error())
			continue
		}
		nodeMetrics := nodeMetricsRes.GetNodeMetrics()

		for _, nodeMetric := range nodeMetrics {
			metricData := nodeMetric.GetMetricData()
			for _, metricDatum := range metricData {
				mData := metricDatum.GetData()
				pData := []*datahub_v1alpha1.Sample{}
				nodeInfo := sender.genNodeInfo(nodeName)
				for _, nodePrediction := range nodePredictions {
					predictRawData := nodePrediction.GetPredictedRawData()
					for _, predictRawDatum := range predictRawData {
						if metricDatum.GetMetricType() == predictRawDatum.GetMetricType() {
							pData = append(pData, predictRawDatum.GetData()...)
						}
					}
				}
				if len(pData) > 0 {
					metricType := metricDatum.GetMetricType()
					scope.Infof("start MAPE calculation for node %s metric %v with granularity %v",
						nodeName, metricType, granularity)
					measurementDataSet := stats.NewMeasurementDataSet(mData, pData, granularity)
					mape, mapeErr := stats.MAPE(measurementDataSet)
					if mapeErr == nil {
						scope.Infof("export MAPE value %v for node %s metric %v with granularity %v", mape,
							nodeName, metricType, granularity)
						sender.metricExporter.SetNodeMetricMAPE(nodeName,
							queue.GetMetricLabel(metricDatum.GetMetricType()), queue.GetGranularityStr(granularity), mape)
					}
					if mapeErr != nil {
						nodeInfo.ModelMetrics = append(nodeInfo.ModelMetrics, metricType)
						scope.Infof(
							"model job for node %s metric %v with granularity %v should be sent due to MAPE calculation failed: %s",
							nodeName, metricType, granularity, mapeErr.Error())
					} else if mape > sender.modelThreshold {
						nodeInfo.ModelMetrics = append(nodeInfo.ModelMetrics, metricType)
						scope.Infof("model job node %s metric %v with granularity %v should be sent due to MAPE %v > %v",
							nodeName, metricType, granularity, mape, sender.modelThreshold)
						shouldDrift = true
					} else {
						scope.Infof("node %s metric %v with granularity %v MAPE %v <= %v, skip sending this model metric",
							nodeName, metricType, granularity, mape, sender.modelThreshold)
					}
				}
				isModeling := sender.modelMapper.IsModeling(pdUnit, dataGranularity, nodeInfo)
				if !isModeling || (isModeling && sender.modelMapper.IsModelTimeout(
					pdUnit, dataGranularity, nodeInfo)) {
					sender.sendJob(node, queueSender, pdUnit, granularity, nodeInfo)
				}
			}
		}
		if shouldDrift {
			scope.Infof("export node %s drift counter with granularity %s",
				nodeName, dataGranularity)
			sender.metricExporter.AddNodeMetricDrift(nodeName, queue.GetGranularityStr(granularity), 1.0)
		}
	}
}

func (sender *nodeModelJobSender) sendJob(node *datahub_v1alpha1.Node, queueSender queue.QueueSender, pdUnit string,
	granularity int64, nodeInfo *modelInfo) {
	nodeName := node.GetName()
	dataGranularity := queue.GetGranularityStr(granularity)
	marshaler := jsonpb.Marshaler{}
	nodeStr, err := marshaler.MarshalToString(node)
	if err != nil {
		scope.Errorf("Encode pb message failed for node %s with granularity seconds %v. %s",
			node.GetName(), granularity, err.Error())
		return
	}
	if len(nodeInfo.ModelMetrics) > 0 && nodeStr != "" {
		jb := queue.NewJobBuilder(pdUnit, granularity, nodeStr)
		jobJSONStr, err := jb.GetJobJSONString()
		if err != nil {
			scope.Errorf(
				"Prepare model job payload failed for node %s with granularity seconds %v. %s",
				nodeName, granularity, err.Error())
			return
		}

		err = queueSender.SendJsonString(modelQueueName, jobJSONStr,
			fmt.Sprintf("%s", nodeName))
		if err == nil {
			sender.modelMapper.AddModelInfo(pdUnit, dataGranularity, nodeInfo)
		} else {
			scope.Errorf(
				"Send model job payload failed for node %s with granularity seconds %v. %s",
				nodeName, granularity, err.Error())
		}
	}
}

func (sender *nodeModelJobSender) genNodeInfo(nodeName string) *modelInfo {
	nodeInfo := new(modelInfo)
	nodeInfo.Name = nodeName
	nodeInfo.ModelMetrics = []datahub_v1alpha1.MetricType{}
	nodeInfo.SetTimeStamp(time.Now().Unix())
	return nodeInfo
}

func (sender *nodeModelJobSender) getLastPrediction(datahubServiceClnt datahub_v1alpha1.DatahubServiceClient,
	node *datahub_v1alpha1.Node, granularity int64) (*datahub_v1alpha1.NodePrediction, int64, error) {
	nodeName := node.GetName()
	nodePredictRes, err := datahubServiceClnt.ListNodePredictions(context.Background(),
		&datahub_v1alpha1.ListNodePredictionsRequest{
			NodeNames:   []string{nodeName},
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
		return nil, 0, err
	}
	if len(nodePredictRes.GetNodePredictions()) > 0 {
		lastNodePrediction := nodePredictRes.GetNodePredictions()[0]
		for _, metricPd := range lastNodePrediction.GetPredictedRawData() {
			for _, metricPdSample := range metricPd.GetData() {
				return lastNodePrediction, metricPdSample.GetTime().GetSeconds(), nil
			}
		}
	}
	return nil, 0, nil
}

func (sender *nodeModelJobSender) getQueryMetricStartTime(descNodePredictions []*datahub_v1alpha1.NodePrediction) int64 {
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
