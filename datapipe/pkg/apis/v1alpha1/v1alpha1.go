package v1alpha1

import (
	RepoApiServer "github.com/containers-ai/alameda/datapipe/pkg/repositories/apiserver"
	Log "github.com/containers-ai/alameda/pkg/utils/log"
	DatahubV1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/genproto/googleapis/rpc/code"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	scope = Log.RegisterScope("datapipe", "datapipe log", 0)
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

	conn, client, err := RepoApiServer.CreateClient(s.Target)
	if err != nil {
		return out, nil
	}
	defer conn.Close()

	// Send to API server
	out, err = client.ListPodMetrics(RepoApiServer.NewContextWithCredential(), in)

	// Check if needs to resend request
	if out != nil {
		if RepoApiServer.NeedResendRequest(out.GetStatus(), err) {
			out, err = client.ListPodMetrics(RepoApiServer.NewContextWithCredential(), in)
		}
	}

	if err != nil {
		return &DatahubV1alpha1.ListPodMetricsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	return out, nil
}

// ListNodeMetrics list nodes' metrics
func (s *ServiceV1alpha1) ListNodeMetrics(ctx context.Context, in *DatahubV1alpha1.ListNodeMetricsRequest) (*DatahubV1alpha1.ListNodeMetricsResponse, error) {
	scope.Debug("Request received from ListNodeMetrics grpc function")

	out := new(DatahubV1alpha1.ListNodeMetricsResponse)

	conn, client, err := RepoApiServer.CreateClient(s.Target)
	if err != nil {
		return out, nil
	}
	defer conn.Close()

	// Send to API server
	out, err = client.ListNodeMetrics(RepoApiServer.NewContextWithCredential(), in)

	// Check if needs to resend request
	if out != nil {
		if RepoApiServer.NeedResendRequest(out.GetStatus(), err) {
			out, err = client.ListNodeMetrics(RepoApiServer.NewContextWithCredential(), in)
		}
	}

	if err != nil {
		return &DatahubV1alpha1.ListNodeMetricsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	return out, nil
}

// ListAlamedaPods returns predicted pods
func (s *ServiceV1alpha1) ListAlamedaPods(ctx context.Context, in *DatahubV1alpha1.ListAlamedaPodsRequest) (*DatahubV1alpha1.ListPodsResponse, error) {
	scope.Debug("Request received from ListAlamedaPods grpc function")

	out := new(DatahubV1alpha1.ListPodsResponse)

	conn, client, err := RepoApiServer.CreateClient(s.Target)
	if err != nil {
		return out, nil
	}
	defer conn.Close()

	// Send to API server
	out, err = client.ListAlamedaPods(RepoApiServer.NewContextWithCredential(), in)

	// Check if needs to resend request
	if out != nil {
		if RepoApiServer.NeedResendRequest(out.GetStatus(), err) {
			out, err = client.ListAlamedaPods(RepoApiServer.NewContextWithCredential(), in)
		}
	}

	if err != nil {
		return &DatahubV1alpha1.ListPodsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	return out, nil
}

// ListAlamedaNodes list nodes in cluster
func (s *ServiceV1alpha1) ListAlamedaNodes(ctx context.Context, in *DatahubV1alpha1.ListAlamedaNodesRequest) (*DatahubV1alpha1.ListNodesResponse, error) {
	scope.Debug("Request received from ListAlamedaNodes grpc function")

	out := new(DatahubV1alpha1.ListNodesResponse)

	conn, client, err := RepoApiServer.CreateClient(s.Target)
	if err != nil {
		return out, nil
	}
	defer conn.Close()

	// Send to API server
	out, err = client.ListAlamedaNodes(RepoApiServer.NewContextWithCredential(), in)

	// Check if needs to resend request
	if out != nil {
		if RepoApiServer.NeedResendRequest(out.GetStatus(), err) {
			out, err = client.ListAlamedaNodes(RepoApiServer.NewContextWithCredential(), in)
		}
	}

	if err != nil {
		return &DatahubV1alpha1.ListNodesResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	return out, nil
}

func (s *ServiceV1alpha1) ListNodes(ctx context.Context, in *DatahubV1alpha1.ListNodesRequest) (*DatahubV1alpha1.ListNodesResponse, error) {
	scope.Debug("Request received from ListNodes grpc function")

	out := new(DatahubV1alpha1.ListNodesResponse)

	conn, client, err := RepoApiServer.CreateClient(s.Target)
	if err != nil {
		return out, nil
	}
	defer conn.Close()

	// Send to API server
	out, err = client.ListNodes(RepoApiServer.NewContextWithCredential(), in)

	// Check if needs to resend request
	if out != nil {
		if RepoApiServer.NeedResendRequest(out.GetStatus(), err) {
			out, err = client.ListNodes(RepoApiServer.NewContextWithCredential(), in)
		}
	}

	if err != nil {
		return &DatahubV1alpha1.ListNodesResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	return out, nil
}

func (s *ServiceV1alpha1) ListControllers(ctx context.Context, in *DatahubV1alpha1.ListControllersRequest) (*DatahubV1alpha1.ListControllersResponse, error) {
	scope.Debug("Request received from ListControllers grpc function")

	out := new(DatahubV1alpha1.ListControllersResponse)

	conn, client, err := RepoApiServer.CreateClient(s.Target)
	if err != nil {
		return out, nil
	}
	defer conn.Close()

	// Send to API server
	out, err = client.ListControllers(RepoApiServer.NewContextWithCredential(), in)

	// Check if needs to resend request
	if out != nil {
		if RepoApiServer.NeedResendRequest(out.GetStatus(), err) {
			out, err = client.ListControllers(RepoApiServer.NewContextWithCredential(), in)
		}
	}

	if err != nil {
		return &DatahubV1alpha1.ListControllersResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	return out, nil
}

// ListPodPredictions list pods' predictions
func (s *ServiceV1alpha1) ListPodPredictions(ctx context.Context, in *DatahubV1alpha1.ListPodPredictionsRequest) (*DatahubV1alpha1.ListPodPredictionsResponse, error) {
	scope.Debug("Request received from ListPodPredictions grpc function")

	out := new(DatahubV1alpha1.ListPodPredictionsResponse)

	conn, client, err := RepoApiServer.CreateClient(s.Target)
	if err != nil {
		return out, nil
	}
	defer conn.Close()

	// Send to API server
	out, err = client.ListPodPredictions(RepoApiServer.NewContextWithCredential(), in)

	// Check if needs to resend request
	if out != nil {
		if RepoApiServer.NeedResendRequest(out.GetStatus(), err) {
			out, err = client.ListPodPredictions(RepoApiServer.NewContextWithCredential(), in)
		}
	}

	if err != nil {
		return &DatahubV1alpha1.ListPodPredictionsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	return out, nil
}

// ListNodePredictions list nodes' predictions
func (s *ServiceV1alpha1) ListNodePredictions(ctx context.Context, in *DatahubV1alpha1.ListNodePredictionsRequest) (*DatahubV1alpha1.ListNodePredictionsResponse, error) {
	scope.Debug("Request received from ListNodePredictions grpc function")

	out := new(DatahubV1alpha1.ListNodePredictionsResponse)

	conn, client, err := RepoApiServer.CreateClient(s.Target)
	if err != nil {
		return out, nil
	}
	defer conn.Close()

	// Send to API server
	out, err = client.ListNodePredictions(RepoApiServer.NewContextWithCredential(), in)

	// Check if needs to resend request
	if out != nil {
		if RepoApiServer.NeedResendRequest(out.GetStatus(), err) {
			out, err = client.ListNodePredictions(RepoApiServer.NewContextWithCredential(), in)
		}
	}

	if err != nil {
		return &DatahubV1alpha1.ListNodePredictionsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	return out, nil
}

// ListPodRecommendations list pod recommendations
func (s *ServiceV1alpha1) ListPodRecommendations(ctx context.Context, in *DatahubV1alpha1.ListPodRecommendationsRequest) (*DatahubV1alpha1.ListPodRecommendationsResponse, error) {
	scope.Debug("Request received from ListPodRecommendations grpc function")

	out := new(DatahubV1alpha1.ListPodRecommendationsResponse)

	conn, client, err := RepoApiServer.CreateClient(s.Target)
	if err != nil {
		return out, nil
	}
	defer conn.Close()

	// Send to API server
	out, err = client.ListPodRecommendations(RepoApiServer.NewContextWithCredential(), in)

	// Check if needs to resend request
	if out != nil {
		if RepoApiServer.NeedResendRequest(out.GetStatus(), err) {
			out, err = client.ListPodRecommendations(RepoApiServer.NewContextWithCredential(), in)
		}
	}

	if err != nil {
		return &DatahubV1alpha1.ListPodRecommendationsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	return out, nil
}

// ListAvailablePodRecommendations list pod recommendations
func (s *ServiceV1alpha1) ListAvailablePodRecommendations(ctx context.Context, in *DatahubV1alpha1.ListPodRecommendationsRequest) (*DatahubV1alpha1.ListPodRecommendationsResponse, error) {
	scope.Debug("Request received from ListAvailablePodRecommendations grpc function")

	out := new(DatahubV1alpha1.ListPodRecommendationsResponse)

	conn, client, err := RepoApiServer.CreateClient(s.Target)
	if err != nil {
		return out, nil
	}
	defer conn.Close()

	// Send to API server
	out, err = client.ListAvailablePodRecommendations(RepoApiServer.NewContextWithCredential(), in)

	// Check if needs to resend request
	if out != nil {
		if RepoApiServer.NeedResendRequest(out.GetStatus(), err) {
			out, err = client.ListAvailablePodRecommendations(RepoApiServer.NewContextWithCredential(), in)
		}
	}

	if err != nil {
		return &DatahubV1alpha1.ListPodRecommendationsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	return out, nil
}

// ListControllerRecommendations list controller recommendations
func (s *ServiceV1alpha1) ListControllerRecommendations(ctx context.Context, in *DatahubV1alpha1.ListControllerRecommendationsRequest) (*DatahubV1alpha1.ListControllerRecommendationsResponse, error) {
	scope.Debug("Request received from ListControllerRecommendations grpc function")

	out := new(DatahubV1alpha1.ListControllerRecommendationsResponse)

	conn, client, err := RepoApiServer.CreateClient(s.Target)
	if err != nil {
		return out, nil
	}
	defer conn.Close()

	// Send to API server
	out, err = client.ListControllerRecommendations(RepoApiServer.NewContextWithCredential(), in)

	// Check if needs to resend request
	if out != nil {
		if RepoApiServer.NeedResendRequest(out.GetStatus(), err) {
			out, err = client.ListControllerRecommendations(RepoApiServer.NewContextWithCredential(), in)
		}
	}

	if err != nil {
		return &DatahubV1alpha1.ListControllerRecommendationsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	return out, nil
}

// ListPodsByNodeName list pods running on specific nodes
func (s *ServiceV1alpha1) ListPodsByNodeName(ctx context.Context, in *DatahubV1alpha1.ListPodsByNodeNamesRequest) (*DatahubV1alpha1.ListPodsResponse, error) {
	scope.Debug("Request received from ListPodsByNodeName grpc function")

	out := new(DatahubV1alpha1.ListPodsResponse)

	conn, client, err := RepoApiServer.CreateClient(s.Target)
	if err != nil {
		return out, nil
	}
	defer conn.Close()

	// Send to API server
	out, err = client.ListPodsByNodeName(RepoApiServer.NewContextWithCredential(), in)

	// Check if needs to resend request
	if out != nil {
		if RepoApiServer.NeedResendRequest(out.GetStatus(), err) {
			out, err = client.ListPodsByNodeName(RepoApiServer.NewContextWithCredential(), in)
		}
	}

	if err != nil {
		return &DatahubV1alpha1.ListPodsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	return out, nil
}

// ListSimulatedSchedulingScores list simulated scheduling scores
func (s *ServiceV1alpha1) ListSimulatedSchedulingScores(ctx context.Context, in *DatahubV1alpha1.ListSimulatedSchedulingScoresRequest) (*DatahubV1alpha1.ListSimulatedSchedulingScoresResponse, error) {
	scope.Debug("Request received from ListSimulatedSchedulingScores grpc function")

	out := new(DatahubV1alpha1.ListSimulatedSchedulingScoresResponse)

	conn, client, err := RepoApiServer.CreateClient(s.Target)
	if err != nil {
		return out, nil
	}
	defer conn.Close()

	// Send to API server
	out, err = client.ListSimulatedSchedulingScores(RepoApiServer.NewContextWithCredential(), in)

	// Check if needs to resend request
	if out != nil {
		if RepoApiServer.NeedResendRequest(out.GetStatus(), err) {
			out, err = client.ListSimulatedSchedulingScores(RepoApiServer.NewContextWithCredential(), in)
		}
	}

	if err != nil {
		return &DatahubV1alpha1.ListSimulatedSchedulingScoresResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	return out, nil
}

// CreatePods add containers information of pods to database
func (s *ServiceV1alpha1) CreatePods(ctx context.Context, in *DatahubV1alpha1.CreatePodsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreatePods grpc function")

	conn, client, err := RepoApiServer.CreateClient(s.Target)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}
	defer conn.Close()

	// Send to API server
	stat, err := client.CreatePods(RepoApiServer.NewContextWithCredential(), in)

	// Check if needs to resend request
	if stat != nil {
		if RepoApiServer.NeedResendRequest(stat, err) {
			stat, err = client.CreatePods(RepoApiServer.NewContextWithCredential(), in)
		}
	}

	stat, _ = RepoApiServer.CheckResponse(stat, err)

	return stat, nil
}

func (s *ServiceV1alpha1) CreateControllers(ctx context.Context, in *DatahubV1alpha1.CreateControllersRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateControllers grpc function")

	conn, client, err := RepoApiServer.CreateClient(s.Target)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}
	defer conn.Close()

	// Send to API server
	stat, err := client.CreateControllers(RepoApiServer.NewContextWithCredential(), in)

	// Check if needs to resend request
	if stat != nil {
		if RepoApiServer.NeedResendRequest(stat, err) {
			stat, err = client.CreateControllers(RepoApiServer.NewContextWithCredential(), in)
		}
	}

	stat, _ = RepoApiServer.CheckResponse(stat, err)

	return stat, nil
}

func (s *ServiceV1alpha1) DeleteControllers(ctx context.Context, in *DatahubV1alpha1.DeleteControllersRequest) (*status.Status, error) {
	scope.Debug("Request received from DeleteControllers grpc function")

	conn, client, err := RepoApiServer.CreateClient(s.Target)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}
	defer conn.Close()

	// Send to API server
	stat, err := client.DeleteControllers(RepoApiServer.NewContextWithCredential(), in)

	// Check if needs to resend request
	if stat != nil {
		if RepoApiServer.NeedResendRequest(stat, err) {
			stat, err = client.DeleteControllers(RepoApiServer.NewContextWithCredential(), in)
		}
	}

	stat, _ = RepoApiServer.CheckResponse(stat, err)

	return stat, nil
}

// DeletePods update containers information of pods to database
func (s *ServiceV1alpha1) DeletePods(ctx context.Context, in *DatahubV1alpha1.DeletePodsRequest) (*status.Status, error) {
	scope.Debug("Request received from DeletePods grpc function")

	conn, client, err := RepoApiServer.CreateClient(s.Target)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}
	defer conn.Close()

	// Send to API server
	stat, err := client.DeletePods(RepoApiServer.NewContextWithCredential(), in)

	// Check if needs to resend request
	if stat != nil {
		if RepoApiServer.NeedResendRequest(stat, err) {
			stat, err = client.DeletePods(RepoApiServer.NewContextWithCredential(), in)
		}
	}

	stat, _ = RepoApiServer.CheckResponse(stat, err)

	return stat, nil
}

// CreateAlamedaNodes add node information to database
func (s *ServiceV1alpha1) CreateAlamedaNodes(ctx context.Context, in *DatahubV1alpha1.CreateAlamedaNodesRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateAlamedaNodes grpc function")

	conn, client, err := RepoApiServer.CreateClient(s.Target)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}
	defer conn.Close()

	// Send to API server
	stat, err := client.CreateAlamedaNodes(RepoApiServer.NewContextWithCredential(), in)

	// Check if needs to resend request
	if stat != nil {
		if RepoApiServer.NeedResendRequest(stat, err) {
			stat, err = client.CreateAlamedaNodes(RepoApiServer.NewContextWithCredential(), in)
		}
	}

	stat, _ = RepoApiServer.CheckResponse(stat, err)

	return stat, nil
}

// CreatePodPredictions add pod predictions information to database
func (s *ServiceV1alpha1) CreatePodPredictions(ctx context.Context, in *DatahubV1alpha1.CreatePodPredictionsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreatePodPredictions grpc function")

	conn, client, err := RepoApiServer.CreateClient(s.Target)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}
	defer conn.Close()

	// Send to API server
	stat, err := client.CreatePodPredictions(RepoApiServer.NewContextWithCredential(), in)

	// Check if needs to resend request
	if stat != nil {
		if RepoApiServer.NeedResendRequest(stat, err) {
			stat, err = client.CreatePodPredictions(RepoApiServer.NewContextWithCredential(), in)
		}
	}

	stat, _ = RepoApiServer.CheckResponse(stat, err)

	return stat, nil
}

// CreateNodePredictions add node predictions information to database
func (s *ServiceV1alpha1) CreateNodePredictions(ctx context.Context, in *DatahubV1alpha1.CreateNodePredictionsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateNodePredictions grpc function")

	conn, client, err := RepoApiServer.CreateClient(s.Target)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}
	defer conn.Close()

	// Send to API server
	stat, err := client.CreateNodePredictions(RepoApiServer.NewContextWithCredential(), in)

	// Check if needs to resend request
	if stat != nil {
		if RepoApiServer.NeedResendRequest(stat, err) {
			stat, err = client.CreateNodePredictions(RepoApiServer.NewContextWithCredential(), in)
		}
	}

	stat, _ = RepoApiServer.CheckResponse(stat, err)

	return stat, nil
}

// CreatePodRecommendations add pod recommendations information to database
func (s *ServiceV1alpha1) CreatePodRecommendations(ctx context.Context, in *DatahubV1alpha1.CreatePodRecommendationsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreatePodRecommendations grpc function")

	conn, client, err := RepoApiServer.CreateClient(s.Target)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}
	defer conn.Close()

	// Send to API server
	stat, err := client.CreatePodRecommendations(RepoApiServer.NewContextWithCredential(), in)

	// Check if needs to resend request
	if stat != nil {
		if RepoApiServer.NeedResendRequest(stat, err) {
			stat, err = client.CreatePodRecommendations(RepoApiServer.NewContextWithCredential(), in)
		}
	}

	stat, _ = RepoApiServer.CheckResponse(stat, err)

	return stat, nil
}

// CreatePodRecommendations add pod recommendations information to database
func (s *ServiceV1alpha1) CreateControllerRecommendations(ctx context.Context, in *DatahubV1alpha1.CreateControllerRecommendationsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateControllerRecommendations grpc function")

	conn, client, err := RepoApiServer.CreateClient(s.Target)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}
	defer conn.Close()

	// Send to API server
	stat, err := client.CreateControllerRecommendations(RepoApiServer.NewContextWithCredential(), in)

	// Check if needs to resend request
	if stat != nil {
		if RepoApiServer.NeedResendRequest(stat, err) {
			stat, err = client.CreateControllerRecommendations(RepoApiServer.NewContextWithCredential(), in)
		}
	}

	stat, _ = RepoApiServer.CheckResponse(stat, err)

	return stat, nil
}

// CreateSimulatedSchedulingScores add simulated scheduling scores to database
func (s *ServiceV1alpha1) CreateSimulatedSchedulingScores(ctx context.Context, in *DatahubV1alpha1.CreateSimulatedSchedulingScoresRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateSimulatedSchedulingScores grpc function")

	conn, client, err := RepoApiServer.CreateClient(s.Target)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}
	defer conn.Close()

	// Send to API server
	stat, err := client.CreateSimulatedSchedulingScores(RepoApiServer.NewContextWithCredential(), in)

	// Check if needs to resend request
	if stat != nil {
		if RepoApiServer.NeedResendRequest(stat, err) {
			stat, err = client.CreateSimulatedSchedulingScores(RepoApiServer.NewContextWithCredential(), in)
		}
	}

	stat, _ = RepoApiServer.CheckResponse(stat, err)

	return stat, nil
}

// DeleteAlamedaNodes remove node information to database
func (s *ServiceV1alpha1) DeleteAlamedaNodes(ctx context.Context, in *DatahubV1alpha1.DeleteAlamedaNodesRequest) (*status.Status, error) {
	scope.Debug("Request received from DeleteAlamedaNodes grpc function")

	conn, client, err := RepoApiServer.CreateClient(s.Target)
	if err != nil {
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}
	defer conn.Close()

	// Send to API server
	stat, err := client.DeleteAlamedaNodes(RepoApiServer.NewContextWithCredential(), in)

	// Check if needs to resend request
	if stat != nil {
		if RepoApiServer.NeedResendRequest(stat, err) {
			stat, err = client.DeleteAlamedaNodes(RepoApiServer.NewContextWithCredential(), in)
		}
	}

	stat, _ = RepoApiServer.CheckResponse(stat, err)

	return stat, nil
}

// Read rawdata from database
func (s *ServiceV1alpha1) ReadRawdata(ctx context.Context, in *DatahubV1alpha1.ReadRawdataRequest) (*DatahubV1alpha1.ReadRawdataResponse, error) {
	scope.Debug("Request received from ReadRawdata grpc function")
	out := new(DatahubV1alpha1.ReadRawdataResponse)

	conn, client, err := RepoApiServer.CreateClient(s.Target)
	if err != nil {
		return out, nil
	}
	defer conn.Close()

	// Send to API server
	out, err = client.ReadRawdata(RepoApiServer.NewContextWithCredential(), in)

	// Check if needs to resend request
	if out != nil {
		if RepoApiServer.NeedResendRequest(out.GetStatus(), err) {
			out, err = client.ReadRawdata(RepoApiServer.NewContextWithCredential(), in)
		}
	}

	if err != nil {
		return &DatahubV1alpha1.ReadRawdataResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	return out, nil
}

// Write rawdata to database
func (s *ServiceV1alpha1) WriteRawdata(ctx context.Context, in *DatahubV1alpha1.WriteRawdataRequest) (*status.Status, error) {
	scope.Debug("Request received from WriteRawdata grpc function")
	out := new(status.Status)

	conn, client, err := RepoApiServer.CreateClient(s.Target)
	if err != nil {
		return out, nil
	}
	defer conn.Close()

	// Send to API server
	stat, err := client.WriteRawdata(RepoApiServer.NewContextWithCredential(), in)

	// Check if needs to resend request
	if stat != nil {
		if RepoApiServer.NeedResendRequest(stat, err) {
			stat, err = client.WriteRawdata(RepoApiServer.NewContextWithCredential(), in)
		}
	}

	stat, _ = RepoApiServer.CheckResponse(stat, err)

	return stat, nil
}

func (s *ServiceV1alpha1) Ping(ctx context.Context, in *empty.Empty) (*status.Status, error) {
	scope.Debug("Request received from Ping grpc function")

	return &status.Status{Code: int32(code.Code_OK)}, nil
}

func (s *ServiceV1alpha1) ListWeaveScopeHosts(ctx context.Context, in *DatahubV1alpha1.ListWeaveScopeHostsRequest) (*DatahubV1alpha1.WeaveScopeResponse, error) {
	scope.Debug("Request received from ListWeaveScopeHosts grpc function")

	out := new(DatahubV1alpha1.WeaveScopeResponse)

	conn, client, err := RepoApiServer.CreateClient(s.Target)
	if err != nil {
		return out, nil
	}
	defer conn.Close()

	// Send to API server
	out, err = client.ListWeaveScopeHosts(RepoApiServer.NewContextWithCredential(), in)

	// Check if needs to resend request
	if out != nil {
		if RepoApiServer.NeedResendRequest(out.GetStatus(), err) {
			out, err = client.ListWeaveScopeHosts(RepoApiServer.NewContextWithCredential(), in)
		}
	}

	if err != nil {
		return &DatahubV1alpha1.WeaveScopeResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	return out, nil
}

func (s *ServiceV1alpha1) GetWeaveScopeHostDetails(ctx context.Context, in *DatahubV1alpha1.ListWeaveScopeHostsRequest) (*DatahubV1alpha1.WeaveScopeResponse, error) {
	scope.Debug("Request received from GetWeaveScopeHostDetails grpc function")

	out := new(DatahubV1alpha1.WeaveScopeResponse)

	conn, client, err := RepoApiServer.CreateClient(s.Target)
	if err != nil {
		return out, nil
	}
	defer conn.Close()

	// Send to API server
	out, err = client.GetWeaveScopeHostDetails(RepoApiServer.NewContextWithCredential(), in)

	// Check if needs to resend request
	if out != nil {
		if RepoApiServer.NeedResendRequest(out.GetStatus(), err) {
			out, err = client.GetWeaveScopeHostDetails(RepoApiServer.NewContextWithCredential(), in)
		}
	}

	if err != nil {
		return &DatahubV1alpha1.WeaveScopeResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	return out, nil
}

func (s *ServiceV1alpha1) ListWeaveScopePods(ctx context.Context, in *DatahubV1alpha1.ListWeaveScopePodsRequest) (*DatahubV1alpha1.WeaveScopeResponse, error) {
	scope.Debug("Request received from ListWeaveScopePods grpc function")

	out := new(DatahubV1alpha1.WeaveScopeResponse)

	conn, client, err := RepoApiServer.CreateClient(s.Target)
	if err != nil {
		return out, nil
	}
	defer conn.Close()

	// Send to API server
	out, err = client.ListWeaveScopePods(RepoApiServer.NewContextWithCredential(), in)

	// Check if needs to resend request
	if out != nil {
		if RepoApiServer.NeedResendRequest(out.GetStatus(), err) {
			out, err = client.ListWeaveScopePods(RepoApiServer.NewContextWithCredential(), in)
		}
	}

	if err != nil {
		return &DatahubV1alpha1.WeaveScopeResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	return out, nil
}

func (s *ServiceV1alpha1) GetWeaveScopePodDetails(ctx context.Context, in *DatahubV1alpha1.ListWeaveScopePodsRequest) (*DatahubV1alpha1.WeaveScopeResponse, error) {
	scope.Debug("Request received from GetWeaveScopePodDetails grpc function")

	out := new(DatahubV1alpha1.WeaveScopeResponse)

	conn, client, err := RepoApiServer.CreateClient(s.Target)
	if err != nil {
		return out, nil
	}
	defer conn.Close()

	// Send to API server
	out, err = client.GetWeaveScopePodDetails(RepoApiServer.NewContextWithCredential(), in)

	// Check if needs to resend request
	if out != nil {
		if RepoApiServer.NeedResendRequest(out.GetStatus(), err) {
			out, err = client.GetWeaveScopePodDetails(RepoApiServer.NewContextWithCredential(), in)
		}
	}

	if err != nil {
		return &DatahubV1alpha1.WeaveScopeResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	return out, nil
}

func (s *ServiceV1alpha1) ListWeaveScopeContainers(ctx context.Context, in *DatahubV1alpha1.ListWeaveScopeContainersRequest) (*DatahubV1alpha1.WeaveScopeResponse, error) {
	scope.Debug("Request received from ListWeaveScopeContainers grpc function")

	out := new(DatahubV1alpha1.WeaveScopeResponse)

	conn, client, err := RepoApiServer.CreateClient(s.Target)
	if err != nil {
		return out, nil
	}
	defer conn.Close()

	// Send to API server
	out, err = client.ListWeaveScopeContainers(RepoApiServer.NewContextWithCredential(), in)

	// Check if needs to resend request
	if out != nil {
		if RepoApiServer.NeedResendRequest(out.GetStatus(), err) {
			out, err = client.ListWeaveScopeContainers(RepoApiServer.NewContextWithCredential(), in)
		}
	}

	if err != nil {
		return &DatahubV1alpha1.WeaveScopeResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	return out, nil
}

func (s *ServiceV1alpha1) ListWeaveScopeContainersByHostname(ctx context.Context, in *DatahubV1alpha1.ListWeaveScopeContainersRequest) (*DatahubV1alpha1.WeaveScopeResponse, error) {
	scope.Debug("Request received from ListWeaveScopeContainersByHostname grpc function")

	out := new(DatahubV1alpha1.WeaveScopeResponse)

	conn, client, err := RepoApiServer.CreateClient(s.Target)
	if err != nil {
		return out, nil
	}
	defer conn.Close()

	// Send to API server
	out, err = client.ListWeaveScopeContainersByHostname(RepoApiServer.NewContextWithCredential(), in)

	// Check if needs to resend request
	if out != nil {
		if RepoApiServer.NeedResendRequest(out.GetStatus(), err) {
			out, err = client.ListWeaveScopeContainersByHostname(RepoApiServer.NewContextWithCredential(), in)
		}
	}

	if err != nil {
		return &DatahubV1alpha1.WeaveScopeResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	return out, nil
}

func (s *ServiceV1alpha1) ListWeaveScopeContainersByImage(ctx context.Context, in *DatahubV1alpha1.ListWeaveScopeContainersRequest) (*DatahubV1alpha1.WeaveScopeResponse, error) {
	scope.Debug("Request received from ListWeaveScopeContainersByImage grpc function")

	out := new(DatahubV1alpha1.WeaveScopeResponse)

	conn, client, err := RepoApiServer.CreateClient(s.Target)
	if err != nil {
		return out, nil
	}
	defer conn.Close()

	// Send to API server
	out, err = client.ListWeaveScopeContainersByImage(RepoApiServer.NewContextWithCredential(), in)

	// Check if needs to resend request
	if out != nil {
		if RepoApiServer.NeedResendRequest(out.GetStatus(), err) {
			out, err = client.ListWeaveScopeContainersByImage(RepoApiServer.NewContextWithCredential(), in)
		}
	}

	if err != nil {
		return &DatahubV1alpha1.WeaveScopeResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	return out, nil
}

func (s *ServiceV1alpha1) GetWeaveScopeContainerDetails(ctx context.Context, in *DatahubV1alpha1.ListWeaveScopeContainersRequest) (*DatahubV1alpha1.WeaveScopeResponse, error) {
	scope.Debug("Request received from GetWeaveScopeContainerDetails grpc function")

	out := new(DatahubV1alpha1.WeaveScopeResponse)

	conn, client, err := RepoApiServer.CreateClient(s.Target)
	if err != nil {
		return out, nil
	}
	defer conn.Close()

	// Send to API server
	out, err = client.GetWeaveScopeContainerDetails(RepoApiServer.NewContextWithCredential(), in)

	// Check if needs to resend request
	if out != nil {
		if RepoApiServer.NeedResendRequest(out.GetStatus(), err) {
			out, err = client.GetWeaveScopeContainerDetails(RepoApiServer.NewContextWithCredential(), in)
		}
	}

	if err != nil {
		return &DatahubV1alpha1.WeaveScopeResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	return out, nil
}
