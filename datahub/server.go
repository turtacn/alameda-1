package datahub

import (
	"fmt"
	"github.com/containers-ai/alameda/datahub/pkg/apis/v1alpha1"
	DatahubConfig "github.com/containers-ai/alameda/datahub/pkg/config"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	OperatorAPIs"github.com/containers-ai/alameda/operator/pkg/apis"
	Log "github.com/containers-ai/alameda/pkg/utils/log"
	DatahubV1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"net"
)

type Server struct {
	err       chan error
	server    *grpc.Server
	Config    DatahubConfig.Config
	K8SClient client.Client
}

var (
	scope         = Log.RegisterScope("gRPC", "gRPC server log", 0)
	tmpTimestamps = []*timestamp.Timestamp{
		&timestamp.Timestamp{Seconds: 1545809000},
		&timestamp.Timestamp{Seconds: 1545809030},
		&timestamp.Timestamp{Seconds: 1545809060},
		&timestamp.Timestamp{Seconds: 1545809090},
		&timestamp.Timestamp{Seconds: 1545809120},
		&timestamp.Timestamp{Seconds: 1545809150},
	}
)

func NewServer(cfg DatahubConfig.Config) (*Server, error) {
	var (
		err error

		server *Server
		k8sCli client.Client
	)

	if err = cfg.Validate(); err != nil {
		return server, errors.New("Configuration validation failed: " + err.Error())
	}
	k8sClientConfig, err := config.GetConfig()
	if err != nil {
		return server, errors.New("Get kubernetes configuration failed: " + err.Error())
	}

	if k8sCli, err = client.New(k8sClientConfig, client.Options{}); err != nil {
		return server, errors.New("Create kubernetes client failed: " + err.Error())
	}

	mgr, err := manager.New(k8sClientConfig, manager.Options{})
	if err != nil {
		scope.Error(err.Error())
	}
	if err := OperatorAPIs.AddToScheme(mgr.GetScheme()); err != nil {
		scope.Error(err.Error())
	}

	server = &Server{
		err: make(chan error),

		Config:    cfg,
		K8SClient: k8sCli,
	}

	return server, nil
}

func (s *Server) Run() error {
	// Build server listener
	scope.Info("starting alameda datahub")
	ln, err := net.Listen("tcp", s.Config.BindAddress)
	if err != nil {
		scope.Error("failed to listen alameda datahub: " + err.Error())
		return fmt.Errorf("GRPC server(datahub) failed to bind address: %s", s.Config.BindAddress)
	}
	scope.Info("datahub listening on port" + s.Config.BindAddress)

	server, err := s.newGRPCServer()
	if err != nil {
		scope.Error(err.Error())
		return err
	}
	s.server = server

	s.Register(server)
	reflection.Register(server)

	if err := server.Serve(ln); err != nil {
		s.err <- fmt.Errorf("GRPC server(datahub) failed to serve: %s", err.Error())
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

func (s *Server) InitInfluxdbDatabase() {
	influxdbClient := InternalInflux.NewClient(&InternalInflux.Config{
		Address:                s.Config.InfluxDB.Address,
		Username:               s.Config.InfluxDB.Username,
		Password:               s.Config.InfluxDB.Password,
		RetentionDuration:      s.Config.InfluxDB.RetentionDuration,
		RetentionShardDuration: s.Config.InfluxDB.RetentionShardDuration,
	})

	databaseList := []string{
		"alameda_prediction",
		"alameda_recommendation",
		"alameda_score",
		"alameda_event",
	}

	for _, db := range databaseList {
		err := influxdbClient.CreateDatabase(db)
		if err != nil {
			scope.Error(err.Error())
		}

		err = influxdbClient.ModifyDefaultRetentionPolicy(db)
		if err != nil {
			scope.Error(err.Error())
		}
	}
}

func (s *Server) newGRPCServer() (*grpc.Server, error) {
	var (
		server *grpc.Server
	)

	server = grpc.NewServer()

	return server, nil
}

func (s *Server) Register(server *grpc.Server) {
	v1alphaSrv := v1alpha1.NewService(&s.Config, s.K8SClient)
	DatahubV1alpha1.RegisterDatahubServiceServer(server, v1alphaSrv)
}
