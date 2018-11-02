package grpc

import (
	"fmt"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/containers-ai/alameda/operator/pkg/utils/log"
	operator_v1alpha1 "github.com/containers-ai/api/operator/v1alpha1"
)

type Service struct {
	Config Config
}

func NewService(c *Config) (*Service, error) {

	// TODO: new metircs repository

	s := &Service{
		Config: *c,
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

func (s *Service) GetMetrics(ctx context.Context, in *operator_v1alpha1.GetMetricsRequest) (*operator_v1alpha1.GetMetricsResponse, error) {

	return &operator_v1alpha1.GetMetricsResponse{}, nil
}

func (s *Service) PostPredictResult(ctx context.Context, in *operator_v1alpha1.PostPredictResultRequest) (*operator_v1alpha1.PostPredictResultResponse, error) {

	return &operator_v1alpha1.PostPredictResultResponse{}, nil
}
