package datahub

import (
	"fmt"
	"net"

	"github.com/containers-ai/alameda/datahub/pkg/apis/keycodes"
	"github.com/containers-ai/alameda/datahub/pkg/apis/v1alpha1"
	DatahubConfig "github.com/containers-ai/alameda/datahub/pkg/config"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	OperatorAPIs "github.com/containers-ai/alameda/operator/api/v1alpha1"
	K8SUtils "github.com/containers-ai/alameda/pkg/utils/kubernetes"
	Log "github.com/containers-ai/alameda/pkg/utils/log"
	DatahubV1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	DatahubKeycodes "github.com/containers-ai/api/datahub/keycodes"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type Server struct {
	err       chan error
	server    *grpc.Server
	Config    DatahubConfig.Config
	K8SClient client.Client
}

var (
	scope = Log.RegisterScope("gRPC", "gRPC server log", 0)
)

func NewServer(cfg DatahubConfig.Config) (*Server, error) {
	var (
		err error

		server *Server
		k8sCli client.Client
	)

	// Validate datahub configuration
	if err = cfg.Validate(); err != nil {
		return server, errors.New("Failed to validate datahub configuration: " + err.Error())
	}

	// Instance kubernetes client
	if k8sCli, err = K8SUtils.NewK8SClient(); err != nil {
		return server, err
	}

	// Get kubernetes configuration
	k8sClientConfig, err := config.GetConfig()
	if err != nil {
		return server, errors.New("Failed to get kubernetes configuration: " + err.Error())
	}

	// Add alameda CR
	mgr, err := manager.New(k8sClientConfig, manager.Options{})
	if err != nil {
		scope.Error(err.Error())
	}
	if err := OperatorAPIs.AddToScheme(mgr.GetScheme()); err != nil {
		scope.Error(err.Error())
	}

	// Get cluster uid and insert into server.Config
	clusterId, err := K8SUtils.GetClusterUID(k8sCli)
	if err != nil {
		scope.Errorf("failed to get cluster id: %s", err.Error())
	}
	cfg.ClusterUID = clusterId

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

	s.register(server)
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
	scope.Info("Initialize database")

	influxdbClient := InternalInflux.NewClient(&InternalInflux.Config{
		Address:                s.Config.InfluxDB.Address,
		Username:               s.Config.InfluxDB.Username,
		Password:               s.Config.InfluxDB.Password,
		RetentionDuration:      s.Config.InfluxDB.RetentionDuration,
		RetentionShardDuration: s.Config.InfluxDB.RetentionShardDuration,
	})

	databaseList := []string{
		"alameda_event",
		"alameda_gpu",
		"alameda_gpu_prediction",
		"alameda_metric",
		"alameda_planning",
		"alameda_prediction",
		"alameda_recommendation",
		"alameda_score",
	}

	for _, db := range databaseList {
		err := influxdbClient.CreateDatabase(db, 0)
		if err != nil {
			scope.Error(err.Error())
		}

		err = influxdbClient.ModifyDefaultRetentionPolicy(db, 0)
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

func (s *Server) register(server *grpc.Server) {
	v1alpha1Srv := v1alpha1.NewService(&s.Config, s.K8SClient)
	DatahubV1alpha1.RegisterDatahubServiceServer(server, v1alpha1Srv)

	keycodesSrv := keycodes.NewService(&s.Config)
	DatahubKeycodes.RegisterKeycodesServiceServer(server, keycodesSrv)
}
