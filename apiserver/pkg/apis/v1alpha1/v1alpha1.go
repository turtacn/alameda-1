package v1alpha1

import (
	Accounts "github.com/containers-ai/alameda/apiserver/pkg/apis/accounts"
	RepoDatahub "github.com/containers-ai/alameda/apiserver/pkg/repositories/datahub"
	Log "github.com/containers-ai/alameda/pkg/utils/log"
	DatahubV1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	scope = Log.RegisterScope("apiserver", "apiserver log", 0)
)

type ServiceV1alpha1 struct {
	Target string
}

func NewServiceV1alpha1() *ServiceV1alpha1 {
	service := ServiceV1alpha1{}
	return &service
}

// ListPodMetrics list pods' metrics
func (s *ServiceV1alpha1) ListPodMetrics(ctx context.Context, in *DatahubV1alpha1.ListPodMetricsRequest) (*DatahubV1alpha1.ListPodMetricsResponse, error) {
	scope.Debug("Request received from ListPodMetrics grpc function")

	out := new(DatahubV1alpha1.ListPodMetricsResponse)

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}
		return out, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}
		return out, nil
	}
	defer conn.Close()

	return client.ListPodMetrics(ctx, in)
}

// ListNodeMetrics list nodes' metrics
func (s *ServiceV1alpha1) ListNodeMetrics(ctx context.Context, in *DatahubV1alpha1.ListNodeMetricsRequest) (*DatahubV1alpha1.ListNodeMetricsResponse, error) {
	scope.Debug("Request received from ListNodeMetrics grpc function")

	out := new(DatahubV1alpha1.ListNodeMetricsResponse)

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}
		return out, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}
		return out, nil
	}
	defer conn.Close()

	return client.ListNodeMetrics(ctx, in)
}

// ListAlamedaPods returns predicted pods
func (s *ServiceV1alpha1) ListAlamedaPods(ctx context.Context, in *DatahubV1alpha1.ListAlamedaPodsRequest) (*DatahubV1alpha1.ListPodsResponse, error) {
	scope.Debug("Request received from ListAlamedaPods grpc function")

	out := new(DatahubV1alpha1.ListPodsResponse)

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}
		return out, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}
		return out, nil
	}
	defer conn.Close()

	return client.ListAlamedaPods(ctx, in)
}

// ListAlamedaNodes list nodes in cluster
func (s *ServiceV1alpha1) ListAlamedaNodes(ctx context.Context, in *DatahubV1alpha1.ListAlamedaNodesRequest) (*DatahubV1alpha1.ListNodesResponse, error) {
	scope.Debug("Request received from ListAlamedaNodes grpc function")

	out := new(DatahubV1alpha1.ListNodesResponse)

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}
		return out, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}
		return out, nil
	}
	defer conn.Close()

	return client.ListAlamedaNodes(ctx, in)
}

func (s *ServiceV1alpha1) ListNodes(ctx context.Context, in *DatahubV1alpha1.ListNodesRequest) (*DatahubV1alpha1.ListNodesResponse, error) {
	scope.Debug("Request received from ListNodes grpc function")

	out := new(DatahubV1alpha1.ListNodesResponse)

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}
		return out, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}
		return out, nil
	}
	defer conn.Close()

	return client.ListNodes(ctx, in)
}

func (s *ServiceV1alpha1) ListControllers(ctx context.Context, in *DatahubV1alpha1.ListControllersRequest) (*DatahubV1alpha1.ListControllersResponse, error) {
	scope.Debug("Request received from ListControllers grpc function")

	out := new(DatahubV1alpha1.ListControllersResponse)

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}
		return out, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}
		return out, nil
	}
	defer conn.Close()

	return client.ListControllers(ctx, in)
}

// ListPodPredictions list pods' predictions
func (s *ServiceV1alpha1) ListPodPredictions(ctx context.Context, in *DatahubV1alpha1.ListPodPredictionsRequest) (*DatahubV1alpha1.ListPodPredictionsResponse, error) {
	scope.Debug("Request received from ListPodPredictions grpc function")

	out := new(DatahubV1alpha1.ListPodPredictionsResponse)

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}
		return out, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}
		return out, nil
	}
	defer conn.Close()

	return client.ListPodPredictions(ctx, in)
}

// ListNodePredictions list nodes' predictions
func (s *ServiceV1alpha1) ListNodePredictions(ctx context.Context, in *DatahubV1alpha1.ListNodePredictionsRequest) (*DatahubV1alpha1.ListNodePredictionsResponse, error) {
	scope.Debug("Request received from ListNodePredictions grpc function")

	out := new(DatahubV1alpha1.ListNodePredictionsResponse)

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}
		return out, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}
		return out, nil
	}
	defer conn.Close()

	return client.ListNodePredictions(ctx, in)
}

// ListPodRecommendations list pod recommendations
func (s *ServiceV1alpha1) ListPodRecommendations(ctx context.Context, in *DatahubV1alpha1.ListPodRecommendationsRequest) (*DatahubV1alpha1.ListPodRecommendationsResponse, error) {
	scope.Debug("Request received from ListPodRecommendations grpc function")

	out := new(DatahubV1alpha1.ListPodRecommendationsResponse)

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}
		return out, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}
		return out, nil
	}
	defer conn.Close()

	return client.ListPodRecommendations(ctx, in)
}

// ListAvailablePodRecommendations list pod recommendations
func (s *ServiceV1alpha1) ListAvailablePodRecommendations(ctx context.Context, in *DatahubV1alpha1.ListPodRecommendationsRequest) (*DatahubV1alpha1.ListPodRecommendationsResponse, error) {
	scope.Debug("Request received from ListAvailablePodRecommendations grpc function")

	out := new(DatahubV1alpha1.ListPodRecommendationsResponse)

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}
		return out, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}
		return out, nil
	}
	defer conn.Close()

	return client.ListAvailablePodRecommendations(ctx, in)
}

// ListControllerRecommendations list controller recommendations
func (s *ServiceV1alpha1) ListControllerRecommendations(ctx context.Context, in *DatahubV1alpha1.ListControllerRecommendationsRequest) (*DatahubV1alpha1.ListControllerRecommendationsResponse, error) {
	scope.Debug("Request received from ListControllerRecommendations grpc function")

	out := new(DatahubV1alpha1.ListControllerRecommendationsResponse)

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}
		return out, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}
		return out, nil
	}
	defer conn.Close()

	return client.ListControllerRecommendations(ctx, in)
}

// ListPodsByNodeName list pods running on specific nodes
func (s *ServiceV1alpha1) ListPodsByNodeName(ctx context.Context, in *DatahubV1alpha1.ListPodsByNodeNamesRequest) (*DatahubV1alpha1.ListPodsResponse, error) {
	scope.Debug("Request received from ListPodsByNodeName grpc function")

	out := new(DatahubV1alpha1.ListPodsResponse)

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}
		return out, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}
		return out, nil
	}
	defer conn.Close()

	return client.ListPodsByNodeName(ctx, in)
}

// ListSimulatedSchedulingScores list simulated scheduling scores
func (s *ServiceV1alpha1) ListSimulatedSchedulingScores(ctx context.Context, in *DatahubV1alpha1.ListSimulatedSchedulingScoresRequest) (*DatahubV1alpha1.ListSimulatedSchedulingScoresResponse, error) {
	scope.Debug("Request received from ListSimulatedSchedulingScores grpc function")

	out := new(DatahubV1alpha1.ListSimulatedSchedulingScoresResponse)

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}
		return out, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}
		return out, nil
	}
	defer conn.Close()

	return client.ListSimulatedSchedulingScores(ctx, in)
}

// CreatePods add containers information of pods to database
func (s *ServiceV1alpha1) CreatePods(ctx context.Context, in *DatahubV1alpha1.CreatePodsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreatePods grpc function")

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}
	defer conn.Close()

	return client.CreatePods(ctx, in)
}

func (s *ServiceV1alpha1) CreateControllers(ctx context.Context, in *DatahubV1alpha1.CreateControllersRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateControllers grpc function")

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}
	defer conn.Close()

	return client.CreateControllers(ctx, in)
}

func (s *ServiceV1alpha1) DeleteControllers(ctx context.Context, in *DatahubV1alpha1.DeleteControllersRequest) (*status.Status, error) {
	scope.Debug("Request received from DeleteControllers grpc function")

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}
	defer conn.Close()

	return client.DeleteControllers(ctx, in)
}

// DeletePods update containers information of pods to database
func (s *ServiceV1alpha1) DeletePods(ctx context.Context, in *DatahubV1alpha1.DeletePodsRequest) (*status.Status, error) {
	scope.Debug("Request received from DeletePods grpc function")

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}
	defer conn.Close()

	return client.DeletePods(ctx, in)
}

// CreateAlamedaNodes add node information to database
func (s *ServiceV1alpha1) CreateAlamedaNodes(ctx context.Context, in *DatahubV1alpha1.CreateAlamedaNodesRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateAlamedaNodes grpc function")

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}
	defer conn.Close()

	return client.CreateAlamedaNodes(ctx, in)
}

// CreatePodPredictions add pod predictions information to database
func (s *ServiceV1alpha1) CreatePodPredictions(ctx context.Context, in *DatahubV1alpha1.CreatePodPredictionsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreatePodPredictions grpc function")

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}
	defer conn.Close()

	return client.CreatePodPredictions(ctx, in)
}

// CreateNodePredictions add node predictions information to database
func (s *ServiceV1alpha1) CreateNodePredictions(ctx context.Context, in *DatahubV1alpha1.CreateNodePredictionsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateNodePredictions grpc function")

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}
	defer conn.Close()

	return client.CreateNodePredictions(ctx, in)
}

// CreatePodRecommendations add pod recommendations information to database
func (s *ServiceV1alpha1) CreatePodRecommendations(ctx context.Context, in *DatahubV1alpha1.CreatePodRecommendationsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreatePodRecommendations grpc function")

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}
	defer conn.Close()

	return client.CreatePodRecommendations(ctx, in)
}

// CreatePodRecommendations add pod recommendations information to database
func (s *ServiceV1alpha1) CreateControllerRecommendations(ctx context.Context, in *DatahubV1alpha1.CreateControllerRecommendationsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateControllerRecommendations grpc function")

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}
	defer conn.Close()

	return client.CreateControllerRecommendations(ctx, in)
}

// CreateSimulatedSchedulingScores add simulated scheduling scores to database
func (s *ServiceV1alpha1) CreateSimulatedSchedulingScores(ctx context.Context, in *DatahubV1alpha1.CreateSimulatedSchedulingScoresRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateSimulatedSchedulingScores grpc function")

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}
	defer conn.Close()

	return client.CreateSimulatedSchedulingScores(ctx, in)
}

// DeleteAlamedaNodes remove node information to database
func (s *ServiceV1alpha1) DeleteAlamedaNodes(ctx context.Context, in *DatahubV1alpha1.DeleteAlamedaNodesRequest) (*status.Status, error) {
	scope.Debug("Request received from DeleteAlamedaNodes grpc function")

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}
	defer conn.Close()

	return client.DeleteAlamedaNodes(ctx, in)
}

// Read rawdata from database
func (s *ServiceV1alpha1) ReadRawdata(ctx context.Context, in *DatahubV1alpha1.ReadRawdataRequest) (*DatahubV1alpha1.ReadRawdataResponse, error) {
	scope.Debug("Request received from ReadRawdata grpc function")

	out := new(DatahubV1alpha1.ReadRawdataResponse)

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}
		return out, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}
		return out, nil
	}
	defer conn.Close()

	return client.ReadRawdata(ctx, in)
}

// Write rawdata to database
func (s *ServiceV1alpha1) WriteRawdata(ctx context.Context, in *DatahubV1alpha1.WriteRawdataRequest) (*status.Status, error) {
	scope.Debug("Request received from WriteRawdata grpc function")

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}
	defer conn.Close()

	return client.WriteRawdata(ctx, in)
}

func (s *ServiceV1alpha1) Ping(ctx context.Context, in *empty.Empty) (*status.Status, error) {
	scope.Debug("Request received from Ping grpc function")

	return &status.Status{Code: int32(code.Code_OK)}, nil
}

func (s *ServiceV1alpha1) ListWeaveScopeHosts(ctx context.Context, in *DatahubV1alpha1.ListWeaveScopeHostsRequest) (*DatahubV1alpha1.WeaveScopeResponse, error) {
	scope.Debug("Request received from ListWeaveScopeHosts grpc function")

	out := new(DatahubV1alpha1.WeaveScopeResponse)

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}
		return out, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}
		return out, nil
	}
	defer conn.Close()

	return client.ListWeaveScopeHosts(ctx, in)
}

func (s *ServiceV1alpha1) GetWeaveScopeHostDetails(ctx context.Context, in *DatahubV1alpha1.ListWeaveScopeHostsRequest) (*DatahubV1alpha1.WeaveScopeResponse, error) {
	scope.Debug("Request received from GetWeaveScopeHostDetails grpc function")

	out := new(DatahubV1alpha1.WeaveScopeResponse)

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}
		return out, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}
		return out, nil
	}
	defer conn.Close()

	return client.GetWeaveScopeHostDetails(ctx, in)
}

func (s *ServiceV1alpha1) ListWeaveScopePods(ctx context.Context, in *DatahubV1alpha1.ListWeaveScopePodsRequest) (*DatahubV1alpha1.WeaveScopeResponse, error) {
	scope.Debug("Request received from ListWeaveScopePods grpc function")

	out := new(DatahubV1alpha1.WeaveScopeResponse)

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}
		return out, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}
		return out, nil
	}
	defer conn.Close()

	return client.ListWeaveScopePods(ctx, in)
}

func (s *ServiceV1alpha1) GetWeaveScopePodDetails(ctx context.Context, in *DatahubV1alpha1.ListWeaveScopePodsRequest) (*DatahubV1alpha1.WeaveScopeResponse, error) {
	scope.Debug("Request received from GetWeaveScopePodDetails grpc function")

	out := new(DatahubV1alpha1.WeaveScopeResponse)

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}
		return out, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}
		return out, nil
	}
	defer conn.Close()

	return client.GetWeaveScopePodDetails(ctx, in)
}

func (s *ServiceV1alpha1) ListWeaveScopeContainers(ctx context.Context, in *DatahubV1alpha1.ListWeaveScopeContainersRequest) (*DatahubV1alpha1.WeaveScopeResponse, error) {
	scope.Debug("Request received from ListWeaveScopeContainers grpc function")

	out := new(DatahubV1alpha1.WeaveScopeResponse)

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}
		return out, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}
		return out, nil
	}
	defer conn.Close()

	return client.ListWeaveScopeContainers(ctx, in)
}

func (s *ServiceV1alpha1) ListWeaveScopeContainersByHostname(ctx context.Context, in *DatahubV1alpha1.ListWeaveScopeContainersRequest) (*DatahubV1alpha1.WeaveScopeResponse, error) {
	scope.Debug("Request received from ListWeaveScopeContainersByHostname grpc function")

	out := new(DatahubV1alpha1.WeaveScopeResponse)

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}
		return out, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}
		return out, nil
	}
	defer conn.Close()

	return client.ListWeaveScopeContainersByHostname(ctx, in)
}

func (s *ServiceV1alpha1) ListWeaveScopeContainersByImage(ctx context.Context, in *DatahubV1alpha1.ListWeaveScopeContainersRequest) (*DatahubV1alpha1.WeaveScopeResponse, error) {
	scope.Debug("Request received from ListWeaveScopeContainersByImage grpc function")

	out := new(DatahubV1alpha1.WeaveScopeResponse)

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}
		return out, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}
		return out, nil
	}
	defer conn.Close()

	return client.ListWeaveScopeContainersByImage(ctx, in)
}

func (s *ServiceV1alpha1) GetWeaveScopeContainerDetails(ctx context.Context, in *DatahubV1alpha1.ListWeaveScopeContainersRequest) (*DatahubV1alpha1.WeaveScopeResponse, error) {
	scope.Debug("Request received from GetWeaveScopeContainerDetails grpc function")

	out := new(DatahubV1alpha1.WeaveScopeResponse)

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}
		return out, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}
		return out, nil
	}
	defer conn.Close()

	return client.GetWeaveScopeContainerDetails(ctx, in)
}

func (s *ServiceV1alpha1) CreateEvents(ctx context.Context, in *DatahubV1alpha1.CreateEventsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateEvents grpc function")

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}
	defer conn.Close()

	return client.CreateEvents(ctx, in)
}

func (s *ServiceV1alpha1) ListEvents(ctx context.Context, in *DatahubV1alpha1.ListEventsRequest) (*DatahubV1alpha1.ListEventsResponse, error) {
	scope.Debug("Request received from ListEvents grpc function")

	out := new(DatahubV1alpha1.ListEventsResponse)

	// Do authentication
	_, err := Accounts.Authenticate(ctx)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: err.Error(),
		}
		return out, nil
	}

	// Call datahub api
	conn, client, err := RepoDatahub.CreateClient(s.Target)
	if err != nil {
		out.Status = &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}
		return out, nil
	}
	defer conn.Close()

	return client.ListEvents(ctx, in)
}
