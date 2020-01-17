package dispatcher

import (
	"fmt"

	"github.com/containers-ai/alameda/ai-dispatcher/consts"
	"github.com/containers-ai/alameda/ai-dispatcher/pkg/queue"
	datahub_common "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
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
		for _, metricType := range []datahub_common.MetricType{
			datahub_common.MetricType_MEMORY_USAGE_BYTES,
			datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
		} {
			jb := queue.NewJobBuilder(node.GetObjectMeta().GetClusterName(),
				pdUnit, granularity, metricType, nodeStr, nil)
			jobJSONStr, err := jb.GetJobJSONString()
			if err != nil {
				scope.Errorf("[NODE][%s][%s] Prepare predict metric %s job payload failed. %s",
					dataGranularity, nodeName, metricType, err.Error())
				continue
			}

			nodeJobStr := fmt.Sprintf("%s/%s/%s/%v/%s", consts.UnitTypeNode,
				node.GetObjectMeta().GetClusterName(), nodeName, granularity, metricType)
			scope.Infof("[NODE][%s][%s] Try to send predict metric %s job: %s",
				dataGranularity, nodeName, metricType, nodeJobStr)
			err = queueSender.SendJsonString(queueName, jobJSONStr, nodeJobStr, granularity)
			if err != nil {
				scope.Errorf("[NODE][%s][%s] Send predict metric %s job failed. %s",
					dataGranularity, nodeName, metricType, err.Error())
			}
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
		for _, ct := range pod.GetContainers() {
			for _, metricType := range []datahub_common.MetricType{
				datahub_common.MetricType_MEMORY_USAGE_BYTES,
				datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
			} {
				jb := queue.NewJobBuilder(pod.GetObjectMeta().GetClusterName(),
					pdUnit, granularity, metricType, podStr, map[string]string{
						"containerName": ct.GetName(),
					})
				jobJSONStr, err := jb.GetJobJSONString()
				if err != nil {
					scope.Errorf("[POD][%s][%s/%s/%s] Prepare predict metric %s job payload failed. %s",
						dataGranularity, podNS, podName, ct.GetName(), metricType, err.Error())
					continue
				}

				podJobStr := fmt.Sprintf("%s/%s/%s/%s/%s/%v/%s", consts.UnitTypePod,
					pod.GetObjectMeta().GetClusterName(), podNS, podName, ct.GetName(), granularity, metricType)
				scope.Infof("[POD][%s][%s/%s/%s] Try to send predict metric %s job: %s",
					dataGranularity, podNS, podName, ct.GetName(), metricType, podJobStr)
				err = queueSender.SendJsonString(queueName, jobJSONStr, podJobStr, granularity)
				if err != nil {
					scope.Errorf("[POD][%s][%s/%s/%s] Send predict metric %s job failed. %s",
						dataGranularity, podNS, podName, ct.GetName(), metricType, err.Error())
				}
			}
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
		for _, metricType := range []datahub_common.MetricType{
			datahub_common.MetricType_MEMORY_USAGE_BYTES,
			datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
		} {
			clusterID := "GPU_CLUSTER_NAME"
			jb := queue.NewJobBuilder(clusterID,
				pdUnit, granularity, metricType, gpuStr, nil)
			jobJSONStr, err := jb.GetJobJSONString()
			if err != nil {
				scope.Errorf(
					"[GPU][%s][%s/%s] Prepare predict metric %s job payload failed. %s",
					dataGranularity, gpuHost, gpuMinorNumber, metricType, err.Error())
				continue
			}

			gpuJobStr := fmt.Sprintf("%s/%s/%s/%s/%v/%s", consts.UnitTypeGPU, clusterID,
				gpuHost, gpuMinorNumber, granularity, metricType)
			scope.Infof("[GPU][%s][%s/%s] Try to send predict metric %s job: %s",
				dataGranularity, gpuHost, gpuMinorNumber, metricType, gpuJobStr)
			err = queueSender.SendJsonString(queueName, jobJSONStr, gpuJobStr, granularity)
			if err != nil {
				scope.Errorf("[GPU][%s][%s/%s] Send predict metric %s job failed. %s",
					dataGranularity, gpuHost, gpuMinorNumber, metricType, err.Error())
			}
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
		for _, metricType := range []datahub_common.MetricType{
			datahub_common.MetricType_MEMORY_USAGE_BYTES,
			datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
		} {
			jb := queue.NewJobBuilder(application.GetObjectMeta().GetClusterName(),
				pdUnit, granularity, metricType, applicationStr, nil)
			jobJSONStr, err := jb.GetJobJSONString()
			if err != nil {
				scope.Errorf(
					"[APPLICATION][%s][%s/%s] Prepare predict metric %s job payload failed. %s",
					dataGranularity, applicationNS, applicationName, metricType, err.Error())
				continue
			}

			appJobStr := fmt.Sprintf("%s/%s/%s/%s/%v/%s", consts.UnitTypeApplication,
				application.GetObjectMeta().GetClusterName(), applicationNS, applicationName, granularity, metricType)
			scope.Infof("[APPLICATION][%s][%s/%s] Try to send predict metric %s job: %s",
				dataGranularity, applicationNS, applicationName, metricType, appJobStr)
			err = queueSender.SendJsonString(queueName, jobJSONStr, appJobStr, granularity)
			if err != nil {
				scope.Errorf("[APPLICATION][%s][%s/%s] Send predict metric %s job failed. %s",
					dataGranularity, applicationNS, applicationName, metricType, err.Error())
			}
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
		for _, metricType := range []datahub_common.MetricType{
			datahub_common.MetricType_MEMORY_USAGE_BYTES,
			datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
		} {
			jb := queue.NewJobBuilder(namespace.GetObjectMeta().GetClusterName(),
				pdUnit, granularity, metricType, namespaceStr, nil)
			jobJSONStr, err := jb.GetJobJSONString()
			if err != nil {
				scope.Errorf("[NAMESPACE][%s][%s] Prepare predict metric %s job payload failed. %s",
					dataGranularity, namespaceName, metricType, err.Error())
				continue
			}

			nsJobStr := fmt.Sprintf("%s/%s/%s/%v/%s", consts.UnitTypeNamespace,
				namespace.GetObjectMeta().GetClusterName(), namespaceName, granularity, metricType)
			scope.Infof("[NAMESPACE][%s][%s] Try to send predict metric %s job: %s",
				dataGranularity, namespaceName, metricType, nsJobStr)
			err = queueSender.SendJsonString(queueName, jobJSONStr, nsJobStr, granularity)
			if err != nil {
				scope.Errorf("[NAMESPACE][%s][%s] Send predict metric %s job failed. %s",
					dataGranularity, namespaceName, metricType, err.Error())
			}
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
		for _, metricType := range []datahub_common.MetricType{
			datahub_common.MetricType_MEMORY_USAGE_BYTES,
			datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
		} {
			jb := queue.NewJobBuilder(cluster.GetObjectMeta().GetClusterName(),
				pdUnit, granularity, metricType, clusterStr, nil)
			jobJSONStr, err := jb.GetJobJSONString()
			if err != nil {
				scope.Errorf("[CLUSTER][%s][%s] Prepare predict metric %s job payload failed. %s",
					dataGranularity, clusterName, metricType, err.Error())
				continue
			}

			clusterJobStr := fmt.Sprintf("%s/%s/%v/%s", consts.UnitTypeCluster, clusterName, granularity, metricType)
			scope.Infof("[CLUSTER][%s][%s] Try to send predict metric %s job: %s",
				dataGranularity, clusterName, metricType, clusterJobStr)
			err = queueSender.SendJsonString(queueName, jobJSONStr, clusterJobStr, granularity)
			if err != nil {
				scope.Errorf("[CLUSTER][%s][%s] Send predict metric %s job failed. %s",
					dataGranularity, clusterName, metricType, err.Error())
			}
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
		controllerKindStr := controller.GetKind().String()
		controllerStr, err := marshaler.MarshalToString(controller)
		if err != nil {
			scope.Errorf(
				"[CONTROLLER][%s][%s][%s/%s] Encode pb message failed. %s",
				controllerKindStr, dataGranularity, controllerNS, controllerName, err.Error())
			continue
		}
		for _, metricType := range []datahub_common.MetricType{
			datahub_common.MetricType_MEMORY_USAGE_BYTES,
			datahub_common.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
		} {
			jb := queue.NewJobBuilder(controller.GetObjectMeta().GetClusterName(),
				pdUnit, granularity, metricType, controllerStr, nil)
			jobJSONStr, err := jb.GetJobJSONString()
			if err != nil {
				scope.Errorf(
					"[CONTROLLER][%s][%s][%s/%s] Prepare predict metric %s job payload failed. %s",
					controllerKindStr, dataGranularity, controllerNS, controllerName, metricType, err.Error())
				continue
			}

			controllerJobStr := fmt.Sprintf("%s/%s/%s/%s/%s/%v/%s", consts.UnitTypeController,
				controller.GetObjectMeta().GetClusterName(), controllerKindStr, controllerNS, controllerName, granularity, metricType)
			scope.Infof("[CONTROLLER][%s][%s][%s/%s] Try to send predict metric %s job: %s",
				controllerKindStr, dataGranularity, controllerNS, controllerName, metricType, controllerJobStr)
			err = queueSender.SendJsonString(queueName, jobJSONStr, controllerJobStr, granularity)
			if err != nil {
				scope.Errorf("[CONTROLLER][%s][%s][%s/%s] Send predict metric %s job failed. %s",
					controllerKindStr, dataGranularity, controllerNS, controllerName, metricType, err.Error())
			}
		}
	}
}
