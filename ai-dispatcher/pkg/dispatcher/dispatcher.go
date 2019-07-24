package dispatcher

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/streadway/amqp"

	"github.com/containers-ai/alameda/ai-dispatcher/pkg/queue"
	"github.com/containers-ai/alameda/pkg/utils/log"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/jsonpb"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

const (
	UNIT_TYPE_NODE = "NODE"
	UNIT_TYPE_POD  = "POD"
)
const queueName = "predict"

var scope = log.RegisterScope("dispatcher", "dispatcher dispatch jobs", 0)

type Dispatcher struct {
	svcGranularities []string
	svcPredictUnits  []string
	datahubGrpcCn    *grpc.ClientConn
	queueConn        *amqp.Connection
}

func NewDispatcher(datahubGrpcCn *grpc.ClientConn, queueConn *amqp.Connection) *Dispatcher {
	dispatcher := &Dispatcher{
		svcGranularities: viper.GetStringSlice("serviceSetting.granularities"),
		svcPredictUnits:  viper.GetStringSlice("serviceSetting.predictUnits"),
		datahubGrpcCn:    datahubGrpcCn,
		queueConn:        queueConn,
	}
	dispatcher.validCfg()
	return dispatcher
}

var wg sync.WaitGroup

func (dispatcher *Dispatcher) Start() {
	// generate len(dispatcher.svcGranularities) senders to publish job,
	// each sender use distinct channel which is not thread safe.
	// all jobs are published to the same queue.
	for _, granularity := range dispatcher.svcGranularities {
		granuSec := viper.GetInt(fmt.Sprintf("granularities.%s.dataGranularitySec", granularity))
		if granuSec == 0 {
			scope.Warnf("Granularity %v is not defined or set incorrect.", granularity)
			continue
		}
		wg.Add(1)
		queueSender := queue.NewRabbitMQSender(dispatcher.queueConn)
		go dispatcher.dispatch(queueSender, int64(granuSec))
	}
	wg.Wait()
}

func (dispatcher *Dispatcher) validCfg() {
	if len(dispatcher.svcGranularities) == 0 {
		scope.Errorf("no setting of granularities of service")
		os.Exit(1)
	}
	if len(dispatcher.svcPredictUnits) == 0 {
		scope.Errorf("no setting of predict units of service")
		os.Exit(1)
	}
}

func (dispatcher *Dispatcher) dispatch(queueSender queue.QueueSender, granularity int64) {
	defer wg.Done()
	for {
		for _, pdUnit := range dispatcher.svcPredictUnits {

			pdUnitType := viper.GetString(fmt.Sprintf("predictUnits.%s.type", pdUnit))

			if pdUnitType == "" {
				scope.Warnf("Unit %s is not defined or set incorrect.", pdUnit)
				continue
			}
			scope.Infof("Start dispatch unit %s with granularity %v seconds", pdUnitType, granularity)
			dispatcher.getAndPushJobs(queueSender, pdUnit, granularity)
		}
		time.Sleep(time.Duration(granularity) * time.Second)
	}
}

func (dispatcher *Dispatcher) getAndPushJobs(queueSender queue.QueueSender, pdUnit string, granularity int64) {

	datahubServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(dispatcher.datahubGrpcCn)
	marshaler := jsonpb.Marshaler{}
	if pdUnit == UNIT_TYPE_NODE {
		res, err := datahubServiceClnt.ListAlamedaNodes(context.Background(), &datahub_v1alpha1.ListAlamedaNodesRequest{})
		if err != nil {
			scope.Errorf("List predict job for nodes failed with granularity %v seconds. %s", granularity, err.Error())
			return
		}
		nodes := res.GetNodes()
		scope.Infof("Start sending %v node jobs to queue with granularity %v seconds.", len(nodes), granularity)
		for _, node := range nodes {
			nodeStr, err := marshaler.MarshalToString(node)
			if err != nil {
				scope.Errorf("Encode pb message failed for node %s with granularity seconds %v. %s", node.GetName(), granularity, err.Error())
				continue
			}
			jb := queue.NewJobBuilder(pdUnit, granularity, nodeStr)
			jobJSONStr, err := jb.GetJobJSONString()
			if err != nil {
				scope.Errorf("Prepare job payload failed for node %s with granularity seconds %v. %s", node.GetName(), granularity, err.Error())
				continue
			}
			err = queueSender.SendJsonString(queueName, jobJSONStr)
			if err != nil {
				scope.Errorf("Send job for node %s failed with granularity %v seconds. %s", node.GetName(), granularity, err.Error())
			}
		}
		scope.Infof("Sending %v node jobs to queue completely with granularity %v seconds.", len(nodes), granularity)
	} else if pdUnit == UNIT_TYPE_POD {
		res, err := datahubServiceClnt.ListAlamedaPods(context.Background(), &datahub_v1alpha1.ListAlamedaPodsRequest{})
		if err != nil {
			scope.Errorf("List predict job for pods failed with granularity %v seconds. %s", granularity, err.Error())
			return
		}
		pods := res.GetPods()
		scope.Infof("Start sending %v pod jobs to queue with granularity %v seconds.", len(pods), granularity)
		for _, pod := range pods {
			podNSN := pod.GetNamespacedName()
			podStr, err := marshaler.MarshalToString(pod)
			if err != nil {
				scope.Errorf("Encode pb message failed for pod %s/%s with granularity %v seconds. %s", podNSN.GetNamespace(), podNSN.GetName(), granularity, err.Error())
				continue
			}
			jb := queue.NewJobBuilder(pdUnit, granularity, podStr)
			jobJSONStr, err := jb.GetJobJSONString()
			if err != nil {
				scope.Errorf("Prepare job payload failed for pod %s/%s with granularity %v seconds. %s", podNSN.GetNamespace(), podNSN.GetName(), granularity, err.Error())
				continue
			}
			err = queueSender.SendJsonString(queueName, jobJSONStr)
			if err != nil {
				scope.Errorf("Send job for pod %s/%s failed with granularity %v seconds. %s", podNSN.GetNamespace(), podNSN.GetName(), granularity, err.Error())
			}
		}
		scope.Infof("Sending %v pod jobs to queue completely with granularity %v seconds.", len(pods), granularity)
	}
}
