package grpc

import (
	go_context "context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	autoscalingv1alpha1 "github.com/containers-ai/alameda/operator/pkg/apis/autoscaling/v1alpha1"
	"github.com/containers-ai/alameda/operator/pkg/controller/alamedaresource"
	"github.com/containers-ai/alameda/operator/pkg/utils/log"
	logUtil "github.com/containers-ai/alameda/operator/pkg/utils/log"
	operator_v1alpha1 "github.com/containers-ai/api/operator/v1alpha1"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	apicorev1 "k8s.io/api/core/v1"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type Service struct {
	Config  Config
	Manager manager.Manager
}

func NewService(c *Config, manager manager.Manager) (*Service, error) {

	// TODO: new metircs repository

	s := &Service{
		Config:  *c,
		Manager: manager,
	}

	return s, nil
}

func (s *Service) Open() error {

	// build server listener
	log.GetLogger().Info("starting gRPC server")
	ln, err := net.Listen("tcp", s.Config.BindAddress)
	if err != nil {
		log.GetLogger().Error(err, "gRPC server failed listen: "+err.Error())
		return fmt.Errorf("GRPC server failed to bind address: %s", s.Config.BindAddress)
	}
	log.GetLogger().Info("gRPC server listening on " + s.Config.BindAddress)

	// build gRPC server
	server, err := s.newGRPCServer()
	if err != nil {
		log.GetLogger().Error(err, err.Error())
		return err
	}

	// register gRPC server
	s.registGRPCServer(server)
	reflection.Register(server)

	// run gRPC server
	if err := server.Serve(ln); err != nil {
		return fmt.Errorf("GRPC server failed to serve: %s", err.Error())
	}

	return nil
}

func (s *Service) newGRPCServer() (*grpc.Server, error) {

	var (
		server *grpc.Server
	)

	server = grpc.NewServer()

	return server, nil
}

func (s *Service) registGRPCServer(server *grpc.Server) {

	operator_v1alpha1.RegisterOperatorServiceServer(server, s)
}

func (s *Service) Close() error {

	return nil
}

func (s *Service) ListMetrics(ctx context.Context, in *operator_v1alpha1.ListMetricsRequest) (*operator_v1alpha1.ListMetricsResponse, error) {
	return nil, nil
}
func (s *Service) ListMetricsSum(ctx context.Context, in *operator_v1alpha1.ListMetricsSumRequest) (*operator_v1alpha1.ListMetricsSumResponse, error) {
	return nil, nil
}

func (s *Service) CreatePredictResult(ctx context.Context, in *operator_v1alpha1.CreatePredictResultRequest) (*operator_v1alpha1.CreatePredictResultResponse, error) {
	// 1. Get namespace list information from predicted pods
	nsRange := map[string]bool{}
	for _, predictPod := range in.GetPredictPods() {
		if _, ok := nsRange[predictPod.GetNamespace()]; !ok {
			nsRange[predictPod.GetNamespace()] = true
		}
	}
	// 2. Get AlamedaResource list from namespace list
	alaListRange := []autoscalingv1alpha1.AlamedaResource{}
	for namespace, _ := range nsRange {
		alamedaresourceList := &autoscalingv1alpha1.AlamedaResourceList{}
		err := s.Manager.GetClient().List(go_context.TODO(), client.InNamespace(namespace), alamedaresourceList)
		if err == nil {
			alaListRange = append(alaListRange, alamedaresourceList.Items...)
		}
	}
	for _, ala := range alaListRange {
		alaAnno := ala.GetAnnotations()
		predictPodsInAla := []*operator_v1alpha1.PredictPod{}
		if alaAnno == nil {
			continue
		}
		if _, ok := alaAnno[alamedaresource.AlamedaK8sController]; !ok {
			continue
		}
		for _, predictPod := range in.GetPredictPods() {
			alaK8sCtrStr := alaAnno[alamedaresource.AlamedaK8sController]
			if isAlamedaPod(alaK8sCtrStr, predictPod.GetUid()) {
				predictPodsInAla = append(predictPodsInAla, predictPod)
			}
		}
		if len(predictPodsInAla) > 0 {
			s.updateAlamedaResourcePredict(ala, predictPodsInAla)
		}
	}

	return &operator_v1alpha1.CreatePredictResultResponse{}, nil
}

func (s *Service) updateAlamedaResourcePredict(ala autoscalingv1alpha1.AlamedaResource, predictPods []*operator_v1alpha1.PredictPod) {
	alamedaresourcePredict := &autoscalingv1alpha1.AlamedaResourcePrediction{}
	s.Manager.GetClient().Get(go_context.TODO(), types.NamespacedName{
		Name:      ala.GetName(),
		Namespace: ala.GetNamespace(),
	}, alamedaresourcePredict)
	for _, predictPod := range predictPods {
		for _, deployment := range alamedaresourcePredict.Status.Prediction.Deployments {
			if _, ok := deployment.Pods[autoscalingv1alpha1.PodUID(predictPod.GetUid())]; ok {
				for _, predictContainer := range predictPod.GetPredictContainers() {
					for containerName, _ := range deployment.Pods[autoscalingv1alpha1.PodUID(predictPod.GetUid())].Containers {
						if predictContainer.GetName() == string(containerName) {
							logUtil.GetLogger().Info(fmt.Sprintf("Update Prediction from AI service. (%s/%s)", ala.GetNamespace(), ala.GetName()))
							recommendation, initialResource := s.updateAlamedaResourcePredictContainer(deployment.Pods[autoscalingv1alpha1.PodUID(predictPod.GetUid())].Containers[containerName], predictContainer)
							alaPredictContainer := deployment.Pods[autoscalingv1alpha1.PodUID(predictPod.GetUid())].Containers[containerName]
							alaPredictContainer.Recommendations = recommendation
							alaPredictContainer.InitialResource = initialResource
							deployment.Pods[autoscalingv1alpha1.PodUID(predictPod.GetUid())].Containers[containerName] = alaPredictContainer
						}
					}
				}
			}
		}
	}

	s.Manager.GetClient().Update(context.TODO(), alamedaresourcePredict)
}

func (s *Service) updateAlamedaResourcePredictContainer(alaPredictContainer autoscalingv1alpha1.PredictContainer, predictContainer *operator_v1alpha1.PredictContainer) ([]autoscalingv1alpha1.Recommendation, apicorev1.ResourceRequirements) {
	for resource, predictData := range predictContainer.RowPredictData {
		tsData := autoscalingv1alpha1.TimeSeriesData{
			PredictData: []autoscalingv1alpha1.PredictData{},
		}
		for _, data := range predictData.GetPredictData() {
			date := time.Unix(data.Time.Seconds, 0)
			tsData.PredictData = append(tsData.PredictData, autoscalingv1alpha1.PredictData{
				Time:  data.Time.Seconds,
				Value: data.Value,
				Date:  date.String(),
			})
		}
		alaPredictContainer.RawPredict[autoscalingv1alpha1.ResourceType(resource)] = tsData
	}
	recommendations := []autoscalingv1alpha1.Recommendation{}
	for _, recommendation := range predictContainer.Recommendations {
		resource := apicorev1.ResourceRequirements{
			Limits:   map[apicorev1.ResourceName]apiresource.Quantity{},
			Requests: map[apicorev1.ResourceName]apiresource.Quantity{},
		}
		for limitKey, limit := range recommendation.Resource.Limit {
			resource.Limits[apicorev1.ResourceName(limitKey)] = apiresource.MustParse(limit)
		}
		for requestKey, request := range recommendation.Resource.Request {
			resource.Requests[apicorev1.ResourceName(requestKey)] = apiresource.MustParse(request)
		}
		date := time.Unix(recommendation.Time.Seconds, 0)
		recommendations = append(recommendations, autoscalingv1alpha1.Recommendation{
			Time:      recommendation.Time.Seconds,
			Date:      date.String(),
			Resources: resource,
		})
	}

	alaPredictContainer.Recommendations = recommendations

	initialResource := apicorev1.ResourceRequirements{
		Limits:   map[apicorev1.ResourceName]apiresource.Quantity{},
		Requests: map[apicorev1.ResourceName]apiresource.Quantity{},
	}
	for limitKey, limit := range predictContainer.InitialResource.Limit {
		initialResource.Limits[apicorev1.ResourceName(limitKey)] = apiresource.MustParse(limit)
	}
	for requestKey, request := range predictContainer.InitialResource.Request {
		initialResource.Requests[apicorev1.ResourceName(requestKey)] = apiresource.MustParse(request)
	}
	alaPredictContainer.InitialResource = initialResource
	return recommendations, initialResource
}

func isAlamedaPod(alaK8sCtrAnnoStr, podUid string) bool {
	akcMap := alamedaresource.GetDefaultAlamedaK8SControllerAnno()
	err := json.Unmarshal([]byte(alaK8sCtrAnnoStr), akcMap)
	if err != nil {
		return false
	}
	for _, deployment := range akcMap.DeploymentMap {
		if _, ok := deployment.PodMap[podUid]; ok {
			return true
		}
	}
	return false
}
