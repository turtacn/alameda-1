package datahub

import (
	"errors"
	"fmt"
	"net"

	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metrics"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metrics/prometheus"
	"github.com/containers-ai/alameda/pkg/utils/log"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type Server struct {
	err    chan error
	server *grpc.Server

	Config    Config
	K8SClient client.Client
	MetricsDB metrics.MetricsDB
}

var (
	scope = log.RegisterScope("gRPC", "gRPC server log", 0)
)

func NewServer(cfg Config) (*Server, error) {

	var (
		err error

		server    *Server
		k8sCli    client.Client
		metricsDB metrics.MetricsDB
	)

	if err = cfg.Validate(); err != nil {
		return server, errors.New("create gRPC server instance failed: " + err.Error())
	}

	k8sClientConfig, err := config.GetConfig()
	if err != nil {
		return server, errors.New("create gRPC server instance failed: " + err.Error())
	}

	if k8sCli, err = client.New(k8sClientConfig, client.Options{}); err != nil {
		return server, errors.New("create gRPC server instance failed: " + err.Error())
	}

	if metricsDB, err = prometheus.New(*cfg.Prometheus); err != nil {
		return server, errors.New("create gRPC server instance failed: " + err.Error())
	}

	server = &Server{
		err: make(chan error),

		Config:    cfg,
		K8SClient: k8sCli,
		MetricsDB: metricsDB,
	}

	return server, nil
}

func (s *Server) Run() error {

	// Open metrics database
	if err := s.MetricsDB.Connect(); err != nil {
		return err
	}

	// build server listener
	scope.Info(("starting gRPC server"))
	ln, err := net.Listen("tcp", s.Config.BindAddress)
	if err != nil {
		scope.Error("gRPC server failed listen: " + err.Error())
		return fmt.Errorf("GRPC server failed to bind address: %s", s.Config.BindAddress)
	}
	scope.Info("gRPC server listening on " + s.Config.BindAddress)

	server, err := s.newGRPCServer()
	if err != nil {
		scope.Error(err.Error())
		return err
	}
	s.server = server

	s.registGRPCServer(server)
	reflection.Register(server)

	if err := server.Serve(ln); err != nil {
		s.err <- fmt.Errorf("GRPC server failed to serve: %s", err.Error())
	}

	return nil
}

func (s *Server) Stop() error {

	if err := s.MetricsDB.Close(); err != nil {
		return err
	}

	s.server.Stop()

	return nil
}

func (s *Server) Err() <-chan error {
	return s.err
}

func (s *Server) newGRPCServer() (*grpc.Server, error) {

	var (
		server *grpc.Server
	)

	server = grpc.NewServer()

	return server, nil
}

func (s *Server) registGRPCServer(server *grpc.Server) {

	datahub_v1alpha1.RegisterDatahubServiceServer(server, s)
}
func (s *Server) ListContainerMetrics(ctx context.Context, in *datahub_v1alpha1.ListContainerMetricsRequest) (*datahub_v1alpha1.ListContainerMetricsResponse, error) {
	return &datahub_v1alpha1.ListContainerMetricsResponse{
		Status: &status.Status{
			Code:    int32(code.Code_UNIMPLEMENTED),
			Message: "Not implemented",
		},
	}, nil
}
func (s *Server) ListNodeMetrics(ctx context.Context, in *datahub_v1alpha1.ListNodeMetricsRequest) (*datahub_v1alpha1.ListNodeMetricsResponse, error) {
	return &datahub_v1alpha1.ListNodeMetricsResponse{
		Status: &status.Status{
			Code:    int32(code.Code_UNIMPLEMENTED),
			Message: "Not implemented",
		},
	}, nil
}
func (s *Server) CreateAlamedaPod(ctx context.Context, in *datahub_v1alpha1.CreateAlamedaPodRequest) (*status.Status, error) {
	return &status.Status{
		Code:    int32(code.Code_UNIMPLEMENTED),
		Message: "Not implemented",
	}, nil
}
func (s *Server) DeleteAlamedaPod(ctx context.Context, in *datahub_v1alpha1.DeleteAlamedaPodRequest) (*status.Status, error) {
	return &status.Status{
		Code:    int32(code.Code_UNIMPLEMENTED),
		Message: "Not implemented",
	}, nil
}
func (s *Server) CreateAlamedaNode(ctx context.Context, in *datahub_v1alpha1.CreateAlamedaNodeRequest) (*status.Status, error) {
	return &status.Status{
		Code:    int32(code.Code_UNIMPLEMENTED),
		Message: "Not implemented",
	}, nil
}
func (s *Server) DeleteAlamedaNode(ctx context.Context, in *datahub_v1alpha1.DeleteAlamedaNodeRequest) (*status.Status, error) {
	return &status.Status{
		Code:    int32(code.Code_UNIMPLEMENTED),
		Message: "Not implemented",
	}, nil
}
func (s *Server) ListAlamedaPods(ctx context.Context, in *empty.Empty) (*datahub_v1alpha1.ListAlamedaPodsResponse, error) {
	return &datahub_v1alpha1.ListAlamedaPodsResponse{
		Status: &status.Status{
			Code:    int32(code.Code_UNIMPLEMENTED),
			Message: "Not implemented",
		},
	}, nil
}
func (s *Server) ListAlamedaNodes(ctx context.Context, in *empty.Empty) (*datahub_v1alpha1.ListAlamedaNodesResponse, error) {
	return &datahub_v1alpha1.ListAlamedaNodesResponse{
		Status: &status.Status{
			Code:    int32(code.Code_UNIMPLEMENTED),
			Message: "Not implemented",
		},
	}, nil
}
func (s *Server) CreatePredictPods(ctx context.Context, in *datahub_v1alpha1.CreatePredictPodsRequest) (*status.Status, error) {
	return &status.Status{
		Code:    int32(code.Code_UNIMPLEMENTED),
		Message: "Not implemented",
	}, nil
}
func (s *Server) CreatePredictNodes(ctx context.Context, in *datahub_v1alpha1.CreatePredictNodesRequest) (*status.Status, error) {
	return &status.Status{
		Code:    int32(code.Code_UNIMPLEMENTED),
		Message: "Not implemented",
	}, nil
}
func (s *Server) GetPodPredictResult(ctx context.Context, in *datahub_v1alpha1.GetPodPredictRequest) (*datahub_v1alpha1.GetPodPredictResponse, error) {
	return &datahub_v1alpha1.GetPodPredictResponse{
		Status: &status.Status{
			Code:    int32(code.Code_UNIMPLEMENTED),
			Message: "Not implemented",
		},
	}, nil
}
func (s *Server) GetNodePredictResult(ctx context.Context, in *datahub_v1alpha1.GetNodePredictRequest) (*datahub_v1alpha1.GetNodePredictResponse, error) {
	return &datahub_v1alpha1.GetNodePredictResponse{
		Status: &status.Status{
			Code:    int32(code.Code_UNIMPLEMENTED),
			Message: "Not implemented",
		},
	}, nil
}
func (s *Server) GetAlamedaPodResourceInfo(ctx context.Context, in *datahub_v1alpha1.GetAlamedaPodResourceInfoRequest) (*datahub_v1alpha1.ListAlamedaPodsResponse, error) {
	return &datahub_v1alpha1.ListAlamedaPodsResponse{
		Status: &status.Status{
			Code:    int32(code.Code_UNIMPLEMENTED),
			Message: "Not implemented",
		},
	}, nil
}
