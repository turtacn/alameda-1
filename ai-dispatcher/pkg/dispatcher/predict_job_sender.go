package dispatcher

import (
	"fmt"

	"github.com/containers-ai/alameda/ai-dispatcher/pkg/queue"
	datahub_gpu "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/gpu"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"github.com/golang/protobuf/jsonpb"
	"github.com/spf13/viper"
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
		if granularity == 30 && !viper.GetBool("hourlyPredict") {
			continue
		}

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
		err = queueSender.SendJsonString(queueName, jobJSONStr,
			fmt.Sprintf("%s/%v", nodeName, granularity))
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
		if granularity == 30 &&
			(!viper.GetBool("hourlyPredict") &&
				pod.GetAlamedaPodSpec().GetScalingTool() != datahub_resources.ScalingTool_VPA) {
			continue
		}

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
		err = queueSender.SendJsonString(queueName, jobJSONStr,
			fmt.Sprintf("%s/%s/%v", podNS, podName, granularity))
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
		if granularity == 30 && !viper.GetBool("hourlyPredict") {
			continue
		}

		gpuHost := gpu.GetMetadata().GetHost()
		gpuMinorNumber := gpu.GetMetadata().GetMinorNumber()
		gpuStr, err := marshaler.MarshalToString(gpu)
		if err != nil {
			scope.Errorf("Encode pb message failed for gpu host: %s, minor number: %s with granularity %v seconds. %s",
				gpuHost, gpuMinorNumber, granularity, err.Error())
			continue
		}
		jb := queue.NewJobBuilder(pdUnit, granularity, gpuStr)
		jobJSONStr, err := jb.GetJobJSONString()
		if err != nil {
			scope.Errorf("Prepare job payload failed for gpu host: %s, minor number: %s with granularity %v seconds. %s",
				gpuHost, gpuMinorNumber, granularity, err.Error())
			continue
		}
		err = queueSender.SendJsonString(queueName, jobJSONStr,
			fmt.Sprintf("%s/%s/%v", gpuHost, gpuMinorNumber, granularity))
		if err != nil {
			scope.Errorf("Send job for gpu host: %s, minor number: %s failed with granularity %v seconds. %s",
				gpuHost, gpuMinorNumber, granularity, err.Error())
		}
	}
}
