package server

import (
	"errors"
	"os"
	"sync"

	"github.com/containers-ai/alameda/operator/services"
	"github.com/containers-ai/alameda/operator/services/grpc"
)

type Server struct {
	err chan error

	Config *Config

	gRPC *grpc.Service

	Services []services.Service

	ServicesByName map[string]bool
}

func NewServer(config *Config) (*Server, error) {

	s := Server{
		err:            make(chan error),
		Services:       make([]services.Service, 0),
		ServicesByName: make(map[string]bool),
	}

	// Validate Server config
	err := config.Validate()
	if err != nil {
		return nil, errors.New("new server instance failed: " + err.Error())
	}
	s.Config = config

	s.appendGRPCService()

	return &s, nil
}

func (s *Server) Start(wg *sync.WaitGroup) {

	for _, service := range s.Services {
		go service.Open()
	}

	go s.watchServices()

	var sigCh = make(chan os.Signal, 1)

	select {
	case <-sigCh:
		s.Close(wg)
	}
}

// Close close the services that server running
func (s *Server) Close(wg *sync.WaitGroup) {
	defer wg.Done()
	for _, service := range s.Services {
		service.Close()
	}
}

func (s *Server) Err() <-chan error {
	return s.err
}

func (s *Server) watchServices() {

	var err error

	select {
	case err = <-s.gRPC.Err():
	}

	s.err <- err
}

func (s *Server) appendGRPCService() {

	config := s.Config
	service := grpc.NewService(config.GRPC, config.Manager)
	s.gRPC = service
	s.AppendService("gRPC", service)
}

func (s *Server) AppendService(serviceName string, service services.Service) {

	if _, exist := s.ServicesByName[serviceName]; exist {
		panic("cannot append service twice")
	}

	s.ServicesByName[serviceName] = true
	s.Services = append(s.Services, service)
}
