package apiserver

import (
	"errors"
	"fmt"
	"github.com/containers-ai/alameda/apiserver/pkg/apis/accounts"
	"github.com/containers-ai/alameda/apiserver/pkg/apis/agents"
	"github.com/containers-ai/alameda/apiserver/pkg/apis/ping"
	"github.com/containers-ai/alameda/apiserver/pkg/apis/rawdata"
	"github.com/containers-ai/alameda/apiserver/pkg/apis/v1alpha1"
	"github.com/containers-ai/alameda/apiserver/pkg/config"
	"github.com/containers-ai/alameda/pkg/utils/log"
	DatahubV1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	Accounts "github.com/containers-ai/federatorai-api/apiserver/accounts"
	Agents "github.com/containers-ai/federatorai-api/apiserver/agents"
	Ping "github.com/containers-ai/federatorai-api/apiserver/ping"
	Rawdata "github.com/containers-ai/federatorai-api/apiserver/rawdata"
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
	scope = log.RegisterScope("apiserver", "API server log", 0)
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
	scope.Info("starting Alameda API server")

	// Build server listener
	ln, err := net.Listen("tcp", s.Config.BindAddress)
	if err != nil {
		scope.Error("failed to listen Alameda API server: " + err.Error())
		return fmt.Errorf("GRPC server(apiserver) failed to bind address: %s", s.Config.BindAddress)
	}

	scope.Info("apiserver listening on " + s.Config.BindAddress)

	server, err := s.newGRPCServer()
	if err != nil {
		scope.Error(err.Error())
		return err
	}
	s.server = server

	s.registerGRPCServer(server)
	reflection.Register(server)

	if err := server.Serve(ln); err != nil {
		s.err <- fmt.Errorf("GRPC server(apiserver) failed to serve: %s", err.Error())
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
	account := accounts.NewServiceAccount(&s.Config)
	Accounts.RegisterAccountsServiceServer(server, account)

	agent := agents.NewServiceAgent(&s.Config)
	Agents.RegisterAgentsServiceServer(server, agent)

	pings := ping.NewServicePing(&s.Config)
	Ping.RegisterPingServiceServer(server, pings)

	rdata := rawdata.NewServiceRawdata(&s.Config)
	Rawdata.RegisterRawdataServiceServer(server, rdata)

	v1alpha1Srv := v1alpha1.NewServiceV1alpha1()
	DatahubV1alpha1.RegisterDatahubServiceServer(server, v1alpha1Srv)
	v1alpha1Srv.Target = s.Config.Datahub.Address
}
