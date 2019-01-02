package datahub

import (
	"errors"
	"fmt"
	"net"
	"time"

	cluster_status_dao "github.com/containers-ai/alameda/datahub/pkg/dao/cluster_status"
	cluster_status_dao_impl "github.com/containers-ai/alameda/datahub/pkg/dao/cluster_status/impl"
	"github.com/containers-ai/alameda/datahub/pkg/dao/metric"
	prometheusMetricDAO "github.com/containers-ai/alameda/datahub/pkg/dao/metric/prometheus"
	"github.com/containers-ai/alameda/pkg/utils/log"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes"
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
}

var (
	scope         = log.RegisterScope("gRPC", "gRPC server log", 0)
	tmpTimestamps = []*timestamp.Timestamp{
		&timestamp.Timestamp{Seconds: 1545809000},
		&timestamp.Timestamp{Seconds: 1545809030},
		&timestamp.Timestamp{Seconds: 1545809060},
		&timestamp.Timestamp{Seconds: 1545809090},
		&timestamp.Timestamp{Seconds: 1545809120},
		&timestamp.Timestamp{Seconds: 1545809150},
	}
)

func NewServer(cfg Config) (*Server, error) {

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

	server = &Server{
		err: make(chan error),

		Config:    cfg,
		K8SClient: k8sCli,
	}

	return server, nil
}

func (s *Server) Run() error {

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

	var (
		err error

		metricDAO metric.MetricsDAO

		requestExt     listPodMetricsRequestExtended
		namespace      = ""
		podName        = ""
		queryStartTime time.Time
		queryEndTime   time.Time

		podsMetricMap     metric.PodsMetricMap
		datahubPodMetrics []*datahub_v1alpha1.PodMetric

		apiInternalServerErrorResponse = datahub_v1alpha1.ListPodMetricsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: "Internal server error.",
			},
		}
	)

	requestExt = listPodMetricsRequestExtended(*in)
	if err = requestExt.validate(); err != nil {
		return &datahub_v1alpha1.ListPodMetricsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INVALID_ARGUMENT),
				Message: err.Error(),
			},
		}, nil
	}

	metricDAO = prometheusMetricDAO.NewWithConfig(*s.Config.Prometheus)

	if in.GetNamespacedName() != nil {
		namespace = in.GetNamespacedName().GetNamespace()
		podName = in.GetNamespacedName().GetName()
	}
	queryStartTime, err = ptypes.Timestamp(in.GetTimeRange().GetStartTime())
	if err != nil {
		return &apiInternalServerErrorResponse, nil
	}
	queryEndTime, err = ptypes.Timestamp(in.GetTimeRange().GetEndTime())
	if err != nil {
		return &apiInternalServerErrorResponse, nil
	}
	listPodMetricsRequest := metric.ListPodMetricsRequest{
		Namespace: namespace,
		PodName:   podName,
		StartTime: queryStartTime,
		EndTime:   queryEndTime,
	}

	podsMetricMap, err = metricDAO.ListPodMetrics(listPodMetricsRequest)
	if err != nil {
		scope.Error("ListPodMetrics failed: " + err.Error())
		return &apiInternalServerErrorResponse, nil
	}

	for _, podMetric := range podsMetricMap {
		podMetricExtended := podMetricExtended(podMetric)
		datahubPodMetric := podMetricExtended.datahubPodMetric()
		datahubPodMetrics = append(datahubPodMetrics, &datahubPodMetric)
	}

	return &datahub_v1alpha1.ListPodMetricsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		PodMetrics: datahubPodMetrics,
	}, nil
}

func (s *Server) ListNodeMetrics(ctx context.Context, in *datahub_v1alpha1.ListNodeMetricsRequest) (*datahub_v1alpha1.ListNodeMetricsResponse, error) {

	var (
		err error

		metricDAO metric.MetricsDAO

		requestExt     listNodeMetricsRequestExtended
		nodeNames      []string
		queryStartTime time.Time
		queryEndTime   time.Time

		nodesMetricMap     metric.NodesMetricMap
		datahubNodeMetrics []*datahub_v1alpha1.NodeMetric

		apiInternalServerErrorResponse = datahub_v1alpha1.ListNodeMetricsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: "Internal server error.",
			},
		}
	)

	requestExt = listNodeMetricsRequestExtended(*in)
	if err = requestExt.validate(); err != nil {
		return &datahub_v1alpha1.ListNodeMetricsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INVALID_ARGUMENT),
				Message: err.Error(),
			},
		}, nil
	}

	metricDAO = prometheusMetricDAO.NewWithConfig(*s.Config.Prometheus)

	nodeNames = in.GetNodeNames()
	queryStartTime, err = ptypes.Timestamp(in.GetTimeRange().GetStartTime())
	if err != nil {
		return &apiInternalServerErrorResponse, nil
	}
	queryEndTime, err = ptypes.Timestamp(in.GetTimeRange().GetEndTime())
	if err != nil {
		return &apiInternalServerErrorResponse, nil
	}
	listNodeMetricsRequest := metric.ListNodeMetricsRequest{
		NodeNames: nodeNames,
		StartTime: queryStartTime,
		EndTime:   queryEndTime,
	}

	nodesMetricMap, err = metricDAO.ListNodesMetric(listNodeMetricsRequest)
	if err != nil {
		scope.Error("ListPodMetrics failed: " + err.Error())
		return &apiInternalServerErrorResponse, nil
	}

	for _, nodeMetric := range nodesMetricMap {
		nodeMetricExtended := nodeMetricExtended(nodeMetric)
		datahubNodeMetric := nodeMetricExtended.datahubNodeMetric()
		datahubNodeMetrics = append(datahubNodeMetrics, &datahubNodeMetric)
	}

	return &datahub_v1alpha1.ListNodeMetricsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		NodeMetrics: datahubNodeMetrics,
	}, nil

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
								NumValue: "25",
							},
							&datahub_v1alpha1.Sample{
								Time:     tmpTimestamps[2],
								NumValue: "30",
							},
							&datahub_v1alpha1.Sample{
								Time:     tmpTimestamps[3],
								NumValue: "35",
							},
							&datahub_v1alpha1.Sample{
								Time:     tmpTimestamps[4],
								NumValue: "40",
							},
							&datahub_v1alpha1.Sample{
								Time:     tmpTimestamps[5],
								NumValue: "45",
							},
						},
					},
					&datahub_v1alpha1.MetricData{
						MetricType: datahub_v1alpha1.MetricType_NODE_MEMORY_USAGE_BYTES,
						Data: []*datahub_v1alpha1.Sample{
							&datahub_v1alpha1.Sample{
								Time:     tmpTimestamps[0],
								NumValue: "64",
							},
							&datahub_v1alpha1.Sample{
								Time:     tmpTimestamps[1],
								NumValue: "128",
							},
							&datahub_v1alpha1.Sample{
								Time:     tmpTimestamps[2],
								NumValue: "152",
							},
							&datahub_v1alpha1.Sample{
								Time:     tmpTimestamps[3],
								NumValue: "176",
							},
							&datahub_v1alpha1.Sample{
								Time:     tmpTimestamps[4],
								NumValue: "200",
							},
							&datahub_v1alpha1.Sample{
								Time:     tmpTimestamps[5],
								NumValue: "224",
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
								NumValue: "25",
							},
							&datahub_v1alpha1.Sample{
								Time:     tmpTimestamps[2],
								NumValue: "30",
							},
							&datahub_v1alpha1.Sample{
								Time:     tmpTimestamps[3],
								NumValue: "35",
							},
							&datahub_v1alpha1.Sample{
								Time:     tmpTimestamps[4],
								NumValue: "40",
							},
							&datahub_v1alpha1.Sample{
								Time:     tmpTimestamps[5],
								NumValue: "45",
							},
						},
					},
					&datahub_v1alpha1.MetricData{
						MetricType: datahub_v1alpha1.MetricType_NODE_MEMORY_USAGE_BYTES,
						Data: []*datahub_v1alpha1.Sample{
							&datahub_v1alpha1.Sample{
								Time:     tmpTimestamps[0],
								NumValue: "64",
							},
							&datahub_v1alpha1.Sample{
								Time:     tmpTimestamps[1],
								NumValue: "128",
							},
							&datahub_v1alpha1.Sample{
								Time:     tmpTimestamps[2],
								NumValue: "152",
							},
							&datahub_v1alpha1.Sample{
								Time:     tmpTimestamps[3],
								NumValue: "176",
							},
							&datahub_v1alpha1.Sample{
								Time:     tmpTimestamps[4],
								NumValue: "200",
							},
							&datahub_v1alpha1.Sample{
								Time:     tmpTimestamps[5],
								NumValue: "224",
							},
						},
					},
				},
			},
		},
	}, nil
}

func (s *Server) ListAlamedaPods(ctx context.Context, in *datahub_v1alpha1.ListAlamedaPodsRequest) (*datahub_v1alpha1.ListPodsResponse, error) {

	return &datahub_v1alpha1.ListPodsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		Pods: []*datahub_v1alpha1.Pod{
			&datahub_v1alpha1.Pod{
				NamespacedName: &datahub_v1alpha1.NamespacedName{
					Namespace: "alameda",
					Name:      "nginx-deployment-f99bb8986-h8rbk",
				},
				ResourceLink: "/namespaces/alameda/deployments/nginx-deployment/replicasets/nginx-deployment-f99bb8986/pods/nginx-deployment-f99bb8986-h8rbk",
				IsAlameda:    true,
				NodeName:     "node1",
			},
			&datahub_v1alpha1.Pod{
				NamespacedName: &datahub_v1alpha1.NamespacedName{
					Namespace: "alameda",
					Name:      "nginx-deployment-f99bb8986-npg2f",
				},
				ResourceLink: "/namespaces/alameda/deployments/nginx-deployment/replicasets/nginx-deployment-f99bb8986/pods/nginx-deployment-f99bb8986-npg2f",
				IsAlameda:    true,
				NodeName:     "node2",
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
	/*
		nodeDAO := &cluster_status_dao_impl.Node{
			InfluxDBConfig: *s.Config.InfluxDB,
		}
		nodes := []*datahub_v1alpha1.Node{}

		if alamedaNodes, err := nodeDAO.ListAlamedaNodes(); err != nil {
			scope.Error(err.Error())
			return &datahub_v1alpha1.ListNodesResponse{
				Status: &status.Status{
					Code: int32(code.Code_INTERNAL),
				},
			}, err
		} else {
			for _, alamedaNode := range alamedaNodes {
				nodes = append(nodes, &datahub_v1alpha1.Node{
					Name: alamedaNode.Name,
				})
			}
			return &datahub_v1alpha1.ListNodesResponse{
				Status: &status.Status{
					Code: int32(code.Code_OK),
				},
				Nodes: nodes,
			}, nil
		}
	*/
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
	nodeDAO := &cluster_status_dao_impl.Node{
		InfluxDBConfig: *s.Config.InfluxDB,
	}
	alamedaNodeList := []*cluster_status_dao.AlamedaNode{}
	for _, alamedaNode := range in.GetAlamedaNodes() {
		alamedaNodeList = append(alamedaNodeList, &cluster_status_dao.AlamedaNode{
			Name: alamedaNode.GetName(),
		})
	}
	if err := nodeDAO.RegisterAlamedaNodes(alamedaNodeList); err != nil {
		scope.Error(err.Error())
		return &status.Status{
			Code: int32(code.Code_INTERNAL),
		}, err
	}

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
	nodeDAO := &cluster_status_dao_impl.Node{
		InfluxDBConfig: *s.Config.InfluxDB,
	}
	alamedaNodeList := []*cluster_status_dao.AlamedaNode{}
	for _, alamedaNode := range in.GetAlamedaNodes() {
		alamedaNodeList = append(alamedaNodeList, &cluster_status_dao.AlamedaNode{
			Name: alamedaNode.GetName(),
		})
	}
	if err := nodeDAO.DeregisterAlamedaNodes(alamedaNodeList); err != nil {
		scope.Error(err.Error())
		return &status.Status{
			Code: int32(code.Code_INTERNAL),
		}, err
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}
