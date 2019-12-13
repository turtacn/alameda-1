package dispatcher

import (
	"fmt"

	"github.com/containers-ai/alameda/ai-dispatcher/pkg/queue"
	datahub_gpu "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/gpu"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"github.com/golang/protobuf/jsonpb"
	"google.golang.org/grpc"
)

type predictJobSender struct {
	datahubGrpcCn *grpc.ClientConn
}

func NewPredictJobSender(datahubGrpcCn *grpc.ClientConn) *predictJobSender {
	return &predictJobSender{
		datahubGrpcCn: datahubGrpcCn,
	}
}

func (dispatcher *predictJobSender) SendNodePredictJobs(nodes []*datahub_resources.Node,
	queueSender queue.QueueSender, pdUnit string, granularity int64) {
	dataGranularity := queue.GetGranularityStr(granularity)
	marshaler := jsonpb.Marshaler{}
	for _, node := range nodes {
		nodeStr, err := marshaler.MarshalToString(node)
		nodeName := node.ObjectMeta.GetName()
		if err != nil {
			scope.Errorf("[NODE][%s][%s] Encode pb message failed. %s",
				dataGranularity, nodeName, err.Error())
			continue
		}
		jb := queue.NewJobBuilder(pdUnit, granularity, nodeStr)
		jobJSONStr, err := jb.GetJobJSONString()
		if err != nil {
			scope.Errorf("[NODE][%s][%s] Prepare predict job payload failed. %s",
				dataGranularity, nodeName, err.Error())
			continue
		}

		nodeJobStr := fmt.Sprintf("%s/%v", nodeName, granularity)
		scope.Infof("[NODE][%s][%s] Try to send predict job: %s", dataGranularity, nodeName, nodeJobStr)
		err = queueSender.SendJsonString(queueName, jobJSONStr, nodeJobStr, granularity)
		if err != nil {
			scope.Errorf("[NODE][%s][%s] Send predict job failed. %s",
				dataGranularity, nodeName, err.Error())
		}
	}
}

func (dispatcher *predictJobSender) SendPodPredictJobs(pods []*datahub_resources.Pod,
	queueSender queue.QueueSender, pdUnit string, granularity int64) {
	dataGranularity := queue.GetGranularityStr(granularity)
	marshaler := jsonpb.Marshaler{}
	for _, pod := range pods {
		podNS := pod.ObjectMeta.GetNamespace()
		podName := pod.ObjectMeta.GetName()
		podStr, err := marshaler.MarshalToString(pod)
		if err != nil {
			scope.Errorf("[POD][%s][%s/%s] Encode pb message failed. %s",
				dataGranularity, podNS, podName, err.Error())
			continue
		}
		jb := queue.NewJobBuilder(pdUnit, granularity, podStr)
		jobJSONStr, err := jb.GetJobJSONString()
		if err != nil {
			scope.Errorf("[POD][%s][%s/%s] Prepare predict job payload failed. %s",
				dataGranularity, podNS, podName, err.Error())
			continue
		}

		podJobStr := fmt.Sprintf("%s/%s/%v", podNS, podName, granularity)
		scope.Infof("[POD][%s][%s/%s] Try to send predict job: %s",
			dataGranularity, podNS, podName, podJobStr)
		err = queueSender.SendJsonString(queueName, jobJSONStr, podJobStr, granularity)
		if err != nil {
			scope.Errorf("[POD][%s][%s/%s] Send predict job failed. %s",
				dataGranularity, podNS, podName, err.Error())
		}
	}
}

func (dispatcher *predictJobSender) SendGPUPredictJobs(gpus []*datahub_gpu.Gpu,
	queueSender queue.QueueSender, pdUnit string, granularity int64) {
	dataGranularity := queue.GetGranularityStr(granularity)
	marshaler := jsonpb.Marshaler{}
	for _, gpu := range gpus {
		gpuHost := gpu.GetMetadata().GetHost()
		gpuMinorNumber := gpu.GetMetadata().GetMinorNumber()
		gpuStr, err := marshaler.MarshalToString(gpu)
		if err != nil {
			scope.Errorf(
				"[GPU][%s][%s/%s] Encode pb message failed. %s",
				dataGranularity, gpuHost, gpuMinorNumber, err.Error())
			continue
		}
		jb := queue.NewJobBuilder(pdUnit, granularity, gpuStr)
		jobJSONStr, err := jb.GetJobJSONString()
		if err != nil {
			scope.Errorf(
				"[GPU][%s][%s/%s] Prepare predict job payload failed. %s",
				dataGranularity, gpuHost, gpuMinorNumber, err.Error())
			continue
		}

		gpuJobStr := fmt.Sprintf("%s/%s/%v", gpuHost, gpuMinorNumber, granularity)
		scope.Infof("[GPU][%s][%s/%s] Try to send predict job: %s",
			dataGranularity, gpuHost, gpuMinorNumber, gpuJobStr)
		err = queueSender.SendJsonString(queueName, jobJSONStr, gpuJobStr, granularity)
		if err != nil {
			scope.Errorf("[GPU][%s][%s/%s] Send predict job failed. %s",
				dataGranularity, gpuHost, gpuMinorNumber, err.Error())
		}
	}
}

func (dispatcher *predictJobSender) SendApplicationPredictJobs(
	applications []*datahub_resources.Application,
	queueSender queue.QueueSender, pdUnit string, granularity int64) {
	dataGranularity := queue.GetGranularityStr(granularity)
	marshaler := jsonpb.Marshaler{}
	for _, application := range applications {
		applicationNS := application.GetObjectMeta().GetNamespace()
		applicationName := application.GetObjectMeta().GetName()
		applicationStr, err := marshaler.MarshalToString(application)
		if err != nil {
			scope.Errorf(
				"[APPLICATION][%s][%s/%s] Encode pb message failed. %s",
				dataGranularity, applicationNS, applicationName, err.Error())
			continue
		}
		jb := queue.NewJobBuilder(pdUnit, granularity, applicationStr)
		jobJSONStr, err := jb.GetJobJSONString()
		if err != nil {
			scope.Errorf(
				"[APPLICATION][%s][%s/%s] Prepare predict job payload failed. %s",
				dataGranularity, applicationNS, applicationName, err.Error())
			continue
		}

		appJobStr := fmt.Sprintf("%s/%s/%v", applicationNS, applicationName, granularity)
		scope.Infof("[APPLICATION][%s][%s/%s] Try to send predict job: %s",
			dataGranularity, applicationNS, applicationName, appJobStr)
		err = queueSender.SendJsonString(queueName, jobJSONStr, appJobStr, granularity)
		if err != nil {
			scope.Errorf("[APPLICATION][%s][%s/%s] Send predict job failed. %s",
				dataGranularity, applicationNS, applicationName, err.Error())
		}
	}
}

func (dispatcher *predictJobSender) SendNamespacePredictJobs(namespaces []*datahub_resources.Namespace,
	queueSender queue.QueueSender, pdUnit string, granularity int64) {
	dataGranularity := queue.GetGranularityStr(granularity)
	marshaler := jsonpb.Marshaler{}
	for _, namespace := range namespaces {
		namespaceStr, err := marshaler.MarshalToString(namespace)
		namespaceName := namespace.GetObjectMeta().GetName()
		if err != nil {
			scope.Errorf("[NAMESPACE][%s][%s] Encode pb message failed. %s",
				dataGranularity, namespaceName, err.Error())
			continue
		}
		jb := queue.NewJobBuilder(pdUnit, granularity, namespaceStr)
		jobJSONStr, err := jb.GetJobJSONString()
		if err != nil {
			scope.Errorf("[NAMESPACE][%s][%s] Prepare predict job payload failed. %s",
				dataGranularity, namespaceName, err.Error())
			continue
		}

		nsJobStr := fmt.Sprintf("%s/%v", namespaceName, granularity)
		scope.Infof("[NAMESPACE][%s][%s] Try to send predict job: %s",
			dataGranularity, namespaceName, nsJobStr)
		err = queueSender.SendJsonString(queueName, jobJSONStr, nsJobStr, granularity)
		if err != nil {
			scope.Errorf("[NAMESPACE][%s][%s] Send predict job failed. %s",
				dataGranularity, namespaceName, err.Error())
		}
	}
}

func (dispatcher *predictJobSender) SendClusterPredictJobs(clusters []*datahub_resources.Cluster,
	queueSender queue.QueueSender, pdUnit string, granularity int64) {
	dataGranularity := queue.GetGranularityStr(granularity)
	marshaler := jsonpb.Marshaler{}
	for _, cluster := range clusters {
		clusterStr, err := marshaler.MarshalToString(cluster)
		clusterName := cluster.ObjectMeta.GetName()
		if err != nil {
			scope.Errorf("[CLUSTER][%s][%s] Encode pb message failed. %s",
				dataGranularity, clusterName, err.Error())
			continue
		}
		jb := queue.NewJobBuilder(pdUnit, granularity, clusterStr)
		jobJSONStr, err := jb.GetJobJSONString()
		if err != nil {
			scope.Errorf("[CLUSTER][%s][%s] Prepare predict job payload failed. %s",
				dataGranularity, clusterName, err.Error())
			continue
		}

		clusterJobStr := fmt.Sprintf("%s/%v", clusterName, granularity)
		scope.Infof("[CLUSTER][%s][%s] Try to send predict job: %s", dataGranularity, clusterName, clusterJobStr)
		err = queueSender.SendJsonString(queueName, jobJSONStr, clusterJobStr, granularity)
		if err != nil {
			scope.Errorf("[CLUSTER][%s][%s] Send predict job failed. %s",
				dataGranularity, clusterName, err.Error())
		}
	}
}

func (dispatcher *predictJobSender) SendControllerPredictJobs(
	controllers []*datahub_resources.Controller,
	queueSender queue.QueueSender, pdUnit string, granularity int64) {
	dataGranularity := queue.GetGranularityStr(granularity)
	marshaler := jsonpb.Marshaler{}
	for _, controller := range controllers {
		controllerNS := controller.GetObjectMeta().GetNamespace()
		controllerName := controller.GetObjectMeta().GetName()
		controllerStr, err := marshaler.MarshalToString(controller)
		if err != nil {
			scope.Errorf(
				"[CONTROLLER][%s][%s][%s/%s] Encode pb message failed. %s",
				controller.GetKind().String(), dataGranularity, controllerNS, controllerName, err.Error())
			continue
		}
		jb := queue.NewJobBuilder(pdUnit, granularity, controllerStr)
		jobJSONStr, err := jb.GetJobJSONString()
		if err != nil {
			scope.Errorf(
				"[CONTROLLER][%s][%s][%s/%s] Prepare predict job payload failed. %s",
				controller.GetKind().String(), dataGranularity, controllerNS, controllerName, err.Error())
			continue
		}

		controllerJobStr := fmt.Sprintf("%s/%s/%v", controllerNS, controllerName, granularity)
		scope.Infof("[CONTROLLER][%s][%s][%s/%s] Try to send predict job: %s",
			controller.GetKind().String(), dataGranularity, controllerNS, controllerName, controllerJobStr)
		err = queueSender.SendJsonString(queueName, jobJSONStr, controllerJobStr, granularity)
		if err != nil {
			scope.Errorf("[CONTROLLER][%s][%s][%s/%s] Send predict job failed. %s",
				controller.GetKind().String(), dataGranularity, controllerNS, controllerName, err.Error())
		}
	}
}
