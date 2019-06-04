package datapipe

import (
	"errors"
	"fmt"
	"github.com/containers-ai/alameda/datapipe/pkg/apis/metrics"
	"github.com/containers-ai/alameda/datapipe/pkg/apis/ping"
	"github.com/containers-ai/alameda/datapipe/pkg/apis/predictions"
	"github.com/containers-ai/alameda/datapipe/pkg/apis/rawdata"
	"github.com/containers-ai/alameda/datapipe/pkg/apis/recommendations"
	"github.com/containers-ai/alameda/datapipe/pkg/apis/resources"
	"github.com/containers-ai/alameda/datapipe/pkg/apis/scores"
	"github.com/containers-ai/alameda/datapipe/pkg/apis/v1alpha1"
	"github.com/containers-ai/alameda/datapipe/pkg/config"
	"github.com/containers-ai/alameda/pkg/utils/log"
	V1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	Metrics "github.com/containers-ai/api/datapipe/metrics"
	Ping "github.com/containers-ai/api/datapipe/ping"
	Predictions "github.com/containers-ai/api/datapipe/predictions"
	Rawdata "github.com/containers-ai/api/datapipe/rawdata"
	Recommendations "github.com/containers-ai/api/datapipe/recommendations"
	Resources "github.com/containers-ai/api/datapipe/resources"
	Scores "github.com/containers-ai/api/datapipe/scores"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"net"
)

type Server struct {
	err    chan error
	server *grpc.Server
	Config config.Config
}

var (
	scope = log.RegisterScope("datapipe", "datapipe log", 0)
)

func NewServer(cfg config.Config) (*Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, errors.New("failed to validate configuration: " + err.Error())
	}

	server := &Server{
		err:    make(chan error),
		Config: cfg,
	}

	return server, nil
}

func (s *Server) Run() error {
	scope.Info("starting Alameda datapipe")

	// Build server listener
	ln, err := net.Listen("tcp", s.Config.BindAddress)
	if err != nil {
		scope.Error("failed to listen Alameda datapipe: " + err.Error())
		return fmt.Errorf("GRPC server(datapipe) failed to bind address: %s", s.Config.BindAddress)
	}

	scope.Info("datapipe listening on " + s.Config.BindAddress)

	server, err := s.newGRPCServer()
	if err != nil {
		scope.Error(err.Error())
		return err
	}
	s.server = server

	s.registerGRPCServer(server)
	reflection.Register(server)

	if err := server.Serve(ln); err != nil {
		s.err <- fmt.Errorf("GRPC server(datapipe) failed to serve: %s", err.Error())
	}

	return nil
}

func (s *Server) Stop() error {
	s.server.Stop()
	return nil
}

func (s *Server) Err() <-chan error {
	return s.err
}

func (s *Server) newGRPCServer() (*grpc.Server, error) {
	server := grpc.NewServer()
	return server, nil
}

func (s *Server) registerGRPCServer(server *grpc.Server) {
	metric := metrics.NewServiceMetric(&s.Config)
	Metrics.RegisterMetricsServiceServer(server, metric)

	pings := ping.NewServicePing(&s.Config)
	Ping.RegisterPingServiceServer(server, pings)

	prediction := predictions.NewServicePrediction(&s.Config)
	Predictions.RegisterPredictionsServiceServer(server, prediction)

	rdata := rawdata.NewServiceRawdata(&s.Config)
	Rawdata.RegisterRawdataServiceServer(server, rdata)

	recommendation := recommendations.NewServiceRecommendation(&s.Config)
	Recommendations.RegisterRecommendationsServiceServer(server, recommendation)

	resource := resources.NewServiceResource(&s.Config)
	Resources.RegisterResourcesServiceServer(server, resource)

	score := scores.NewServiceScore(&s.Config)
	Scores.RegisterScoresServiceServer(server, score)

	v1alpha1Srv := v1alpha1.NewServiceV1alpha1()
	V1alpha1.RegisterDatahubServiceServer(server, v1alpha1Srv)
	v1alpha1Srv.Target = s.Config.APIServer.Address
}
