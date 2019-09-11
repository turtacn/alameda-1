package dispatcher

import (
	"github.com/containers-ai/alameda/ai-dispatcher/pkg/queue"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
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

func (dispatcher *predictJobSender) sendNodePredictJobs(nodes []*datahub_v1alpha1.Node,
	queueSender queue.QueueSender, pdUnit string, granularity int64) {
	marshaler := jsonpb.Marshaler{}
	for _, node := range nodes {
		nodeStr, err := marshaler.MarshalToString(node)
		if err != nil {
			scope.Errorf("Encode pb message failed for node %s with granularity seconds %v. %s",
				node.GetName(), granularity, err.Error())
			continue
		}
		jb := queue.NewJobBuilder(pdUnit, granularity, nodeStr)
		jobJSONStr, err := jb.GetJobJSONString()
		if err != nil {
			scope.Errorf("Prepare job payload failed for node %s with granularity seconds %v. %s",
				node.GetName(), granularity, err.Error())
			continue
		}
		err = queueSender.SendJsonString(queueName, jobJSONStr)
		if err != nil {
			scope.Errorf("Send job for node %s failed with granularity %v seconds. %s",
				node.GetName(), granularity, err.Error())
		}
	}
}

func (dispatcher *predictJobSender) sendPodPredictJobs(pods []*datahub_v1alpha1.Pod,
	queueSender queue.QueueSender, pdUnit string, granularity int64) {
	marshaler := jsonpb.Marshaler{}
	for _, pod := range pods {
		podNSN := pod.GetNamespacedName()
		podStr, err := marshaler.MarshalToString(pod)
		if err != nil {
			scope.Errorf("Encode pb message failed for pod %s/%s with granularity %v seconds. %s",
				podNSN.GetNamespace(), podNSN.GetName(), granularity, err.Error())
			continue
		}
		jb := queue.NewJobBuilder(pdUnit, granularity, podStr)
		jobJSONStr, err := jb.GetJobJSONString()
		if err != nil {
			scope.Errorf("Prepare job payload failed for pod %s/%s with granularity %v seconds. %s",
				podNSN.GetNamespace(), podNSN.GetName(), granularity, err.Error())
			continue
		}
		err = queueSender.SendJsonString(queueName, jobJSONStr)
		if err != nil {
			scope.Errorf("Send job for pod %s/%s failed with granularity %v seconds. %s",
				podNSN.GetNamespace(), podNSN.GetName(), granularity, err.Error())
		}
	}
}
