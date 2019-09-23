package dispatcher

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/containers-ai/alameda/ai-dispatcher/pkg/queue"
	"github.com/containers-ai/alameda/pkg/utils/log"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
)

const (
	UnitTypeNode = "NODE"
	UnitTypePod  = "POD"
	UnitTypeGPU  = "GPU"
)
const queueName = "predict"
const modelQueueName = "model"

var scope = log.RegisterScope("dispatcher", "dispatcher dispatch jobs", 0)

type Dispatcher struct {
	svcGranularities []string
	svcPredictUnits  []string
	datahubGrpcCn    *grpc.ClientConn
	queueConn        *amqp.Connection

	modelJobSender   *modelJobSender
	predictJobSender *predictJobSender
}

func NewDispatcher(datahubGrpcCn *grpc.ClientConn, granularities []string,
	predictUnits []string, modelMapper *ModelMapper) *Dispatcher {
	modelJobSender := NewModelJobSender(datahubGrpcCn, modelMapper)
	predictJobSender := NewPredictJobSender(datahubGrpcCn)
	dispatcher := &Dispatcher{
		svcGranularities: granularities,
		svcPredictUnits:  predictUnits,
		datahubGrpcCn:    datahubGrpcCn,
		modelJobSender:   modelJobSender,
		predictJobSender: predictJobSender,
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
		granuSec := viper.GetInt(
			fmt.Sprintf("granularities.%s.dataGranularitySec", granularity))
		queueJobSendIntervalSec := viper.GetInt(
			fmt.Sprintf("granularities.%s.queueJobSendIntervalSec", granularity))
		if granuSec == 0 {
			scope.Warnf("Granularity %v is not defined or set incorrect.", granularity)
			continue
		}
		predictionStep := viper.GetInt(fmt.Sprintf("granularities.%s.predictionSteps",
			granularity))
		if predictionStep == 0 {
			scope.Warnf("Prediction step of Granularity %v is not defined or set incorrect.",
				granularity)
			continue
		}
		wg.Add(1)
		go dispatcher.dispatch(int64(granuSec), int64(predictionStep),
			int64(queueJobSendIntervalSec))
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

func (dispatcher *Dispatcher) dispatch(granularity int64, predictionStep int64,
	queueJobSendIntervalSec int64) {
	defer wg.Done()
	queueURL := viper.GetString("queue.url")
	queueConnRetryItvMS := viper.GetInt64("queue.retry.connectIntervalMs")
	if queueConnRetryItvMS == 0 {
		queueConnRetryItvMS = 3000
	}
	for {
		queueConn := queue.GetQueueConn(queueURL, queueConnRetryItvMS)
		queueSender := queue.NewRabbitMQSender(queueConn)
		for _, pdUnit := range dispatcher.svcPredictUnits {

			pdUnitType := viper.GetString(fmt.Sprintf("predictUnits.%s.type", pdUnit))

			if pdUnitType == "" {
				scope.Warnf("Unit %s is not defined or set incorrect.", pdUnit)
				continue
			}
			scope.Infof("Start dispatch unit %s with granularity %v seconds",
				pdUnitType, granularity)
			dispatcher.getAndPushJobs(queueSender, pdUnit, granularity, predictionStep)
		}
		queueConn.Close()
		time.Sleep(time.Duration(queueJobSendIntervalSec) * time.Second)
	}
}

func (dispatcher *Dispatcher) getAndPushJobs(queueSender queue.QueueSender,
	pdUnit string, granularity int64, predictionStep int64) {

	datahubServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(dispatcher.datahubGrpcCn)

	if pdUnit == UnitTypeNode {
		res, err := datahubServiceClnt.ListAlamedaNodes(context.Background(),
			&datahub_v1alpha1.ListAlamedaNodesRequest{})
		if err != nil {
			scope.Errorf("List nodes for model/predict job failed with granularity %v seconds. %s",
				granularity, err.Error())
			return
		}
		nodes := res.GetNodes()
		// send predict jobs
		scope.Infof("Start sending %v node jobs to queue with granularity %v seconds.",
			len(nodes), granularity)
		dispatcher.predictJobSender.SendNodePredictJobs(nodes, queueSender, pdUnit, granularity)
		if viper.GetBool("model.enabled") {
			dispatcher.modelJobSender.sendNodeModelJobs(nodes, queueSender, pdUnit, granularity,
				predictionStep)
		}
		scope.Infof("Sending %v node jobs to queue completely with granularity %v seconds.",
			len(nodes), granularity)

	} else if pdUnit == UnitTypePod {
		res, err := datahubServiceClnt.ListAlamedaPods(context.Background(),
			&datahub_v1alpha1.ListAlamedaPodsRequest{})
		if err != nil {
			scope.Errorf("List pods for model/predict job failed with granularity %v seconds. %s",
				granularity, err.Error())
			return
		}
		pods := res.GetPods()
		// send predict jobs
		scope.Infof("Start sending %v pod jobs to queue with granularity %v seconds.",
			len(pods), granularity)
		dispatcher.predictJobSender.SendPodPredictJobs(pods, queueSender, pdUnit, granularity)
		if viper.GetBool("model.enabled") {
			dispatcher.modelJobSender.sendPodModelJobs(pods, queueSender, pdUnit, granularity,
				predictionStep)
		}
		scope.Infof("Sending %v pod jobs to queue completely with granularity %v seconds.",
			len(pods), granularity)
	} else if pdUnit == UnitTypeGPU && granularity == 3600 {
		res, err := datahubServiceClnt.ListGpus(context.Background(),
			&datahub_v1alpha1.ListGpusRequest{})
		if err != nil {
			scope.Errorf("List gpus for model/predict job failed with granularity %v seconds. %s",
				granularity, err.Error())
			return
		}
		gpus := res.GetGpus()
		// send predict jobs
		scope.Infof("Start sending %v gpu jobs to queue with granularity %v seconds.",
			len(gpus), granularity)
		dispatcher.predictJobSender.SendGPUPredictJobs(gpus, queueSender, pdUnit, granularity)
		if viper.GetBool("model.enabled") {
			dispatcher.modelJobSender.sendGPUModelJobs(gpus, queueSender, pdUnit, granularity,
				predictionStep)
		}
		scope.Infof("Sending %v gpu jobs to queue completely with granularity %v seconds.",
			len(gpus), granularity)
	}
}
