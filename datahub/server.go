package datahub

import (
	"fmt"
	DaoClusterStatus "github.com/containers-ai/alameda/datahub/pkg/dao/cluster_status"
	DaoClusterStatusImpl "github.com/containers-ai/alameda/datahub/pkg/dao/cluster_status/impl"
	DaoMetric "github.com/containers-ai/alameda/datahub/pkg/dao/metric"
	DaoMetricInflux "github.com/containers-ai/alameda/datahub/pkg/dao/metric/influxdb"
	DaoMetricPromth "github.com/containers-ai/alameda/datahub/pkg/dao/metric/prometheus"
	DaoPredictionImpl "github.com/containers-ai/alameda/datahub/pkg/dao/prediction/impl"
	DaoRecommendation "github.com/containers-ai/alameda/datahub/pkg/dao/recommendation"
	DaoRecommendationImpl "github.com/containers-ai/alameda/datahub/pkg/dao/recommendation/impl"
	DaoScore "github.com/containers-ai/alameda/datahub/pkg/dao/score"
	DaoScoreImplInflux "github.com/containers-ai/alameda/datahub/pkg/dao/score/impl/influxdb"
	DaoWeaveScope "github.com/containers-ai/alameda/datahub/pkg/dao/weavescope"
	DatahubUtils "github.com/containers-ai/alameda/datahub/pkg/utils"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
	OperatorAPIs "github.com/containers-ai/alameda/operator/pkg/apis"
	OperatorAPIsAutoScalingV1Alpha1 "github.com/containers-ai/alameda/operator/pkg/apis/autoscaling/v1alpha1"
	OperatorReconcilerAlamedaRecommendation "github.com/containers-ai/alameda/operator/pkg/reconciler/alamedarecommendation"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	AlamedaLog "github.com/containers-ai/alameda/pkg/utils/log"
	DatahubV1Alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	Common "github.com/containers-ai/api/common"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	K8SErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"net"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type Server struct {
	err    chan error
	server *grpc.Server

	Config    Config
	K8SClient client.Client
}

var (
	scope         = AlamedaLog.RegisterScope("gRPC", "gRPC server log", 0)
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

	// build server listener
	scope.Info("starting gRPC server")
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
		"alameda_metric",
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

func (s *Server) registGRPCServer(server *grpc.Server) {

	DatahubV1Alpha1.RegisterDatahubServiceServer(server, s)
}

func (s *Server) ListPodMetrics(ctx context.Context, in *DatahubV1Alpha1.ListPodMetricsRequest) (*DatahubV1Alpha1.ListPodMetricsResponse, error) {
	scope.Debug("Request received from ListPodMetrics grpc function: " + AlamedaUtils.InterfaceToString(in))

	//--------------------------------------------------------
	_, err := os.Stat("metric_cpu.csv")
	if !os.IsNotExist(err) {
		return s.ListPodMetricsDemo(ctx, in)
	}

	//--------------------------------------------------------
	metricDAO := DaoMetricInflux.NewWithConfig(*s.Config.InfluxDB)
	podMetricList, err := metricDAO.ListContainerMetrics(in)

	if err != nil {
		return &DatahubV1Alpha1.ListPodMetricsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	return &DatahubV1Alpha1.ListPodMetricsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		PodMetrics: podMetricList,
	}, nil
}

// ListPodMetrics list pods' metrics
func (s *Server) ListPodMetricsPrometheus(ctx context.Context, in *DatahubV1Alpha1.ListPodMetricsRequest) (*DatahubV1Alpha1.ListPodMetricsResponse, error) {
	//func (s *Server) ListPodMetrics(ctx context.Context, in *DatahubV1Alpha1.ListPodMetricsRequest) (*DatahubV1Alpha1.ListPodMetricsResponse, error) {
	scope.Debug("Request received from ListPodMetrics grpc function: " + AlamedaUtils.InterfaceToString(in))

	//--------------------------------------------------------
	_, err := os.Stat("metric_cpu.csv")
	if !os.IsNotExist(err) {
		return s.ListPodMetricsDemo(ctx, in)
	}

	//--------------------------------------------------------
	var (
		metricDAO DaoMetric.MetricsDAO

		requestExt     datahubListPodMetricsRequestExtended
		namespace      = ""
		podName        = ""
		queryCondition DBCommon.QueryCondition

		podsMetricMap     DaoMetric.PodsMetricMap
		datahubPodMetrics []*DatahubV1Alpha1.PodMetric
	)

	requestExt = datahubListPodMetricsRequestExtended{*in}
	if err = requestExt.validate(); err != nil {
		return &DatahubV1Alpha1.ListPodMetricsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INVALID_ARGUMENT),
				Message: err.Error(),
			},
		}, nil
	}

	metricDAO = DaoMetricPromth.NewWithConfig(*s.Config.Prometheus)

	if in.GetNamespacedName() != nil {
		namespace = in.GetNamespacedName().GetNamespace()
		podName = in.GetNamespacedName().GetName()
	}
	queryCondition = datahubQueryConditionExtend{queryCondition: in.GetQueryCondition()}.daoQueryCondition()
	listPodMetricsRequest := DaoMetric.ListPodMetricsRequest{
		Namespace:      namespace,
		PodName:        podName,
		QueryCondition: queryCondition,
	}

	podsMetricMap, err = metricDAO.ListPodMetrics(listPodMetricsRequest)
	if err != nil {
		scope.Errorf("ListPodMetrics failed: %+v", err)
		return &DatahubV1Alpha1.ListPodMetricsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	for _, podMetric := range podsMetricMap {
		podMetricExtended := daoPodMetricExtended{podMetric}
		datahubPodMetric := podMetricExtended.datahubPodMetric()
		datahubPodMetrics = append(datahubPodMetrics, datahubPodMetric)
	}

	return &DatahubV1Alpha1.ListPodMetricsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		PodMetrics: datahubPodMetrics,
	}, nil
}

// ListPodMetrics list pods' metrics for demo
func (s *Server) ListPodMetricsDemo(ctx context.Context, in *DatahubV1Alpha1.ListPodMetricsRequest) (*DatahubV1Alpha1.ListPodMetricsResponse, error) {
	scope.Debug("Request received from ListPodMetricsDemo grpc function: " + AlamedaUtils.InterfaceToString(in))

	demoPodMetricList := make([]*DatahubV1Alpha1.PodMetric, 0)
	endTime := in.GetQueryCondition().GetTimeRange().GetEndTime().GetSeconds()

	if endTime == 0 {
		return &DatahubV1Alpha1.ListPodMetricsResponse{
			Status: &status.Status{
				Code: int32(code.Code_INVALID_ARGUMENT),
			},
			PodMetrics: demoPodMetricList,
		}, errors.Errorf("Invalid EndTime")
	}

	if endTime%3600 != 0 {
		endTime = endTime - (endTime % 3600) + 3600
	}

	//step := int(in.GetQueryCondition().GetTimeRange().GetStep().GetSeconds())
	step := 3600
	if step == 0 {
		step = 3600
	}

	tempNamespacedName := DatahubV1Alpha1.NamespacedName{
		Namespace: in.NamespacedName.Namespace,
		Name:      in.NamespacedName.Name,
	}

	demoContainerMetricList := make([]*DatahubV1Alpha1.ContainerMetric, 0)
	demoContainerMetric := DatahubV1Alpha1.ContainerMetric{
		Name:       in.NamespacedName.Name,
		MetricData: make([]*DatahubV1Alpha1.MetricData, 0),
	}
	demoContainerMetricList = append(demoContainerMetricList, &demoContainerMetric)

	demoMetricDataCPU := DatahubV1Alpha1.MetricData{
		MetricType: DatahubV1Alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
		Data:       make([]*DatahubV1Alpha1.Sample, 0),
	}

	demoMetricDataMem := DatahubV1Alpha1.MetricData{
		MetricType: DatahubV1Alpha1.MetricType_MEMORY_USAGE_BYTES,
		Data:       make([]*DatahubV1Alpha1.Sample, 0),
	}

	demoDataMapCPU, _ := DatahubUtils.ReadCSV("metric_cpu.csv")
	demoDataMapMem, _ := DatahubUtils.ReadCSV("metric_memory.csv")

	demoKey := in.NamespacedName.Namespace + "_" + in.NamespacedName.Name

	startTime := endTime - int64(step*len(demoDataMapCPU[demoKey]))
	for index, value := range demoDataMapCPU[demoKey] {
		second := startTime + int64(index*step)
		demoMetricDataCPU.Data = append(demoMetricDataCPU.Data, &DatahubV1Alpha1.Sample{
			Time:     &timestamp.Timestamp{Seconds: int64(second)},
			NumValue: value,
		})
	}

	for index, value := range demoDataMapMem[demoKey] {
		second := startTime + int64(index*step)
		demoMetricDataMem.Data = append(demoMetricDataMem.Data, &DatahubV1Alpha1.Sample{
			Time:     &timestamp.Timestamp{Seconds: int64(second)},
			NumValue: value,
		})
	}

	demoContainerMetric.MetricData = append(demoContainerMetric.MetricData, &demoMetricDataCPU)
	demoContainerMetric.MetricData = append(demoContainerMetric.MetricData, &demoMetricDataMem)

	demoPodMetric := DatahubV1Alpha1.PodMetric{
		NamespacedName:   &tempNamespacedName,
		ContainerMetrics: demoContainerMetricList,
	}
	demoPodMetricList = append(demoPodMetricList, &demoPodMetric)

	return &DatahubV1Alpha1.ListPodMetricsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		PodMetrics: demoPodMetricList,
	}, nil
}

func (s *Server) ListNodeMetrics(ctx context.Context, in *DatahubV1Alpha1.ListNodeMetricsRequest) (*DatahubV1Alpha1.ListNodeMetricsResponse, error) {
	scope.Debug("Request received from ListNodeMetrics grpc function: " + AlamedaUtils.InterfaceToString(in))

	metricDAO := DaoMetricInflux.NewWithConfig(*s.Config.InfluxDB)
	nodeMetricList, err := metricDAO.ListNodeMetrics(in)

	if err != nil {
		return &DatahubV1Alpha1.ListNodeMetricsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	return &DatahubV1Alpha1.ListNodeMetricsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		NodeMetrics: nodeMetricList,
	}, nil
}

// ListNodeMetrics list nodes' metrics
func (s *Server) ListNodeMetricsPromethues(ctx context.Context, in *DatahubV1Alpha1.ListNodeMetricsRequest) (*DatahubV1Alpha1.ListNodeMetricsResponse, error) {
	scope.Debug("Request received from ListNodeMetrics grpc function: " + AlamedaUtils.InterfaceToString(in))

	var (
		err error

		metricDAO DaoMetric.MetricsDAO

		requestExt     datahubListNodeMetricsRequestExtended
		nodeNames      []string
		queryCondition DBCommon.QueryCondition

		nodesMetricMap     DaoMetric.NodesMetricMap
		datahubNodeMetrics []*DatahubV1Alpha1.NodeMetric
	)

	requestExt = datahubListNodeMetricsRequestExtended{*in}
	if err = requestExt.validate(); err != nil {
		return &DatahubV1Alpha1.ListNodeMetricsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INVALID_ARGUMENT),
				Message: err.Error(),
			},
		}, nil
	}

	metricDAO = DaoMetricPromth.NewWithConfig(*s.Config.Prometheus)

	nodeNames = in.GetNodeNames()
	queryCondition = datahubQueryConditionExtend{queryCondition: in.GetQueryCondition()}.daoQueryCondition()
	listNodeMetricsRequest := DaoMetric.ListNodeMetricsRequest{
		NodeNames:      nodeNames,
		QueryCondition: queryCondition,
	}

	nodesMetricMap, err = metricDAO.ListNodesMetric(listNodeMetricsRequest)
	if err != nil {
		scope.Errorf("ListNodeMetrics failed: %+v", err)
		return &DatahubV1Alpha1.ListNodeMetricsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	for _, nodeMetric := range nodesMetricMap {
		nodeMetricExtended := daoNodeMetricExtended{nodeMetric}
		datahubNodeMetric := nodeMetricExtended.datahubNodeMetric()
		datahubNodeMetrics = append(datahubNodeMetrics, datahubNodeMetric)
	}

	return &DatahubV1Alpha1.ListNodeMetricsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		NodeMetrics: datahubNodeMetrics,
	}, nil
}

// ListAlamedaPods returns predicted pods
func (s *Server) ListAlamedaPods(ctx context.Context, in *DatahubV1Alpha1.ListAlamedaPodsRequest) (*DatahubV1Alpha1.ListPodsResponse, error) {

	scope.Debug("Request received from ListAlamedaPods grpc function: " + AlamedaUtils.InterfaceToString(in))

	var containerDAO DaoClusterStatus.ContainerOperation = &DaoClusterStatusImpl.Container{
		InfluxDBConfig: *s.Config.InfluxDB,
	}

	namespace, name := "", ""
	if namespacedName := in.GetNamespacedName(); namespacedName != nil {
		namespace = namespacedName.GetNamespace()
		name = namespacedName.GetName()
	}
	kind := in.GetKind()
	timeRange := in.GetTimeRange()

	if alamedaPods, err := containerDAO.ListAlamedaPods(namespace, name, kind, timeRange); err != nil {
		scope.Error(err.Error())
		return &DatahubV1Alpha1.ListPodsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	} else {
		res := &DatahubV1Alpha1.ListPodsResponse{
			Pods: alamedaPods,
			Status: &status.Status{
				Code: int32(code.Code_OK),
			},
		}
		scope.Debug("Request sent from ListAlamedaPods grpc function: " + AlamedaUtils.InterfaceToString(res))
		return res, nil
	}
}

// ListAlamedaNodes list nodes in cluster
func (s *Server) ListAlamedaNodes(ctx context.Context, in *DatahubV1Alpha1.ListAlamedaNodesRequest) (*DatahubV1Alpha1.ListNodesResponse, error) {
	scope.Debug("Request received from ListAlamedaNodes grpc function: " + AlamedaUtils.InterfaceToString(in))

	var nodeDAO DaoClusterStatus.NodeOperation = &DaoClusterStatusImpl.Node{
		InfluxDBConfig: *s.Config.InfluxDB,
	}

	timeRange := in.GetTimeRange()

	if alamedaNodes, err := nodeDAO.ListAlamedaNodes(timeRange); err != nil {
		scope.Error(err.Error())
		return &DatahubV1Alpha1.ListNodesResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	} else {
		return &DatahubV1Alpha1.ListNodesResponse{
			Status: &status.Status{
				Code: int32(code.Code_OK),
			},
			Nodes: alamedaNodes,
		}, nil
	}
}

func (s *Server) ListNodes(ctx context.Context, in *DatahubV1Alpha1.ListNodesRequest) (*DatahubV1Alpha1.ListNodesResponse, error) {
	scope.Debug("Request received from ListNodes grpc function: " + AlamedaUtils.InterfaceToString(in))

	var nodeDAO DaoClusterStatus.NodeOperation = &DaoClusterStatusImpl.Node{
		InfluxDBConfig: *s.Config.InfluxDB,
	}

	req := DaoClusterStatus.ListNodesRequest{
		NodeNames: in.GetNodeNames(),
		InCluster: true,
	}
	if nodes, err := nodeDAO.ListNodes(req); err != nil {
		scope.Error(err.Error())
		return &DatahubV1Alpha1.ListNodesResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	} else {
		return &DatahubV1Alpha1.ListNodesResponse{
			Status: &status.Status{
				Code: int32(code.Code_OK),
			},
			Nodes: nodes,
		}, nil
	}
}

func (s *Server) ListControllers(ctx context.Context, in *DatahubV1Alpha1.ListControllersRequest) (*DatahubV1Alpha1.ListControllersResponse, error) {
	scope.Debug("Request received from ListControllers grpc function: " + AlamedaUtils.InterfaceToString(in))

	controllerDAO := &DaoClusterStatusImpl.Controller{
		InfluxDBConfig: *s.Config.InfluxDB,
	}

	controllers, err := controllerDAO.ListControllers(in)
	if err != nil {
		scope.Error(err.Error())
		return &DatahubV1Alpha1.ListControllersResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	response := DatahubV1Alpha1.ListControllersResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		Controllers: controllers,
	}
	return &response, nil
}

// ListPodPredictions list pods' predictions
func (s *Server) ListPodPredictions(ctx context.Context, in *DatahubV1Alpha1.ListPodPredictionsRequest) (*DatahubV1Alpha1.ListPodPredictionsResponse, error) {
	scope.Debug("Request received from ListPodPredictions grpc function: " + AlamedaUtils.InterfaceToString(in))

	//--------------------------------------------------------
	_, err := os.Stat("prediction_cpu.csv")
	if !os.IsNotExist(err) {
		return s.ListPodPredictionsDemo(ctx, in)
	}

	//--------------------------------------------------------
	predictionDAO := DaoPredictionImpl.NewInfluxDBWithConfig(*s.Config.InfluxDB)

	datahubListPodPredictionsRequestExtended := datahubListPodPredictionsRequestExtended{in}
	listPodPredictionsRequest := datahubListPodPredictionsRequestExtended.daoListPodPredictionsRequest()

	podsPredictions, err := predictionDAO.ListPodPredictions(listPodPredictionsRequest)

	if err != nil {
		scope.Errorf("ListPodPrediction failed: %+v", err)
		return &DatahubV1Alpha1.ListPodPredictionsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	if in.GetFillDays() > 0 {
		predictionDAO.FillPodPredictions(podsPredictions, in.GetFillDays())
	}

	return &DatahubV1Alpha1.ListPodPredictionsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		PodPredictions: podsPredictions,
	}, nil
}

// ListPodPredictions list pods' predictions for demo
func (s *Server) ListPodPredictionsDemo(ctx context.Context, in *DatahubV1Alpha1.ListPodPredictionsRequest) (*DatahubV1Alpha1.ListPodPredictionsResponse, error) {
	scope.Debug("Request received from ListPodPredictionsDemo grpc function: " + AlamedaUtils.InterfaceToString(in))

	demoPodPredictionList := make([]*DatahubV1Alpha1.PodPrediction, 0)
	endTime := in.GetQueryCondition().GetTimeRange().GetEndTime().GetSeconds()

	if endTime == 0 {
		return &DatahubV1Alpha1.ListPodPredictionsResponse{
			Status: &status.Status{
				Code: int32(code.Code_INVALID_ARGUMENT),
			},
			PodPredictions: demoPodPredictionList,
		}, errors.Errorf("Invalid EndTime")
	}

	if endTime%3600 != 0 {
		endTime = endTime - (endTime % 3600) + 3600
	}

	//step := int(in.GetQueryCondition().GetTimeRange().GetStep().GetSeconds())
	step := 3600
	if step == 0 {
		step = 3600
	}

	if endTime == 0 {
		return &DatahubV1Alpha1.ListPodPredictionsResponse{
			Status: &status.Status{
				Code: int32(code.Code_INVALID_ARGUMENT),
			},
			PodPredictions: demoPodPredictionList,
		}, errors.Errorf("Invalid EndTime")
	}

	tempNamespacedName := DatahubV1Alpha1.NamespacedName{
		Namespace: in.NamespacedName.Namespace,
		Name:      in.NamespacedName.Name,
	}

	demoContainerPredictionList := make([]*DatahubV1Alpha1.ContainerPrediction, 0)
	demoContainerPrediction := DatahubV1Alpha1.ContainerPrediction{
		Name:             in.NamespacedName.Name,
		PredictedRawData: make([]*DatahubV1Alpha1.MetricData, 0),
	}
	demoContainerPredictionList = append(demoContainerPredictionList, &demoContainerPrediction)

	demoPredictionDataCPU := DatahubV1Alpha1.MetricData{
		MetricType: DatahubV1Alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
		Data:       make([]*DatahubV1Alpha1.Sample, 0),
	}

	demoPredictionDataMem := DatahubV1Alpha1.MetricData{
		MetricType: DatahubV1Alpha1.MetricType_MEMORY_USAGE_BYTES,
		Data:       make([]*DatahubV1Alpha1.Sample, 0),
	}

	demoDataMapCPU, _ := DatahubUtils.ReadCSV("prediction_cpu.csv")
	demoDataMapMem, _ := DatahubUtils.ReadCSV("prediction_memory.csv")

	demoKey := in.NamespacedName.Namespace + "_" + in.NamespacedName.Name
	startTime := endTime - int64(step*len(demoDataMapCPU[demoKey]))

	for index, value := range demoDataMapCPU[demoKey] {
		second := startTime + int64(index*step)
		demoPredictionDataCPU.Data = append(demoPredictionDataCPU.Data, &DatahubV1Alpha1.Sample{
			Time:     &timestamp.Timestamp{Seconds: int64(second)},
			NumValue: value,
		})
	}

	for index, value := range demoDataMapMem[demoKey] {
		second := startTime + int64(index*step)
		demoPredictionDataMem.Data = append(demoPredictionDataMem.Data, &DatahubV1Alpha1.Sample{
			Time:     &timestamp.Timestamp{Seconds: int64(second)},
			NumValue: value,
		})
	}

	demoContainerPrediction.PredictedRawData = append(demoContainerPrediction.PredictedRawData, &demoPredictionDataCPU)
	demoContainerPrediction.PredictedRawData = append(demoContainerPrediction.PredictedRawData, &demoPredictionDataMem)

	demoPodMetric := DatahubV1Alpha1.PodPrediction{
		NamespacedName:       &tempNamespacedName,
		ContainerPredictions: demoContainerPredictionList,
	}
	demoPodPredictionList = append(demoPodPredictionList, &demoPodMetric)

	return &DatahubV1Alpha1.ListPodPredictionsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		PodPredictions: demoPodPredictionList,
	}, nil
}

// ListNodePredictions list nodes' predictions
func (s *Server) ListNodePredictions(ctx context.Context, in *DatahubV1Alpha1.ListNodePredictionsRequest) (*DatahubV1Alpha1.ListNodePredictionsResponse, error) {
	scope.Debug("Request received from ListNodePredictions grpc function: " + AlamedaUtils.InterfaceToString(in))

	predictionDAO := DaoPredictionImpl.NewInfluxDBWithConfig(*s.Config.InfluxDB)

	datahubListNodePredictionsRequestExtended := datahubListNodePredictionsRequestExtended{in}
	listNodePredictionRequest := datahubListNodePredictionsRequestExtended.daoListNodePredictionsRequest()
	nodePredictions, err := predictionDAO.ListNodePredictions(listNodePredictionRequest)
	if err != nil {
		scope.Errorf("ListNodePredictions failed: %+v", err)
		return &DatahubV1Alpha1.ListNodePredictionsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	return &DatahubV1Alpha1.ListNodePredictionsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		NodePredictions: nodePredictions,
	}, nil
}

// ListPodRecommendations list pod recommendations
func (s *Server) ListPodRecommendations(ctx context.Context, in *DatahubV1Alpha1.ListPodRecommendationsRequest) (*DatahubV1Alpha1.ListPodRecommendationsResponse, error) {
	scope.Debug("Request received from ListPodRecommendations grpc function: " + AlamedaUtils.InterfaceToString(in))
	var containerDAO DaoRecommendation.ContainerOperation = &DaoRecommendationImpl.Container{
		InfluxDBConfig: *s.Config.InfluxDB,
	}

	podRecommendations, err := containerDAO.ListPodRecommendations(in)
	if err != nil {
		scope.Error(err.Error())
		return &DatahubV1Alpha1.ListPodRecommendationsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	res := &DatahubV1Alpha1.ListPodRecommendationsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		PodRecommendations: podRecommendations,
	}
	scope.Debug("Response sent from ListPodRecommendations grpc function: " + AlamedaUtils.InterfaceToString(res))
	return res, nil
}

// ListAvailablePodRecommendations list pod recommendations
func (s *Server) ListAvailablePodRecommendations(ctx context.Context, in *DatahubV1Alpha1.ListPodRecommendationsRequest) (*DatahubV1Alpha1.ListPodRecommendationsResponse, error) {
	scope.Debug("Request received from ListAvailablePodRecommendations grpc function: " + AlamedaUtils.InterfaceToString(in))
	containerDAO := &DaoRecommendationImpl.Container{
		InfluxDBConfig: *s.Config.InfluxDB,
	}

	podRecommendations, err := containerDAO.ListAvailablePodRecommendations(in)
	if err != nil {
		scope.Error(err.Error())
		return &DatahubV1Alpha1.ListPodRecommendationsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	res := &DatahubV1Alpha1.ListPodRecommendationsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		PodRecommendations: podRecommendations,
	}
	scope.Debug("Response sent from ListPodRecommendations grpc function: " + AlamedaUtils.InterfaceToString(res))
	return res, nil

}

// ListControllerRecommendations list controller recommendations
func (s *Server) ListControllerRecommendations(ctx context.Context, in *DatahubV1Alpha1.ListControllerRecommendationsRequest) (*DatahubV1Alpha1.ListControllerRecommendationsResponse, error) {
	scope.Debug("Request received from ListControllerRecommendations grpc function: " + AlamedaUtils.InterfaceToString(in))

	controllerDAO := &DaoRecommendationImpl.Controller{
		InfluxDBConfig: *s.Config.InfluxDB,
	}

	controllerRecommendations, err := controllerDAO.ListControllerRecommendations(in)
	if err != nil {
		scope.Errorf("api ListControllerRecommendations failed: %v", err)
		response := &DatahubV1Alpha1.ListControllerRecommendationsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
			ControllerRecommendations: controllerRecommendations,
		}
		return response, nil
	}

	response := &DatahubV1Alpha1.ListControllerRecommendationsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		ControllerRecommendations: controllerRecommendations,
	}

	scope.Debug("Response sent from ListControllerRecommendations grpc function: " + AlamedaUtils.InterfaceToString(response))
	return response, nil

}

// ListPodsByNodeName list pods running on specific nodes
func (s *Server) ListPodsByNodeName(ctx context.Context, in *DatahubV1Alpha1.ListPodsByNodeNamesRequest) (*DatahubV1Alpha1.ListPodsResponse, error) {
	scope.Debug("Request received from ListPodsByNodeName grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &DatahubV1Alpha1.ListPodsResponse{
		Status: &status.Status{
			Code:    int32(code.Code_OK),
			Message: "This function is deprecated.",
		},
	}, nil
}

// ListSimulatedSchedulingScores list simulated scheduling scores
func (s *Server) ListSimulatedSchedulingScores(ctx context.Context, in *DatahubV1Alpha1.ListSimulatedSchedulingScoresRequest) (*DatahubV1Alpha1.ListSimulatedSchedulingScoresResponse, error) {
	scope.Debug("Request received from ListSimulatedSchedulingScores grpc function: " + AlamedaUtils.InterfaceToString(in))

	var (
		err error

		scoreDAO                          DaoScore.DAO
		scoreDAOListRequest               DaoScore.ListRequest
		scoreDAOSimulatedSchedulingScores = make([]*DaoScore.SimulatedSchedulingScore, 0)

		datahubScores = make([]*DatahubV1Alpha1.SimulatedSchedulingScore, 0)
	)

	scoreDAO = DaoScoreImplInflux.NewWithConfig(*s.Config.InfluxDB)

	datahubListSimulatedSchedulingScoresRequestExtended := datahubListSimulatedSchedulingScoresRequestExtended{in}
	scoreDAOListRequest = datahubListSimulatedSchedulingScoresRequestExtended.daoLisRequest()
	scoreDAOSimulatedSchedulingScores, err = scoreDAO.ListSimulatedScheduingScores(scoreDAOListRequest)
	if err != nil {
		scope.Errorf("api ListSimulatedSchedulingScores failed: %v", err)
		return &DatahubV1Alpha1.ListSimulatedSchedulingScoresResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
			Scores: datahubScores,
		}, nil
	}

	for _, daoSimulatedSchedulingScore := range scoreDAOSimulatedSchedulingScores {

		t, err := ptypes.TimestampProto(daoSimulatedSchedulingScore.Timestamp)
		if err != nil {
			scope.Warnf("api ListSimulatedSchedulingScores warn: time convert failed: %s", err.Error())
		}
		datahubScore := DatahubV1Alpha1.SimulatedSchedulingScore{
			Time:        t,
			ScoreBefore: float32(daoSimulatedSchedulingScore.ScoreBefore),
			ScoreAfter:  float32(daoSimulatedSchedulingScore.ScoreAfter),
		}
		datahubScores = append(datahubScores, &datahubScore)
	}

	return &DatahubV1Alpha1.ListSimulatedSchedulingScoresResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		Scores: datahubScores,
	}, nil
}

// CreatePods add containers information of pods to database
func (s *Server) CreatePods(ctx context.Context, in *DatahubV1Alpha1.CreatePodsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreatePods grpc function: " + AlamedaUtils.InterfaceToString(in))
	var containerDAO DaoClusterStatus.ContainerOperation = &DaoClusterStatusImpl.Container{
		InfluxDBConfig: *s.Config.InfluxDB,
	}

	if err := containerDAO.AddPods(in.GetPods()); err != nil {
		scope.Error(err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}
	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *Server) CreateControllers(ctx context.Context, in *DatahubV1Alpha1.CreateControllersRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateControllers grpc function: " + AlamedaUtils.InterfaceToString(in))

	controllerDAO := &DaoClusterStatusImpl.Controller{
		InfluxDBConfig: *s.Config.InfluxDB,
	}

	err := controllerDAO.CreateControllers(in.GetControllers())
	if err != nil {
		scope.Error(err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *Server) DeleteControllers(ctx context.Context, in *DatahubV1Alpha1.DeleteControllersRequest) (*status.Status, error) {
	scope.Debug("Request received from DeleteControllers grpc function: " + AlamedaUtils.InterfaceToString(in))

	controllerDAO := &DaoClusterStatusImpl.Controller{
		InfluxDBConfig: *s.Config.InfluxDB,
	}

	err := controllerDAO.DeleteControllers(in)
	if err != nil {
		scope.Error(err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

// DeletePods update containers information of pods to database
func (s *Server) DeletePods(ctx context.Context, in *DatahubV1Alpha1.DeletePodsRequest) (*status.Status, error) {
	scope.Debug("Request received from DeletePods grpc function: " + AlamedaUtils.InterfaceToString(in))

	var containerDAO DaoClusterStatus.ContainerOperation = &DaoClusterStatusImpl.Container{
		InfluxDBConfig: *s.Config.InfluxDB,
	}
	if err := containerDAO.DeletePods(in.GetPods()); err != nil {
		scope.Errorf("DeletePods failed: %+v", err)
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}
	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

// CreateAlamedaNodes add node information to database
func (s *Server) CreateAlamedaNodes(ctx context.Context, in *DatahubV1Alpha1.CreateAlamedaNodesRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateAlamedaNodes grpc function: " + AlamedaUtils.InterfaceToString(in))

	var nodeDAO DaoClusterStatus.NodeOperation = &DaoClusterStatusImpl.Node{
		InfluxDBConfig: *s.Config.InfluxDB,
	}
	if err := nodeDAO.RegisterAlamedaNodes(in.GetAlamedaNodes()); err != nil {
		scope.Error(err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

// CreatePodPredictions add pod predictions information to database
func (s *Server) CreatePodPredictions(ctx context.Context, in *DatahubV1Alpha1.CreatePodPredictionsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreatePodPredictions grpc function: " + AlamedaUtils.InterfaceToString(in))

	predictionDAO := DaoPredictionImpl.NewInfluxDBWithConfig(*s.Config.InfluxDB)
	err := predictionDAO.CreateContainerPredictions(in)
	if err != nil {
		scope.Errorf("create pod predictions failed: %+v", err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

// CreateNodePredictions add node predictions information to database
func (s *Server) CreateNodePredictions(ctx context.Context, in *DatahubV1Alpha1.CreateNodePredictionsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateNodePredictions grpc function: " + AlamedaUtils.InterfaceToString(in))

	predictionDAO := DaoPredictionImpl.NewInfluxDBWithConfig(*s.Config.InfluxDB)
	err := predictionDAO.CreateNodePredictions(in)
	if err != nil {
		scope.Errorf("create node predictions failed: %+v", err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

// CreatePodRecommendations add pod recommendations information to database
func (s *Server) CreatePodRecommendations(ctx context.Context, in *DatahubV1Alpha1.CreatePodRecommendationsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreatePodRecommendations grpc function: " + AlamedaUtils.InterfaceToString(in))
	var containerDAO DaoRecommendation.ContainerOperation = &DaoRecommendationImpl.Container{
		InfluxDBConfig: *s.Config.InfluxDB,
	}

	podRecommendations := in.GetPodRecommendations()
	for _, podRecommendation := range podRecommendations {
		podNS := podRecommendation.GetNamespacedName().Namespace
		podName := podRecommendation.GetNamespacedName().Name
		alamedaRecommendation := &OperatorAPIsAutoScalingV1Alpha1.AlamedaRecommendation{}

		if err := s.K8SClient.Get(context.TODO(), types.NamespacedName{
			Namespace: podNS,
			Name:      podName,
		}, alamedaRecommendation); err == nil {
			alamedarecommendationReconciler := OperatorReconcilerAlamedaRecommendation.NewReconciler(s.K8SClient, alamedaRecommendation)
			if alamedaRecommendation, err = alamedarecommendationReconciler.UpdateResourceRecommendation(podRecommendation); err == nil {
				if err = s.K8SClient.Update(context.TODO(), alamedaRecommendation); err != nil {
					scope.Error(err.Error())
				}
			}
		} else if !K8SErrors.IsNotFound(err) {
			scope.Error(err.Error())
		}
	}

	if err := containerDAO.AddPodRecommendations(in); err != nil {
		scope.Error(err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, err
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

// CreatePodRecommendations add pod recommendations information to database
func (s *Server) CreateControllerRecommendations(ctx context.Context, in *DatahubV1Alpha1.CreateControllerRecommendationsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateControllerRecommendations grpc function: " + AlamedaUtils.InterfaceToString(in))

	controllerDAO := DaoRecommendationImpl.Controller{
		InfluxDBConfig: *s.Config.InfluxDB,
	}

	controllerRecommendationList := in.GetControllerRecommendations()
	err := controllerDAO.AddControllerRecommendations(controllerRecommendationList)

	if err != nil {
		scope.Error(err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, err
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

// CreateSimulatedSchedulingScores add simulated scheduling scores to database
func (s *Server) CreateSimulatedSchedulingScores(ctx context.Context, in *DatahubV1Alpha1.CreateSimulatedSchedulingScoresRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateSimulatedSchedulingScores grpc function: " + AlamedaUtils.InterfaceToString(in))

	var (
		err error

		scoreDAO                           DaoScore.DAO
		daoSimulatedSchedulingScoreEntites = make([]*DaoScore.SimulatedSchedulingScore, 0)
	)

	scoreDAO = DaoScoreImplInflux.NewWithConfig(*s.Config.InfluxDB)

	for _, scoreEntity := range in.GetScores() {

		if scoreEntity == nil {
			continue
		}

		timestamp, _ := ptypes.Timestamp(scoreEntity.GetTime())
		daoSimulatedSchedulingScoreEntity := DaoScore.SimulatedSchedulingScore{
			Timestamp:   timestamp,
			ScoreBefore: float64(scoreEntity.GetScoreBefore()),
			ScoreAfter:  float64(scoreEntity.GetScoreAfter()),
		}
		daoSimulatedSchedulingScoreEntites = append(daoSimulatedSchedulingScoreEntites, &daoSimulatedSchedulingScoreEntity)
	}

	err = scoreDAO.CreateSimulatedScheduingScores(daoSimulatedSchedulingScoreEntites)
	if err != nil {
		scope.Errorf("api CreateSimulatedSchedulingScores failed: %+v", err)
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

// DeleteAlamedaNodes remove node information to database
func (s *Server) DeleteAlamedaNodes(ctx context.Context, in *DatahubV1Alpha1.DeleteAlamedaNodesRequest) (*status.Status, error) {
	scope.Debug("Request received from DeleteAlamedaNodes grpc function: " + AlamedaUtils.InterfaceToString(in))

	var nodeDAO DaoClusterStatus.NodeOperation = &DaoClusterStatusImpl.Node{
		InfluxDBConfig: *s.Config.InfluxDB,
	}
	alamedaNodeList := []*DatahubV1Alpha1.Node{}
	for _, alamedaNode := range in.GetAlamedaNodes() {
		alamedaNodeList = append(alamedaNodeList, &DatahubV1Alpha1.Node{
			Name: alamedaNode.GetName(),
		})
	}
	if err := nodeDAO.DeregisterAlamedaNodes(alamedaNodeList); err != nil {
		scope.Error(err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

// Read rawdata from database
func (s *Server) ReadRawdata(ctx context.Context, in *DatahubV1Alpha1.ReadRawdataRequest) (*DatahubV1Alpha1.ReadRawdataResponse, error) {
	scope.Debug("Request received from ReadRawdata grpc function")

	var (
		err     error
		rawdata = make([]*Common.ReadRawdata, 0)
	)

	switch in.GetDatabaseType() {
	case Common.DatabaseType_INFLUXDB:
		rawdata, err = InternalInflux.ReadRawdata(s.Config.InfluxDB, in.GetQueries())
	case Common.DatabaseType_PROMETHEUS:
		rawdata, err = InternalPromth.ReadRawdata(s.Config.Prometheus, in.GetQueries())
	default:
		err = errors.New(fmt.Sprintf("database type(%s) is not supported", Common.DatabaseType_name[int32(in.GetDatabaseType())]))
	}

	if err != nil {
		scope.Errorf("api ReadRawdata failed: %v", err)
		response := &DatahubV1Alpha1.ReadRawdataResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
			Rawdata: rawdata,
		}
		return response, err
	}

	response := &DatahubV1Alpha1.ReadRawdataResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		Rawdata: rawdata,
	}

	return response, nil
}

// Write rawdata to database
func (s *Server) WriteRawdata(ctx context.Context, in *DatahubV1Alpha1.WriteRawdataRequest) (*status.Status, error) {
	scope.Debug("Request received from WriteRawdata grpc function")

	var (
		err error
	)

	switch in.GetDatabaseType() {
	case Common.DatabaseType_INFLUXDB:
		err = InternalInflux.WriteRawdata(s.Config.InfluxDB, in.GetRawdata())
	case Common.DatabaseType_PROMETHEUS:
		err = errors.New(fmt.Sprintf("database type(%s) is not supported yet", Common.DatabaseType_name[int32(in.GetDatabaseType())]))
	default:
		err = errors.New(fmt.Sprintf("database type(%s) is not supported", Common.DatabaseType_name[int32(in.GetDatabaseType())]))
	}

	if err != nil {
		scope.Errorf("api WriteRawdata failed: %v", err)
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, err
	}

	return &status.Status{Code: int32(code.Code_OK)}, nil
}

func (s *Server) Ping(ctx context.Context, in *empty.Empty) (*status.Status, error) {
	scope.Debug("Request received from Ping grpc function")

	return &status.Status{Code: int32(code.Code_OK)}, nil
}

func (s *Server) ListWeaveScopeHosts(ctx context.Context, in *DatahubV1Alpha1.ListWeaveScopeHostsRequest) (*DatahubV1Alpha1.WeaveScopeResponse, error) {
	response := &DatahubV1Alpha1.WeaveScopeResponse{}

	weaveScopeDAO := DaoWeaveScope.WeaveScope{
		WeaveScopeConfig: s.Config.WeaveScope,
	}

	rawdata, err := weaveScopeDAO.ListWeaveScopeHosts(in)

	if err != nil {
		scope.Error(err.Error())
		return &DatahubV1Alpha1.WeaveScopeResponse{
			Status: &status.Status{
				Code: int32(code.Code_OK),
			},
			Rawdata: rawdata,
		}, nil
	}

	response.Rawdata = rawdata
	return response, nil
}

func (s *Server) GetWeaveScopeHostDetails(ctx context.Context, in *DatahubV1Alpha1.ListWeaveScopeHostsRequest) (*DatahubV1Alpha1.WeaveScopeResponse, error) {
	response := &DatahubV1Alpha1.WeaveScopeResponse{}

	weaveScopeDAO := DaoWeaveScope.WeaveScope{
		WeaveScopeConfig: s.Config.WeaveScope,
	}
	rawdata, err := weaveScopeDAO.GetWeaveScopeHostDetails(in)

	if err != nil {
		scope.Error(err.Error())
		return &DatahubV1Alpha1.WeaveScopeResponse{
			Status: &status.Status{
				Code: int32(code.Code_OK),
			},
			Rawdata: rawdata,
		}, nil
	}

	response.Rawdata = rawdata
	return response, nil
}

func (s *Server) ListWeaveScopePods(ctx context.Context, in *DatahubV1Alpha1.ListWeaveScopePodsRequest) (*DatahubV1Alpha1.WeaveScopeResponse, error) {
	response := &DatahubV1Alpha1.WeaveScopeResponse{}

	weaveScopeDAO := DaoWeaveScope.WeaveScope{
		WeaveScopeConfig: s.Config.WeaveScope,
	}
	rawdata, err := weaveScopeDAO.ListWeaveScopePods(in)

	if err != nil {
		scope.Error(err.Error())
		return &DatahubV1Alpha1.WeaveScopeResponse{
			Status: &status.Status{
				Code: int32(code.Code_OK),
			},
			Rawdata: rawdata,
		}, nil
	}

	response.Rawdata = rawdata
	return response, nil
}

func (s *Server) GetWeaveScopePodDetails(ctx context.Context, in *DatahubV1Alpha1.ListWeaveScopePodsRequest) (*DatahubV1Alpha1.WeaveScopeResponse, error) {
	response := &DatahubV1Alpha1.WeaveScopeResponse{}

	weaveScopeDAO := DaoWeaveScope.WeaveScope{
		WeaveScopeConfig: s.Config.WeaveScope,
	}
	rawdata, err := weaveScopeDAO.GetWeaveScopePodDetails(in)

	if err != nil {
		scope.Error(err.Error())
		return &DatahubV1Alpha1.WeaveScopeResponse{
			Status: &status.Status{
				Code: int32(code.Code_OK),
			},
			Rawdata: rawdata,
		}, nil
	}

	response.Rawdata = rawdata
	return response, nil
}

func (s *Server) ListWeaveScopeContainers(ctx context.Context, in *DatahubV1Alpha1.ListWeaveScopeContainersRequest) (*DatahubV1Alpha1.WeaveScopeResponse, error) {
	response := &DatahubV1Alpha1.WeaveScopeResponse{}

	weaveScopeDAO := DaoWeaveScope.WeaveScope{
		WeaveScopeConfig: s.Config.WeaveScope,
	}
	rawdata, err := weaveScopeDAO.ListWeaveScopeContainers(in)

	if err != nil {
		scope.Error(err.Error())
		return &DatahubV1Alpha1.WeaveScopeResponse{
			Status: &status.Status{
				Code: int32(code.Code_OK),
			},
			Rawdata: rawdata,
		}, nil
	}

	response.Rawdata = rawdata
	return response, nil
}

func (s *Server) ListWeaveScopeContainersByHostname(ctx context.Context, in *DatahubV1Alpha1.ListWeaveScopeContainersRequest) (*DatahubV1Alpha1.WeaveScopeResponse, error) {
	response := &DatahubV1Alpha1.WeaveScopeResponse{}

	weaveScopeDAO := DaoWeaveScope.WeaveScope{
		WeaveScopeConfig: s.Config.WeaveScope,
	}
	rawdata, err := weaveScopeDAO.ListWeaveScopeContainersByHostname(in)

	if err != nil {
		scope.Error(err.Error())
		return &DatahubV1Alpha1.WeaveScopeResponse{
			Status: &status.Status{
				Code: int32(code.Code_OK),
			},
			Rawdata: rawdata,
		}, nil
	}

	response.Rawdata = rawdata
	return response, nil
}

func (s *Server) ListWeaveScopeContainersByImage(ctx context.Context, in *DatahubV1Alpha1.ListWeaveScopeContainersRequest) (*DatahubV1Alpha1.WeaveScopeResponse, error) {
	response := &DatahubV1Alpha1.WeaveScopeResponse{}

	weaveScopeDAO := DaoWeaveScope.WeaveScope{
		WeaveScopeConfig: s.Config.WeaveScope,
	}
	rawdata, err := weaveScopeDAO.ListWeaveScopeContainersByImage(in)

	if err != nil {
		scope.Error(err.Error())
		return &DatahubV1Alpha1.WeaveScopeResponse{
			Status: &status.Status{
				Code: int32(code.Code_OK),
			},
			Rawdata: rawdata,
		}, nil
	}

	response.Rawdata = rawdata
	return response, nil
}

func (s *Server) GetWeaveScopeContainerDetails(ctx context.Context, in *DatahubV1Alpha1.ListWeaveScopeContainersRequest) (*DatahubV1Alpha1.WeaveScopeResponse, error) {
	response := &DatahubV1Alpha1.WeaveScopeResponse{}

	weaveScopeDAO := DaoWeaveScope.WeaveScope{
		WeaveScopeConfig: s.Config.WeaveScope,
	}
	rawdata, err := weaveScopeDAO.GetWeaveScopeContainerDetails(in)

	if err != nil {
		scope.Error(err.Error())
		return &DatahubV1Alpha1.WeaveScopeResponse{
			Status: &status.Status{
				Code: int32(code.Code_OK),
			},
			Rawdata: rawdata,
		}, nil
	}

	response.Rawdata = rawdata
	return response, nil
}
