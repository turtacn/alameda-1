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
	marshaler := jsonpb.Marshaler{}
	for _, node := range nodes {
		nodeStr, err := marshaler.MarshalToString(node)
		nodeName := node.ObjectMeta.GetName()
		if err != nil {
			scope.Errorf("Encode pb message failed for node %s with granularity seconds %v. %s",
				nodeName, granularity, err.Error())
			continue
		}
		jb := queue.NewJobBuilder(pdUnit, granularity, nodeStr)
		jobJSONStr, err := jb.GetJobJSONString()
		if err != nil {
			scope.Errorf("Prepare job payload failed for node %s with granularity seconds %v. %s",
				nodeName, granularity, err.Error())
			continue
		}

		nodeJobStr := fmt.Sprintf("%s/%v", nodeName, granularity)
		scope.Infof("Try to send node predict job: %s", nodeJobStr)
		err = queueSender.SendJsonString(queueName, jobJSONStr, nodeJobStr)
		if err != nil {
			scope.Errorf("Send job for node %s failed with granularity %v seconds. %s",
				nodeName, granularity, err.Error())
		}
	}
}

func (dispatcher *predictJobSender) SendPodPredictJobs(pods []*datahub_resources.Pod,
	queueSender queue.QueueSender, pdUnit string, granularity int64) {
	marshaler := jsonpb.Marshaler{}
	for _, pod := range pods {
		podNS := pod.ObjectMeta.GetNamespace()
		podName := pod.ObjectMeta.GetName()
		podStr, err := marshaler.MarshalToString(pod)
		if err != nil {
			scope.Errorf("Encode pb message failed for pod %s/%s with granularity %v seconds. %s",
				podNS, podName, granularity, err.Error())
			continue
		}
		jb := queue.NewJobBuilder(pdUnit, granularity, podStr)
		jobJSONStr, err := jb.GetJobJSONString()
		if err != nil {
			scope.Errorf("Prepare job payload failed for pod %s/%s with granularity %v seconds. %s",
				podNS, podName, granularity, err.Error())
			continue
		}

		podJobStr := fmt.Sprintf("%s/%s/%v", podNS, podName, granularity)
		scope.Infof("Try to send pod predict job: %s", podJobStr)
		err = queueSender.SendJsonString(queueName, jobJSONStr, podJobStr)
		if err != nil {
			scope.Errorf("Send job for pod %s/%s failed with granularity %v seconds. %s",
				podNS, podName, granularity, err.Error())
		}
	}
}

func (dispatcher *predictJobSender) SendGPUPredictJobs(gpus []*datahub_gpu.Gpu,
	queueSender queue.QueueSender, pdUnit string, granularity int64) {
	marshaler := jsonpb.Marshaler{}
	for _, gpu := range gpus {
		gpuHost := gpu.GetMetadata().GetHost()
		gpuMinorNumber := gpu.GetMetadata().GetMinorNumber()
		gpuStr, err := marshaler.MarshalToString(gpu)
		if err != nil {
			scope.Errorf(
				"Encode pb message failed for gpu host: %s, minor number: %s with granularity %v seconds. %s",
				gpuHost, gpuMinorNumber, granularity, err.Error())
			continue
		}
		jb := queue.NewJobBuilder(pdUnit, granularity, gpuStr)
		jobJSONStr, err := jb.GetJobJSONString()
		if err != nil {
			scope.Errorf(
				"Prepare job payload failed for gpu host: %s, minor number: %s with granularity %v seconds. %s",
				gpuHost, gpuMinorNumber, granularity, err.Error())
			continue
		}

		gpuJobStr := fmt.Sprintf("%s/%s/%v", gpuHost, gpuMinorNumber, granularity)
		scope.Infof("Try to send gpu predict job: %s", gpuJobStr)
		err = queueSender.SendJsonString(queueName, jobJSONStr, gpuJobStr)
		if err != nil {
			scope.Errorf("Send job for gpu host: %s, minor number: %s failed with granularity %v seconds. %s",
				gpuHost, gpuMinorNumber, granularity, err.Error())
		}
	}
}

func (dispatcher *predictJobSender) SendApplicationPredictJobs(
	applications []*datahub_resources.Application,
	queueSender queue.QueueSender, pdUnit string, granularity int64) {
	marshaler := jsonpb.Marshaler{}
	for _, application := range applications {
		applicationNS := application.GetObjectMeta().GetNamespace()
		applicationName := application.GetObjectMeta().GetName()
		applicationStr, err := marshaler.MarshalToString(application)
		if err != nil {
			scope.Errorf(
				"Encode pb message failed for application %s/%s with granularity %v seconds. %s",
				applicationNS, applicationName, granularity, err.Error())
			continue
		}
		jb := queue.NewJobBuilder(pdUnit, granularity, applicationStr)
		jobJSONStr, err := jb.GetJobJSONString()
		if err != nil {
			scope.Errorf(
				"Prepare job payload failed for application %s/%s with granularity %v seconds. %s",
				applicationNS, applicationName, granularity, err.Error())
			continue
		}

		appJobStr := fmt.Sprintf("%s/%s/%v", applicationNS, applicationName, granularity)
		scope.Infof("Try to send application predict job: %s", appJobStr)
		err = queueSender.SendJsonString(queueName, jobJSONStr, appJobStr)
		if err != nil {
			scope.Errorf("Send job for application %s/%s failed with granularity %v seconds. %s",
				applicationNS, applicationName, granularity, err.Error())
		}
	}
}

func (dispatcher *predictJobSender) SendNamespacePredictJobs(namespaces []*datahub_resources.Namespace,
	queueSender queue.QueueSender, pdUnit string, granularity int64) {
	marshaler := jsonpb.Marshaler{}
	for _, namespace := range namespaces {
		namespaceStr, err := marshaler.MarshalToString(namespace)
		namespaceName := namespace.GetObjectMeta().GetNamespace()
		if err != nil {
			scope.Errorf("Encode pb message failed for namespace %s with granularity seconds %v. %s",
				namespaceName, granularity, err.Error())
			continue
		}
		jb := queue.NewJobBuilder(pdUnit, granularity, namespaceStr)
		jobJSONStr, err := jb.GetJobJSONString()
		if err != nil {
			scope.Errorf("Prepare job payload failed for namespace %s with granularity seconds %v. %s",
				namespaceName, granularity, err.Error())
			continue
		}

		nsJobStr := fmt.Sprintf("%s/%v", namespaceName, granularity)
		scope.Infof("Try to send namespace predict job: %s", nsJobStr)
		err = queueSender.SendJsonString(queueName, jobJSONStr, nsJobStr)
		if err != nil {
			scope.Errorf("Send job for namespace %s failed with granularity %v seconds. %s",
				namespaceName, granularity, err.Error())
		}
	}
}

func (dispatcher *predictJobSender) SendClusterPredictJobs(clusters []*datahub_resources.Cluster,
	queueSender queue.QueueSender, pdUnit string, granularity int64) {
	marshaler := jsonpb.Marshaler{}
	for _, cluster := range clusters {
		clusterStr, err := marshaler.MarshalToString(cluster)
		clusterName := cluster.ObjectMeta.GetName()
		if err != nil {
			scope.Errorf("Encode pb message failed for cluster %s with granularity seconds %v. %s",
				clusterName, granularity, err.Error())
			continue
		}
		jb := queue.NewJobBuilder(pdUnit, granularity, clusterStr)
		jobJSONStr, err := jb.GetJobJSONString()
		if err != nil {
			scope.Errorf("Prepare job payload failed for cluster %s with granularity seconds %v. %s",
				clusterName, granularity, err.Error())
			continue
		}

		clusterJobStr := fmt.Sprintf("%s/%v", clusterName, granularity)
		scope.Infof("Try to send cluster predict job: %s", clusterJobStr)
		err = queueSender.SendJsonString(queueName, jobJSONStr, clusterJobStr)
		if err != nil {
			scope.Errorf("Send job for cluster %s failed with granularity %v seconds. %s",
				clusterName, granularity, err.Error())
		}
	}
}

func (dispatcher *predictJobSender) SendControllerPredictJobs(
	controllers []*datahub_resources.Controller,
	queueSender queue.QueueSender, pdUnit string, granularity int64) {
	marshaler := jsonpb.Marshaler{}
	for _, controller := range controllers {
		controllerNS := controller.GetObjectMeta().GetNamespace()
		controllerName := controller.GetObjectMeta().GetName()
		controllerStr, err := marshaler.MarshalToString(controller)
		if err != nil {
			scope.Errorf(
				"Encode pb message failed for controller %s %s/%s with granularity %v seconds. %s",
				controller.GetKind().String(), controllerNS, controllerName, granularity, err.Error())
			continue
		}
		jb := queue.NewJobBuilder(pdUnit, granularity, controllerStr)
		jobJSONStr, err := jb.GetJobJSONString()
		if err != nil {
			scope.Errorf(
				"Prepare job payload failed for controller %s %s/%s with granularity %v seconds. %s",
				controller.GetKind().String(), controllerNS, controllerName, granularity, err.Error())
			continue
		}

		controllerJobStr := fmt.Sprintf("%s/%s/%v", controllerNS, controllerName, granularity)
		scope.Infof("Try to send controller predict job: %s", controllerJobStr)
		err = queueSender.SendJsonString(queueName, jobJSONStr, controllerJobStr)
		if err != nil {
			scope.Errorf("Send job for controller %s %s/%s failed with granularity %v seconds. %s",
				controller.GetKind().String(), controllerNS, controllerName, granularity, err.Error())
		}
	}
}
