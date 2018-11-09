package server

import (
	"os"
	"sync"

	"github.com/containers-ai/alameda/operator/services"
	"github.com/containers-ai/alameda/operator/services/grpc"
)

type Server struct {
	Config *Config

	gRPC *grpc.Service

	Services []services.Service
}

func NewServer(config *Config) (*Server, error) {

	s := Server{}

	// Validate Server config
	err := config.Validate()
	if err != nil {
		return nil, err
	}
	s.Config = config

	s.gRPC, err = grpc.NewService(config.GRPC, config.Manager)
	if err != nil {
		return nil, err
	}
	s.Services = append(s.Services, s.gRPC)

	return &s, nil
}

func (s *Server) Start(wg *sync.WaitGroup) {
	defer wg.Done()

	for _, service := range s.Services {
		go service.Open()
	}

	var sigCh = make(chan os.Signal, 1)

	select {
	case <-sigCh:
		s.Close()
	}

}

// Close close the services that server running
func (s *Server) Close() {
	for _, service := range s.Services {
		service.Close()
	}
}
