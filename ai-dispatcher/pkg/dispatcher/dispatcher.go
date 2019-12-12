package dispatcher

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/containers-ai/alameda/ai-dispatcher/pkg/metrics"
	"github.com/containers-ai/alameda/ai-dispatcher/pkg/queue"
	"github.com/containers-ai/alameda/pkg/utils/log"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	datahub_gpu "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/gpu"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
)

const (
	UnitTypeNode        = "NODE"
	UnitTypePod         = "POD"
	UnitTypeGPU         = "GPU"
	UnitTypeNamespace   = "NAMESPACE"
	UnitTypeApplication = "APPLICATION"
	UnitTypeCluster     = "CLUSTER"
	UnitTypeController  = "CONTROLLER"
)
const queueName = "predict"
const modelQueueName = "model"

var (
	modelHasVPA   = false
	predictHasVPA = false
)

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
	predictUnits []string, modelMapper *ModelMapper, metricExporter *metrics.Exporter) *Dispatcher {
	modelJobSender := NewModelJobSender(datahubGrpcCn, modelMapper, metricExporter)
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
		predictionStep := viper.GetInt(fmt.Sprintf("granularities.%s.predictionSteps",
			granularity))
		if predictionStep == 0 {
			scope.Warnf("Prediction step of Granularity %v is not defined or set incorrect.",
				granularity)
			continue
		}
		wg.Add(1)
		go dispatcher.dispatch(granularity, int64(predictionStep),
			"predictionJobSendIntervalSec")
		wg.Add(1)
		go dispatcher.dispatch(granularity, int64(predictionStep),
			"modelJobSendIntervalSec")
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

func (dispatcher *Dispatcher) dispatch(granularity string, predictionStep int64,
	queueJobType string) {
	defer wg.Done()
	granularitySec := int64(viper.GetInt(
		fmt.Sprintf("granularities.%s.dataGranularitySec", granularity)))
	if granularitySec == 0 {
		scope.Warnf("Granularity %v is not defined or set incorrect.", granularitySec)
		return
	}
	queueJobSendIntervalSec := viper.GetInt(
		fmt.Sprintf("granularities.%s.%s", granularity, queueJobType))
	queueURL := viper.GetString("queue.url")
	queueConnRetryItvMS := viper.GetInt64("queue.retry.connectIntervalMs")
	if queueConnRetryItvMS == 0 {
		queueConnRetryItvMS = 3000
	}
	for {
		queueSender, queueConn := queue.NewRabbitMQSender(queueURL, queueConnRetryItvMS)
		// Node will send model/predict job with granularity 30s if modelHasVPA/predictHasVPA is true
		if granularitySec == 30 {
			modelHasVPA = false
			predictHasVPA = false
		}
		for _, pdUnit := range dispatcher.svcPredictUnits {
			if dispatcher.skipJobSending(pdUnit, granularitySec) {
				continue
			}

			pdUnitType := viper.GetString(fmt.Sprintf("predictUnits.%s.type", pdUnit))

			if pdUnitType == "" {
				scope.Warnf("Unit %s is not defined or set incorrect.", pdUnit)
				continue
			}

			if queueJobType == "predictionJobSendIntervalSec" {
				scope.Infof(
					"Start dispatching prediction unit %s with granularity %v seconds and cycle %v seconds",
					pdUnitType, granularitySec, queueJobSendIntervalSec)
			} else if queueJobType == "modelJobSendIntervalSec" {
				scope.Infof(
					"Start dispatching model unit %s with granularity %v seconds and cycle %v seconds",
					pdUnitType, granularitySec, queueJobSendIntervalSec)
			}

			dispatcher.getAndPushJobs(queueSender, pdUnit, granularitySec,
				predictionStep, queueJobType)
		}
		queueConn.Close()
		time.Sleep(time.Duration(queueJobSendIntervalSec) * time.Second)
	}
}

func (dispatcher *Dispatcher) getAndPushJobs(queueSender queue.QueueSender,
	pdUnit string, granularity int64, predictionStep int64, queueJobType string) {

	datahubServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(dispatcher.datahubGrpcCn)

	if pdUnit == UnitTypeNode {
		res, err := datahubServiceClnt.ListNodes(context.Background(),
			&datahub_resources.ListNodesRequest{})
		if err != nil {
			scope.Errorf(
				"List nodes for model/predict job failed with granularity %v seconds. %s",
				granularity, err.Error())
			return
		}

		nodes := []*datahub_resources.Node{}
		if queueJobType == "predictionJobSendIntervalSec" {
			for _, no := range res.GetNodes() {
				if (granularity == 30 && !viper.GetBool("hourlyPredict")) && !predictHasVPA {
					continue
				}
				nodes = append(nodes, no)
			}
			scope.Infof(
				"Start sending %v node prediction jobs to queue with granularity %v seconds.",
				len(nodes), granularity)
			dispatcher.predictJobSender.SendNodePredictJobs(nodes, queueSender, pdUnit, granularity)
		}
		if viper.GetBool("model.enabled") && queueJobType == "modelJobSendIntervalSec" {
			for _, no := range res.GetNodes() {
				if (granularity == 30 && !viper.GetBool("hourlyPredict")) && !modelHasVPA {
					continue
				}
				nodes = append(nodes, no)
			}
			scope.Infof(
				"Start sending %v node model jobs to queue with granularity %v seconds.",
				len(nodes), granularity)
			dispatcher.modelJobSender.SendNodeModelJobs(nodes, queueSender, pdUnit, granularity,
				predictionStep)
		}
		scope.Infof(
			"Sending %v node jobs to queue completely with granularity %v seconds.",
			len(nodes), granularity)

	} else if pdUnit == UnitTypePod {
		res, err := datahubServiceClnt.ListPods(context.Background(),
			&datahub_resources.ListPodsRequest{
				ScalingTool: datahub_resources.ScalingTool_VPA,
			})
		if err != nil {
			scope.Errorf(
				"List pods for model/predict job failed with granularity %v seconds. %s",
				granularity, err.Error())
			return
		}

		pods := []*datahub_resources.Pod{}
		hasVPA := false
		for _, pod := range res.GetPods() {
			if granularity == 30 && (!viper.GetBool("hourlyPredict") &&
				pod.GetAlamedaPodSpec().GetScalingTool() != datahub_resources.ScalingTool_VPA) {
				continue
			}
			if pod.GetAlamedaPodSpec().GetScalingTool() == datahub_resources.ScalingTool_VPA {
				hasVPA = true
			}
			pods = append(pods, pod)
		}

		if queueJobType == "predictionJobSendIntervalSec" {
			if hasVPA {
				predictHasVPA = true
			}
			scope.Infof(
				"Start sending %v pod prediction jobs to queue with granularity %v seconds.",
				len(pods), granularity)
			dispatcher.predictJobSender.SendPodPredictJobs(pods, queueSender, pdUnit, granularity)
		}
		if viper.GetBool("model.enabled") && queueJobType == "modelJobSendIntervalSec" {
			if hasVPA {
				modelHasVPA = true
			}
			scope.Infof(
				"Start sending %v pod model jobs to queue with granularity %v seconds.",
				len(pods), granularity)
			dispatcher.modelJobSender.SendPodModelJobs(pods, queueSender, pdUnit, granularity,
				predictionStep)
		}
		scope.Infof(
			"Sending %v pod jobs to queue completely with granularity %v seconds.",
			len(pods), granularity)
	} else if pdUnit == UnitTypeGPU {
		res, err := datahubServiceClnt.ListGpus(context.Background(),
			&datahub_gpu.ListGpusRequest{})
		if err != nil {
			scope.Errorf(
				"List gpus for model/predict job failed with granularity %v seconds. %s",
				granularity, err.Error())
			return
		}
		gpus := res.GetGpus()
		if queueJobType == "predictionJobSendIntervalSec" {
			scope.Infof(
				"Start sending %v gpu prediction jobs to queue with granularity %v seconds.",
				len(gpus), granularity)
			dispatcher.predictJobSender.SendGPUPredictJobs(gpus, queueSender, pdUnit, granularity)
		}
		if viper.GetBool("model.enabled") && queueJobType == "modelJobSendIntervalSec" {
			scope.Infof(
				"Start sending %v gpu model jobs to queue with granularity %v seconds.",
				len(gpus), granularity)
			dispatcher.modelJobSender.SendGPUModelJobs(gpus, queueSender, pdUnit, granularity,
				predictionStep)
		}
		scope.Infof("Sending %v gpu jobs to queue completely with granularity %v seconds.",
			len(gpus), granularity)
	} else if pdUnit == UnitTypeApplication {
		res, err := datahubServiceClnt.ListApplications(context.Background(),
			&datahub_resources.ListApplicationsRequest{})
		if err != nil {
			scope.Errorf(
				"List applications for model/predict job failed with granularity %v seconds. %s",
				granularity, err.Error())
			return
		}
		applications := []*datahub_resources.Application{}
		for _, app := range res.GetApplications() {
			if granularity == 30 && (!viper.GetBool("hourlyPredict") &&
				app.GetAlamedaApplicationSpec().GetScalingTool() !=
					datahub_resources.ScalingTool_VPA) {
				continue
			}
			applications = append(applications, app)
		}
		if queueJobType == "predictionJobSendIntervalSec" {
			scope.Infof(
				"Start sending %v application prediction jobs to queue with granularity %v seconds.",
				len(applications), granularity)
			dispatcher.predictJobSender.SendApplicationPredictJobs(applications, queueSender, pdUnit, granularity)
		}
		if viper.GetBool("model.enabled") && queueJobType == "modelJobSendIntervalSec" {
			scope.Infof(
				"Start sending %v application model jobs to queue with granularity %v seconds.",
				len(applications), granularity)
			dispatcher.modelJobSender.SendApplicationModelJobs(applications, queueSender, pdUnit, granularity,
				predictionStep)
		}
		scope.Infof("Sending %v application jobs to queue completely with granularity %v seconds.",
			len(applications), granularity)
	} else if pdUnit == UnitTypeNamespace {
		res, err := datahubServiceClnt.ListNamespaces(context.Background(),
			&datahub_resources.ListNamespacesRequest{})
		if err != nil {
			scope.Errorf(
				"List namespaces for model/predict job failed with granularity %v seconds. %s",
				granularity, err.Error())
			return
		}

		namespaces := res.GetNamespaces()
		if queueJobType == "predictionJobSendIntervalSec" {
			scope.Infof(
				"Start sending %v namespace prediction jobs to queue with granularity %v seconds.",
				len(namespaces), granularity)
			dispatcher.predictJobSender.SendNamespacePredictJobs(namespaces, queueSender, pdUnit, granularity)
		}
		if viper.GetBool("model.enabled") && queueJobType == "modelJobSendIntervalSec" {
			scope.Infof(
				"Start sending %v namespace model jobs to queue with granularity %v seconds.",
				len(namespaces), granularity)
			dispatcher.modelJobSender.SendNamespaceModelJobs(namespaces, queueSender, pdUnit, granularity,
				predictionStep)
		}
		scope.Infof(
			"Sending %v namespace jobs to queue completely with granularity %v seconds.",
			len(namespaces), granularity)
	} else if pdUnit == UnitTypeCluster {
		res, err := datahubServiceClnt.ListClusters(context.Background(),
			&datahub_resources.ListClustersRequest{})
		if err != nil {
			scope.Errorf(
				"List clusters for model/predict job failed with granularity %v seconds. %s",
				granularity, err.Error())
			return
		}

		clusters := res.GetClusters()
		if queueJobType == "predictionJobSendIntervalSec" {
			scope.Infof(
				"Start sending %v cluster prediction jobs to queue with granularity %v seconds.",
				len(clusters), granularity)
			dispatcher.predictJobSender.SendClusterPredictJobs(clusters, queueSender, pdUnit, granularity)
		}
		if viper.GetBool("model.enabled") && queueJobType == "modelJobSendIntervalSec" {
			scope.Infof(
				"Start sending %v cluster model jobs to queue with granularity %v seconds.",
				len(clusters), granularity)
			dispatcher.modelJobSender.SendClusterModelJobs(clusters, queueSender, pdUnit, granularity,
				predictionStep)
		}
		scope.Infof(
			"Sending %v cluster jobs to queue completely with granularity %v seconds.",
			len(clusters), granularity)
	} else if pdUnit == UnitTypeController {
		res, err := datahubServiceClnt.ListControllers(context.Background(),
			&datahub_resources.ListControllersRequest{})
		if err != nil {
			scope.Errorf(
				"List controllers for model/predict job failed with granularity %v seconds. %s",
				granularity, err.Error())
			return
		}
		controllers := []*datahub_resources.Controller{}
		for _, ctrl := range res.GetControllers() {
			if granularity == 30 && (!viper.GetBool("hourlyPredict") &&
				ctrl.GetAlamedaControllerSpec().GetScalingTool() !=
					datahub_resources.ScalingTool_VPA) {
				continue
			}
			controllers = append(controllers, ctrl)
		}
		if queueJobType == "predictionJobSendIntervalSec" {
			scope.Infof(
				"Start sending %v controller prediction jobs to queue with granularity %v seconds.",
				len(controllers), granularity)
			dispatcher.predictJobSender.SendControllerPredictJobs(controllers, queueSender, pdUnit, granularity)
		}
		if viper.GetBool("model.enabled") && queueJobType == "modelJobSendIntervalSec" {
			scope.Infof(
				"Start sending %v controller model jobs to queue with granularity %v seconds.",
				len(controllers), granularity)
			dispatcher.modelJobSender.SendControllerModelJobs(controllers, queueSender, pdUnit, granularity,
				predictionStep)
		}
		scope.Infof("Sending %v controller jobs to queue completely with granularity %v seconds.",
			len(controllers), granularity)
	}
}

func (dispatcher *Dispatcher) skipJobSending(pdUnit string, granularitySec int64) bool {
	if pdUnit == UnitTypeGPU && granularitySec != 3600 {
		return true
	}

	return (pdUnit == UnitTypeCluster || pdUnit == UnitTypeNamespace) &&
		(granularitySec == 30 && !viper.GetBool("hourlyPredict"))
}
