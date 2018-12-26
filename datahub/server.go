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
	"github.com/golang/protobuf/ptypes/timestamp"
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

	tmpTimestamps = []*timestamp.Timestamp{
		&timestamp.Timestamp{Seconds: 1545809867},
		&timestamp.Timestamp{Seconds: 1545809897},
	}
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

func (s *Server) ListPodMetrics(ctx context.Context, in *datahub_v1alpha1.ListPodMetricsRequest) (*datahub_v1alpha1.ListPodMetricsResponse, error) {

	return &datahub_v1alpha1.ListPodMetricsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		PodMetrics: []*datahub_v1alpha1.PodMetric{
			&datahub_v1alpha1.PodMetric{
				NamespacedName: &datahub_v1alpha1.NamespacedName{
					Namespace: "openshit-monitoring",
					Name:      "prometheus-k8s-0",
				},
				ContainerMetrics: []*datahub_v1alpha1.ContainerMetric{
					&datahub_v1alpha1.ContainerMetric{
						Name: "prometheus",
						MetricData: []*datahub_v1alpha1.MetricData{
							&datahub_v1alpha1.MetricData{
								MetricType: datahub_v1alpha1.MetricType_CONTAINER_CPU_USAGE_SECONDS_PERCENTAGE,
								Data: []*datahub_v1alpha1.Sample{
									&datahub_v1alpha1.Sample{
										Time:     tmpTimestamps[0],
										NumValue: "20",
									},
									&datahub_v1alpha1.Sample{
										Time:     tmpTimestamps[1],
										NumValue: "50",
									},
								},
							},
							&datahub_v1alpha1.MetricData{
								MetricType: datahub_v1alpha1.MetricType_CONTAINER_MEMORY_USAGE_BYTES,
								Data: []*datahub_v1alpha1.Sample{
									&datahub_v1alpha1.Sample{
										Time:     tmpTimestamps[0],
										NumValue: "512",
									},
									&datahub_v1alpha1.Sample{
										Time:     tmpTimestamps[1],
										NumValue: "1024",
									},
								},
							},
						},
					},
				},
			},
			&datahub_v1alpha1.PodMetric{
				NamespacedName: &datahub_v1alpha1.NamespacedName{
					Namespace: "openshit-monitoring",
					Name:      "prometheus-k8s-1",
				},
				ContainerMetrics: []*datahub_v1alpha1.ContainerMetric{
					&datahub_v1alpha1.ContainerMetric{
						Name: "prometheus",
						MetricData: []*datahub_v1alpha1.MetricData{
							&datahub_v1alpha1.MetricData{
								MetricType: datahub_v1alpha1.MetricType_CONTAINER_CPU_USAGE_SECONDS_PERCENTAGE,
								Data: []*datahub_v1alpha1.Sample{
									&datahub_v1alpha1.Sample{
										Time:     tmpTimestamps[0],
										NumValue: "20",
									},
									&datahub_v1alpha1.Sample{
										Time:     tmpTimestamps[1],
										NumValue: "50",
									},
								},
							},
							&datahub_v1alpha1.MetricData{
								MetricType: datahub_v1alpha1.MetricType_CONTAINER_MEMORY_USAGE_BYTES,
								Data: []*datahub_v1alpha1.Sample{
									&datahub_v1alpha1.Sample{
										Time:     tmpTimestamps[0],
										NumValue: "512",
									},
									&datahub_v1alpha1.Sample{
										Time:     tmpTimestamps[1],
										NumValue: "1024",
									},
								},
							},
						},
					},
				},
			},
		},
	}, nil
}

func (s *Server) ListNodeMetrics(ctx context.Context, in *datahub_v1alpha1.ListNodeMetricsRequest) (*datahub_v1alpha1.ListNodeMetricsResponse, error) {
	return &datahub_v1alpha1.ListNodeMetricsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		NodeMetrics: []*datahub_v1alpha1.NodeMetric{
			&datahub_v1alpha1.NodeMetric{
				Name: "node1",
				MetricData: []*datahub_v1alpha1.MetricData{
					&datahub_v1alpha1.MetricData{
						MetricType: datahub_v1alpha1.MetricType_NODE_CPU_USAGE_SECONDS_PERCENTAGE,
						Data: []*datahub_v1alpha1.Sample{
							&datahub_v1alpha1.Sample{
								Time:     tmpTimestamps[0],
								NumValue: "20",
							},
							&datahub_v1alpha1.Sample{
								Time:     tmpTimestamps[1],
								NumValue: "50",
							},
						},
					},
					&datahub_v1alpha1.MetricData{
						MetricType: datahub_v1alpha1.MetricType_NODE_MEMORY_USAGE_BYTES,
						Data: []*datahub_v1alpha1.Sample{
							&datahub_v1alpha1.Sample{
								Time:     tmpTimestamps[0],
								NumValue: "512",
							},
							&datahub_v1alpha1.Sample{
								Time:     tmpTimestamps[1],
								NumValue: "1024",
							},
						},
					},
				},
			},
			&datahub_v1alpha1.NodeMetric{
				Name: "node2",
				MetricData: []*datahub_v1alpha1.MetricData{
					&datahub_v1alpha1.MetricData{
						MetricType: datahub_v1alpha1.MetricType_NODE_CPU_USAGE_SECONDS_PERCENTAGE,
						Data: []*datahub_v1alpha1.Sample{
							&datahub_v1alpha1.Sample{
								Time:     tmpTimestamps[0],
								NumValue: "20",
							},
							&datahub_v1alpha1.Sample{
								Time:     tmpTimestamps[1],
								NumValue: "50",
							},
						},
					},
					&datahub_v1alpha1.MetricData{
						MetricType: datahub_v1alpha1.MetricType_NODE_MEMORY_USAGE_BYTES,
						Data: []*datahub_v1alpha1.Sample{
							&datahub_v1alpha1.Sample{
								Time:     tmpTimestamps[0],
								NumValue: "512",
							},
							&datahub_v1alpha1.Sample{
								Time:     tmpTimestamps[1],
								NumValue: "1024",
							},
						},
					},
				},
			},
		},
	}, nil
}

func (s *Server) ListAlamedaPods(ctx context.Context, in *datahub_v1alpha1.ListAlamedaPodsRequest) (*datahub_v1alpha1.ListPodsResponse, error) {

	var tmpMetricsData = []*datahub_v1alpha1.MetricData{
		&datahub_v1alpha1.MetricData{
			MetricType: datahub_v1alpha1.MetricType_CONTAINER_CPU_USAGE_SECONDS_PERCENTAGE,
			Data: []*datahub_v1alpha1.Sample{
				&datahub_v1alpha1.Sample{
					Time:     tmpTimestamps[0],
					NumValue: "20",
				},
				&datahub_v1alpha1.Sample{
					Time:     tmpTimestamps[1],
					NumValue: "50",
				},
			},
		},
		&datahub_v1alpha1.MetricData{
			MetricType: datahub_v1alpha1.MetricType_CONTAINER_MEMORY_USAGE_BYTES,
			Data: []*datahub_v1alpha1.Sample{
				&datahub_v1alpha1.Sample{
					Time:     tmpTimestamps[0],
					NumValue: "512",
				},
				&datahub_v1alpha1.Sample{
					Time:     tmpTimestamps[1],
					NumValue: "1024",
				},
			},
		},
	}

	return &datahub_v1alpha1.ListPodsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		Pods: []*datahub_v1alpha1.Pod{
			&datahub_v1alpha1.Pod{
				NamespacedName: &datahub_v1alpha1.NamespacedName{
					Namespace: "openshit-monitoring",
					Name:      "prometheus-k8s-0",
				},
				ResourceLink: "",
				Containers: []*datahub_v1alpha1.Container{
					&datahub_v1alpha1.Container{
						Name:                                 "prometheus",
						LimitResource:                        tmpMetricsData,
						RequestResource:                      tmpMetricsData,
						LimitResourceRecommendation:          tmpMetricsData,
						InitialLimitResourceRecommendation:   tmpMetricsData,
						InitialRequestResourceRecommendation: tmpMetricsData,
					},
					&datahub_v1alpha1.Container{
						Name:                                 "another-container",
						LimitResource:                        tmpMetricsData,
						RequestResource:                      tmpMetricsData,
						LimitResourceRecommendation:          tmpMetricsData,
						InitialLimitResourceRecommendation:   tmpMetricsData,
						InitialRequestResourceRecommendation: tmpMetricsData,
					},
				},
			},
			&datahub_v1alpha1.Pod{
				NamespacedName: &datahub_v1alpha1.NamespacedName{
					Namespace: "openshit-monitoring",
					Name:      "prometheus-k8s-1",
				},
				ResourceLink: "",
				Containers: []*datahub_v1alpha1.Container{
					&datahub_v1alpha1.Container{
						Name:                                 "prometheus",
						LimitResource:                        tmpMetricsData,
						RequestResource:                      tmpMetricsData,
						LimitResourceRecommendation:          tmpMetricsData,
						InitialLimitResourceRecommendation:   tmpMetricsData,
						InitialRequestResourceRecommendation: tmpMetricsData,
					},
					&datahub_v1alpha1.Container{
						Name:                                 "another-container",
						LimitResource:                        tmpMetricsData,
						RequestResource:                      tmpMetricsData,
						LimitResourceRecommendation:          tmpMetricsData,
						InitialLimitResourceRecommendation:   tmpMetricsData,
						InitialRequestResourceRecommendation: tmpMetricsData,
					},
				},
			},
		},
	}, nil
}

func (s *Server) ListAlamedaNodes(ctx context.Context, in *empty.Empty) (*datahub_v1alpha1.ListNodesResponse, error) {
	return &datahub_v1alpha1.ListNodesResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		Nodes: []*datahub_v1alpha1.Node{
			&datahub_v1alpha1.Node{Name: "node1"},
			&datahub_v1alpha1.Node{Name: "node2"},
		},
	}, nil
}

func (s *Server) ListPodPredictions(ctx context.Context, in *datahub_v1alpha1.ListPodPredictionsRequest) (*datahub_v1alpha1.ListPodPredictionsResponse, error) {

	var tmpMetricsData = []*datahub_v1alpha1.MetricData{
		&datahub_v1alpha1.MetricData{
			MetricType: datahub_v1alpha1.MetricType_CONTAINER_CPU_USAGE_SECONDS_PERCENTAGE,
			Data: []*datahub_v1alpha1.Sample{
				&datahub_v1alpha1.Sample{
					Time:     tmpTimestamps[0],
					NumValue: "20",
				},
				&datahub_v1alpha1.Sample{
					Time:     tmpTimestamps[1],
					NumValue: "50",
				},
			},
		},
		&datahub_v1alpha1.MetricData{
			MetricType: datahub_v1alpha1.MetricType_CONTAINER_MEMORY_USAGE_BYTES,
			Data: []*datahub_v1alpha1.Sample{
				&datahub_v1alpha1.Sample{
					Time:     tmpTimestamps[0],
					NumValue: "512",
				},
				&datahub_v1alpha1.Sample{
					Time:     tmpTimestamps[1],
					NumValue: "1024",
				},
			},
		},
	}

	return &datahub_v1alpha1.ListPodPredictionsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		PodPredictions: []*datahub_v1alpha1.PodPrediction{
			&datahub_v1alpha1.PodPrediction{
				NamespacedName: &datahub_v1alpha1.NamespacedName{
					Namespace: "openshift-monitoring",
					Name:      "prometheus-k8s-0",
				},
				ContainerPredictions: []*datahub_v1alpha1.ContainerPrediction{
					&datahub_v1alpha1.ContainerPrediction{
						Name:                            "prometheus",
						PredictedRawData:                tmpMetricsData,
						PredictedLimitData:              tmpMetricsData,
						PredictedRequestData:            tmpMetricsData,
						PredictedInitialLimitResource:   tmpMetricsData,
						PredictedInitialRequestResource: tmpMetricsData,
					},
					&datahub_v1alpha1.ContainerPrediction{
						Name:                            "another-container",
						PredictedRawData:                tmpMetricsData,
						PredictedLimitData:              tmpMetricsData,
						PredictedRequestData:            tmpMetricsData,
						PredictedInitialLimitResource:   tmpMetricsData,
						PredictedInitialRequestResource: tmpMetricsData,
					},
				},
			},
			&datahub_v1alpha1.PodPrediction{
				NamespacedName: &datahub_v1alpha1.NamespacedName{
					Namespace: "openshift-monitoring",
					Name:      "prometheus-k8s-1",
				},
				ContainerPredictions: []*datahub_v1alpha1.ContainerPrediction{
					&datahub_v1alpha1.ContainerPrediction{
						Name:                            "prometheus",
						PredictedRawData:                tmpMetricsData,
						PredictedLimitData:              tmpMetricsData,
						PredictedRequestData:            tmpMetricsData,
						PredictedInitialLimitResource:   tmpMetricsData,
						PredictedInitialRequestResource: tmpMetricsData,
					},
					&datahub_v1alpha1.ContainerPrediction{
						Name:                            "another-container",
						PredictedRawData:                tmpMetricsData,
						PredictedLimitData:              tmpMetricsData,
						PredictedRequestData:            tmpMetricsData,
						PredictedInitialLimitResource:   tmpMetricsData,
						PredictedInitialRequestResource: tmpMetricsData,
					},
				},
			},
		},
	}, nil
}

func (s *Server) ListNodePredictions(ctx context.Context, in *datahub_v1alpha1.ListNodePredictionsRequest) (*datahub_v1alpha1.ListNodePredictionsResponse, error) {

	var tmpNodePredictionsData = []*datahub_v1alpha1.MetricData{
		&datahub_v1alpha1.MetricData{
			MetricType: datahub_v1alpha1.MetricType_NODE_CPU_USAGE_SECONDS_PERCENTAGE,
			Data: []*datahub_v1alpha1.Sample{
				&datahub_v1alpha1.Sample{
					Time:     tmpTimestamps[0],
					NumValue: "20",
				},
				&datahub_v1alpha1.Sample{
					Time:     tmpTimestamps[1],
					NumValue: "50",
				},
			},
		},
		&datahub_v1alpha1.MetricData{
			MetricType: datahub_v1alpha1.MetricType_NODE_MEMORY_USAGE_BYTES,
			Data: []*datahub_v1alpha1.Sample{
				&datahub_v1alpha1.Sample{
					Time:     tmpTimestamps[0],
					NumValue: "512",
				},
				&datahub_v1alpha1.Sample{
					Time:     tmpTimestamps[1],
					NumValue: "1024",
				},
			},
		},
	}

	return &datahub_v1alpha1.ListNodePredictionsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		NodePredictions: []*datahub_v1alpha1.NodePrediction{
			&datahub_v1alpha1.NodePrediction{
				Name:             "node1",
				PredictedRawData: tmpNodePredictionsData,
			},
			&datahub_v1alpha1.NodePrediction{
				Name:             "node2",
				PredictedRawData: tmpNodePredictionsData,
			},
		},
	}, nil
}

func (s *Server) ListPodRecommendations(ctx context.Context, in *datahub_v1alpha1.ListPodRecommendationsRequest) (*datahub_v1alpha1.ListPodRecommendationsResponse, error) {

	var tmpRecommendationsData = []*datahub_v1alpha1.MetricData{
		&datahub_v1alpha1.MetricData{
			MetricType: datahub_v1alpha1.MetricType_CONTAINER_CPU_USAGE_SECONDS_PERCENTAGE,
			Data: []*datahub_v1alpha1.Sample{
				&datahub_v1alpha1.Sample{
					Time:     tmpTimestamps[0],
					NumValue: "20",
				},
				&datahub_v1alpha1.Sample{
					Time:     tmpTimestamps[1],
					NumValue: "50",
				},
			},
		},
		&datahub_v1alpha1.MetricData{
			MetricType: datahub_v1alpha1.MetricType_CONTAINER_MEMORY_USAGE_BYTES,
			Data: []*datahub_v1alpha1.Sample{
				&datahub_v1alpha1.Sample{
					Time:     tmpTimestamps[0],
					NumValue: "512",
				},
				&datahub_v1alpha1.Sample{
					Time:     tmpTimestamps[1],
					NumValue: "1024",
				},
			},
		},
	}

	return &datahub_v1alpha1.ListPodRecommendationsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		PodRecommendations: []*datahub_v1alpha1.PodRecommendation{
			&datahub_v1alpha1.PodRecommendation{
				NamespacedName:         &datahub_v1alpha1.NamespacedName{Namespace: "openshift-monitoring", Name: "prometheus-k8s-0"},
				ApplyRecommendationNow: false,
				AssignPodPolicy: &datahub_v1alpha1.AssignPodPolicy{
					Time:   tmpTimestamps[0],
					Policy: &datahub_v1alpha1.AssignPodPolicy_NodeName{NodeName: "node1"},
				},
				ContainerRecommendations: []*datahub_v1alpha1.ContainerRecommendation{
					&datahub_v1alpha1.ContainerRecommendation{
						Name:                   "prometheus",
						LimitRecommendations:   tmpRecommendationsData,
						RequestRecommendations: tmpRecommendationsData,
					},
					&datahub_v1alpha1.ContainerRecommendation{
						Name:                   "another-container",
						LimitRecommendations:   tmpRecommendationsData,
						RequestRecommendations: tmpRecommendationsData,
					},
				},
			},
			&datahub_v1alpha1.PodRecommendation{
				NamespacedName:         &datahub_v1alpha1.NamespacedName{Namespace: "openshift-monitoring", Name: "prometheus-k8s-1"},
				ApplyRecommendationNow: false,
				AssignPodPolicy: &datahub_v1alpha1.AssignPodPolicy{
					Time:   tmpTimestamps[0],
					Policy: &datahub_v1alpha1.AssignPodPolicy_NodeName{NodeName: "node2"},
				},
				ContainerRecommendations: []*datahub_v1alpha1.ContainerRecommendation{
					&datahub_v1alpha1.ContainerRecommendation{
						Name:                   "prometheus",
						LimitRecommendations:   tmpRecommendationsData,
						RequestRecommendations: tmpRecommendationsData,
					},
					&datahub_v1alpha1.ContainerRecommendation{
						Name:                   "another-container",
						LimitRecommendations:   tmpRecommendationsData,
						RequestRecommendations: tmpRecommendationsData,
					},
				},
			},
		},
	}, nil
}

func (s *Server) ListPodsByNodeName(ctx context.Context, in *datahub_v1alpha1.ListPodsByNodeNameRequest) (*datahub_v1alpha1.ListPodsResponse, error) {

	var tmpMetricsData = []*datahub_v1alpha1.MetricData{
		&datahub_v1alpha1.MetricData{
			MetricType: datahub_v1alpha1.MetricType_CONTAINER_CPU_USAGE_SECONDS_PERCENTAGE,
			Data: []*datahub_v1alpha1.Sample{
				&datahub_v1alpha1.Sample{
					Time:     tmpTimestamps[0],
					NumValue: "20",
				},
				&datahub_v1alpha1.Sample{
					Time:     tmpTimestamps[1],
					NumValue: "50",
				},
			},
		},
		&datahub_v1alpha1.MetricData{
			MetricType: datahub_v1alpha1.MetricType_CONTAINER_MEMORY_USAGE_BYTES,
			Data: []*datahub_v1alpha1.Sample{
				&datahub_v1alpha1.Sample{
					Time:     tmpTimestamps[0],
					NumValue: "512",
				},
				&datahub_v1alpha1.Sample{
					Time:     tmpTimestamps[1],
					NumValue: "1024",
				},
			},
		},
	}

	return &datahub_v1alpha1.ListPodsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		Pods: []*datahub_v1alpha1.Pod{
			&datahub_v1alpha1.Pod{
				NamespacedName: &datahub_v1alpha1.NamespacedName{
					Namespace: "openshit-monitoring",
					Name:      "prometheus-k8s-0",
				},
				ResourceLink: "",
				Containers: []*datahub_v1alpha1.Container{
					&datahub_v1alpha1.Container{
						Name:                                 "prometheus",
						LimitResource:                        tmpMetricsData,
						RequestResource:                      tmpMetricsData,
						LimitResourceRecommendation:          tmpMetricsData,
						InitialLimitResourceRecommendation:   tmpMetricsData,
						InitialRequestResourceRecommendation: tmpMetricsData,
					},
					&datahub_v1alpha1.Container{
						Name:                                 "another-container",
						LimitResource:                        tmpMetricsData,
						RequestResource:                      tmpMetricsData,
						LimitResourceRecommendation:          tmpMetricsData,
						InitialLimitResourceRecommendation:   tmpMetricsData,
						InitialRequestResourceRecommendation: tmpMetricsData,
					},
				},
			},
			&datahub_v1alpha1.Pod{
				NamespacedName: &datahub_v1alpha1.NamespacedName{
					Namespace: "openshit-monitoring",
					Name:      "prometheus-k8s-1",
				},
				ResourceLink: "",
				Containers: []*datahub_v1alpha1.Container{
					&datahub_v1alpha1.Container{
						Name:                                 "prometheus",
						LimitResource:                        tmpMetricsData,
						RequestResource:                      tmpMetricsData,
						LimitResourceRecommendation:          tmpMetricsData,
						InitialLimitResourceRecommendation:   tmpMetricsData,
						InitialRequestResourceRecommendation: tmpMetricsData,
					},
					&datahub_v1alpha1.Container{
						Name:                                 "another-container",
						LimitResource:                        tmpMetricsData,
						RequestResource:                      tmpMetricsData,
						LimitResourceRecommendation:          tmpMetricsData,
						InitialLimitResourceRecommendation:   tmpMetricsData,
						InitialRequestResourceRecommendation: tmpMetricsData,
					},
				},
			},
		},
	}, nil
}

func (s *Server) CreateAlamedaPods(ctx context.Context, in *datahub_v1alpha1.CreateAlamedaPodsRequest) (*status.Status, error) {
	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *Server) CreateAlamedaNodes(ctx context.Context, in *datahub_v1alpha1.CreateAlamedaNodesRequest) (*status.Status, error) {
	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}
func (s *Server) CreatePodPredictions(ctx context.Context, in *datahub_v1alpha1.CreatePodPredictionsRequest) (*status.Status, error) {
	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *Server) CreateNodePredictions(ctx context.Context, in *datahub_v1alpha1.CreateNodePredictionsRequest) (*status.Status, error) {
	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *Server) CreatePodRecommendations(ctx context.Context, in *datahub_v1alpha1.CreatePodRecommendationsRequest) (*status.Status, error) {
	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *Server) DeleteAlamedaPods(ctx context.Context, in *datahub_v1alpha1.DeleteAlamedaPodsRequest) (*status.Status, error) {
	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *Server) DeleteAlamedaNodes(ctx context.Context, in *datahub_v1alpha1.DeleteAlamedaNodesRequest) (*status.Status, error) {
	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}
