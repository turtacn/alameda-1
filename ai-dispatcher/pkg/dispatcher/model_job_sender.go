package dispatcher

import (
	"github.com/containers-ai/alameda/ai-dispatcher/pkg/metrics"
	"github.com/containers-ai/alameda/ai-dispatcher/pkg/queue"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

type modelJobSender struct {
	datahubGrpcCn      *grpc.ClientConn
	modelThreshold     float64
	modelMapper        *ModelMapper
	metricExporter     *metrics.Exporter
	podModelJobSender  *podModelJobSender
	nodeModelJobSender *nodeModelJobSender
	gpuModelJobSender  *gpuModelJobSender
}

func NewModelJobSender(datahubGrpcCn *grpc.ClientConn, modelMapper *ModelMapper,
	metricExporter *metrics.Exporter) *modelJobSender {
	return &modelJobSender{
		datahubGrpcCn:  datahubGrpcCn,
		modelThreshold: viper.GetFloat64("model.threshold"),
		modelMapper:    modelMapper,
		metricExporter: metricExporter,
		podModelJobSender: NewPodModelJobSender(datahubGrpcCn, modelMapper,
			metricExporter),
		nodeModelJobSender: NewNodeModelJobSender(datahubGrpcCn, modelMapper,
			metricExporter),
		gpuModelJobSender: NewGPUModelJobSender(datahubGrpcCn, modelMapper,
			metricExporter),
	}
}

func (dispatcher *modelJobSender) SendNodeModelJobs(nodes []*datahub_v1alpha1.Node,
	queueSender queue.QueueSender, pdUnit string, granularity int64, predictionStep int64) {
	dispatcher.nodeModelJobSender.sendModelJobs(nodes, queueSender, pdUnit, granularity, predictionStep)
}

func (dispatcher *modelJobSender) SendPodModelJobs(pods []*datahub_v1alpha1.Pod, queueSender queue.QueueSender,
	pdUnit string, granularity int64, predictionStep int64) {
	dispatcher.podModelJobSender.sendModelJobs(pods, queueSender,
		pdUnit, granularity, predictionStep)
}

func (dispatcher *modelJobSender) SendGPUModelJobs(gpus []*datahub_v1alpha1.Gpu,
	queueSender queue.QueueSender, pdUnit string, granularity int64, predictionStep int64) {
	dispatcher.gpuModelJobSender.sendModelJobs(gpus,
		queueSender, pdUnit, granularity, predictionStep)
}
