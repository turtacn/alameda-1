package v1alpha1

import (
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/status"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	DatapipeConfig "github.com/containers-ai/alameda/datapipe/pkg/config"
	Log "github.com/containers-ai/alameda/pkg/utils/log"
)

var (
	scope = Log.RegisterScope("datapipe", "datapipe log", 0)
)

type ServiceV1alpha1 struct {
	Config *DatapipeConfig.Config
}

func NewServiceV1alpha1(cfg *DatapipeConfig.Config) *ServiceV1alpha1 {
	service := ServiceV1alpha1{}
	service.Config = cfg
	return &service
}

// ListPodMetrics list pods' metrics
func (s *ServiceV1alpha1) ListPodMetrics(ctx context.Context, in *datahub_v1alpha1.ListPodMetricsRequest) (*datahub_v1alpha1.ListPodMetricsResponse, error) {
	out := new(datahub_v1alpha1.ListPodMetricsResponse)
	return out, nil
}

// ListNodeMetrics list nodes' metrics
func (s *ServiceV1alpha1) ListNodeMetrics(ctx context.Context, in *datahub_v1alpha1.ListNodeMetricsRequest) (*datahub_v1alpha1.ListNodeMetricsResponse, error) {
	out := new(datahub_v1alpha1.ListNodeMetricsResponse)
	return out, nil
}

// ListAlamedaPods returns predicted pods
func (s *ServiceV1alpha1) ListAlamedaPods(ctx context.Context, in *datahub_v1alpha1.ListAlamedaPodsRequest) (*datahub_v1alpha1.ListPodsResponse, error) {
	out := new(datahub_v1alpha1.ListPodsResponse)
	return out, nil
}

// ListAlamedaNodes list nodes in cluster
func (s *ServiceV1alpha1) ListAlamedaNodes(ctx context.Context, in *datahub_v1alpha1.ListAlamedaNodesRequest) (*datahub_v1alpha1.ListNodesResponse, error) {
	out := new(datahub_v1alpha1.ListNodesResponse)
	return out, nil
}

func (s *ServiceV1alpha1) ListNodes(ctx context.Context, in *datahub_v1alpha1.ListNodesRequest) (*datahub_v1alpha1.ListNodesResponse, error) {
	out := new(datahub_v1alpha1.ListNodesResponse)
	return out, nil
}

func (s *ServiceV1alpha1) ListControllers(ctx context.Context, in *datahub_v1alpha1.ListControllersRequest) (*datahub_v1alpha1.ListControllersResponse, error) {
	out := new(datahub_v1alpha1.ListControllersResponse)
	return out, nil
}

// ListPodPredictions list pods' predictions
func (s *ServiceV1alpha1) ListPodPredictions(ctx context.Context, in *datahub_v1alpha1.ListPodPredictionsRequest) (*datahub_v1alpha1.ListPodPredictionsResponse, error) {
	out := new(datahub_v1alpha1.ListPodPredictionsResponse)
	return out, nil
}

// ListPodPredictions list pods' predictions for demo
func (s *ServiceV1alpha1) ListPodPredictionsDemo(ctx context.Context, in *datahub_v1alpha1.ListPodPredictionsRequest) (*datahub_v1alpha1.ListPodPredictionsResponse, error) {
	out := new(datahub_v1alpha1.ListPodPredictionsResponse)
	return out, nil
}

// ListNodePredictions list nodes' predictions
func (s *ServiceV1alpha1) ListNodePredictions(ctx context.Context, in *datahub_v1alpha1.ListNodePredictionsRequest) (*datahub_v1alpha1.ListNodePredictionsResponse, error) {
	out := new(datahub_v1alpha1.ListNodePredictionsResponse)
	return out, nil
}

// ListPodRecommendations list pod recommendations
func (s *ServiceV1alpha1) ListPodRecommendations(ctx context.Context, in *datahub_v1alpha1.ListPodRecommendationsRequest) (*datahub_v1alpha1.ListPodRecommendationsResponse, error) {
	out := new(datahub_v1alpha1.ListPodRecommendationsResponse)
	return out, nil
}

// ListAvailablePodRecommendations list pod recommendations
func (s *ServiceV1alpha1) ListAvailablePodRecommendations(ctx context.Context, in *datahub_v1alpha1.ListPodRecommendationsRequest) (*datahub_v1alpha1.ListPodRecommendationsResponse, error) {
	out := new(datahub_v1alpha1.ListPodRecommendationsResponse)
	return out, nil
}

// ListControllerRecommendations list controller recommendations
func (s *ServiceV1alpha1) ListControllerRecommendations(ctx context.Context, in *datahub_v1alpha1.ListControllerRecommendationsRequest) (*datahub_v1alpha1.ListControllerRecommendationsResponse, error) {
	out := new(datahub_v1alpha1.ListControllerRecommendationsResponse)
	return out, nil
}

// ListPodsByNodeName list pods running on specific nodes
func (s *ServiceV1alpha1) ListPodsByNodeName(ctx context.Context, in *datahub_v1alpha1.ListPodsByNodeNamesRequest) (*datahub_v1alpha1.ListPodsResponse, error) {
	out := new(datahub_v1alpha1.ListPodsResponse)
	return out, nil
}

// ListSimulatedSchedulingScores list simulated scheduling scores
func (s *ServiceV1alpha1) ListSimulatedSchedulingScores(ctx context.Context, in *datahub_v1alpha1.ListSimulatedSchedulingScoresRequest) (*datahub_v1alpha1.ListSimulatedSchedulingScoresResponse, error) {
	out := new(datahub_v1alpha1.ListSimulatedSchedulingScoresResponse)
	return out, nil
}

// CreatePods add containers information of pods to database
func (s *ServiceV1alpha1) CreatePods(ctx context.Context, in *datahub_v1alpha1.CreatePodsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreatePods grpc function")

	return s.CreatePodsImpl(in)
}

func (s *ServiceV1alpha1) CreateControllers(ctx context.Context, in *datahub_v1alpha1.CreateControllersRequest) (*status.Status, error) {
	out := new(status.Status)
	return out, nil
}

func (s *ServiceV1alpha1) DeleteControllers(ctx context.Context, in *datahub_v1alpha1.DeleteControllersRequest) (*status.Status, error) {
	out := new(status.Status)
	return out, nil
}

// DeletePods update containers information of pods to database
func (s *ServiceV1alpha1) DeletePods(ctx context.Context, in *datahub_v1alpha1.DeletePodsRequest) (*status.Status, error) {
	out := new(status.Status)
	return out, nil
}

// CreateAlamedaNodes add node information to database
func (s *ServiceV1alpha1) CreateAlamedaNodes(ctx context.Context, in *datahub_v1alpha1.CreateAlamedaNodesRequest) (*status.Status, error) {
	out := new(status.Status)
	return out, nil
}

// CreatePodPredictions add pod predictions information to database
func (s *ServiceV1alpha1) CreatePodPredictions(ctx context.Context, in *datahub_v1alpha1.CreatePodPredictionsRequest) (*status.Status, error) {
	out := new(status.Status)
	return out, nil
}

// CreateNodePredictions add node predictions information to database
func (s *ServiceV1alpha1) CreateNodePredictions(ctx context.Context, in *datahub_v1alpha1.CreateNodePredictionsRequest) (*status.Status, error) {
	out := new(status.Status)
	return out, nil
}

// CreatePodRecommendations add pod recommendations information to database
func (s *ServiceV1alpha1) CreatePodRecommendations(ctx context.Context, in *datahub_v1alpha1.CreatePodRecommendationsRequest) (*status.Status, error) {
	out := new(status.Status)
	return out, nil
}

// CreatePodRecommendations add pod recommendations information to database
func (s *ServiceV1alpha1) CreateControllerRecommendations(ctx context.Context, in *datahub_v1alpha1.CreateControllerRecommendationsRequest) (*status.Status, error) {
	out := new(status.Status)
	return out, nil
}

// CreateSimulatedSchedulingScores add simulated scheduling scores to database
func (s *ServiceV1alpha1) CreateSimulatedSchedulingScores(ctx context.Context, in *datahub_v1alpha1.CreateSimulatedSchedulingScoresRequest) (*status.Status, error) {
	out := new(status.Status)
	return out, nil
}

// DeleteAlamedaNodes remove node information to database
func (s *ServiceV1alpha1) DeleteAlamedaNodes(ctx context.Context, in *datahub_v1alpha1.DeleteAlamedaNodesRequest) (*status.Status, error) {
	out := new(status.Status)
	return out, nil
}

// Read rawdata from database
func (s *ServiceV1alpha1) ReadRawdata(ctx context.Context, in *datahub_v1alpha1.ReadRawdataRequest) (*datahub_v1alpha1.ReadRawdataResponse, error) {
	out := new(datahub_v1alpha1.ReadRawdataResponse)
	return out, nil
}

// Write rawdata to database
func (s *ServiceV1alpha1) WriteRawdata(ctx context.Context, in *datahub_v1alpha1.WriteRawdataRequest) (*status.Status, error) {
	out := new(status.Status)
	return out, nil
}
