package dispatcher

import (
	"github.com/containers-ai/alameda/ai-dispatcher/pkg/metrics"
	"github.com/containers-ai/alameda/ai-dispatcher/pkg/queue"
	datahub_gpu "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/gpu"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"google.golang.org/grpc"
)

type modelJobSender struct {
	datahubGrpcCn  *grpc.ClientConn
	modelMapper    *ModelMapper
	metricExporter *metrics.Exporter

	podModelJobSender         *podModelJobSender
	nodeModelJobSender        *nodeModelJobSender
	gpuModelJobSender         *gpuModelJobSender
	applicationModelJobSender *applicationModelJobSender
	namespaceModelJobSender   *namespaceModelJobSender
	clusterModelJobSender     *clusterModelJobSender
	controllerModelJobSender  *controllerModelJobSender
}

func NewModelJobSender(datahubGrpcCn *grpc.ClientConn, modelMapper *ModelMapper,
	metricExporter *metrics.Exporter) *modelJobSender {

	return &modelJobSender{
		datahubGrpcCn:  datahubGrpcCn,
		modelMapper:    modelMapper,
		metricExporter: metricExporter,

		podModelJobSender: NewPodModelJobSender(datahubGrpcCn, modelMapper,
			metricExporter),
		nodeModelJobSender: NewNodeModelJobSender(datahubGrpcCn, modelMapper,
			metricExporter),
		gpuModelJobSender: NewGPUModelJobSender(datahubGrpcCn, modelMapper,
			metricExporter),
		applicationModelJobSender: NewApplicationModelJobSender(datahubGrpcCn, modelMapper,
			metricExporter),
		namespaceModelJobSender: NewNamespaceModelJobSender(datahubGrpcCn, modelMapper,
			metricExporter),
		clusterModelJobSender: NewClusterModelJobSender(datahubGrpcCn, modelMapper,
			metricExporter),
		controllerModelJobSender: NewControllerModelJobSender(datahubGrpcCn, modelMapper,
			metricExporter),
	}
}

func (dispatcher *modelJobSender) SendNodeModelJobs(nodes []*datahub_resources.Node,
	queueSender queue.QueueSender, pdUnit string, granularity int64, predictionStep int64) {
	dispatcher.nodeModelJobSender.sendModelJobs(nodes, queueSender, pdUnit, granularity, predictionStep)
}

func (dispatcher *modelJobSender) SendPodModelJobs(pods []*datahub_resources.Pod, queueSender queue.QueueSender,
	pdUnit string, granularity int64, predictionStep int64) {
	dispatcher.podModelJobSender.sendModelJobs(pods, queueSender,
		pdUnit, granularity, predictionStep)
}

func (dispatcher *modelJobSender) SendGPUModelJobs(gpus []*datahub_gpu.Gpu,
	queueSender queue.QueueSender, pdUnit string, granularity int64, predictionStep int64) {
	dispatcher.gpuModelJobSender.sendModelJobs(gpus,
		queueSender, pdUnit, granularity, predictionStep)
}

func (dispatcher *modelJobSender) SendApplicationModelJobs(applications []*datahub_resources.Application,
	queueSender queue.QueueSender, pdUnit string, granularity int64, predictionStep int64) {
	dispatcher.applicationModelJobSender.sendModelJobs(applications,
		queueSender, pdUnit, granularity, predictionStep)
}

func (dispatcher *modelJobSender) SendNamespaceModelJobs(namespaces []*datahub_resources.Namespace,
	queueSender queue.QueueSender, pdUnit string, granularity int64, predictionStep int64) {
	dispatcher.namespaceModelJobSender.sendModelJobs(namespaces,
		queueSender, pdUnit, granularity, predictionStep)
}

func (dispatcher *modelJobSender) SendClusterModelJobs(clusters []*datahub_resources.Cluster,
	queueSender queue.QueueSender, pdUnit string, granularity int64, predictionStep int64) {
	dispatcher.clusterModelJobSender.sendModelJobs(clusters,
		queueSender, pdUnit, granularity, predictionStep)
}

func (dispatcher *modelJobSender) SendControllerModelJobs(controllers []*datahub_resources.Controller,
	queueSender queue.QueueSender, pdUnit string, granularity int64, predictionStep int64) {
	dispatcher.controllerModelJobSender.sendModelJobs(controllers,
		queueSender, pdUnit, granularity, predictionStep)
}
